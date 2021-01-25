package metricmaker

import (
	"strings"

	"github.com/bosh-prometheus/firehose_exporter/metrics"
	"github.com/bosh-prometheus/firehose_exporter/utils"
	dto "github.com/prometheus/client_model/go"
)

func RetroCompatMetricNames(metric *metrics.RawMetric) {
	isContainerMetric := utils.MetricIsContainerMetric(metric)

	switch metric.MetricName() {
	case "cpu":
		metric.SetHelp("Cloud Foundry Firehose container metric: CPU used, on a scale of 0 to 100.")
	case "memory":
		metric.SetHelp("Cloud Foundry Firehose container metric: bytes of memory used.")
	case "disk":
		metric.SetHelp("Cloud Foundry Firehose container metric: bytes of disk used.")
	case "memory_quota":
		metric.SetHelp("Cloud Foundry Firehose container metric: maximum bytes of memory allocated to container.")
	case "disk_quota":
		metric.SetHelp("Cloud Foundry Firehose container metric: maximum bytes of disk allocated to container.")
	}
	FindAndReplaceByName("cpu", "container_metric_cpu_percentage")(metric)
	FindAndReplaceByName("memory", "container_metric_memory_bytes")(metric)
	FindAndReplaceByName("disk", "container_metric_disk_bytes")(metric)
	FindAndReplaceByName("memory_quota", "container_metric_memory_bytes_quota")(metric)
	FindAndReplaceByName("disk_quota", "container_metric_disk_bytes_quota")(metric)

	if isContainerMetric {
		return
	}

	if *metric.MetricType() == dto.MetricType_COUNTER &&
		!strings.HasSuffix(metric.MetricName(), "_delta") &&
		metric.MetricName() != "http_total" {
		metric.SetMetricName("counter_event_" + metric.Origin() + "_" + metric.MetricName() + "_total")
		metric.SetHelp("Cloud Foundry Firehose counter metrics.")
	}
	if *metric.MetricType() == dto.MetricType_GAUGE {
		metric.SetMetricName("value_metric_" + metric.Origin() + "_" + metric.MetricName())
		metric.SetHelp("Cloud Foundry Firehose value metrics.")
	}
}
