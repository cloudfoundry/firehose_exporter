package collectors

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/cloudfoundry-community/firehose_exporter/metrics"
	"github.com/cloudfoundry-community/firehose_exporter/utils"
)

type ValueMetricsCollector struct {
	namespace                 string
	metricsStore              *metrics.Store
	deploymentsFilter         map[string]struct{}
	valueMetricsCollectorDesc *prometheus.Desc
}

func NewValueMetricsCollector(
	namespace string,
	metricsStore *metrics.Store,
	dopplerDeployments []string,
) *ValueMetricsCollector {
	valueMetricsCollectorDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, value_metrics_subsystem, "collector"),
		"Cloud Foundry Firehose value metrics collector.",
		nil,
		nil,
	)

	deploymentsFilter := map[string]struct{}{}
	for _, deployment := range dopplerDeployments {
		deploymentsFilter[deployment] = struct{}{}
	}

	collector := &ValueMetricsCollector{
		namespace:                 namespace,
		metricsStore:              metricsStore,
		deploymentsFilter:         deploymentsFilter,
		valueMetricsCollectorDesc: valueMetricsCollectorDesc,
	}
	return collector
}

func (c ValueMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	for _, valueMetric := range c.metricsStore.GetValueMetrics() {
		_, ok := c.deploymentsFilter[valueMetric.Deployment]
		if len(c.deploymentsFilter) == 0 || ok {
			metricName := utils.NormalizeName(valueMetric.Origin) + "_" + utils.NormalizeName(valueMetric.Name)
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(c.namespace, value_metrics_subsystem, metricName),
					fmt.Sprintf("Cloud Foundry Firehose '%s' value metric.", valueMetric.Name),
					[]string{"origin", "bosh_deployment", "bosh_job", "bosh_index", "bosh_ip", "unit"},
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
}

func (c ValueMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.valueMetricsCollectorDesc
}
