package collectors

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/cloudfoundry-community/firehose_exporter/metrics"
)

type internalMetricsCollector struct {
	namespace                            string
	metricsStore                         *metrics.Store
	totalEnvelopesReceivedDesc           *prometheus.Desc
	lastReceivedEnvelopeTimestampDesc    *prometheus.Desc
	totalMetricsReceivedDesc             *prometheus.Desc
	lastReceivedMetricTimestampDesc      *prometheus.Desc
	totalContainerMetricsReceivedDesc    *prometheus.Desc
	lastReceivedContainerMetricTimestamp *prometheus.Desc
	totalCounterEventsReceivedDesc       *prometheus.Desc
	lastReceivedCounterEventTimestamp    *prometheus.Desc
	totalValueMetricsReceivedDesc        *prometheus.Desc
	lastReceivedValueMetricTimestamp     *prometheus.Desc
	slowConsumerAlertDesc                *prometheus.Desc
}

func NewInternalMetricsCollector(
	namespace string,
	metricsStore *metrics.Store,
) *internalMetricsCollector {
	totalEnvelopesReceivedDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "total_envelopes_received"),
		"Total number of envelopes received from Cloud Foundry Firehose.",
		[]string{},
		nil,
	)

	lastReceivedEnvelopeTimestampDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "last_received_envelope_timestamp"),
		"Number of seconds since 1970 of last envelope received from Cloud Foundry Firehose.",
		[]string{},
		nil,
	)

	totalMetricsReceivedDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "total_metrics_received"),
		"Total number of metrics received from Cloud Foundry Firehose.",
		[]string{},
		nil,
	)

	lastReceivedMetricTimestampDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "last_received_metric_timestamp"),
		"Number of seconds since 1970 of last metric received from Cloud Foundry Firehose.",
		[]string{},
		nil,
	)

	totalContainerMetricsReceivedDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "total_container_metrics_received"),
		"Total number of container metrics received from Cloud Foundry Firehose.",
		[]string{},
		nil,
	)

	lastReceivedContainerMetricTimestamp := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "last_received_container_metric_timestamp"),
		"Number of seconds since 1970 of last container metric received from Cloud Foundry Firehose.",
		[]string{},
		nil,
	)

	totalCounterEventsReceivedDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "total_counter_events_received"),
		"Total number of counter events received from Cloud Foundry Firehose.",
		[]string{},
		nil,
	)

	lastReceivedCounterEventTimestamp := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "last_received_counter_event_timestamp"),
		"Number of seconds since 1970 of last counter event received from Cloud Foundry Firehose.",
		[]string{},
		nil,
	)

	totalValueMetricsReceivedDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "total_value_metrics_received"),
		"Total number of value metrics received from Cloud Foundry Firehose.",
		[]string{},
		nil,
	)

	lastReceivedValueMetricTimestamp := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "last_received_value_metric_timestamp"),
		"Number of seconds since 1970 of last value metric received from Cloud Foundry Firehose.",
		[]string{},
		nil,
	)

	slowConsumerAlertDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "slow_consumer"),
		"Nozzle could not keep up with Cloud Foundry Firehose.",
		[]string{},
		nil,
	)

	collector := &internalMetricsCollector{
		namespace:                            namespace,
		metricsStore:                         metricsStore,
		totalEnvelopesReceivedDesc:           totalEnvelopesReceivedDesc,
		lastReceivedEnvelopeTimestampDesc:    lastReceivedEnvelopeTimestampDesc,
		totalMetricsReceivedDesc:             totalMetricsReceivedDesc,
		lastReceivedMetricTimestampDesc:      lastReceivedMetricTimestampDesc,
		totalContainerMetricsReceivedDesc:    totalContainerMetricsReceivedDesc,
		lastReceivedContainerMetricTimestamp: lastReceivedContainerMetricTimestamp,
		totalCounterEventsReceivedDesc:       totalCounterEventsReceivedDesc,
		lastReceivedCounterEventTimestamp:    lastReceivedCounterEventTimestamp,
		totalValueMetricsReceivedDesc:        totalValueMetricsReceivedDesc,
		lastReceivedValueMetricTimestamp:     lastReceivedValueMetricTimestamp,
		slowConsumerAlertDesc:                slowConsumerAlertDesc,
	}
	return collector
}

func (c internalMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	internalMetrics := c.metricsStore.GetInternalMetrics()

	ch <- prometheus.MustNewConstMetric(
		c.totalEnvelopesReceivedDesc,
		prometheus.CounterValue,
		internalMetrics.TotalEnvelopesReceived,
	)

	ch <- prometheus.MustNewConstMetric(
		c.lastReceivedEnvelopeTimestampDesc,
		prometheus.GaugeValue,
		float64(internalMetrics.LastReceivedEnvelopTimestamp),
	)

	ch <- prometheus.MustNewConstMetric(
		c.totalMetricsReceivedDesc,
		prometheus.CounterValue,
		internalMetrics.TotalMetricsReceived,
	)

	ch <- prometheus.MustNewConstMetric(
		c.lastReceivedMetricTimestampDesc,
		prometheus.GaugeValue,
		float64(internalMetrics.LastReceivedMetricTimestamp),
	)

	ch <- prometheus.MustNewConstMetric(
		c.totalContainerMetricsReceivedDesc,
		prometheus.CounterValue,
		internalMetrics.TotalContainerMetricsReceived,
	)

	ch <- prometheus.MustNewConstMetric(
		c.lastReceivedContainerMetricTimestamp,
		prometheus.GaugeValue,
		float64(internalMetrics.LastReceivedContainerMetricTimestamp),
	)

	ch <- prometheus.MustNewConstMetric(
		c.totalCounterEventsReceivedDesc,
		prometheus.CounterValue,
		internalMetrics.TotalCounterEventsReceived,
	)

	ch <- prometheus.MustNewConstMetric(
		c.lastReceivedCounterEventTimestamp,
		prometheus.GaugeValue,
		float64(internalMetrics.LastReceivedCounterEventTimestamp),
	)

	ch <- prometheus.MustNewConstMetric(
		c.totalValueMetricsReceivedDesc,
		prometheus.CounterValue,
		internalMetrics.TotalValueMetricsReceived,
	)

	ch <- prometheus.MustNewConstMetric(
		c.lastReceivedValueMetricTimestamp,
		prometheus.GaugeValue,
		float64(internalMetrics.LastReceivedValueMetricTimestamp),
	)

	if internalMetrics.SlowConsumerAlert {
		ch <- prometheus.MustNewConstMetric(
			c.slowConsumerAlertDesc,
			prometheus.UntypedValue,
			1,
		)
	}
}

func (c internalMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.totalEnvelopesReceivedDesc
	ch <- c.lastReceivedEnvelopeTimestampDesc
	ch <- c.totalMetricsReceivedDesc
	ch <- c.lastReceivedMetricTimestampDesc
	ch <- c.totalContainerMetricsReceivedDesc
	ch <- c.lastReceivedContainerMetricTimestamp
	ch <- c.totalCounterEventsReceivedDesc
	ch <- c.lastReceivedCounterEventTimestamp
	ch <- c.totalValueMetricsReceivedDesc
	ch <- c.lastReceivedValueMetricTimestamp
	ch <- c.slowConsumerAlertDesc
}
