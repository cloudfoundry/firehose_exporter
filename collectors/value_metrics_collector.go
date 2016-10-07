package collectors

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/cloudfoundry-community/firehose_exporter/metrics"
	"github.com/cloudfoundry-community/firehose_exporter/utils"
)

type valueMetricsCollector struct {
	namespace                 string
	metricsStore              *metrics.Store
	valueMetricsCollectorDesc *prometheus.Desc
}

func NewValueMetricsCollector(
	namespace string,
	metricsStore *metrics.Store,
) *valueMetricsCollector {
	valueMetricsCollectorDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, value_metrics_subsystem, "collector"),
		"Cloud Foundry Firehose value metrics collector.",
		nil,
		nil,
	)

	collector := &valueMetricsCollector{
		namespace:                 namespace,
		metricsStore:              metricsStore,
		valueMetricsCollectorDesc: valueMetricsCollectorDesc,
	}
	return collector
}

func (c valueMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	for _, valueMetric := range c.metricsStore.GetValueMetrics() {
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(c.namespace, value_metrics_subsystem, utils.NormalizeName(valueMetric.Name)),
				fmt.Sprintf("Cloud Foundry firehose '%s' value metric.", valueMetric.Name),
				[]string{"origin", "deployment", "job", "index", "ip", "unit"},
				nil,
			),
			prometheus.GaugeValue,
			float64(valueMetric.Value),
			valueMetric.Origin,
			valueMetric.Deployment,
			valueMetric.Job,
			valueMetric.Index,
			valueMetric.IP,
			valueMetric.Unit,
		)
	}
}

func (c valueMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.valueMetricsCollectorDesc
}
