package collectors_test

import (
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

		requestTotalDesc                 *prometheus.Desc
		responseSizeBytesDesc            *prometheus.Desc
		lastRequestTimestampDesc         *prometheus.Desc
		clientRequestDurationSecondsDesc *prometheus.Desc
		serverRequestDurationSecondsDesc *prometheus.Desc
	)

	BeforeEach(func() {
		namespace = "test_exporter"
		deploymentFilter = filters.NewDeploymentFilter([]string{})
		eventFilter, _ = filters.NewEventFilter([]string{})
		metricsStore = metrics.NewStore(metricsExpiration, metricsCleanupInterval, deploymentFilter, eventFilter)

		requestTotalDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "http_start_stop", "request_total"),
			"Cloud Foundry Firehose http start stop total requests.",
			[]string{"application_id", "instance_id", "uri", "method", "status_code"},
			nil,
		)

		responseSizeBytesDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "http_start_stop", "response_size_bytes"),
			"Summary of Cloud Foundry Firehose http start stop request size in bytes.",
			[]string{"application_id", "instance_id", "uri", "method"},
			nil,
		)

		lastRequestTimestampDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "http_start_stop", "last_request_timestamp"),
			"Number of seconds since 1970 since last http start stop received from Cloud Foundry Firehose.",
			[]string{"application_id", "instance_id", "uri", "method"},
			nil,
		)

		clientRequestDurationSecondsDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "http_start_stop", "client_request_duration_seconds"),
			"Summary of Cloud Foundry Firehose http start stop client request duration in seconds.",
			[]string{"application_id", "instance_id", "uri", "method"},
			nil,
		)

		serverRequestDurationSecondsDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "http_start_stop", "server_request_duration_seconds"),
			"Summary of Cloud Foundry Firehose http start stop server request duration in seconds.",
			[]string{"application_id", "instance_id", "uri", "method"},
			nil,
		)
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

		It("returns a request_total metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(requestTotalDesc)))
		})

		It("returns a response_size_bytes metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(responseSizeBytesDesc)))
		})

		It("returns a last_request_timestamp metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastRequestTimestampDesc)))
		})

		It("returns a client_request_duration_seconds metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(clientRequestDurationSecondsDesc)))
		})

		It("returns a server_request_duration_seconds metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(serverRequestDurationSecondsDesc)))
		})
	})

	Describe("Collect", func() {
		var (
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
			httpStartStopUri                  = "FakeURI"
			httpStartStopRemoteAddress        = "FakeRemoteAddress"
			httpStartStopUserAgent            = "FakeUserAgent"
			httpStartStopStatusCode           = int32(200)
			httpStartStopContentLength        = int64(32)
			httpStartStopApplicationId        = "8060986d-43aa-4097-8989-1c292accbeb3"
			httpStartStopInstanceIndex        = int32(1)
			httpStartStopInstanceId           = "FakeInstanceId"

			httpStartStopMetricsChan           chan prometheus.Metric
			requestTotalMetric                 prometheus.Metric
			responseSizeBytesMetric            prometheus.Metric
			lastRequestTimestampMetric         prometheus.Metric
			clientRequestDurationSecondsMetric prometheus.Metric
			serverRequestDurationSecondsMetric prometheus.Metric
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

			requestTotalMetric = prometheus.MustNewConstMetric(
				requestTotalDesc,
				prometheus.CounterValue,
				float64(1),
				httpStartStopApplicationId,
				httpStartStopInstanceId,
				httpStartStopUri,
				httpStartStopMethod,
				strconv.Itoa(int(httpStartStopStatusCode)),
			)

			responseSizeBytesMetric = prometheus.MustNewConstSummary(
				responseSizeBytesDesc,
				uint64(1),
				float64(httpStartStopContentLength),
				map[float64]float64{
					float64(0.5):  float64(httpStartStopContentLength),
					float64(0.9):  float64(httpStartStopContentLength),
					float64(0.99): float64(httpStartStopContentLength),
				},
				httpStartStopApplicationId,
				httpStartStopInstanceId,
				httpStartStopUri,
				httpStartStopMethod,
			)

			lastRequestTimestampMetric = prometheus.MustNewConstMetric(
				lastRequestTimestampDesc,
				prometheus.GaugeValue,
				utils.NanosecondsToSeconds(httpStartStopClientStartTimestamp),
				httpStartStopApplicationId,
				httpStartStopInstanceId,
				httpStartStopUri,
				httpStartStopMethod,
			)

			clientRequestDurationSecondsMetric = prometheus.MustNewConstSummary(
				clientRequestDurationSecondsDesc,
				uint64(1),
				httpStartStopClientDuration,
				map[float64]float64{
					float64(0.5):  httpStartStopClientDuration,
					float64(0.9):  httpStartStopClientDuration,
					float64(0.99): httpStartStopClientDuration,
				},
				httpStartStopApplicationId,
				httpStartStopInstanceId,
				httpStartStopUri,
				httpStartStopMethod,
			)

			serverRequestDurationSecondsMetric = prometheus.MustNewConstSummary(
				serverRequestDurationSecondsDesc,
				uint64(1),
				httpStartStopServerDuration,
				map[float64]float64{
					float64(0.5):  httpStartStopServerDuration,
					float64(0.9):  httpStartStopServerDuration,
					float64(0.99): httpStartStopServerDuration,
				},
				httpStartStopApplicationId,
				httpStartStopInstanceId,
				httpStartStopUri,
				httpStartStopMethod,
			)
		})

		JustBeforeEach(func() {
			go httpStartStopCollector.Collect(httpStartStopMetricsChan)
		})

		It("returns a request_total metric", func() {
			Eventually(httpStartStopMetricsChan).Should(Receive(Equal(requestTotalMetric)))
		})

		It("returns a response_size_bytes metric", func() {
			Eventually(httpStartStopMetricsChan).Should(Receive(Equal(responseSizeBytesMetric)))
		})

		It("returns a last_request_timestamp metric", func() {
			Eventually(httpStartStopMetricsChan).Should(Receive(Equal(lastRequestTimestampMetric)))
		})

		It("returns a client_request_duration_seconds metric", func() {
			Eventually(httpStartStopMetricsChan).Should(Receive(Equal(clientRequestDurationSecondsMetric)))
		})

		It("returns a server_request_duration_seconds metric", func() {
			Eventually(httpStartStopMetricsChan).Should(Receive(Equal(serverRequestDurationSecondsMetric)))
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
