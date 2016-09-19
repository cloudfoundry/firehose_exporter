package collectors

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/frodenas/firehose_exporter/metrics"
)

var (
	containerMetricsCollectorDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, container_metrics_subsystem, "collector"),
		"Cloud Foundry firehose container metrics collector.",
		nil,
		nil,
	)

	cpuPercentageMetricDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, container_metrics_subsystem, "cpu_percentage"),
		"Cloud Foundry firehose container metric: CPU used, on a scale of 0 to 100.",
		[]string{"origin", "deployment", "job", "index", "ip", "application_id", "instance_id"},
		nil,
	)

	memoryBytesMetricDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, container_metrics_subsystem, "memory_bytes"),
		"Cloud Foundry firehose container metric: bytes of memory used.",
		[]string{"origin", "deployment", "job", "index", "ip", "application_id", "instance_id"},
		nil,
	)

	diskBytesMetricDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, container_metrics_subsystem, "disk_bytes"),
		"Cloud Foundry firehose container metric: bytes of disk used.",
		[]string{"origin", "deployment", "job", "index", "ip", "application_id", "instance_id"},
		nil,
	)

	memoryBytesQuotaMetricDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, container_metrics_subsystem, "memory_bytes_quota"),
		"Cloud Foundry firehose container metric: maximum bytes of memory allocated to container.",
		[]string{"origin", "deployment", "job", "index", "ip", "application_id", "instance_id"},
		nil,
	)

	diskBytesQuotaMetricDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, container_metrics_subsystem, "disk_bytes_quota"),
		"Cloud Foundry firehose container metric: maximum bytes of disk allocated to container.",
		[]string{"origin", "deployment", "job", "index", "ip", "application_id", "instance_id"},
		nil,
	)
)

type containerMetricsCollector struct {
	metricsStore *metrics.Store
}

func NewContainerMetricsCollector(metricsStore *metrics.Store) *containerMetricsCollector {
	collector := &containerMetricsCollector{
		metricsStore: metricsStore,
	}
	return collector
}

func (c containerMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	for _, containerMetric := range c.metricsStore.GetContainerMetrics() {
		ch <- prometheus.MustNewConstMetric(
			cpuPercentageMetricDesc,
			prometheus.GaugeValue,
			containerMetric.CpuPercentage,
			containerMetric.Origin,
			containerMetric.Deployment,
			containerMetric.Job,
			containerMetric.Index,
			containerMetric.IP,
			containerMetric.ApplicationId,
			strconv.Itoa(int(containerMetric.InstanceIndex)),
		)
		ch <- prometheus.MustNewConstMetric(
			memoryBytesMetricDesc,
			prometheus.GaugeValue,
			float64(containerMetric.MemoryBytes),
			containerMetric.Origin,
			containerMetric.Deployment,
			containerMetric.Job,
			containerMetric.Index,
			containerMetric.IP,
			containerMetric.ApplicationId,
			strconv.Itoa(int(containerMetric.InstanceIndex)),
		)
		ch <- prometheus.MustNewConstMetric(
			diskBytesMetricDesc,
			prometheus.GaugeValue,
			float64(containerMetric.DiskBytes),
			containerMetric.Origin,
			containerMetric.Deployment,
			containerMetric.Job,
			containerMetric.Index,
			containerMetric.IP,
			containerMetric.ApplicationId,
			strconv.Itoa(int(containerMetric.InstanceIndex)),
		)
		ch <- prometheus.MustNewConstMetric(
			memoryBytesQuotaMetricDesc,
			prometheus.GaugeValue,
			float64(containerMetric.MemoryBytesQuota),
			containerMetric.Origin,
			containerMetric.Deployment,
			containerMetric.Job,
			containerMetric.Index,
			containerMetric.IP,
			containerMetric.ApplicationId,
			strconv.Itoa(int(containerMetric.InstanceIndex)),
		)
		ch <- prometheus.MustNewConstMetric(
			diskBytesQuotaMetricDesc,
			prometheus.GaugeValue,
			float64(containerMetric.DiskBytesQuota),
			containerMetric.Origin,
			containerMetric.Deployment,
			containerMetric.Job,
			containerMetric.Index,
			containerMetric.IP,
			containerMetric.ApplicationId,
			strconv.Itoa(int(containerMetric.InstanceIndex)),
		)
	}
}

func (c containerMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- containerMetricsCollectorDesc
	ch <- cpuPercentageMetricDesc
	ch <- memoryBytesMetricDesc
	ch <- diskBytesMetricDesc
	ch <- memoryBytesQuotaMetricDesc
	ch <- diskBytesQuotaMetricDesc
}
