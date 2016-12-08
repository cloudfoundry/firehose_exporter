package collectors

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/mjseid/firehose_exporter/cfinstanceinfoapi"
	"github.com/mjseid/firehose_exporter/metrics"
)

type ContainerMetricsCollector struct {
	namespace                  string
	metricsStore               *metrics.Store
	cpuPercentageMetricDesc    *prometheus.Desc
	memoryBytesMetricDesc      *prometheus.Desc
	diskBytesMetricDesc        *prometheus.Desc
	memoryBytesQuotaMetricDesc *prometheus.Desc
	diskBytesQuotaMetricDesc   *prometheus.Desc
	appinfo                    map[string]cfinstanceinfoapi.AppInfo
}

func NewContainerMetricsCollector(
	namespace string,
	metricsStore *metrics.Store,
	appinfo map[string]cfinstanceinfoapi.AppInfo,
) *ContainerMetricsCollector {
	cpuPercentageMetricDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, container_metrics_subsystem, "cpu_percentage"),
		"Cloud Foundry Firehose container metric: CPU used, on a scale of 0 to 100.",
		[]string{"bosh_job_ip", "application_id", "instance_id", "app_name", "space", "org"},
		nil,
	)

	memoryBytesMetricDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, container_metrics_subsystem, "memory_bytes"),
		"Cloud Foundry Firehose container metric: bytes of memory used.",
		[]string{"bosh_job_ip", "application_id", "instance_id", "app_name", "space", "org"},
		nil,
	)

	diskBytesMetricDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, container_metrics_subsystem, "disk_bytes"),
		"Cloud Foundry Firehose container metric: bytes of disk used.",
		[]string{"bosh_job_ip", "application_id", "instance_id", "app_name", "space", "org"},
		nil,
	)

	memoryBytesQuotaMetricDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, container_metrics_subsystem, "memory_bytes_quota"),
		"Cloud Foundry Firehose container metric: maximum bytes of memory allocated to container.",
		[]string{"bosh_job_ip", "application_id", "instance_id", "app_name", "space", "org"},
		nil,
	)

	diskBytesQuotaMetricDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, container_metrics_subsystem, "disk_bytes_quota"),
		"Cloud Foundry Firehose container metric: maximum bytes of disk allocated to container.",
		[]string{"bosh_job_ip", "application_id", "instance_id", "app_name", "space", "org"},
		nil,
	)

	return &ContainerMetricsCollector{
		namespace:                  namespace,
		metricsStore:               metricsStore,
		cpuPercentageMetricDesc:    cpuPercentageMetricDesc,
		memoryBytesMetricDesc:      memoryBytesMetricDesc,
		diskBytesMetricDesc:        diskBytesMetricDesc,
		memoryBytesQuotaMetricDesc: memoryBytesQuotaMetricDesc,
		diskBytesQuotaMetricDesc:   diskBytesQuotaMetricDesc,
		appinfo:                    appinfo,
	}
}

func (c ContainerMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	for _, containerMetric := range c.metricsStore.GetContainerMetrics() {
		ch <- prometheus.MustNewConstMetric(
			c.cpuPercentageMetricDesc,
			prometheus.GaugeValue,
			containerMetric.CpuPercentage,
			containerMetric.IP,
			containerMetric.ApplicationId,
			strconv.Itoa(int(containerMetric.InstanceIndex)),
			c.appinfo[containerMetric.ApplicationId].Name,
			c.appinfo[containerMetric.ApplicationId].Space,
			c.appinfo[containerMetric.ApplicationId].Org,
		)
		ch <- prometheus.MustNewConstMetric(
			c.memoryBytesMetricDesc,
			prometheus.GaugeValue,
			float64(containerMetric.MemoryBytes),
			containerMetric.IP,
			containerMetric.ApplicationId,
			strconv.Itoa(int(containerMetric.InstanceIndex)),
			c.appinfo[containerMetric.ApplicationId].Name,
			c.appinfo[containerMetric.ApplicationId].Space,
			c.appinfo[containerMetric.ApplicationId].Org,
		)
		ch <- prometheus.MustNewConstMetric(
			c.diskBytesMetricDesc,
			prometheus.GaugeValue,
			float64(containerMetric.DiskBytes),
			containerMetric.IP,
			containerMetric.ApplicationId,
			strconv.Itoa(int(containerMetric.InstanceIndex)),
			c.appinfo[containerMetric.ApplicationId].Name,
			c.appinfo[containerMetric.ApplicationId].Space,
			c.appinfo[containerMetric.ApplicationId].Org,
		)
		ch <- prometheus.MustNewConstMetric(
			c.memoryBytesQuotaMetricDesc,
			prometheus.GaugeValue,
			float64(containerMetric.MemoryBytesQuota),
			containerMetric.IP,
			containerMetric.ApplicationId,
			strconv.Itoa(int(containerMetric.InstanceIndex)),
			c.appinfo[containerMetric.ApplicationId].Name,
			c.appinfo[containerMetric.ApplicationId].Space,
			c.appinfo[containerMetric.ApplicationId].Org,
		)
		ch <- prometheus.MustNewConstMetric(
			c.diskBytesQuotaMetricDesc,
			prometheus.GaugeValue,
			float64(containerMetric.DiskBytesQuota),
			containerMetric.IP,
			containerMetric.ApplicationId,
			strconv.Itoa(int(containerMetric.InstanceIndex)),
			c.appinfo[containerMetric.ApplicationId].Name,
			c.appinfo[containerMetric.ApplicationId].Space,
			c.appinfo[containerMetric.ApplicationId].Org,
		)
	}
}

func (c ContainerMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.cpuPercentageMetricDesc
	ch <- c.memoryBytesMetricDesc
	ch <- c.diskBytesMetricDesc
	ch <- c.memoryBytesQuotaMetricDesc
	ch <- c.diskBytesQuotaMetricDesc
}
