package collectors

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/frodenas/firehose_exporter/metrics"
	"github.com/frodenas/firehose_exporter/utils"
)

var (
	counterMetricsCollectorDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, counter_events_subsystem, "collector"),
		"Cloud Foundry firehose counter metrics collector.",
		nil,
		nil,
	)
)

type counterMetricsCollector struct {
	metricsStore *metrics.Store
}

func NewCounterMetricsCollector(metricsStore *metrics.Store) *counterMetricsCollector {
	collector := &counterMetricsCollector{
		metricsStore: metricsStore,
	}
	return collector
}

func (c counterMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	for _, counterMetric := range c.metricsStore.GetCounterMetrics() {
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, counter_events_subsystem, utils.NormalizeName(counterMetric.Name)),
				fmt.Sprintf("Cloud Foundry firehose '%s' counter event.", counterMetric.Name),
				[]string{"origin", "deployment", "job", "index", "ip"},
				nil,
			),
			prometheus.CounterValue,
			float64(counterMetric.Total),
			counterMetric.Origin,
			counterMetric.Deployment,
			counterMetric.Job,
			counterMetric.Index,
			counterMetric.IP,
		)
	}
}

func (c counterMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- counterMetricsCollectorDesc
}
