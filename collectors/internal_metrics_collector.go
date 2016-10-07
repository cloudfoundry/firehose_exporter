package collectors

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/cloudfoundry-community/firehose_exporter/metrics"
)

type internalMetricsCollector struct {
	namespace                         string
	metricsStore                      *metrics.Store
	totalEnvelopesReceivedDesc        *prometheus.Desc
	totalMetricsReceivedDesc          *prometheus.Desc
	totalContainerMetricsReceivedDesc *prometheus.Desc
	totalCounterEventsReceivedDesc    *prometheus.Desc
	totalValueMetricsReceivedDesc     *prometheus.Desc
	slowConsumerAlertDesc             *prometheus.Desc
	lastReceivedMetricTimestampDesc   *prometheus.Desc
}

func NewInternalMetricsCollector(
	namespace string,
	metricsStore *metrics.Store,
) *internalMetricsCollector {
	totalEnvelopesReceivedDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "total_envelopes_received"),
		"Total number of envelopes received from Cloud Foundry firehose.",
		[]string{},
		nil,
	)

	totalMetricsReceivedDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "total_metrics_received"),
		"Total number of metrics received from Cloud Foundry firehose.",
		[]string{},
		nil,
	)

	totalContainerMetricsReceivedDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "total_container_metrics_received"),
		"Total number of container metrics received from Cloud Foundry firehose.",
		[]string{},
		nil,
	)

	totalCounterEventsReceivedDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "total_counter_events_received"),
		"Total number of counter events received from Cloud Foundry firehose.",
		[]string{},
		nil,
	)

	totalValueMetricsReceivedDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "total_value_metrics_received"),
		"Total number of value metrics received from Cloud Foundry firehose.",
		[]string{},
		nil,
	)

	slowConsumerAlertDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "slow_consumer"),
		"Nozzle could not keep up with Cloud Foundry Firehose.",
		[]string{},
		nil,
	)

	lastReceivedMetricTimestampDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "last_received_metric_timestamp"),
		"Last received metric timestamp (milliseconds since epoch).",
		[]string{},
		nil,
	)

	collector := &internalMetricsCollector{
		namespace:                         namespace,
		metricsStore:                      metricsStore,
		totalEnvelopesReceivedDesc:        totalEnvelopesReceivedDesc,
		totalMetricsReceivedDesc:          totalMetricsReceivedDesc,
		totalContainerMetricsReceivedDesc: totalContainerMetricsReceivedDesc,
		totalCounterEventsReceivedDesc:    totalCounterEventsReceivedDesc,
		totalValueMetricsReceivedDesc:     totalValueMetricsReceivedDesc,
		slowConsumerAlertDesc:             slowConsumerAlertDesc,
		lastReceivedMetricTimestampDesc:   lastReceivedMetricTimestampDesc,
	}
	return collector
}

func (c internalMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(
		c.totalEnvelopesReceivedDesc,
		prometheus.CounterValue,
		c.metricsStore.GetInternalMetrics().TotalEnvelopesReceived,
	)
	ch <- prometheus.MustNewConstMetric(
		c.totalMetricsReceivedDesc,
		prometheus.CounterValue,
		c.metricsStore.GetInternalMetrics().TotalMetricsReceived,
	)
	ch <- prometheus.MustNewConstMetric(
		c.totalContainerMetricsReceivedDesc,
		prometheus.CounterValue,
		c.metricsStore.GetInternalMetrics().TotalContainerMetricsReceived,
	)
	ch <- prometheus.MustNewConstMetric(
		c.totalCounterEventsReceivedDesc,
		prometheus.CounterValue,
		c.metricsStore.GetInternalMetrics().TotalCounterEventsReceived,
	)
	ch <- prometheus.MustNewConstMetric(
		c.totalValueMetricsReceivedDesc,
		prometheus.CounterValue,
		c.metricsStore.GetInternalMetrics().TotalValueMetricsReceived,
	)

	if c.metricsStore.GetInternalMetrics().SlowConsumerAlert {
		ch <- prometheus.MustNewConstMetric(
			c.slowConsumerAlertDesc,
			prometheus.UntypedValue,
			1,
		)
	}
}

func (c internalMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.totalEnvelopesReceivedDesc
	ch <- c.totalMetricsReceivedDesc
	ch <- c.totalContainerMetricsReceivedDesc
	ch <- c.totalCounterEventsReceivedDesc
	ch <- c.totalValueMetricsReceivedDesc
	ch <- c.slowConsumerAlertDesc
	ch <- c.lastReceivedMetricTimestampDesc
}
