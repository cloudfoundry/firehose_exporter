package logstream

import (
	"code.cloudfoundry.org/go-loggregator"
	"net/http"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/prometheus/common/log"

	"github.com/bosh-prometheus/firehose_exporter/metrics"
)

type LogStream struct {
	url               string
	skipSSLValidation bool
	subscriptionID    string
	metricsStore      *metrics.Store
	messages          <-chan *events.Envelope
	consumer          *V2Adapter
	httpClient        doer
}

type doer interface {
	Do(req *http.Request) (*http.Response, error)
}

func New(
	url string,
	skipSSLValidation bool,
	subscriptionID string,
	metricsStore *metrics.Store,
	httpClient doer,
) *LogStream {
	return &LogStream{
		url:               url,
		skipSSLValidation: skipSSLValidation,
		subscriptionID:    subscriptionID,
		metricsStore:      metricsStore,
		messages:          make(<-chan *events.Envelope),
		httpClient:        httpClient,
	}
}

// Start processes both errors and messages until both channels are closed
// It then closes the underlying consumer.
func (n *LogStream) Start() {
	log.Info("Starting Firehose Nozzle...")
	defer log.Info("Firehose Nozzle shutting down...")
	n.consumeLogstream()
	n.parseEnvelopes()
}

func (n *LogStream) consumeLogstream() {
	rlpGatewayClient := loggregator.NewRLPGatewayClient(
		n.url,
		loggregator.WithRLPGatewayHTTPClient(n.httpClient),
	)
	a := NewV2Adapter(rlpGatewayClient)
	n.messages = a.Firehose(n.subscriptionID)
}

// parseEnvelopes will read and process both errs and messages, until
// both are closed, at which time it will close the consumer and return
func (n *LogStream) parseEnvelopes() {
	defer n.consumer.Close()

	for {
		select {
		case envelope, ok := <-n.messages:
			if !ok {
				return
			}
			n.metricsStore.AddMetric(envelope)
		}
	}
}
