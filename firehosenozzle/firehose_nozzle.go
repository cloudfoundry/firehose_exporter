package firehosenozzle

import (
	"crypto/tls"
	"time"

	"github.com/cloudfoundry/noaa/consumer"
	noaerrors "github.com/cloudfoundry/noaa/errors"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gorilla/websocket"
	"github.com/prometheus/common/log"

	"github.com/bosh-prometheus/firehose_exporter/metrics"
)

type FirehoseNozzle struct {
	url                string
	skipSSLValidation  bool
	subscriptionID     string
	idleTimeout        time.Duration
	minRetryDelay      time.Duration
	maxRetryDelay      time.Duration
	maxRetryCount      int
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
	maxRetryCount int,
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
		maxRetryCount:      maxRetryCount,
		authTokenRefresher: authTokenRefresher,
		metricsStore:       metricsStore,
		errs:               make(<-chan error),
		messages:           make(<-chan *events.Envelope),
	}
}

// Start processes both errors and messages until both channels are closed
// It then closes the underlying consumer.
func (n *FirehoseNozzle) Start() {
	log.Info("Starting Firehose Nozzle...")
	defer log.Info("Firehose Nozzle shutting down...")
	n.consumeFirehose()
	n.parseEnvelopes()
}

func (n *FirehoseNozzle) consumeFirehose() {
	n.consumer = consumer.New(
		n.url,
		&tls.Config{InsecureSkipVerify: n.skipSSLValidation},
		nil,
	)
	n.consumer.RefreshTokenFrom(n.authTokenRefresher)
	n.consumer.SetDebugPrinter(DebugPrinter{})
	if n.idleTimeout > 0 {
		n.consumer.SetIdleTimeout(n.idleTimeout)
	}
	if n.minRetryDelay > 0 {
		n.consumer.SetMinRetryDelay(n.minRetryDelay)
	}
	if n.maxRetryDelay > 0 {
		n.consumer.SetMaxRetryDelay(n.maxRetryDelay)
	}
	if n.maxRetryCount > 0 {
		n.consumer.SetMaxRetryCount(n.maxRetryCount)
	}
	n.messages, n.errs = n.consumer.FilteredFirehose(n.subscriptionID, "", consumer.Metrics)
}

// parseEnvelopes will read and process both errs and messages, until
// both are closed, at which time it will close the consumer and return
func (n *FirehoseNozzle) parseEnvelopes() {
	defer n.consumer.Close()

	for messages, errs := n.messages, n.errs; messages != nil || errs != nil; {
		select {
		case envelope, ok := <-messages:
			if !ok {
				messages = nil
				continue
			}
			n.handleMessage(envelope)
			n.metricsStore.AddMetric(envelope)
		case err, ok := <-errs:
			if !ok {
				errs = nil
				continue
			}
			n.handleError(err)
		}
	}
}

func (n *FirehoseNozzle) handleMessage(envelope *events.Envelope) {
	if envelope.GetEventType() == events.Envelope_CounterEvent && envelope.CounterEvent.GetName() == "TruncatingBuffer.DroppedMessages" && envelope.GetOrigin() == "doppler" {
		log.Infof("We've intercepted an upstream message which indicates that the Nozzle or the TrafficController is not keeping up. Please try scaling up the Nozzle.")
		n.metricsStore.AlertSlowConsumerError()
	}
}

func (n *FirehoseNozzle) handleError(err error) {
	log.Errorf("Error while reading from the Firehose: %v", err)

	switch err.(type) {
	case noaerrors.RetryError:
		switch noaRetryError := err.(noaerrors.RetryError).Err.(type) {
		case *websocket.CloseError:
			switch noaRetryError.Code {
			case websocket.ClosePolicyViolation:
				log.Errorf("Nozzle couldn't keep up. Please try scaling up the Nozzle.")
				n.metricsStore.AlertSlowConsumerError()
			}
		}
	}
}

type DebugPrinter struct{}

func (dp DebugPrinter) Print(title, dump string) {
	log.Debugf("%s: %s", title, dump)
}
