package firehosenozzle_test

import (
	"flag"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	firehosefakes "github.com/cloudfoundry-community/firehose_exporter/firehosenozzle/fakes"
	"github.com/cloudfoundry-community/firehose_exporter/metrics"
	"github.com/cloudfoundry-community/firehose_exporter/uaatokenrefresher"
	"github.com/cloudfoundry-community/firehose_exporter/uaatokenrefresher/fakes"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/websocket"

	. "github.com/cloudfoundry-community/firehose_exporter/firehosenozzle"
)

func init() {
	flag.Set("log.level", "fatal")
}

var _ = Describe("FirehoseNozzle", func() {
	var (
		skipSSLValidation  bool
		subscriptionID     string
		idleTimeoutSeconds uint32

		fakeUAA   *fakes.FakeUAA
		fakeToken string

		fakeFirehose *firehosefakes.FakeFirehose

		authTokenRefresher *uaatokenrefresher.UAATokenRefresher

		metricsExpiration      time.Duration
		metricsCleanupInterval time.Duration
		metricsStore           *metrics.Store

		firehoseNozzle *FirehoseNozzle

		envelope     events.Envelope
		numEnvelopes = 10
	)

	BeforeEach(func() {
		skipSSLValidation = true
		subscriptionID = "fake-subscription-id"
		idleTimeoutSeconds = 5

		fakeUAA = fakes.NewFakeUAA("bearer", "123456789")
		fakeToken = fakeUAA.AuthToken()
		fakeUAA.Start()

		fakeFirehose = firehosefakes.NewFakeFirehose(fakeToken)
		fakeFirehose.Start()

		authTokenRefresher, _ = uaatokenrefresher.New(
			fakeUAA.URL(), "client-id", "client-secret", true,
		)

		metricsStore = metrics.NewStore(metricsExpiration, metricsCleanupInterval)

		for i := 0; i < numEnvelopes; i++ {
			envelope = events.Envelope{
				Origin:     proto.String("fake-origin"),
				EventType:  events.Envelope_ValueMetric.Enum(),
				Timestamp:  proto.Int64(time.Now().Unix()),
				Deployment: proto.String("fake-deployment-name"),
				Job:        proto.String("fake-job-name"),
				Index:      proto.String("0"),
				Ip:         proto.String("1.2.3.4"),
				ValueMetric: &events.ValueMetric{
					Name:  proto.String(fmt.Sprintf("fake-metric-%d", i)),
					Value: proto.Float64(float64(i)),
					Unit:  proto.String("counter"),
				},
			}
			fakeFirehose.AddEvent(envelope)
		}
	})

	JustBeforeEach(func() {
		firehoseNozzle = New(
			strings.Replace(fakeFirehose.URL(), "http:", "ws:", 1),
			skipSSLValidation,
			subscriptionID,
			idleTimeoutSeconds,
			authTokenRefresher,
			metricsStore,
		)
		go firehoseNozzle.Start()
	})

	AfterEach(func() {
		fakeFirehose.Close()
		fakeUAA.Close()
	})

	It("receives data from the firehose", func() {
		Eventually(fakeFirehose.Requested).Should(BeTrue())
		Consistently(metricsStore.GetInternalMetrics().TotalEnvelopesReceived).Should(Equal(int64(numEnvelopes)))
	})

	Context("when receives a TruncatingBuffer.DroppedMessages value metric", func() {
		var (
			slowConsumerError events.Envelope
		)

		BeforeEach(func() {
			slowConsumerError = events.Envelope{
				Origin:    proto.String("doppler"),
				Timestamp: proto.Int64(time.Now().Unix()),
				EventType: events.Envelope_CounterEvent.Enum(),
				CounterEvent: &events.CounterEvent{
					Name:  proto.String("TruncatingBuffer.DroppedMessages"),
					Delta: proto.Uint64(1),
					Total: proto.Uint64(1),
				},
				Deployment: proto.String("deployment-name"),
				Job:        proto.String("doppler"),
			}

			fakeFirehose.AddEvent(slowConsumerError)
		})

		It("sets a SlowConsumerAlert", func() {
			Eventually(fakeFirehose.Requested).Should(BeTrue())
			Consistently(metricsStore.GetInternalMetrics().SlowConsumerAlert).Should(BeTrue())
		})
	})

	Context("when when the server disconnects abnormally", func() {
		var (
			closeMessage []byte
		)

		JustBeforeEach(func() {
			fakeFirehose.SetCloseMessage(closeMessage)
		})

		Context("abnormally", func() {
			BeforeEach(func() {
				closeMessage = websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "websocket-close-policy-violation")
			})

			It("sets a SlowConsumerAlert", func() {
				Eventually(fakeFirehose.Requested).Should(BeTrue())
				Consistently(metricsStore.GetInternalMetrics().SlowConsumerAlert).Should(BeTrue())
			})
		})

		Context("for other reasons", func() {
			BeforeEach(func() {
				closeMessage = websocket.FormatCloseMessage(websocket.CloseInvalidFramePayloadData, "websocket-close-invalid-frame-payload-data")
			})

			It("does not sets a SlowConsumerAlert", func() {
				Eventually(fakeFirehose.Requested).Should(BeTrue())
				Consistently(metricsStore.GetInternalMetrics().SlowConsumerAlert).Should(BeFalse())
			})
		})
	})
})
