package logstream_test

import (
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"fmt"
	"github.com/bosh-prometheus/firehose_exporter/authclient"
	"github.com/bosh-prometheus/firehose_exporter/logstream"
	"github.com/cloudfoundry-incubator/uaago"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/bosh-prometheus/firehose_exporter/filters"
	logstreamfakes "github.com/bosh-prometheus/firehose_exporter/logstream/fakes"
	"github.com/bosh-prometheus/firehose_exporter/metrics"
	"github.com/bosh-prometheus/firehose_exporter/uaatokenrefresher/fakes"
	"github.com/prometheus/common/log"
)

func init() {
	log.Base().SetLevel("fatal")
}

var _ = Describe("LogStream", func() {
	var (
		skipSSLValidation bool
		subscriptionID    string

		fakeUAA   *fakes.FakeUAA
		fakeToken string

		metricsExpiration      time.Duration
		metricsCleanupInterval time.Duration
		deploymentFilter       *filters.DeploymentFilter
		eventFilter            *filters.EventFilter
		metricsStore           *metrics.Store

		ls            *logstream.LogStream
		fakeLogStream *logstreamfakes.FakeLogStream

		envelope     *loggregator_v2.Envelope
		numEnvelopes = 10
	)

	BeforeEach(func() {
		skipSSLValidation = true
		subscriptionID = "fake-subscription-id"

		fakeUAA = fakes.NewFakeUAA("bearer", "123456789")
		fakeToken = fakeUAA.AuthToken()
		fakeUAA.Start()

		fakeLogStream = logstreamfakes.NewFakeLogStream(fakeToken)
		fakeLogStream.Start()

		deploymentFilter = filters.NewDeploymentFilter([]string{})
		eventFilter, _ = filters.NewEventFilter([]string{})
		metricsStore = metrics.NewStore(metricsExpiration, metricsCleanupInterval, deploymentFilter, eventFilter)

		for i := 0; i < numEnvelopes; i++ {
			envelope = &loggregator_v2.Envelope{
				SourceId: "fake-origin",
				Message: &loggregator_v2.Envelope_Gauge{
					Gauge: &loggregator_v2.Gauge{
						Metrics: map[string]*loggregator_v2.GaugeValue{
							fmt.Sprintf("fake-metric-%d", i): {Unit: "counter", Value: float64(i)},
						},
					},
				},
				Timestamp: time.Now().Unix(),
			}

			fakeLogStream.AddEvent(envelope)
		}
	})

	JustBeforeEach(func() {
		uaa, err := uaago.NewClient(fakeUAA.URL())
		if err != nil {
			log.Errorln(fmt.Sprint("Failed connecting to Get token from UAA..", err), "")
		}

		ac := authclient.NewHttp(uaa, "", "", skipSSLValidation)

		ls = logstream.New(
			fakeLogStream.URL(),
			skipSSLValidation,
			subscriptionID,
			metricsStore,
			ac,
		)
		go ls.Start()
	})

	AfterEach(func() {
		fakeLogStream.Close()
		fakeUAA.Close()
	})

	It("receives data from the firehose", func() {
		Eventually(fakeLogStream.Requested).Should(BeTrue())
		Eventually(func() int64 { return metricsStore.GetInternalMetrics().TotalEnvelopesReceived }).Should(Equal(int64(numEnvelopes)))
	})
})
