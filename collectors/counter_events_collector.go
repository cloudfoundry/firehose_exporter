package collectors

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/cloudfoundry-community/firehose_exporter/metrics"
	"github.com/cloudfoundry-community/firehose_exporter/utils"
)

type CounterEventsCollector struct {
	namespace                  string
	metricsStore               *metrics.Store
	counterEventsCollectorDesc *prometheus.Desc
}

func NewCounterEventsCollector(
	namespace string,
	metricsStore *metrics.Store,
) *CounterEventsCollector {
	counterEventsCollectorDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, counter_events_subsystem, "collector"),
		"Cloud Foundry Firehose counter metrics collector.",
		nil,
		nil,
	)

	return &CounterEventsCollector{
		namespace:                  namespace,
		metricsStore:               metricsStore,
		counterEventsCollectorDesc: counterEventsCollectorDesc,
	}
}

func (c CounterEventsCollector) Collect(ch chan<- prometheus.Metric) {
	for _, counterEvent := range c.metricsStore.GetCounterEvents() {
		metricName := utils.NormalizeName(counterEvent.Origin) + utils.NormalizeName(counterEvent.Name) + "_total"
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(c.namespace, counter_events_subsystem, metricName),
				fmt.Sprintf("Cloud Foundry Firehose '%s' total counter event.", counterEvent.Name),
				[]string{"origin", "bosh_deployment", "bosh_job", "bosh_index", "bosh_ip"},
				nil,
			),
			prometheus.CounterValue,
			float64(counterEvent.Total),
			counterEvent.Origin,
			counterEvent.Deployment,
			counterEvent.Job,
			counterEvent.Index,
			counterEvent.IP,
		)

		metricName = utils.NormalizeName(counterEvent.Origin) + utils.NormalizeName(counterEvent.Name) + "_delta"
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(c.namespace, counter_events_subsystem, metricName),
				fmt.Sprintf("Cloud Foundry Firehose '%s' delta counter event.", counterEvent.Name),
				[]string{"origin", "bosh_deployment", "bosh_job", "bosh_index", "bosh_ip"},
				nil,
			),
			prometheus.GaugeValue,
			float64(counterEvent.Delta),
			counterEvent.Origin,
			counterEvent.Deployment,
			counterEvent.Job,
			counterEvent.Index,
			counterEvent.IP,
		)
	}
}

func (c CounterEventsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.counterEventsCollectorDesc
}
