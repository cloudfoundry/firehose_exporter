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

var _ = Describe("InternalMetricsCollector", func() {
	var (
		namespace                string
		metricsStore             *metrics.Store
		metricsExpiration        time.Duration
		metricsCleanupInterval   time.Duration
		deploymentFilter         *filters.DeploymentFilter
		eventFilter              *filters.EventFilter
		internalMetricsCollector *InternalMetricsCollector

		totalEnvelopesReceivedDesc               *prometheus.Desc
		lastEnvelopeReceivedTimestampDesc        *prometheus.Desc
		totalMetricsReceivedDesc                 *prometheus.Desc
		lastMetricReceivedTimestampDesc          *prometheus.Desc
		totalContainerMetricsReceivedDesc        *prometheus.Desc
		totalContainerMetricsProcessedDesc       *prometheus.Desc
		containerMetricsCachedDesc               *prometheus.Desc
		lastContainerMetricReceivedTimestampDesc *prometheus.Desc
		totalCounterEventsReceivedDesc           *prometheus.Desc
		totalCounterEventsProcessedDesc          *prometheus.Desc
		counterEventsCachedDesc                  *prometheus.Desc
		lastCounterEventReceivedTimestampDesc    *prometheus.Desc
		totalHttpStartStopReceivedDesc           *prometheus.Desc
		totalHttpStartStopProcessedDesc          *prometheus.Desc
		httpStartStopCachedDesc                  *prometheus.Desc
		lastHttpStartStopReceivedTimestampDesc   *prometheus.Desc
		totalValueMetricsReceivedDesc            *prometheus.Desc
		totalValueMetricsProcessedDesc           *prometheus.Desc
		valueMetricsCachedDesc                   *prometheus.Desc
		lastValueMetricReceivedTimestampDesc     *prometheus.Desc
		slowConsumerAlertDesc                    *prometheus.Desc
		lastSlowConsumerAlertTimestampDesc       *prometheus.Desc
	)

	BeforeEach(func() {
		namespace = "test_exporter"
		deploymentFilter = filters.NewDeploymentFilter([]string{})
		eventFilter, _ = filters.NewEventFilter([]string{})
		metricsStore = metrics.NewStore(metricsExpiration, metricsCleanupInterval, deploymentFilter, eventFilter)

		totalEnvelopesReceivedDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "total_envelopes_received"),
			"Total number of envelopes received from Cloud Foundry Firehose.",
			[]string{},
			nil,
		)

		lastEnvelopeReceivedTimestampDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "last_envelope_received_timestamp"),
			"Number of seconds since 1970 since last envelope received from Cloud Foundry Firehose.",
			[]string{},
			nil,
		)

		totalMetricsReceivedDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "total_metrics_received"),
			"Total number of metrics received from Cloud Foundry Firehose.",
			[]string{},
			nil,
		)

		lastMetricReceivedTimestampDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "last_metric_received_timestamp"),
			"Number of seconds since 1970 since last metric received from Cloud Foundry Firehose.",
			[]string{},
			nil,
		)

		totalContainerMetricsReceivedDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "total_container_metrics_received"),
			"Total number of container metrics received from Cloud Foundry Firehose.",
			[]string{},
			nil,
		)

		totalContainerMetricsProcessedDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "total_container_metrics_processed"),
			"Total number of container metrics processed from Cloud Foundry Firehose.",
			[]string{},
			nil,
		)

		containerMetricsCachedDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "container_metrics_cached"),
			"Number of container metrics cached from Cloud Foundry Firehose.",
			[]string{},
			nil,
		)

		lastContainerMetricReceivedTimestampDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "last_container_metric_received_timestamp"),
			"Number of seconds since 1970 since last container metric received from Cloud Foundry Firehose.",
			[]string{},
			nil,
		)

		totalCounterEventsReceivedDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "total_counter_events_received"),
			"Total number of counter events received from Cloud Foundry Firehose.",
			[]string{},
			nil,
		)

		totalCounterEventsProcessedDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "total_counter_events_processed"),
			"Total number of counter events processed from Cloud Foundry Firehose.",
			[]string{},
			nil,
		)

		counterEventsCachedDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "counter_events_cached"),
			"Number of counter events cached from Cloud Foundry Firehose.",
			[]string{},
			nil,
		)

		lastCounterEventReceivedTimestampDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "last_counter_event_received_timestamp"),
			"Number of seconds since 1970 since last counter event received from Cloud Foundry Firehose.",
			[]string{},
			nil,
		)

		totalHttpStartStopReceivedDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "total_http_start_stop_received"),
			"Total number of http start stop received from Cloud Foundry Firehose.",
			[]string{},
			nil,
		)

		totalHttpStartStopProcessedDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "total_http_start_stop_processed"),
			"Total number of http start stop processed from Cloud Foundry Firehose.",
			[]string{},
			nil,
		)

		httpStartStopCachedDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "http_start_stop_cached"),
			"Number of http start stop cached from Cloud Foundry Firehose.",
			[]string{},
			nil,
		)

		lastHttpStartStopReceivedTimestampDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "last_http_start_stop_received_timestamp"),
			"Number of seconds since 1970 since last http start stop received from Cloud Foundry Firehose.",
			[]string{},
			nil,
		)

		totalValueMetricsReceivedDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "total_value_metrics_received"),
			"Total number of value metrics received from Cloud Foundry Firehose.",
			[]string{},
			nil,
		)

		totalValueMetricsProcessedDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "total_value_metrics_processed"),
			"Total number of value metrics processed from Cloud Foundry Firehose.",
			[]string{},
			nil,
		)

		valueMetricsCachedDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "value_metrics_cached"),
			"Number of value metrics cached from Cloud Foundry Firehose.",
			[]string{},
			nil,
		)

		lastValueMetricReceivedTimestampDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "last_value_metric_received_timestamp"),
			"Number of seconds since 1970 since last value metric received from Cloud Foundry Firehose.",
			[]string{},
			nil,
		)

		slowConsumerAlertDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "slow_consumer_alert"),
			"Nozzle could not keep up with Cloud Foundry Firehose.",
			[]string{},
			nil,
		)

		lastSlowConsumerAlertTimestampDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "last_slow_consumer_alert_timestamp"),
			"Number of seconds since 1970 since last slow consumer alert received from Cloud Foundry Firehose.",
			[]string{},
			nil,
		)
	})

	JustBeforeEach(func() {
		internalMetricsCollector = NewInternalMetricsCollector(namespace, metricsStore)
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
			Eventually(descriptions).Should(Receive(Equal(totalEnvelopesReceivedDesc)))
		})

		It("returns a last_envelope_received_timestamp metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastEnvelopeReceivedTimestampDesc)))
		})

		It("returns a total_metrics_received metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(totalMetricsReceivedDesc)))
		})

		It("returns a last_metric_received_timestamp metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastMetricReceivedTimestampDesc)))
		})

		It("returns a total_container_metrics_received metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(totalContainerMetricsReceivedDesc)))
		})

		It("returns a total_container_metrics_processed metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(totalContainerMetricsProcessedDesc)))
		})

		It("returns a container_metrics_cached metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(containerMetricsCachedDesc)))
		})

		It("returns a last_container_metric_received_timestamp metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastContainerMetricReceivedTimestampDesc)))
		})

		It("returns a total_counter_events_received metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(totalCounterEventsReceivedDesc)))
		})

		It("returns a total_counter_events_processed metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(totalCounterEventsProcessedDesc)))
		})

		It("returns a counter_events_cached metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(counterEventsCachedDesc)))
		})

		It("returns a last_counter_event_received_timestamp metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastCounterEventReceivedTimestampDesc)))
		})

		It("returns a total_http_start_stop_received metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(totalHttpStartStopReceivedDesc)))
		})

		It("returns a total_http_start_stop_processed metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(totalHttpStartStopProcessedDesc)))
		})

		It("returns a http_start_stop_cached metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(httpStartStopCachedDesc)))
		})

		It("returns a last_http_start_stop_received_timestamp metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastHttpStartStopReceivedTimestampDesc)))
		})

		It("returns a total_value_metrics_received metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(totalValueMetricsReceivedDesc)))
		})

		It("returns a total_value_metrics_processed metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(totalValueMetricsProcessedDesc)))
		})

		It("returns a value_metrics_cached metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(valueMetricsCachedDesc)))
		})

		It("returns a last_value_metric_received_timestamp metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastValueMetricReceivedTimestampDesc)))
		})

		It("returns a slow_consumer_alert metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(slowConsumerAlertDesc)))
		})

		It("returns a last_slow_consumer_alert_timestamp metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastSlowConsumerAlertTimestampDesc)))
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

			internalMetricsChan                        chan prometheus.Metric
			totalEnvelopesReceivedMetric               prometheus.Metric
			lastEnvelopeReceivedTimestampMetric        prometheus.Metric
			totalMetricsReceivedMetric                 prometheus.Metric
			lastMetricReceivedTimestampMetric          prometheus.Metric
			totalContainerMetricsReceivedMetric        prometheus.Metric
			totalContainerMetricsProcessedMetric       prometheus.Metric
			containerMetricsCachedMetric               prometheus.Metric
			lastContainerMetricReceivedTimestampMetric prometheus.Metric
			totalCounterEventsReceivedMetric           prometheus.Metric
			totalCounterEventsProcessedMetric          prometheus.Metric
			counterEventsCachedMetric                  prometheus.Metric
			lastCounterEventReceivedTimestampMetric    prometheus.Metric
			totalHttpStartStopReceivedMetric           prometheus.Metric
			totalHttpStartStopProcessedMetric          prometheus.Metric
			httpStartStopCachedMetric                  prometheus.Metric
			lastHttpStartStopReceivedTimestampMetric   prometheus.Metric
			totalValueMetricsReceivedMetric            prometheus.Metric
			totalValueMetricsProcessedMetric           prometheus.Metric
			valueMetricsCachedMetric                   prometheus.Metric
			lastValueMetricReceivedTimestampMetric     prometheus.Metric
			slowConsumerAlertMetric                    prometheus.Metric
			lastSlowConsumerAlertTimestampMetric       prometheus.Metric
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

			totalEnvelopesReceivedMetric = prometheus.MustNewConstMetric(
				totalEnvelopesReceivedDesc,
				prometheus.CounterValue,
				float64(totalEnvelopesReceived),
			)

			lastEnvelopeReceivedTimestampMetric = prometheus.MustNewConstMetric(
				lastEnvelopeReceivedTimestampDesc,
				prometheus.GaugeValue,
				float64(lastEnvelopeReceivedTimestamp),
			)

			totalMetricsReceivedMetric = prometheus.MustNewConstMetric(
				totalMetricsReceivedDesc,
				prometheus.CounterValue,
				float64(totalMetricsReceived),
			)

			lastMetricReceivedTimestampMetric = prometheus.MustNewConstMetric(
				lastMetricReceivedTimestampDesc,
				prometheus.GaugeValue,
				float64(lastMetricReceivedTimestamp),
			)

			totalContainerMetricsReceivedMetric = prometheus.MustNewConstMetric(
				totalContainerMetricsReceivedDesc,
				prometheus.CounterValue,
				float64(totalContainerMetricsReceived),
			)

			totalContainerMetricsProcessedMetric = prometheus.MustNewConstMetric(
				totalContainerMetricsProcessedDesc,
				prometheus.CounterValue,
				float64(totalContainerMetricsProcessed),
			)

			containerMetricsCachedMetric = prometheus.MustNewConstMetric(
				containerMetricsCachedDesc,
				prometheus.GaugeValue,
				float64(0),
			)

			lastContainerMetricReceivedTimestampMetric = prometheus.MustNewConstMetric(
				lastContainerMetricReceivedTimestampDesc,
				prometheus.GaugeValue,
				float64(lastContainerMetricReceivedTimestamp),
			)

			totalCounterEventsReceivedMetric = prometheus.MustNewConstMetric(
				totalCounterEventsReceivedDesc,
				prometheus.CounterValue,
				float64(totalCounterEventsReceived),
			)

			totalCounterEventsProcessedMetric = prometheus.MustNewConstMetric(
				totalCounterEventsProcessedDesc,
				prometheus.CounterValue,
				float64(totalCounterEventsProcessed),
			)

			counterEventsCachedMetric = prometheus.MustNewConstMetric(
				counterEventsCachedDesc,
				prometheus.GaugeValue,
				float64(0),
			)

			lastCounterEventReceivedTimestampMetric = prometheus.MustNewConstMetric(
				lastCounterEventReceivedTimestampDesc,
				prometheus.GaugeValue,
				float64(lastCounterEventReceivedTimestamp),
			)

			totalHttpStartStopReceivedMetric = prometheus.MustNewConstMetric(
				totalHttpStartStopReceivedDesc,
				prometheus.CounterValue,
				float64(totalHttpStartStopReceived),
			)

			totalHttpStartStopProcessedMetric = prometheus.MustNewConstMetric(
				totalHttpStartStopProcessedDesc,
				prometheus.CounterValue,
				float64(totalHttpStartStopProcessed),
			)

			httpStartStopCachedMetric = prometheus.MustNewConstMetric(
				httpStartStopCachedDesc,
				prometheus.GaugeValue,
				float64(0),
			)

			lastHttpStartStopReceivedTimestampMetric = prometheus.MustNewConstMetric(
				lastHttpStartStopReceivedTimestampDesc,
				prometheus.GaugeValue,
				float64(lastHttpStartStopReceivedTimestamp),
			)

			totalValueMetricsReceivedMetric = prometheus.MustNewConstMetric(
				totalValueMetricsReceivedDesc,
				prometheus.CounterValue,
				float64(totalValueMetricsReceived),
			)

			totalValueMetricsProcessedMetric = prometheus.MustNewConstMetric(
				totalValueMetricsProcessedDesc,
				prometheus.CounterValue,
				float64(totalValueMetricsProcessed),
			)

			valueMetricsCachedMetric = prometheus.MustNewConstMetric(
				valueMetricsCachedDesc,
				prometheus.GaugeValue,
				float64(0),
			)

			lastValueMetricReceivedTimestampMetric = prometheus.MustNewConstMetric(
				lastValueMetricReceivedTimestampDesc,
				prometheus.GaugeValue,
				float64(lastValueMetricReceivedTimestamp),
			)

			slowConsumerAlertMetric = prometheus.MustNewConstMetric(
				slowConsumerAlertDesc,
				prometheus.UntypedValue,
				0,
			)

			lastSlowConsumerAlertTimestampMetric = prometheus.MustNewConstMetric(
				lastSlowConsumerAlertTimestampDesc,
				prometheus.GaugeValue,
				float64(lastSlowConsumerAlertTimestamp),
			)
		})

		JustBeforeEach(func() {
			metricsStore.SetInternalMetrics(internalMetrics)
			go internalMetricsCollector.Collect(internalMetricsChan)
		})

		It("returns a total_envelopes_received metric", func() {
			Eventually(internalMetricsChan).Should(Receive(Equal(totalEnvelopesReceivedMetric)))
		})

		It("returns a last_envelope_received_timestamp metric", func() {
			Eventually(internalMetricsChan).Should(Receive(Equal(lastEnvelopeReceivedTimestampMetric)))
		})

		It("returns a total_metrics_received metric", func() {
			Eventually(internalMetricsChan).Should(Receive(Equal(totalMetricsReceivedMetric)))
		})

		It("returns a last_metric_received_timestamp metric", func() {
			Eventually(internalMetricsChan).Should(Receive(Equal(lastMetricReceivedTimestampMetric)))
		})

		It("returns a total_container_metrics_received metric", func() {
			Eventually(internalMetricsChan).Should(Receive(Equal(totalContainerMetricsReceivedMetric)))
		})

		It("returns a total_container_metrics_processed metric", func() {
			Eventually(internalMetricsChan).Should(Receive(Equal(totalContainerMetricsProcessedMetric)))
		})

		It("returns a container_metrics_cached metric", func() {
			Eventually(internalMetricsChan).Should(Receive(Equal(containerMetricsCachedMetric)))
		})

		It("returns a last_container_metric_received_timestamp", func() {
			Eventually(internalMetricsChan).Should(Receive(Equal(lastContainerMetricReceivedTimestampMetric)))
		})

		It("returns a total_counter_events_received metric", func() {
			Eventually(internalMetricsChan).Should(Receive(Equal(totalCounterEventsReceivedMetric)))
		})

		It("returns a total_counter_events_processed metric", func() {
			Eventually(internalMetricsChan).Should(Receive(Equal(totalCounterEventsProcessedMetric)))
		})

		It("returns a counter_events_cached metric", func() {
			Eventually(internalMetricsChan).Should(Receive(Equal(counterEventsCachedMetric)))
		})

		It("returns a last_counter_event_received_timestamp metric", func() {
			Eventually(internalMetricsChan).Should(Receive(Equal(lastCounterEventReceivedTimestampMetric)))
		})

		It("returns a total_http_start_stop_received metric", func() {
			Eventually(internalMetricsChan).Should(Receive(Equal(totalHttpStartStopReceivedMetric)))
		})

		It("returns a total_http_start_stop_processed metric", func() {
			Eventually(internalMetricsChan).Should(Receive(Equal(totalHttpStartStopProcessedMetric)))
		})

		It("returns a http_start_stop_cached metric", func() {
			Eventually(internalMetricsChan).Should(Receive(Equal(httpStartStopCachedMetric)))
		})

		It("returns a last_http_start_stop_received_timestamp metric", func() {
			Eventually(internalMetricsChan).Should(Receive(Equal(lastHttpStartStopReceivedTimestampMetric)))
		})

		It("returns a total_value_metrics_received metric", func() {
			Eventually(internalMetricsChan).Should(Receive(Equal(totalValueMetricsReceivedMetric)))
		})

		It("returns a total_value_metrics_processed metric", func() {
			Eventually(internalMetricsChan).Should(Receive(Equal(totalValueMetricsProcessedMetric)))
		})

		It("returns a value_metrics_cached metric", func() {
			Eventually(internalMetricsChan).Should(Receive(Equal(valueMetricsCachedMetric)))
		})

		It("returns a last_value_metric_received_timestamp metric", func() {
			Eventually(internalMetricsChan).Should(Receive(Equal(lastValueMetricReceivedTimestampMetric)))
		})

		It("returns a slow_consumer_alert metric", func() {
			Eventually(internalMetricsChan).Should(Receive(Equal(slowConsumerAlertMetric)))
		})

		Context("when SlowConsumerAlert is true", func() {
			BeforeEach(func() {
				slowConsumerAlert = true
				internalMetrics.SlowConsumerAlert = slowConsumerAlert

				slowConsumerAlertMetric = prometheus.MustNewConstMetric(
					slowConsumerAlertDesc,
					prometheus.UntypedValue,
					1,
				)
			})

			It("returns a slow_consumer_alert metric", func() {
				Eventually(internalMetricsChan).Should(Receive(Equal(slowConsumerAlertMetric)))
			})

		})

		It("returns a last_slow_consumer_alert_timestamp metric", func() {
			Eventually(internalMetricsChan).Should(Receive(Equal(lastSlowConsumerAlertTimestampMetric)))
		})
	})
})
