package collectors_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-community/firehose_exporter/filters"
	"github.com/cloudfoundry-community/firehose_exporter/metrics"
	"github.com/prometheus/client_golang/prometheus"

	. "github.com/cloudfoundry-community/firehose_exporter/collectors"
	. "github.com/cloudfoundry-community/firehose_exporter/utils/test_matchers"
)

var _ = Describe("InternalMetricsCollector", func() {
	var (
		namespace                string
		environment              string
		metricsStore             *metrics.Store
		metricsExpiration        time.Duration
		metricsCleanupInterval   time.Duration
		deploymentFilter         *filters.DeploymentFilter
		eventFilter              *filters.EventFilter
		internalMetricsCollector *InternalMetricsCollector

		totalEnvelopesReceivedMetric               prometheus.Gauge
		lastEnvelopeReceivedTimestampMetric        prometheus.Gauge
		totalMetricsReceivedMetric                 prometheus.Gauge
		lastMetricReceivedTimestampMetric          prometheus.Gauge
		totalContainerMetricsReceivedMetric        prometheus.Gauge
		totalContainerMetricsProcessedMetric       prometheus.Gauge
		containerMetricsCachedMetric               prometheus.Gauge
		lastContainerMetricReceivedTimestampMetric prometheus.Gauge
		totalCounterEventsReceivedMetric           prometheus.Gauge
		totalCounterEventsProcessedMetric          prometheus.Gauge
		counterEventsCachedMetric                  prometheus.Gauge
		lastCounterEventReceivedTimestampMetric    prometheus.Gauge
		totalHttpStartStopReceivedMetric           prometheus.Gauge
		totalHttpStartStopProcessedMetric          prometheus.Gauge
		httpStartStopCachedMetric                  prometheus.Gauge
		lastHttpStartStopReceivedTimestampMetric   prometheus.Gauge
		totalValueMetricsReceivedMetric            prometheus.Gauge
		totalValueMetricsProcessedMetric           prometheus.Gauge
		valueMetricsCachedMetric                   prometheus.Gauge
		lastValueMetricReceivedTimestampMetric     prometheus.Gauge
		slowConsumerAlertMetric                    prometheus.Gauge
		lastSlowConsumerAlertTimestampMetric       prometheus.Gauge
	)

	BeforeEach(func() {
		namespace = "test_exporter"
		environment = "test_environment"
		deploymentFilter = filters.NewDeploymentFilter([]string{})
		eventFilter, _ = filters.NewEventFilter([]string{})
		metricsStore = metrics.NewStore(metricsExpiration, metricsCleanupInterval, deploymentFilter, eventFilter)

		totalEnvelopesReceivedMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "total_envelopes_received",
				Help:        "Total number of envelopes received from Cloud Foundry Firehose.",
				ConstLabels: prometheus.Labels{"environment": environment},
			},
		)

		lastEnvelopeReceivedTimestampMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_envelope_received_timestamp",
				Help:        "Number of seconds since 1970 since last envelope received from Cloud Foundry Firehose.",
				ConstLabels: prometheus.Labels{"environment": environment},
			},
		)

		totalMetricsReceivedMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "total_metrics_received",
				Help:        "Total number of metrics received from Cloud Foundry Firehose.",
				ConstLabels: prometheus.Labels{"environment": environment},
			},
		)

		lastMetricReceivedTimestampMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_metric_received_timestamp",
				Help:        "Number of seconds since 1970 since last metric received from Cloud Foundry Firehose.",
				ConstLabels: prometheus.Labels{"environment": environment},
			},
		)

		totalContainerMetricsReceivedMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "total_container_metrics_received",
				Help:        "Total number of container metrics received from Cloud Foundry Firehose.",
				ConstLabels: prometheus.Labels{"environment": environment},
			},
		)

		totalContainerMetricsProcessedMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "total_container_metrics_processed",
				Help:        "Total number of container metrics processed from Cloud Foundry Firehose.",
				ConstLabels: prometheus.Labels{"environment": environment},
			},
		)

		containerMetricsCachedMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "container_metrics_cached",
				Help:        "Number of container metrics cached from Cloud Foundry Firehose.",
				ConstLabels: prometheus.Labels{"environment": environment},
			},
		)

		lastContainerMetricReceivedTimestampMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_container_metric_received_timestamp",
				Help:        "Number of seconds since 1970 since last container metric received from Cloud Foundry Firehose.",
				ConstLabels: prometheus.Labels{"environment": environment},
			},
		)

		totalCounterEventsReceivedMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "total_counter_events_received",
				Help:        "Total number of counter events received from Cloud Foundry Firehose.",
				ConstLabels: prometheus.Labels{"environment": environment},
			},
		)

		totalCounterEventsProcessedMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "total_counter_events_processed",
				Help:        "Total number of counter events processed from Cloud Foundry Firehose.",
				ConstLabels: prometheus.Labels{"environment": environment},
			},
		)

		counterEventsCachedMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "counter_events_cached",
				Help:        "Number of counter events cached from Cloud Foundry Firehose.",
				ConstLabels: prometheus.Labels{"environment": environment},
			},
		)

		lastCounterEventReceivedTimestampMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_counter_event_received_timestamp",
				Help:        "Number of seconds since 1970 since last counter event received from Cloud Foundry Firehose.",
				ConstLabels: prometheus.Labels{"environment": environment},
			},
		)

		totalHttpStartStopReceivedMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "total_http_start_stop_received",
				Help:        "Total number of http start stop received from Cloud Foundry Firehose.",
				ConstLabels: prometheus.Labels{"environment": environment},
			},
		)

		totalHttpStartStopProcessedMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "total_http_start_stop_processed",
				Help:        "Total number of http start stop processed from Cloud Foundry Firehose.",
				ConstLabels: prometheus.Labels{"environment": environment},
			},
		)

		httpStartStopCachedMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "http_start_stop_cached",
				Help:        "Number of http start stop cached from Cloud Foundry Firehose.",
				ConstLabels: prometheus.Labels{"environment": environment},
			},
		)

		lastHttpStartStopReceivedTimestampMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_http_start_stop_received_timestamp",
				Help:        "Number of seconds since 1970 since last http start stop received from Cloud Foundry Firehose.",
				ConstLabels: prometheus.Labels{"environment": environment},
			},
		)

		totalValueMetricsReceivedMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "total_value_metrics_received",
				Help:        "Total number of value metrics received from Cloud Foundry Firehose.",
				ConstLabels: prometheus.Labels{"environment": environment},
			},
		)

		totalValueMetricsProcessedMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "total_value_metrics_processed",
				Help:        "Total number of value metrics processed from Cloud Foundry Firehose.",
				ConstLabels: prometheus.Labels{"environment": environment},
			},
		)

		valueMetricsCachedMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "value_metrics_cached",
				Help:        "Number of value metrics cached from Cloud Foundry Firehose.",
				ConstLabels: prometheus.Labels{"environment": environment},
			},
		)

		lastValueMetricReceivedTimestampMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_value_metric_received_timestamp",
				Help:        "Number of seconds since 1970 since last value metric received from Cloud Foundry Firehose.",
				ConstLabels: prometheus.Labels{"environment": environment},
			},
		)

		slowConsumerAlertMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "slow_consumer_alert",
				Help:        "Nozzle could not keep up with Cloud Foundry Firehose.",
				ConstLabels: prometheus.Labels{"environment": environment},
			},
		)

		lastSlowConsumerAlertTimestampMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_slow_consumer_alert_timestamp",
				Help:        "Number of seconds since 1970 since last slow consumer alert received from Cloud Foundry Firehose.",
				ConstLabels: prometheus.Labels{"environment": environment},
			},
		)
	})

	JustBeforeEach(func() {
		internalMetricsCollector = NewInternalMetricsCollector(namespace, environment, metricsStore)
	})

	Describe("Describe", func() {
		var (
			descriptions chan *prometheus.Desc
		)

		BeforeEach(func() {
			descriptions = make(chan *prometheus.Desc)
		})

		JustBeforeEach(func() {
			go internalMetricsCollector.Describe(descriptions)
		})

		It("returns a total_envelopes_received metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(totalEnvelopesReceivedMetric.Desc())))
		})

		It("returns a last_envelope_received_timestamp metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastEnvelopeReceivedTimestampMetric.Desc())))
		})

		It("returns a total_metrics_received metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(totalMetricsReceivedMetric.Desc())))
		})

		It("returns a last_metric_received_timestamp metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastMetricReceivedTimestampMetric.Desc())))
		})

		It("returns a total_container_metrics_received metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(totalContainerMetricsReceivedMetric.Desc())))
		})

		It("returns a total_container_metrics_processed metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(totalContainerMetricsProcessedMetric.Desc())))
		})

		It("returns a container_metrics_cached metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(containerMetricsCachedMetric.Desc())))
		})

		It("returns a last_container_metric_received_timestamp metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastContainerMetricReceivedTimestampMetric.Desc())))
		})

		It("returns a total_counter_events_received metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(totalCounterEventsReceivedMetric.Desc())))
		})

		It("returns a total_counter_events_processed metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(totalCounterEventsProcessedMetric.Desc())))
		})

		It("returns a counter_events_cached metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(counterEventsCachedMetric.Desc())))
		})

		It("returns a last_counter_event_received_timestamp metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastCounterEventReceivedTimestampMetric.Desc())))
		})

		It("returns a total_http_start_stop_received metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(totalHttpStartStopReceivedMetric.Desc())))
		})

		It("returns a total_http_start_stop_processed metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(totalHttpStartStopProcessedMetric.Desc())))
		})

		It("returns a http_start_stop_cached metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(httpStartStopCachedMetric.Desc())))
		})

		It("returns a last_http_start_stop_received_timestamp metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastHttpStartStopReceivedTimestampMetric.Desc())))
		})

		It("returns a total_value_metrics_received metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(totalValueMetricsReceivedMetric.Desc())))
		})

		It("returns a total_value_metrics_processed metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(totalValueMetricsProcessedMetric.Desc())))
		})

		It("returns a value_metrics_cached metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(valueMetricsCachedMetric.Desc())))
		})

		It("returns a last_value_metric_received_timestamp metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastValueMetricReceivedTimestampMetric.Desc())))
		})

		It("returns a slow_consumer_alert metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(slowConsumerAlertMetric.Desc())))
		})

		It("returns a last_slow_consumer_alert_timestamp metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastSlowConsumerAlertTimestampMetric.Desc())))
		})
	})

	Describe("Collect", func() {
		var (
			internalMetrics                      metrics.InternalMetrics
			totalEnvelopesReceived               = int64(1000)
			lastEnvelopeReceivedTimestamp        = time.Now().Unix()
			totalMetricsReceived                 = int64(500)
			lastMetricReceivedTimestamp          = time.Now().Unix()
			totalContainerMetricsReceived        = int64(100)
			totalContainerMetricsProcessed       = int64(50)
			lastContainerMetricReceivedTimestamp = time.Now().Unix()
			totalCounterEventsReceived           = int64(200)
			totalCounterEventsProcessed          = int64(100)
			lastCounterEventReceivedTimestamp    = time.Now().Unix()
			totalHttpStartStopReceived           = int64(300)
			totalHttpStartStopProcessed          = int64(150)
			lastHttpStartStopReceivedTimestamp   = time.Now().Unix()
			totalValueMetricsReceived            = int64(400)
			totalValueMetricsProcessed           = int64(200)
			lastValueMetricReceivedTimestamp     = time.Now().Unix()
			slowConsumerAlert                    = false
			lastSlowConsumerAlertTimestamp       = time.Now().Unix()

			internalMetricsChan chan prometheus.Metric
		)

		BeforeEach(func() {
			internalMetrics = metrics.InternalMetrics{
				TotalEnvelopesReceived:               totalEnvelopesReceived,
				LastEnvelopReceivedTimestamp:         lastEnvelopeReceivedTimestamp,
				TotalMetricsReceived:                 totalMetricsReceived,
				LastMetricReceivedTimestamp:          lastMetricReceivedTimestamp,
				TotalContainerMetricsReceived:        totalContainerMetricsReceived,
				TotalContainerMetricsProcessed:       totalContainerMetricsProcessed,
				LastContainerMetricReceivedTimestamp: lastContainerMetricReceivedTimestamp,
				TotalCounterEventsReceived:           totalCounterEventsReceived,
				TotalCounterEventsProcessed:          totalCounterEventsProcessed,
				LastCounterEventReceivedTimestamp:    lastCounterEventReceivedTimestamp,
				TotalHttpStartStopReceived:           totalHttpStartStopReceived,
				TotalHttpStartStopProcessed:          totalHttpStartStopProcessed,
				LastHttpStartStopReceivedTimestamp:   lastHttpStartStopReceivedTimestamp,
				TotalValueMetricsReceived:            totalValueMetricsReceived,
				TotalValueMetricsProcessed:           totalValueMetricsProcessed,
				LastValueMetricReceivedTimestamp:     lastValueMetricReceivedTimestamp,
				SlowConsumerAlert:                    slowConsumerAlert,
				LastSlowConsumerAlertTimestamp:       lastSlowConsumerAlertTimestamp,
			}

			internalMetricsChan = make(chan prometheus.Metric)

			totalEnvelopesReceivedMetric.Set(
				float64(totalEnvelopesReceived),
			)

			lastEnvelopeReceivedTimestampMetric.Set(float64(lastEnvelopeReceivedTimestamp))

			totalMetricsReceivedMetric.Set(float64(totalMetricsReceived))

			lastMetricReceivedTimestampMetric.Set(float64(lastMetricReceivedTimestamp))

			totalContainerMetricsReceivedMetric.Set(float64(totalContainerMetricsReceived))

			totalContainerMetricsProcessedMetric.Set(float64(totalContainerMetricsProcessed))

			containerMetricsCachedMetric.Set(float64(0))

			lastContainerMetricReceivedTimestampMetric.Set(float64(lastContainerMetricReceivedTimestamp))

			totalCounterEventsReceivedMetric.Set(float64(totalCounterEventsReceived))

			totalCounterEventsProcessedMetric.Set(float64(totalCounterEventsProcessed))

			counterEventsCachedMetric.Set(float64(0))

			lastCounterEventReceivedTimestampMetric.Set(float64(lastCounterEventReceivedTimestamp))

			totalHttpStartStopReceivedMetric.Set(float64(totalHttpStartStopReceived))

			totalHttpStartStopProcessedMetric.Set(float64(totalHttpStartStopProcessed))

			httpStartStopCachedMetric.Set(float64(0))

			lastHttpStartStopReceivedTimestampMetric.Set(float64(lastHttpStartStopReceivedTimestamp))

			totalValueMetricsReceivedMetric.Set(float64(totalValueMetricsReceived))

			totalValueMetricsProcessedMetric.Set(float64(totalValueMetricsProcessed))

			valueMetricsCachedMetric.Set(float64(0))

			lastValueMetricReceivedTimestampMetric.Set(float64(lastValueMetricReceivedTimestamp))

			slowConsumerAlertMetric.Set(0)

			lastSlowConsumerAlertTimestampMetric.Set(float64(lastSlowConsumerAlertTimestamp))
		})

		JustBeforeEach(func() {
			metricsStore.SetInternalMetrics(internalMetrics)
			go internalMetricsCollector.Collect(internalMetricsChan)
		})

		It("returns a total_envelopes_received metric", func() {
			Eventually(internalMetricsChan).Should(Receive(PrometheusMetric(totalEnvelopesReceivedMetric)))
		})

		It("returns a last_envelope_received_timestamp metric", func() {
			Eventually(internalMetricsChan).Should(Receive(PrometheusMetric(lastEnvelopeReceivedTimestampMetric)))
		})

		It("returns a total_metrics_received metric", func() {
			Eventually(internalMetricsChan).Should(Receive(PrometheusMetric(totalMetricsReceivedMetric)))
		})

		It("returns a last_metric_received_timestamp metric", func() {
			Eventually(internalMetricsChan).Should(Receive(PrometheusMetric(lastMetricReceivedTimestampMetric)))
		})

		It("returns a total_container_metrics_received metric", func() {
			Eventually(internalMetricsChan).Should(Receive(PrometheusMetric(totalContainerMetricsReceivedMetric)))
		})

		It("returns a total_container_metrics_processed metric", func() {
			Eventually(internalMetricsChan).Should(Receive(PrometheusMetric(totalContainerMetricsProcessedMetric)))
		})

		It("returns a container_metrics_cached metric", func() {
			Eventually(internalMetricsChan).Should(Receive(PrometheusMetric(containerMetricsCachedMetric)))
		})

		It("returns a last_container_metric_received_timestamp", func() {
			Eventually(internalMetricsChan).Should(Receive(PrometheusMetric(lastContainerMetricReceivedTimestampMetric)))
		})

		It("returns a total_counter_events_received metric", func() {
			Eventually(internalMetricsChan).Should(Receive(PrometheusMetric(totalCounterEventsReceivedMetric)))
		})

		It("returns a total_counter_events_processed metric", func() {
			Eventually(internalMetricsChan).Should(Receive(PrometheusMetric(totalCounterEventsProcessedMetric)))
		})

		It("returns a counter_events_cached metric", func() {
			Eventually(internalMetricsChan).Should(Receive(PrometheusMetric(counterEventsCachedMetric)))
		})

		It("returns a last_counter_event_received_timestamp metric", func() {
			Eventually(internalMetricsChan).Should(Receive(PrometheusMetric(lastCounterEventReceivedTimestampMetric)))
		})

		It("returns a total_http_start_stop_received metric", func() {
			Eventually(internalMetricsChan).Should(Receive(PrometheusMetric(totalHttpStartStopReceivedMetric)))
		})

		It("returns a total_http_start_stop_processed metric", func() {
			Eventually(internalMetricsChan).Should(Receive(PrometheusMetric(totalHttpStartStopProcessedMetric)))
		})

		It("returns a http_start_stop_cached metric", func() {
			Eventually(internalMetricsChan).Should(Receive(PrometheusMetric(httpStartStopCachedMetric)))
		})

		It("returns a last_http_start_stop_received_timestamp metric", func() {
			Eventually(internalMetricsChan).Should(Receive(PrometheusMetric(lastHttpStartStopReceivedTimestampMetric)))
		})

		It("returns a total_value_metrics_received metric", func() {
			Eventually(internalMetricsChan).Should(Receive(PrometheusMetric(totalValueMetricsReceivedMetric)))
		})

		It("returns a total_value_metrics_processed metric", func() {
			Eventually(internalMetricsChan).Should(Receive(PrometheusMetric(totalValueMetricsProcessedMetric)))
		})

		It("returns a value_metrics_cached metric", func() {
			Eventually(internalMetricsChan).Should(Receive(PrometheusMetric(valueMetricsCachedMetric)))
		})

		It("returns a last_value_metric_received_timestamp metric", func() {
			Eventually(internalMetricsChan).Should(Receive(PrometheusMetric(lastValueMetricReceivedTimestampMetric)))
		})

		It("returns a slow_consumer_alert metric", func() {
			Eventually(internalMetricsChan).Should(Receive(PrometheusMetric(slowConsumerAlertMetric)))
		})

		Context("when SlowConsumerAlert is true", func() {
			BeforeEach(func() {
				slowConsumerAlert = true
				internalMetrics.SlowConsumerAlert = slowConsumerAlert

				slowConsumerAlertMetric.Set(1)
			})

			It("returns a slow_consumer_alert metric", func() {
				Eventually(internalMetricsChan).Should(Receive(PrometheusMetric(slowConsumerAlertMetric)))
			})

		})

		It("returns a last_slow_consumer_alert_timestamp metric", func() {
			Eventually(internalMetricsChan).Should(Receive(PrometheusMetric(lastSlowConsumerAlertTimestampMetric)))
		})
	})
})
