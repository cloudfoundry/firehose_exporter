package collectors

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/bosh-prometheus/firehose_exporter/metrics"
)

type ContainerMetricsCollector struct {
	namespace              string
	environment            string
	metricsStore           *metrics.Store
	cpuPercentageMetric    *prometheus.GaugeVec
	memoryBytesMetric      *prometheus.GaugeVec
	diskBytesMetric        *prometheus.GaugeVec
	memoryBytesQuotaMetric *prometheus.GaugeVec
	diskBytesQuotaMetric   *prometheus.GaugeVec
}

func NewContainerMetricsCollector(
	namespace string,
	environment string,
	metricsStore *metrics.Store,
) *ContainerMetricsCollector {
	cpuPercentageMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   container_metrics_subsystem,
			Name:        "cpu_percentage",
			Help:        "Cloud Foundry Firehose container metric: CPU used, on a scale of 0 to 100.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
		[]string{"origin", "bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_ip", "application_id", "instance_index"},
	)

	memoryBytesMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   container_metrics_subsystem,
			Name:        "memory_bytes",
			Help:        "Cloud Foundry Firehose container metric: bytes of memory used.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
		[]string{"origin", "bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_ip", "application_id", "instance_index"},
	)

	diskBytesMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   container_metrics_subsystem,
			Name:        "disk_bytes",
			Help:        "Cloud Foundry Firehose container metric: bytes of disk used.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
		[]string{"origin", "bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_ip", "application_id", "instance_index"},
	)

	memoryBytesQuotaMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   container_metrics_subsystem,
			Name:        "memory_bytes_quota",
			Help:        "Cloud Foundry Firehose container metric: maximum bytes of memory allocated to container.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
		[]string{"origin", "bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_ip", "application_id", "instance_index"},
	)

	diskBytesQuotaMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   container_metrics_subsystem,
			Name:        "disk_bytes_quota",
			Help:        "Cloud Foundry Firehose container metric: maximum bytes of disk allocated to container.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
		[]string{"origin", "bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_ip", "application_id", "instance_index"},
	)

	return &ContainerMetricsCollector{
		namespace:              namespace,
		environment:            environment,
		metricsStore:           metricsStore,
		cpuPercentageMetric:    cpuPercentageMetric,
		memoryBytesMetric:      memoryBytesMetric,
		diskBytesMetric:        diskBytesMetric,
		memoryBytesQuotaMetric: memoryBytesQuotaMetric,
		diskBytesQuotaMetric:   diskBytesQuotaMetric,
	}
}

func (c ContainerMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	c.cpuPercentageMetric.Reset()
	c.memoryBytesMetric.Reset()
	c.diskBytesMetric.Reset()
	c.memoryBytesQuotaMetric.Reset()
	c.diskBytesQuotaMetric.Reset()

	for _, containerMetric := range c.metricsStore.GetContainerMetrics() {
		c.cpuPercentageMetric.WithLabelValues(
			containerMetric.Origin,
			containerMetric.Deployment,
			containerMetric.Job,
			containerMetric.Index,
			containerMetric.IP,
			containerMetric.ApplicationId,
			strconv.Itoa(int(containerMetric.InstanceIndex)),
		).Set(containerMetric.CpuPercentage)

		c.memoryBytesMetric.WithLabelValues(
			containerMetric.Origin,
			containerMetric.Deployment,
			containerMetric.Job,
			containerMetric.Index,
			containerMetric.IP,
			containerMetric.ApplicationId,
			strconv.Itoa(int(containerMetric.InstanceIndex)),
		).Set(float64(containerMetric.MemoryBytes))

		c.diskBytesMetric.WithLabelValues(
			containerMetric.Origin,
			containerMetric.Deployment,
			containerMetric.Job,
			containerMetric.Index,
			containerMetric.IP,
			containerMetric.ApplicationId,
			strconv.Itoa(int(containerMetric.InstanceIndex)),
		).Set(float64(containerMetric.DiskBytes))

		c.memoryBytesQuotaMetric.WithLabelValues(
			containerMetric.Origin,
			containerMetric.Deployment,
			containerMetric.Job,
			containerMetric.Index,
			containerMetric.IP,
			containerMetric.ApplicationId,
			strconv.Itoa(int(containerMetric.InstanceIndex)),
		).Set(float64(containerMetric.MemoryBytesQuota))

		c.diskBytesQuotaMetric.WithLabelValues(
			containerMetric.Origin,
			containerMetric.Deployment,
			containerMetric.Job,
			containerMetric.Index,
			containerMetric.IP,
			containerMetric.ApplicationId,
			strconv.Itoa(int(containerMetric.InstanceIndex)),
		).Set(float64(containerMetric.DiskBytesQuota))
	}

	c.cpuPercentageMetric.Collect(ch)
	c.memoryBytesMetric.Collect(ch)
	c.diskBytesMetric.Collect(ch)
	c.memoryBytesQuotaMetric.Collect(ch)
	c.diskBytesQuotaMetric.Collect(ch)
}

func (c ContainerMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	c.cpuPercentageMetric.Describe(ch)
	c.memoryBytesMetric.Describe(ch)
	c.diskBytesMetric.Describe(ch)
	c.memoryBytesQuotaMetric.Describe(ch)
	c.diskBytesQuotaMetric.Describe(ch)
}
