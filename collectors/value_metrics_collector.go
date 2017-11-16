package collectors

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/mjseid/firehose_exporter/metrics"
	"github.com/mjseid/firehose_exporter/utils"
)

type ValueMetricsCollector struct {
	namespace                 string
	environment               string
	metricsStore              *metrics.Store
	valueMetricsCollectorDesc *prometheus.Desc
}

func NewValueMetricsCollector(
	namespace string,
	environment string,
	metricsStore *metrics.Store,
) *ValueMetricsCollector {
	valueMetricsCollectorDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, value_metrics_subsystem, "collector"),
		"Cloud Foundry Firehose value metrics collector.",
		nil,
		prometheus.Labels{"environment": environment},
	)

	return &ValueMetricsCollector{
		namespace:                 namespace,
		environment:               environment,
		metricsStore:              metricsStore,
		valueMetricsCollectorDesc: valueMetricsCollectorDesc,
	}
}

func (c ValueMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	for _, valueMetric := range c.metricsStore.GetValueMetrics() {
		metricName := utils.NormalizeName(valueMetric.Origin) + "_" + utils.NormalizeName(valueMetric.Name)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(c.namespace, value_metrics_subsystem, metricName),
				fmt.Sprintf("Cloud Foundry Firehose '%s' value metric from '%s'.", valueMetric.Name, utils.NormalizeDesc(valueMetric.Origin)),
				[]string{"origin", "bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_ip", "unit"},
				prometheus.Labels{"environment": c.environment},
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

func (c ValueMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.valueMetricsCollectorDesc
}
