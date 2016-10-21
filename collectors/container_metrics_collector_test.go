package collectors_test

import (
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-community/firehose_exporter/metrics"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"

	. "github.com/cloudfoundry-community/firehose_exporter/collectors"
)

var _ = Describe("ContainerMetricsCollector", func() {
	var (
		namespace                 string
		metricsStore              *metrics.Store
		metricsExpiration         time.Duration
		metricsCleanupInterval    time.Duration
		dopplerDeployments        []string
		containerMetricsCollector *ContainerMetricsCollector

		cpuPercentageMetricDesc    *prometheus.Desc
		memoryBytesMetricDesc      *prometheus.Desc
		diskBytesMetricDesc        *prometheus.Desc
		memoryBytesQuotaMetricDesc *prometheus.Desc
		diskBytesQuotaMetricDesc   *prometheus.Desc
	)

	BeforeEach(func() {
		namespace = "test_exporter"
		metricsStore = metrics.NewStore(metricsExpiration, metricsCleanupInterval)
		dopplerDeployments = []string{}

		cpuPercentageMetricDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "container_metric", "cpu_percentage"),
			"Cloud Foundry Firehose container metric: CPU used, on a scale of 0 to 100.",
			[]string{"origin", "bosh_deployment", "bosh_job", "bosh_index", "bosh_ip", "application_id", "instance_id"},
			nil,
		)

		memoryBytesMetricDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "container_metric", "memory_bytes"),
			"Cloud Foundry Firehose container metric: bytes of memory used.",
			[]string{"origin", "bosh_deployment", "bosh_job", "bosh_index", "bosh_ip", "application_id", "instance_id"},
			nil,
		)

		diskBytesMetricDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "container_metric", "disk_bytes"),
			"Cloud Foundry Firehose container metric: bytes of disk used.",
			[]string{"origin", "bosh_deployment", "bosh_job", "bosh_index", "bosh_ip", "application_id", "instance_id"},
			nil,
		)

		memoryBytesQuotaMetricDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "container_metric", "memory_bytes_quota"),
			"Cloud Foundry Firehose container metric: maximum bytes of memory allocated to container.",
			[]string{"origin", "bosh_deployment", "bosh_job", "bosh_index", "bosh_ip", "application_id", "instance_id"},
			nil,
		)

		diskBytesQuotaMetricDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "container_metric", "disk_bytes_quota"),
			"Cloud Foundry Firehose container metric: maximum bytes of disk allocated to container.",
			[]string{"origin", "bosh_deployment", "bosh_job", "bosh_index", "bosh_ip", "application_id", "instance_id"},
			nil,
		)
	})

	JustBeforeEach(func() {
		containerMetricsCollector = NewContainerMetricsCollector(namespace, metricsStore, dopplerDeployments)
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
			Eventually(descriptions).Should(Receive(Equal(cpuPercentageMetricDesc)))
		})

		It("returns a container_metric_memory_bytes metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(memoryBytesMetricDesc)))
		})

		It("returns a container_metric_disk_bytes metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(diskBytesMetricDesc)))
		})

		It("returns a container_metric_memory_bytes_quota metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(memoryBytesQuotaMetricDesc)))
		})

		It("returns a container_metric_disk_bytes_quota metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(diskBytesQuotaMetricDesc)))
		})
	})

	Describe("Collect", func() {
		var (
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

			containerMetricsChan    chan prometheus.Metric
			cpuPercentageMetric1    prometheus.Metric
			memoryBytesMetric1      prometheus.Metric
			diskBytesMetric1        prometheus.Metric
			memoryBytesQuotaMetric1 prometheus.Metric
			diskBytesQuotaMetric1   prometheus.Metric
			cpuPercentageMetric2    prometheus.Metric
			memoryBytesMetric2      prometheus.Metric
			diskBytesMetric2        prometheus.Metric
			memoryBytesQuotaMetric2 prometheus.Metric
			diskBytesQuotaMetric2   prometheus.Metric
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

			cpuPercentageMetric1 = prometheus.MustNewConstMetric(
				cpuPercentageMetricDesc,
				prometheus.GaugeValue,
				containerMetric1CpuPercentage,
				origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				containerMetric1ApplicationId,
				strconv.Itoa(int(containerMetric1InstanceIndex)),
			)

			memoryBytesMetric1 = prometheus.MustNewConstMetric(
				memoryBytesMetricDesc,
				prometheus.GaugeValue,
				float64(containerMetric1MemoryBytes),
				origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				containerMetric1ApplicationId,
				strconv.Itoa(int(containerMetric1InstanceIndex)),
			)

			diskBytesMetric1 = prometheus.MustNewConstMetric(
				diskBytesMetricDesc,
				prometheus.GaugeValue,
				float64(containerMetric1DiskBytes),
				origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				containerMetric1ApplicationId,
				strconv.Itoa(int(containerMetric1InstanceIndex)),
			)

			memoryBytesQuotaMetric1 = prometheus.MustNewConstMetric(
				memoryBytesQuotaMetricDesc,
				prometheus.GaugeValue,
				float64(containerMetric1MemoryBytesQuota),
				origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				containerMetric1ApplicationId,
				strconv.Itoa(int(containerMetric1InstanceIndex)),
			)

			diskBytesQuotaMetric1 = prometheus.MustNewConstMetric(
				diskBytesQuotaMetricDesc,
				prometheus.GaugeValue,
				float64(containerMetric1DiskBytesQuota),
				origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				containerMetric1ApplicationId,
				strconv.Itoa(int(containerMetric1InstanceIndex)),
			)

			cpuPercentageMetric2 = prometheus.MustNewConstMetric(
				cpuPercentageMetricDesc,
				prometheus.GaugeValue,
				containerMetric2CpuPercentage,
				origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				containerMetric2ApplicationId,
				strconv.Itoa(int(containerMetric2InstanceIndex)),
			)

			memoryBytesMetric2 = prometheus.MustNewConstMetric(
				memoryBytesMetricDesc,
				prometheus.GaugeValue,
				float64(containerMetric2MemoryBytes),
				origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				containerMetric2ApplicationId,
				strconv.Itoa(int(containerMetric2InstanceIndex)),
			)

			diskBytesMetric2 = prometheus.MustNewConstMetric(
				diskBytesMetricDesc,
				prometheus.GaugeValue,
				float64(containerMetric2DiskBytes),
				origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				containerMetric2ApplicationId,
				strconv.Itoa(int(containerMetric2InstanceIndex)),
			)

			memoryBytesQuotaMetric2 = prometheus.MustNewConstMetric(
				memoryBytesQuotaMetricDesc,
				prometheus.GaugeValue,
				float64(containerMetric2MemoryBytesQuota),
				origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				containerMetric2ApplicationId,
				strconv.Itoa(int(containerMetric2InstanceIndex)),
			)

			diskBytesQuotaMetric2 = prometheus.MustNewConstMetric(
				diskBytesQuotaMetricDesc,
				prometheus.GaugeValue,
				float64(containerMetric2DiskBytesQuota),
				origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				containerMetric2ApplicationId,
				strconv.Itoa(int(containerMetric2InstanceIndex)),
			)
		})

		JustBeforeEach(func() {
			go containerMetricsCollector.Collect(containerMetricsChan)
		})

		It("returns a container_metric_cpu_percentage metric for FakeApplicationId1", func() {
			Eventually(containerMetricsChan).Should(Receive(Equal(cpuPercentageMetric1)))
		})

		It("returns a container_metric_memory_bytes metric for FakeApplicationId1", func() {
			Eventually(containerMetricsChan).Should(Receive(Equal(memoryBytesMetric1)))
		})

		It("returns a container_metric_disk_bytes metric for FakeApplicationId1", func() {
			Eventually(containerMetricsChan).Should(Receive(Equal(diskBytesMetric1)))
		})

		It("returns a container_metric_memory_bytes_quota metric for FakeApplicationId1", func() {
			Eventually(containerMetricsChan).Should(Receive(Equal(memoryBytesQuotaMetric1)))
		})

		It("returns a container_metric_disk_bytes_quota metric for FakeApplicationId1", func() {
			Eventually(containerMetricsChan).Should(Receive(Equal(diskBytesQuotaMetric1)))
		})

		It("returns a container_metric_cpu_percentage metric for FakeApplicationId2", func() {
			Eventually(containerMetricsChan).Should(Receive(Equal(cpuPercentageMetric2)))
		})

		It("returns a container_metric_memory_bytes metric for FakeApplicationId2", func() {
			Eventually(containerMetricsChan).Should(Receive(Equal(memoryBytesMetric2)))
		})

		It("returns a container_metric_disk_bytes metric for FakeApplicationId2", func() {
			Eventually(containerMetricsChan).Should(Receive(Equal(diskBytesMetric2)))
		})

		It("returns a container_metric_memory_bytes_quota metric for FakeApplicationId2", func() {
			Eventually(containerMetricsChan).Should(Receive(Equal(memoryBytesQuotaMetric2)))
		})

		It("returns a container_metric_disk_bytes_quota metric for FakeApplicationId2", func() {
			Eventually(containerMetricsChan).Should(Receive(Equal(diskBytesQuotaMetric2)))
		})

		Context("when there is no container metrics", func() {
			BeforeEach(func() {
				metricsStore.FlushContainerMetrics()
			})

			It("does not return any metric", func() {
				Consistently(containerMetricsChan).ShouldNot(Receive())
			})
		})

		Context("when there is a deployment filter", func() {
			BeforeEach(func() {
				dopplerDeployments = []string{"fake-deployment-name"}
			})

			It("returns a container_metric_cpu_percentage metric for FakeApplicationId1", func() {
				Eventually(containerMetricsChan).Should(Receive(Equal(cpuPercentageMetric1)))
			})

			It("returns a container_metric_memory_bytes metric for FakeApplicationId1", func() {
				Eventually(containerMetricsChan).Should(Receive(Equal(memoryBytesMetric1)))
			})

			It("returns a container_metric_disk_bytes metric for FakeApplicationId1", func() {
				Eventually(containerMetricsChan).Should(Receive(Equal(diskBytesMetric1)))
			})

			It("returns a container_metric_memory_bytes_quota metric for FakeApplicationId1", func() {
				Eventually(containerMetricsChan).Should(Receive(Equal(memoryBytesQuotaMetric1)))
			})

			It("returns a container_metric_disk_bytes_quota metric for FakeApplicationId1", func() {
				Eventually(containerMetricsChan).Should(Receive(Equal(diskBytesQuotaMetric1)))
			})

			It("returns a container_metric_cpu_percentage metric for FakeApplicationId2", func() {
				Eventually(containerMetricsChan).Should(Receive(Equal(cpuPercentageMetric2)))
			})

			It("returns a container_metric_memory_bytes metric for FakeApplicationId2", func() {
				Eventually(containerMetricsChan).Should(Receive(Equal(memoryBytesMetric2)))
			})

			It("returns a container_metric_disk_bytes metric for FakeApplicationId2", func() {
				Eventually(containerMetricsChan).Should(Receive(Equal(diskBytesMetric2)))
			})

			It("returns a container_metric_memory_bytes_quota metric for FakeApplicationId2", func() {
				Eventually(containerMetricsChan).Should(Receive(Equal(memoryBytesQuotaMetric2)))
			})

			It("returns a container_metric_disk_bytes_quota metric for FakeApplicationId2", func() {
				Eventually(containerMetricsChan).Should(Receive(Equal(diskBytesQuotaMetric2)))
			})

			Context("and the metrics deployment does not match", func() {
				BeforeEach(func() {
					dopplerDeployments = []string{"another-fake-deployment-name"}
				})

				It("does not return any metric", func() {
					Consistently(containerMetricsChan).ShouldNot(Receive())
				})
			})
		})
	})
})
