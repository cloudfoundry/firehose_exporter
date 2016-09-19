package firehosenozzle

import (
	"crypto/tls"
	"time"

	"github.com/cloudfoundry/noaa/consumer"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gorilla/websocket"
	"github.com/prometheus/common/log"

	"github.com/frodenas/firehose_exporter/metrics"
)

type FirehoseNozzle struct {
	url                string
	skipSSLValidation  bool
	subscriptionID     string
	idleTimeoutSeconds uint32
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
	idleTimeoutSeconds uint32,
	authTokenRefresher consumer.TokenRefresher,
	metricsStore *metrics.Store,
) *FirehoseNozzle {
	return &FirehoseNozzle{
		url:                url,
		skipSSLValidation:  skipSSLValidation,
		subscriptionID:     subscriptionID,
		idleTimeoutSeconds: idleTimeoutSeconds,
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
	n.consumer.SetIdleTimeout(time.Duration(n.idleTimeoutSeconds) * time.Second)
	n.messages, n.errs = n.consumer.Firehose(n.subscriptionID, "")
}

func (n *FirehoseNozzle) parseEnvelopes() error {
	for {
		select {
		case envelope := <-n.messages:
			n.handleMessage(envelope)
			n.metricsStore.AddMetric(envelope)
		case err := <-n.errs:
			n.handleError(err)
			return err
		}
	}
}

func (n *FirehoseNozzle) handleMessage(envelope *events.Envelope) {
	if envelope.GetEventType() == events.Envelope_CounterEvent && envelope.CounterEvent.GetName() == "TruncatingBuffer.DroppedMessages" && envelope.GetOrigin() == "doppler" {
		log.Infof("We've intercepted an upstream message which indicates that the nozzle or the TrafficController is not keeping up. Please try scaling up the nozzle.")
		n.metricsStore.AlertSlowConsumerError()
	}
}

func (n *FirehoseNozzle) handleError(err error) {
	switch closeErr := err.(type) {
	case *websocket.CloseError:
		switch closeErr.Code {
		case websocket.CloseNormalClosure:
		// no op
		case websocket.ClosePolicyViolation:
			log.Errorf("Error while reading from the firehose: %v", err)
			log.Errorf("Disconnected because nozzle couldn't keep up. Please try scaling up the nozzle.")
			n.metricsStore.AlertSlowConsumerError()
		default:
			log.Errorf("Error while reading from the firehose: %v", err)
		}
	default:
		log.Errorf("Error while reading from the firehose: %v", err)
	}

	log.Infof("Closing connection with traffic controller due to %v", err)
	n.consumer.Close()
}
