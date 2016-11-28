package collectors_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-community/firehose_exporter/filters"
	"github.com/cloudfoundry-community/firehose_exporter/metrics"
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
			"Cloud Foundry Firehose http start stop request size in bytes.",
			[]string{"application_id", "instance_id", "uri", "method"},
			nil,
		)

		clientRequestDurationSecondsDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "http_start_stop", "client_request_duration_seconds"),
			"Cloud Foundry Firehose http start stop client request duration in seconds.",
			[]string{"application_id", "instance_id", "uri", "method"},
			nil,
		)

		serverRequestDurationSecondsDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "http_start_stop", "server_request_duration_seconds"),
			"Cloud Foundry Firehose http start stop server request duration in seconds.",
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

		It("returns a client_request_duration_seconds metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(clientRequestDurationSecondsDesc)))
		})

		It("returns a server_request_duration_seconds metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(serverRequestDurationSecondsDesc)))
		})
	})

	Describe("Collect", func() {
	})
})
