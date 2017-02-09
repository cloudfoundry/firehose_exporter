package collectors_test

import (
	"fmt"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-community/firehose_exporter/filters"
	"github.com/cloudfoundry-community/firehose_exporter/metrics"
	"github.com/cloudfoundry-community/firehose_exporter/utils"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"

	. "github.com/cloudfoundry-community/firehose_exporter/collectors"
	. "github.com/cloudfoundry-community/firehose_exporter/utils/test_matchers"
)

var _ = Describe("HttpStartStopCollector", func() {
	var (
		namespace              string
		metricsStore           *metrics.Store
		metricsExpiration      time.Duration
		metricsCleanupInterval time.Duration
		deploymentFilter       *filters.DeploymentFilter
		eventFilter            *filters.EventFilter
		httpStartStopCollector *HttpStartStopCollector

		requestsMetric                     *prometheus.GaugeVec
		responseSizeBytesMetric            *prometheus.SummaryVec
		lastRequestTimestampMetric         *prometheus.GaugeVec
		clientRequestDurationSecondsMetric *prometheus.SummaryVec
		serverRequestDurationSecondsMetric *prometheus.SummaryVec

		origin         = "fake-origin"
		boshDeployment = "fake-deployment-name"
		boshJob        = "fake-job-name"
		boshIndex      = "0"
		boshIP         = "1.2.3.4"

		httpStartStopClientStartTimestamp = int64(2)
		httpStartStopClientStopTimestamp  = int64(20)
		httpStartStopServerStartTimestamp = int64(1)
		httpStartStopServerStopTimestamp  = int64(10)
		httpStartStopClientDuration       = utils.NanosecondsToSeconds(httpStartStopClientStopTimestamp - httpStartStopClientStartTimestamp)
		httpStartStopServerDuration       = utils.NanosecondsToSeconds(httpStartStopServerStopTimestamp - httpStartStopServerStartTimestamp)
		httpStartStopRequestId            = "1beb4072-acaa-483f-5a8b-425dc080af13"
		httpStartStopClientPeerType       = events.PeerType_Client
		httpStartStopServerPeerType       = events.PeerType_Server
		httpStartStopMethod               = "GET"
		httpStartStopScheme               = "http"
		httpStartStopHost                 = "www.example.com"
		httpStartStopUri                  = fmt.Sprintf("%s://user:password@%s/foo?bar", httpStartStopScheme, httpStartStopHost)
		httpStartStopRemoteAddress        = "FakeRemoteAddress"
		httpStartStopUserAgent            = "FakeUserAgent"
		httpStartStopStatusCode           = int32(200)
		httpStartStopContentLength        = int64(32)
		httpStartStopApplicationId        = "8060986d-43aa-4097-8989-1c292accbeb3"
		httpStartStopInstanceIndex        = int32(1)
		httpStartStopInstanceId           = "FakeInstanceId"
	)

	BeforeEach(func() {
		namespace = "test_exporter"
		deploymentFilter = filters.NewDeploymentFilter([]string{})
		eventFilter, _ = filters.NewEventFilter([]string{})
		metricsStore = metrics.NewStore(metricsExpiration, metricsCleanupInterval, deploymentFilter, eventFilter)

		requestsMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "http_start_stop",
				Name:      "requests",
				Help:      "Cloud Foundry Firehose http start stop requests.",
			},
			[]string{"application_id", "instance_id", "method", "scheme", "host", "status_code"},
		)

		requestsMetric.WithLabelValues(
			httpStartStopApplicationId,
			httpStartStopInstanceId,
			httpStartStopMethod,
			httpStartStopScheme,
			httpStartStopHost,
			strconv.Itoa(int(httpStartStopStatusCode)),
		).Inc()

		responseSizeBytesMetric = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace: namespace,
				Subsystem: "http_start_stop",
				Name:      "response_size_bytes",
				Help:      "Summary of Cloud Foundry Firehose http start stop request size in bytes.",
			},
			[]string{"application_id", "instance_id", "method", "scheme", "host"},
		)

		responseSizeBytesMetric.WithLabelValues(
			httpStartStopApplicationId,
			httpStartStopInstanceId,
			httpStartStopMethod,
			httpStartStopScheme,
			httpStartStopHost,
		).Observe(float64(httpStartStopContentLength))

		lastRequestTimestampMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "http_start_stop",
				Name:      "last_request_timestamp",
				Help:      "Number of seconds since 1970 since last http start stop received from Cloud Foundry Firehose.",
			},
			[]string{"application_id", "instance_id", "method", "scheme", "host"},
		)

		lastRequestTimestampMetric.WithLabelValues(
			httpStartStopApplicationId,
			httpStartStopInstanceId,
			httpStartStopMethod,
			httpStartStopScheme,
			httpStartStopHost,
		).Set(utils.NanosecondsToSeconds(httpStartStopClientStartTimestamp))

		clientRequestDurationSecondsMetric = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace: namespace,
				Subsystem: "http_start_stop",
				Name:      "client_request_duration_seconds",
				Help:      "Summary of Cloud Foundry Firehose http start stop client request duration in seconds.",
			},
			[]string{"application_id", "instance_id", "method", "scheme", "host"},
		)

		clientRequestDurationSecondsMetric.WithLabelValues(
			httpStartStopApplicationId,
			httpStartStopInstanceId,
			httpStartStopMethod,
			httpStartStopScheme,
			httpStartStopHost,
		).Observe(httpStartStopClientDuration)

		serverRequestDurationSecondsMetric = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace: namespace,
				Subsystem: "http_start_stop",
				Name:      "server_request_duration_seconds",
				Help:      "Summary of Cloud Foundry Firehose http start stop server request duration in seconds.",
			},
			[]string{"application_id", "instance_id", "method", "scheme", "host"},
		)

		serverRequestDurationSecondsMetric.WithLabelValues(
			httpStartStopApplicationId,
			httpStartStopInstanceId,
			httpStartStopMethod,
			httpStartStopScheme,
			httpStartStopHost,
		).Observe(httpStartStopServerDuration)
	})

	JustBeforeEach(func() {
		httpStartStopCollector = NewHttpStartStopCollector(namespace, metricsStore)
	})

	Describe("Describe", func() {
		var (
			descriptions chan *prometheus.Desc
		)

		BeforeEach(func() {
			descriptions = make(chan *prometheus.Desc)
		})

		JustBeforeEach(func() {
			go httpStartStopCollector.Describe(descriptions)
		})

		It("returns a requests metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(requestsMetric.WithLabelValues(
				httpStartStopApplicationId,
				httpStartStopInstanceId,
				httpStartStopMethod,
				httpStartStopScheme,
				httpStartStopHost,
				strconv.Itoa(int(httpStartStopStatusCode)),
			).Desc())))
		})

		It("returns a response_size_bytes metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(responseSizeBytesMetric.WithLabelValues(
				httpStartStopApplicationId,
				httpStartStopInstanceId,
				httpStartStopMethod,
				httpStartStopScheme,
				httpStartStopHost,
			).Desc())))
		})

		It("returns a last_request_timestamp metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastRequestTimestampMetric.WithLabelValues(
				httpStartStopApplicationId,
				httpStartStopInstanceId,
				httpStartStopMethod,
				httpStartStopScheme,
				httpStartStopHost,
			).Desc())))
		})

		It("returns a client_request_duration_seconds metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(clientRequestDurationSecondsMetric.WithLabelValues(
				httpStartStopApplicationId,
				httpStartStopInstanceId,
				httpStartStopMethod,
				httpStartStopScheme,
				httpStartStopHost,
			).Desc())))
		})

		It("returns a server_request_duration_seconds metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(serverRequestDurationSecondsMetric.WithLabelValues(
				httpStartStopApplicationId,
				httpStartStopInstanceId,
				httpStartStopMethod,
				httpStartStopScheme,
				httpStartStopHost,
			).Desc())))
		})
	})

	Describe("Collect", func() {
		var (
			httpStartStopMetricsChan chan prometheus.Metric
		)

		BeforeEach(func() {
			metricsStore.AddMetric(
				&events.Envelope{
					Origin:     proto.String(origin),
					EventType:  events.Envelope_HttpStartStop.Enum(),
					Timestamp:  proto.Int64(time.Now().Unix() * 1000),
					Deployment: proto.String(boshDeployment),
					Job:        proto.String(boshJob),
					Index:      proto.String(boshIndex),
					Ip:         proto.String(boshIP),
					Tags:       map[string]string{},
					HttpStartStop: &events.HttpStartStop{
						StartTimestamp: proto.Int64(httpStartStopClientStartTimestamp),
						StopTimestamp:  proto.Int64(httpStartStopClientStopTimestamp),
						RequestId:      utils.StringToUUID(httpStartStopRequestId),
						PeerType:       &httpStartStopClientPeerType,
						Method:         events.Method(events.Method_value[httpStartStopMethod]).Enum(),
						Uri:            proto.String(httpStartStopUri),
						RemoteAddress:  proto.String(httpStartStopRemoteAddress),
						UserAgent:      proto.String(httpStartStopUserAgent),
						StatusCode:     proto.Int32(httpStartStopStatusCode),
						ContentLength:  proto.Int64(httpStartStopContentLength),
						ApplicationId:  utils.StringToUUID(httpStartStopApplicationId),
						InstanceIndex:  proto.Int32(httpStartStopInstanceIndex),
						InstanceId:     proto.String(httpStartStopInstanceId),
					},
				},
			)

			metricsStore.AddMetric(
				&events.Envelope{
					Origin:     proto.String(origin),
					EventType:  events.Envelope_HttpStartStop.Enum(),
					Timestamp:  proto.Int64(time.Now().Unix() * 1000),
					Deployment: proto.String(boshDeployment),
					Job:        proto.String(boshJob),
					Index:      proto.String(boshIndex),
					Ip:         proto.String(boshIP),
					Tags:       map[string]string{},
					HttpStartStop: &events.HttpStartStop{
						StartTimestamp: proto.Int64(httpStartStopServerStartTimestamp),
						StopTimestamp:  proto.Int64(httpStartStopServerStopTimestamp),
						RequestId:      utils.StringToUUID(httpStartStopRequestId),
						PeerType:       &httpStartStopServerPeerType,
						Method:         events.Method(events.Method_value[httpStartStopMethod]).Enum(),
						Uri:            proto.String(httpStartStopUri),
						RemoteAddress:  proto.String(httpStartStopRemoteAddress),
						UserAgent:      proto.String(httpStartStopUserAgent),
						StatusCode:     proto.Int32(httpStartStopStatusCode),
						ContentLength:  proto.Int64(httpStartStopContentLength),
					},
				},
			)

			httpStartStopMetricsChan = make(chan prometheus.Metric)
		})

		JustBeforeEach(func() {
			go httpStartStopCollector.Collect(httpStartStopMetricsChan)
		})

		It("returns a requests metric", func() {
			Eventually(httpStartStopMetricsChan).Should(Receive(PrometheusMetric(requestsMetric.WithLabelValues(
				httpStartStopApplicationId,
				httpStartStopInstanceId,
				httpStartStopMethod,
				httpStartStopScheme,
				httpStartStopHost,
				strconv.Itoa(int(httpStartStopStatusCode)),
			))))
		})

		It("returns a response_size_bytes metric", func() {
			Eventually(httpStartStopMetricsChan).Should(Receive(PrometheusMetric(responseSizeBytesMetric.WithLabelValues(
				httpStartStopApplicationId,
				httpStartStopInstanceId,
				httpStartStopMethod,
				httpStartStopScheme,
				httpStartStopHost,
			))))
		})

		It("returns a last_request_timestamp metric", func() {
			Eventually(httpStartStopMetricsChan).Should(Receive(PrometheusMetric(lastRequestTimestampMetric.WithLabelValues(
				httpStartStopApplicationId,
				httpStartStopInstanceId,
				httpStartStopMethod,
				httpStartStopScheme,
				httpStartStopHost,
			))))
		})

		It("returns a client_request_duration_seconds metric", func() {
			Eventually(httpStartStopMetricsChan).Should(Receive(PrometheusMetric(clientRequestDurationSecondsMetric.WithLabelValues(
				httpStartStopApplicationId,
				httpStartStopInstanceId,
				httpStartStopMethod,
				httpStartStopScheme,
				httpStartStopHost,
			))))
		})

		It("returns a server_request_duration_seconds metric", func() {
			Eventually(httpStartStopMetricsChan).Should(Receive(PrometheusMetric(serverRequestDurationSecondsMetric.WithLabelValues(
				httpStartStopApplicationId,
				httpStartStopInstanceId,
				httpStartStopMethod,
				httpStartStopScheme,
				httpStartStopHost,
			))))
		})

		Context("when there is no http start stop metrics", func() {
			BeforeEach(func() {
				metricsStore.FlushHttpStartStops()
			})

			It("does not return any metric", func() {
				Consistently(httpStartStopMetricsChan).ShouldNot(Receive())
			})
		})
	})
})
