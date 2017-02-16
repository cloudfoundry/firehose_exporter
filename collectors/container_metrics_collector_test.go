package collectors_test

import (
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-community/firehose_exporter/filters"
	"github.com/cloudfoundry-community/firehose_exporter/metrics"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"

	. "github.com/cloudfoundry-community/firehose_exporter/collectors"
	. "github.com/cloudfoundry-community/firehose_exporter/utils/test_matchers"
)

var _ = Describe("ContainerMetricsCollector", func() {
	var (
		namespace                 string
		environment               string
		metricsStore              *metrics.Store
		metricsExpiration         time.Duration
		metricsCleanupInterval    time.Duration
		deploymentFilter          *filters.DeploymentFilter
		eventFilter               *filters.EventFilter
		containerMetricsCollector *ContainerMetricsCollector

		cpuPercentageMetric    *prometheus.GaugeVec
		memoryBytesMetric      *prometheus.GaugeVec
		diskBytesMetric        *prometheus.GaugeVec
		memoryBytesQuotaMetric *prometheus.GaugeVec
		diskBytesQuotaMetric   *prometheus.GaugeVec

		origin         = "fake-origin"
		boshDeployment = "fake-deployment-name"
		boshJob        = "fake-job-name"
		boshIndex      = "0"
		boshIP         = "1.2.3.4"

		containerMetric1ApplicationId    = "FakeApplicationId1"
		containerMetric1InstanceIndex    = int32(1)
		containerMetric1CpuPercentage    = float64(0.5)
		containerMetric1MemoryBytes      = uint64(1000)
		containerMetric1DiskBytes        = uint64(1500)
		containerMetric1MemoryBytesQuota = uint64(2000)
		containerMetric1DiskBytesQuota   = uint64(3000)

		containerMetric2ApplicationId    = "FakeApplicationId2"
		containerMetric2InstanceIndex    = int32(2)
		containerMetric2CpuPercentage    = float64(1.5)
		containerMetric2MemoryBytes      = uint64(2000)
		containerMetric2DiskBytes        = uint64(2500)
		containerMetric2MemoryBytesQuota = uint64(4000)
		containerMetric2DiskBytesQuota   = uint64(5000)
	)

	BeforeEach(func() {
		namespace = "test_exporter"
		environment = "test_environment"
		deploymentFilter = filters.NewDeploymentFilter([]string{})
		eventFilter, _ = filters.NewEventFilter([]string{})
		metricsStore = metrics.NewStore(metricsExpiration, metricsCleanupInterval, deploymentFilter, eventFilter)

		cpuPercentageMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "container_metric",
				Name:        "cpu_percentage",
				Help:        "Cloud Foundry Firehose container metric: CPU used, on a scale of 0 to 100.",
				ConstLabels: prometheus.Labels{"environment": environment},
			},
			[]string{"origin", "bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_ip", "application_id", "instance_index"},
		)

		cpuPercentageMetric.WithLabelValues(
			origin,
			boshDeployment,
			boshJob,
			boshIndex,
			boshIP,
			containerMetric1ApplicationId,
			strconv.Itoa(int(containerMetric1InstanceIndex)),
		).Set(containerMetric1CpuPercentage)

		cpuPercentageMetric.WithLabelValues(
			origin,
			boshDeployment,
			boshJob,
			boshIndex,
			boshIP,
			containerMetric2ApplicationId,
			strconv.Itoa(int(containerMetric2InstanceIndex)),
		).Set(containerMetric2CpuPercentage)

		memoryBytesMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "container_metric",
				Name:        "memory_bytes",
				Help:        "Cloud Foundry Firehose container metric: bytes of memory used.",
				ConstLabels: prometheus.Labels{"environment": environment},
			},
			[]string{"origin", "bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_ip", "application_id", "instance_index"},
		)

		memoryBytesMetric.WithLabelValues(
			origin,
			boshDeployment,
			boshJob,
			boshIndex,
			boshIP,
			containerMetric1ApplicationId,
			strconv.Itoa(int(containerMetric1InstanceIndex)),
		).Set(float64(containerMetric1MemoryBytes))

		memoryBytesMetric.WithLabelValues(
			origin,
			boshDeployment,
			boshJob,
			boshIndex,
			boshIP,
			containerMetric2ApplicationId,
			strconv.Itoa(int(containerMetric2InstanceIndex)),
		).Set(float64(containerMetric2MemoryBytes))

		diskBytesMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "container_metric",
				Name:        "disk_bytes",
				Help:        "Cloud Foundry Firehose container metric: bytes of disk used.",
				ConstLabels: prometheus.Labels{"environment": environment},
			},
			[]string{"origin", "bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_ip", "application_id", "instance_index"},
		)

		diskBytesMetric.WithLabelValues(
			origin,
			boshDeployment,
			boshJob,
			boshIndex,
			boshIP,
			containerMetric1ApplicationId,
			strconv.Itoa(int(containerMetric1InstanceIndex)),
		).Set(float64(containerMetric1DiskBytes))

		diskBytesMetric.WithLabelValues(
			origin,
			boshDeployment,
			boshJob,
			boshIndex,
			boshIP,
			containerMetric2ApplicationId,
			strconv.Itoa(int(containerMetric2InstanceIndex)),
		).Set(float64(containerMetric2DiskBytes))

		memoryBytesQuotaMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "container_metric",
				Name:        "memory_bytes_quota",
				Help:        "Cloud Foundry Firehose container metric: maximum bytes of memory allocated to container.",
				ConstLabels: prometheus.Labels{"environment": environment},
			},
			[]string{"origin", "bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_ip", "application_id", "instance_index"},
		)

		memoryBytesQuotaMetric.WithLabelValues(
			origin,
			boshDeployment,
			boshJob,
			boshIndex,
			boshIP,
			containerMetric1ApplicationId,
			strconv.Itoa(int(containerMetric1InstanceIndex)),
		).Set(float64(containerMetric1MemoryBytesQuota))

		memoryBytesQuotaMetric.WithLabelValues(
			origin,
			boshDeployment,
			boshJob,
			boshIndex,
			boshIP,
			containerMetric2ApplicationId,
			strconv.Itoa(int(containerMetric2InstanceIndex)),
		).Set(float64(containerMetric2MemoryBytesQuota))

		diskBytesQuotaMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "container_metric",
				Name:        "disk_bytes_quota",
				Help:        "Cloud Foundry Firehose container metric: maximum bytes of disk allocated to container.",
				ConstLabels: prometheus.Labels{"environment": environment},
			},
			[]string{"origin", "bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_ip", "application_id", "instance_index"},
		)

		diskBytesQuotaMetric.WithLabelValues(
			origin,
			boshDeployment,
			boshJob,
			boshIndex,
			boshIP,
			containerMetric1ApplicationId,
			strconv.Itoa(int(containerMetric1InstanceIndex)),
		).Set(float64(containerMetric1DiskBytesQuota))

		diskBytesQuotaMetric.WithLabelValues(
			origin,
			boshDeployment,
			boshJob,
			boshIndex,
			boshIP,
			containerMetric2ApplicationId,
			strconv.Itoa(int(containerMetric2InstanceIndex)),
		).Set(float64(containerMetric2DiskBytesQuota))
	})

	JustBeforeEach(func() {
		containerMetricsCollector = NewContainerMetricsCollector(namespace, environment, metricsStore)
	})

	Describe("Describe", func() {
		var (
			descriptions chan *prometheus.Desc
		)

		BeforeEach(func() {
			descriptions = make(chan *prometheus.Desc)
		})

		JustBeforeEach(func() {
			go containerMetricsCollector.Describe(descriptions)
		})

		It("returns a container_metric_cpu_percentage metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(cpuPercentageMetric.WithLabelValues(
				origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				containerMetric1ApplicationId,
				strconv.Itoa(int(containerMetric1InstanceIndex)),
			).Desc())))
		})

		It("returns a container_metric_memory_bytes metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(memoryBytesMetric.WithLabelValues(
				origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				containerMetric1ApplicationId,
				strconv.Itoa(int(containerMetric1InstanceIndex)),
			).Desc())))
		})

		It("returns a container_metric_disk_bytes metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(diskBytesMetric.WithLabelValues(
				origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				containerMetric1ApplicationId,
				strconv.Itoa(int(containerMetric1InstanceIndex)),
			).Desc())))
		})

		It("returns a container_metric_memory_bytes_quota metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(memoryBytesQuotaMetric.WithLabelValues(
				origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				containerMetric1ApplicationId,
				strconv.Itoa(int(containerMetric1InstanceIndex)),
			).Desc())))
		})

		It("returns a container_metric_disk_bytes_quota metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(diskBytesQuotaMetric.WithLabelValues(
				origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				containerMetric1ApplicationId,
				strconv.Itoa(int(containerMetric1InstanceIndex)),
			).Desc())))
		})
	})

	Describe("Collect", func() {
		var (
			containerMetricsChan chan prometheus.Metric
		)

		BeforeEach(func() {
			metricsStore.AddMetric(
				&events.Envelope{
					Origin:     proto.String(origin),
					EventType:  events.Envelope_ContainerMetric.Enum(),
					Timestamp:  proto.Int64(time.Now().Unix() * 1000),
					Deployment: proto.String(boshDeployment),
					Job:        proto.String(boshJob),
					Index:      proto.String(boshIndex),
					Ip:         proto.String(boshIP),
					ContainerMetric: &events.ContainerMetric{
						ApplicationId:    proto.String(containerMetric1ApplicationId),
						InstanceIndex:    proto.Int32(containerMetric1InstanceIndex),
						CpuPercentage:    proto.Float64(containerMetric1CpuPercentage),
						MemoryBytes:      proto.Uint64(containerMetric1MemoryBytes),
						DiskBytes:        proto.Uint64(containerMetric1DiskBytes),
						MemoryBytesQuota: proto.Uint64(containerMetric1MemoryBytesQuota),
						DiskBytesQuota:   proto.Uint64(containerMetric1DiskBytesQuota),
					},
				},
			)

			metricsStore.AddMetric(
				&events.Envelope{
					Origin:     proto.String(origin),
					EventType:  events.Envelope_ContainerMetric.Enum(),
					Timestamp:  proto.Int64(time.Now().Unix() * 1000),
					Deployment: proto.String(boshDeployment),
					Job:        proto.String(boshJob),
					Index:      proto.String(boshIndex),
					Ip:         proto.String(boshIP),
					ContainerMetric: &events.ContainerMetric{
						ApplicationId:    proto.String(containerMetric2ApplicationId),
						InstanceIndex:    proto.Int32(containerMetric2InstanceIndex),
						CpuPercentage:    proto.Float64(containerMetric2CpuPercentage),
						MemoryBytes:      proto.Uint64(containerMetric2MemoryBytes),
						DiskBytes:        proto.Uint64(containerMetric2DiskBytes),
						MemoryBytesQuota: proto.Uint64(containerMetric2MemoryBytesQuota),
						DiskBytesQuota:   proto.Uint64(containerMetric2DiskBytesQuota),
					},
				},
			)

			containerMetricsChan = make(chan prometheus.Metric)
		})

		JustBeforeEach(func() {
			go containerMetricsCollector.Collect(containerMetricsChan)
		})

		It("returns a container_metric_cpu_percentage metric for FakeApplicationId1", func() {
			Eventually(containerMetricsChan).Should(Receive(PrometheusMetric(cpuPercentageMetric.WithLabelValues(
				origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				containerMetric1ApplicationId,
				strconv.Itoa(int(containerMetric1InstanceIndex)),
			))))
		})

		It("returns a container_metric_memory_bytes metric for FakeApplicationId1", func() {
			Eventually(containerMetricsChan).Should(Receive(PrometheusMetric(memoryBytesMetric.WithLabelValues(
				origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				containerMetric1ApplicationId,
				strconv.Itoa(int(containerMetric1InstanceIndex)),
			))))
		})

		It("returns a container_metric_disk_bytes metric for FakeApplicationId1", func() {
			Eventually(containerMetricsChan).Should(Receive(PrometheusMetric(diskBytesMetric.WithLabelValues(
				origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				containerMetric1ApplicationId,
				strconv.Itoa(int(containerMetric1InstanceIndex)),
			))))
		})

		It("returns a container_metric_memory_bytes_quota metric for FakeApplicationId1", func() {
			Eventually(containerMetricsChan).Should(Receive(PrometheusMetric(memoryBytesQuotaMetric.WithLabelValues(
				origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				containerMetric1ApplicationId,
				strconv.Itoa(int(containerMetric1InstanceIndex)),
			))))
		})

		It("returns a container_metric_disk_bytes_quota metric for FakeApplicationId1", func() {
			Eventually(containerMetricsChan).Should(Receive(PrometheusMetric(diskBytesQuotaMetric.WithLabelValues(
				origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				containerMetric1ApplicationId,
				strconv.Itoa(int(containerMetric1InstanceIndex)),
			))))
		})

		It("returns a container_metric_cpu_percentage metric for FakeApplicationId2", func() {
			Eventually(containerMetricsChan).Should(Receive(PrometheusMetric(cpuPercentageMetric.WithLabelValues(
				origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				containerMetric2ApplicationId,
				strconv.Itoa(int(containerMetric2InstanceIndex)),
			))))
		})

		It("returns a container_metric_memory_bytes metric for FakeApplicationId2", func() {
			Eventually(containerMetricsChan).Should(Receive(PrometheusMetric(memoryBytesMetric.WithLabelValues(
				origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				containerMetric2ApplicationId,
				strconv.Itoa(int(containerMetric2InstanceIndex)),
			))))
		})

		It("returns a container_metric_disk_bytes metric for FakeApplicationId2", func() {
			Eventually(containerMetricsChan).Should(Receive(PrometheusMetric(diskBytesMetric.WithLabelValues(
				origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				containerMetric2ApplicationId,
				strconv.Itoa(int(containerMetric2InstanceIndex)),
			))))
		})

		It("returns a container_metric_memory_bytes_quota metric for FakeApplicationId2", func() {
			Eventually(containerMetricsChan).Should(Receive(PrometheusMetric(memoryBytesQuotaMetric.WithLabelValues(
				origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				containerMetric2ApplicationId,
				strconv.Itoa(int(containerMetric2InstanceIndex)),
			))))
		})

		It("returns a container_metric_disk_bytes_quota metric for FakeApplicationId2", func() {
			Eventually(containerMetricsChan).Should(Receive(PrometheusMetric(diskBytesQuotaMetric.WithLabelValues(
				origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				containerMetric2ApplicationId,
				strconv.Itoa(int(containerMetric2InstanceIndex)),
			))))
		})

		Context("when there is no container metrics", func() {
			BeforeEach(func() {
				metricsStore.FlushContainerMetrics()
			})

			It("does not return any metric", func() {
				Consistently(containerMetricsChan).ShouldNot(Receive())
			})
		})
	})
})
