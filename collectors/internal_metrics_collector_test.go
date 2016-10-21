package collectors_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

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
		internalMetricsCollector *InternalMetricsCollector

		totalEnvelopesReceivedDesc               *prometheus.Desc
		lastEnvelopeReceivedTimestampDesc        *prometheus.Desc
		totalMetricsReceivedDesc                 *prometheus.Desc
		lastMetricReceivedTimestampDesc          *prometheus.Desc
		totalContainerMetricsReceivedDesc        *prometheus.Desc
		lastContainerMetricReceivedTimestampDesc *prometheus.Desc
		totalCounterEventsReceivedDesc           *prometheus.Desc
		lastCounterEventReceivedTimestampDesc    *prometheus.Desc
		totalValueMetricsReceivedDesc            *prometheus.Desc
		lastValueMetricReceivedTimestampDesc     *prometheus.Desc
		slowConsumerAlertDesc                    *prometheus.Desc
		lastSlowConsumerAlertTimestampDesc       *prometheus.Desc
	)

	BeforeEach(func() {
		namespace = "test_exporter"
		metricsStore = metrics.NewStore(metricsExpiration, metricsCleanupInterval)

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

		lastCounterEventReceivedTimestampDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "last_counter_event_received_timestamp"),
			"Number of seconds since 1970 since last counter event received from Cloud Foundry Firehose.",
			[]string{},
			nil,
		)

		totalValueMetricsReceivedDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "total_value_metrics_received"),
			"Total number of value metrics received from Cloud Foundry Firehose.",
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

		It("returns a last_container_metric_received_timestamp metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastContainerMetricReceivedTimestampDesc)))
		})

		It("returns a total_counter_events_received metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(totalCounterEventsReceivedDesc)))
		})

		It("returns a last_counter_event_received_timestamp metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastCounterEventReceivedTimestampDesc)))
		})

		It("returns a total_value_metrics_received metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(totalValueMetricsReceivedDesc)))
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
			lastEnvelopeReceivedTimestamp        = time.Now().UnixNano()
			totalMetricsReceived                 = int64(500)
			lastMetricReceivedTimestamp          = time.Now().UnixNano()
			totalContainerMetricsReceived        = int64(100)
			lastContainerMetricReceivedTimestamp = time.Now().UnixNano()
			totalCounterEventsReceived           = int64(200)
			lastCounterEventReceivedTimestamp    = time.Now().UnixNano()
			totalValueMetricsReceived            = int64(300)
			lastValueMetricReceivedTimestamp     = time.Now().UnixNano()
			slowConsumerAlert                    = false
			lastSlowConsumerAlertTimestamp       = time.Now().UnixNano()

			internalMetricsChan                        chan prometheus.Metric
			totalEnvelopesReceivedMetric               prometheus.Metric
			lastEnvelopeReceivedTimestampMetric        prometheus.Metric
			totalMetricsReceivedMetric                 prometheus.Metric
			lastMetricReceivedTimestampMetric          prometheus.Metric
			totalContainerMetricsReceivedMetric        prometheus.Metric
			lastContainerMetricReceivedTimestampMetric prometheus.Metric
			totalCounterEventsReceivedMetric           prometheus.Metric
			lastCounterEventReceivedTimestampMetric    prometheus.Metric
			totalValueMetricsReceivedMetric            prometheus.Metric
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
				LastContainerMetricReceivedTimestamp: lastContainerMetricReceivedTimestamp,
				TotalCounterEventsReceived:           totalCounterEventsReceived,
				LastCounterEventReceivedTimestamp:    lastCounterEventReceivedTimestamp,
				TotalValueMetricsReceived:            totalValueMetricsReceived,
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

			lastCounterEventReceivedTimestampMetric = prometheus.MustNewConstMetric(
				lastCounterEventReceivedTimestampDesc,
				prometheus.GaugeValue,
				float64(lastCounterEventReceivedTimestamp),
			)

			totalValueMetricsReceivedMetric = prometheus.MustNewConstMetric(
				totalValueMetricsReceivedDesc,
				prometheus.CounterValue,
				float64(totalValueMetricsReceived),
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

		It("returns a last_container_metric_received_timestamp", func() {
			Eventually(internalMetricsChan).Should(Receive(Equal(lastContainerMetricReceivedTimestampMetric)))
		})

		It("returns a total_counter_events_received metric", func() {
			Eventually(internalMetricsChan).Should(Receive(Equal(totalCounterEventsReceivedMetric)))
		})

		It("returns a last_counter_event_received_timestamp metric", func() {
			Eventually(internalMetricsChan).Should(Receive(Equal(lastCounterEventReceivedTimestampMetric)))
		})

		It("returns a total_value_metrics_received metric", func() {
			Eventually(internalMetricsChan).Should(Receive(Equal(totalValueMetricsReceivedMetric)))
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
