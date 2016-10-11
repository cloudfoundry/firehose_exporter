package collectors

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/cloudfoundry-community/firehose_exporter/metrics"
	"github.com/cloudfoundry-community/firehose_exporter/utils"
)

type counterMetricsCollector struct {
	namespace                   string
	metricsStore                *metrics.Store
	deploymentsFilter           map[string]struct{}
	counterMetricsCollectorDesc *prometheus.Desc
}

func NewCounterMetricsCollector(
	namespace string,
	metricsStore *metrics.Store,
	dopplerDeployments []string,
) *counterMetricsCollector {
	counterMetricsCollectorDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, counter_events_subsystem, "collector"),
		"Cloud Foundry Firehose counter metrics collector.",
		nil,
		nil,
	)

	deploymentsFilter := map[string]struct{}{}
	for _, deployment := range dopplerDeployments {
		deploymentsFilter[deployment] = struct{}{}
	}

	collector := &counterMetricsCollector{
		namespace:                   namespace,
		metricsStore:                metricsStore,
		deploymentsFilter:           deploymentsFilter,
		counterMetricsCollectorDesc: counterMetricsCollectorDesc,
	}
	return collector
}

func (c counterMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	for _, counterMetric := range c.metricsStore.GetCounterMetrics() {
		_, ok := c.deploymentsFilter[counterMetric.Deployment]
		if len(c.deploymentsFilter) == 0 || ok {
			metricName := utils.NormalizeName(counterMetric.Origin) + "_total_" + utils.NormalizeName(counterMetric.Name)
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(c.namespace, counter_events_subsystem, metricName),
					fmt.Sprintf("Cloud Foundry Firehose '%s' total counter event.", counterMetric.Name),
					[]string{"origin", "bosh_deployment", "bosh_job", "bosh_index", "bosh_ip"},
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

			metricName = utils.NormalizeName(counterMetric.Origin) + "_delta_" + utils.NormalizeName(counterMetric.Name)
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(c.namespace, counter_events_subsystem, metricName),
					fmt.Sprintf("Cloud Foundry Firehose '%s' delta counter event.", counterMetric.Name),
					[]string{"origin", "bosh_deployment", "bosh_job", "bosh_index", "bosh_ip"},
					nil,
				),
				prometheus.GaugeValue,
				float64(counterMetric.Delta),
				counterMetric.Origin,
				counterMetric.Deployment,
				counterMetric.Job,
				counterMetric.Index,
				counterMetric.IP,
			)
		}
	}
}

func (c counterMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.counterMetricsCollectorDesc
}
