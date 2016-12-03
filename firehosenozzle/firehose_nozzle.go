package firehosenozzle

import (
	"crypto/tls"
	"time"

	"github.com/cloudfoundry/noaa/consumer"
	noaerrors "github.com/cloudfoundry/noaa/errors"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gorilla/websocket"
	"github.com/prometheus/common/log"

	"github.com/cloudfoundry-community/firehose_exporter/metrics"
)

type FirehoseNozzle struct {
	url                string
	skipSSLValidation  bool
	subscriptionID     string
	idleTimeout        time.Duration
	minRetryDelay      time.Duration
	maxRetryDelay      time.Duration
	authTokenRefresher consumer.TokenRefresher
	metricsStore       *metrics.Store
	errs               <-chan error
	messages           <-chan *events.Envelope
	consumer           *consumer.Consumer
}

func New(
	url string,
	skipSSLValidation bool,
	subscriptionID string,
	idleTimeout time.Duration,
	minRetryDelay time.Duration,
	maxRetryDelay time.Duration,
	authTokenRefresher consumer.TokenRefresher,
	metricsStore *metrics.Store,
) *FirehoseNozzle {
	return &FirehoseNozzle{
		url:                url,
		skipSSLValidation:  skipSSLValidation,
		subscriptionID:     subscriptionID,
		idleTimeout:        idleTimeout,
		minRetryDelay:      minRetryDelay,
		maxRetryDelay:      maxRetryDelay,
		authTokenRefresher: authTokenRefresher,
		metricsStore:       metricsStore,
		errs:               make(<-chan error),
		messages:           make(<-chan *events.Envelope),
	}
}

func (n *FirehoseNozzle) Start() error {
	log.Info("Starting Firehose Nozzle...")
	n.consumeFirehose()
	err := n.parseEnvelopes()
	log.Info("Firehose Nozzle shutting down...")
	return err
}

func (n *FirehoseNozzle) consumeFirehose() {
	n.consumer = consumer.New(
		n.url,
		&tls.Config{InsecureSkipVerify: n.skipSSLValidation},
		nil,
	)
	n.consumer.RefreshTokenFrom(n.authTokenRefresher)
	if n.idleTimeout > 0 {
		n.consumer.SetIdleTimeout(n.idleTimeout)
	}
	if n.minRetryDelay > 0 {
		n.consumer.SetMinRetryDelay(n.minRetryDelay)
	}
	if n.maxRetryDelay > 0 {
		n.consumer.SetMaxRetryDelay(n.maxRetryDelay)
	}
	n.messages, n.errs = n.consumer.Firehose(n.subscriptionID, "")
}

func (n *FirehoseNozzle) parseEnvelopes() error {
	for {
		select {
		case envelope := <-n.messages:
			n.handleMessage(envelope)
			n.metricsStore.AddMetric(envelope)
		case err := <-n.errs:
			retryError := n.handleError(err)
			if !retryError {
				return err
			}
		}
	}
}

func (n *FirehoseNozzle) handleMessage(envelope *events.Envelope) {
	if envelope.GetEventType() == events.Envelope_CounterEvent && envelope.CounterEvent.GetName() == "TruncatingBuffer.DroppedMessages" && envelope.GetOrigin() == "doppler" {
		log.Infof("We've intercepted an upstream message which indicates that the Nozzle or the TrafficController is not keeping up. Please try scaling up the Nozzle.")
		n.metricsStore.AlertSlowConsumerError()
	}
}

func (n *FirehoseNozzle) handleError(err error) bool {
	log.Errorf("Error while reading from the Firehose: %v", err)

	switch err.(type) {
	case noaerrors.RetryError:
		switch noaRetryError := err.(noaerrors.RetryError).Err.(type) {
		case *websocket.CloseError:
			switch noaRetryError.Code {
			case websocket.CloseNormalClosure:
			// no op
			case websocket.ClosePolicyViolation:
				log.Errorf("Nozzle couldn't keep up. Please try scaling up the Nozzle.")
				n.metricsStore.AlertSlowConsumerError()
			}
		}
		return true
	}

	log.Info("Closing connection with Firehose...")
	n.consumer.Close()

	return false
}
