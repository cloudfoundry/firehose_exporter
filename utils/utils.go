package utils

import (
	"strings"

	"github.com/cloudfoundry/firehose_exporter/metrics"
)

func MetricIsContainerMetric(metric *metrics.RawMetric) bool {
	return MetricNameIsContainerMetric(metric.MetricName())
}

func MetricNameIsContainerMetric(metricName string) bool {
	return metricName == "cpu" || metricName == "memory" || metricName == "disk" ||
		metricName == "memory_quota" || metricName == "disk_quota" ||
		strings.HasSuffix(metricName, "_cpu") || strings.HasSuffix(metricName, "_memory") ||
		strings.HasSuffix(metricName, "_disk") || strings.HasSuffix(metricName, "_memory_quota") ||
		strings.HasSuffix(metricName, "_disk_quota")
}

func MetricIsHTTPMetric(metric *metrics.RawMetric) bool {
	return strings.Contains(metric.MetricName(), metrics.GorouterHTTPSummaryMetricName) ||
		strings.Contains(metric.MetricName(), metrics.GorouterHTTPHistogramMetricName) ||
		strings.Contains(metric.MetricName(), metrics.GorouterHTTPCounterMetricName)
}
