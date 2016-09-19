package collectors

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/frodenas/firehose_exporter/metrics"
)

var (
	totalEnvelopesReceivedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "total_envelopes_received"),
		"Total number of envelopes received from Cloud Foundry firehose.",
		[]string{},
		nil,
	)

	totalMetricsReceivedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "total_metrics_received"),
		"Total number of metrics received from Cloud Foundry firehose.",
		[]string{},
		nil,
	)

	totalContainerMetricsReceivedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "total_container_metrics_received"),
		"Total number of container metrics received from Cloud Foundry firehose.",
		[]string{},
		nil,
	)

	totalCounterEventsReceivedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "total_counter_events_received"),
		"Total number of counter events received from Cloud Foundry firehose.",
		[]string{},
		nil,
	)

	totalValueMetricsReceivedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "total_value_metrics_received"),
		"Total number of value metrics received from Cloud Foundry firehose.",
		[]string{},
		nil,
	)

	slowConsumerAlertDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "slow_consumer"),
		"Nozzle could not keep up with Cloud Foundry firehose.",
		[]string{},
		nil,
	)
)

type internalMetricsCollector struct {
	metricsStore *metrics.Store
}

func NewInternalMetricsCollector(metricsStore *metrics.Store) *internalMetricsCollector {
	collector := &internalMetricsCollector{
		metricsStore: metricsStore,
	}
	return collector
}

func (c internalMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(
		totalEnvelopesReceivedDesc,
		prometheus.CounterValue,
		c.metricsStore.GetInternalMetrics().TotalEnvelopesReceived,
	)
	ch <- prometheus.MustNewConstMetric(
		totalMetricsReceivedDesc,
		prometheus.CounterValue,
		c.metricsStore.GetInternalMetrics().TotalMetricsReceived,
	)
	ch <- prometheus.MustNewConstMetric(
		totalContainerMetricsReceivedDesc,
		prometheus.CounterValue,
		c.metricsStore.GetInternalMetrics().TotalContainerMetricsReceived,
	)
	ch <- prometheus.MustNewConstMetric(
		totalCounterEventsReceivedDesc,
		prometheus.CounterValue,
		c.metricsStore.GetInternalMetrics().TotalCounterEventsReceived,
	)
	ch <- prometheus.MustNewConstMetric(
		totalValueMetricsReceivedDesc,
		prometheus.CounterValue,
		c.metricsStore.GetInternalMetrics().TotalValueMetricsReceived,
	)

	if c.metricsStore.GetInternalMetrics().SlowConsumerAlert {
		ch <- prometheus.MustNewConstMetric(
			slowConsumerAlertDesc,
			prometheus.UntypedValue,
			1,
		)
	}
}

func (c internalMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- totalEnvelopesReceivedDesc
	ch <- totalMetricsReceivedDesc
	ch <- totalContainerMetricsReceivedDesc
	ch <- totalCounterEventsReceivedDesc
	ch <- totalValueMetricsReceivedDesc
	ch <- slowConsumerAlertDesc
}
