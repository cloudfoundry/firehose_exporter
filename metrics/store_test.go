package metrics_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"

	"github.com/cloudfoundry-community/firehose_exporter/metrics"
)

var _ = Describe("Store", func() {
	var (
		metricsStore           *metrics.Store
		metricsExpiration      time.Duration
		metricsCleanupInterval time.Duration

		origin          = "fake-origin"
		boshDeployment  = "fake-deployment-name"
		boshJob         = "fake-job-name"
		boshIndex       = "0"
		boshIP          = "1.2.3.4"
		metricTimestamp = time.Now().Unix() * 1000

		containerMetricApplicationId    = "FakeApplicationId1"
		containerMetricInstanceIndex    = int32(1)
		containerMetricCpuPercentage    = float64(0.5)
		containerMetricMemoryBytes      = uint64(1000)
		containerMetricDiskBytes        = uint64(1500)
		containerMetricMemoryBytesQuota = uint64(2000)
		containerMetricDiskBytesQuota   = uint64(3000)

		counterEventName  = "FakeCounterEvent1"
		counterEventDelta = uint64(5)
		counterEventTotal = uint64(1000)

		valueMetricName  = "FakeValueMetric1"
		valueMetricValue = float64(2000)
		valueMetricUnit  = "kb"

		containerMetric metrics.ContainerMetric
		counterEvent    metrics.CounterEvent
		valueMetric     metrics.ValueMetric

		internalMetrics  metrics.InternalMetrics
		containerMetrics metrics.ContainerMetrics
		counterEvents    metrics.CounterEvents
		valueMetrics     metrics.ValueMetrics
	)

	BeforeEach(func() {
		metricsStore = metrics.NewStore(metricsExpiration, metricsCleanupInterval)
	})

	Describe("GetInternalMetrics", func() {
		BeforeEach(func() {
			internalMetrics = metricsStore.GetInternalMetrics()
		})

		It("returns the TotalEnvelopesReceived", func() {
			Expect(internalMetrics.TotalEnvelopesReceived).To(Equal(int64(0)))
		})

		It("returns the LastEnvelopReceivedTimestamp", func() {
			Expect(internalMetrics.LastEnvelopReceivedTimestamp).To(Equal(int64(0)))
		})

		It("returns the TotalMetricsReceived", func() {
			Expect(internalMetrics.TotalMetricsReceived).To(Equal(int64(0)))
		})

		It("returns the LastMetricReceivedTimestamp", func() {
			Expect(internalMetrics.LastMetricReceivedTimestamp).To(Equal(int64(0)))
		})

		It("returns the TotalContainerMetricsReceived", func() {
			Expect(internalMetrics.TotalContainerMetricsReceived).To(Equal(int64(0)))
		})

		It("returns the LastContainerMetricReceivedTimestamp", func() {
			Expect(internalMetrics.LastContainerMetricReceivedTimestamp).To(Equal(int64(0)))
		})

		It("returns the TotalCounterEventsReceived", func() {
			Expect(internalMetrics.TotalCounterEventsReceived).To(Equal(int64(0)))
		})

		It("returns the LastCounterEventReceivedTimestamp", func() {
			Expect(internalMetrics.LastCounterEventReceivedTimestamp).To(Equal(int64(0)))
		})

		It("returns the TotalValueMetricsReceived", func() {
			Expect(internalMetrics.TotalValueMetricsReceived).To(Equal(int64(0)))
		})

		It("returns the LastValueMetricReceivedTimestamp", func() {
			Expect(internalMetrics.LastValueMetricReceivedTimestamp).To(Equal(int64(0)))
		})

		It("returns the SlowConsumerAlert", func() {
			Expect(internalMetrics.SlowConsumerAlert).To(BeFalse())
		})

		It("returns the LastSlowConsumerAlertTimestamp", func() {
			Expect(internalMetrics.LastSlowConsumerAlertTimestamp).To(Equal(int64(0)))
		})
	})

	Describe("SetInternalMetrics", func() {
		var (
			totalEnvelopesReceived               = int64(1000)
			lastEnvelopeReceivedTimestamp        = time.Now().UnixNano()
			totalMetricsReceived                 = int64(500)
			lastMetricReceivedTimestamp          = time.Now().UnixNano()
			totalContainerMetricsReceived        = int64(100)
			lastContainerMetricReceivedTimestamp = time.Now().UnixNano()
			totalCounterEventsReceived           = int64(200)
			lastCounterEventReceivedTimestamp    = time.Now().UnixNano()
			totalValueMetricsReceived            = int64(300)
			lastValueMetricReceivedTimestamp     = time.Now().UnixNano()
			slowConsumerAlert                    = true
			lastSlowConsumerAlertTimestamp       = time.Now().UnixNano()
		)

		BeforeEach(func() {
			metricsStore.SetInternalMetrics(metrics.InternalMetrics{
				TotalEnvelopesReceived:               totalEnvelopesReceived,
				LastEnvelopReceivedTimestamp:         lastEnvelopeReceivedTimestamp,
				TotalMetricsReceived:                 totalMetricsReceived,
				LastMetricReceivedTimestamp:          lastMetricReceivedTimestamp,
				TotalContainerMetricsReceived:        totalContainerMetricsReceived,
				LastContainerMetricReceivedTimestamp: lastContainerMetricReceivedTimestamp,
				TotalCounterEventsReceived:           totalCounterEventsReceived,
				LastCounterEventReceivedTimestamp:    lastCounterEventReceivedTimestamp,
				TotalValueMetricsReceived:            totalValueMetricsReceived,
				LastValueMetricReceivedTimestamp:     lastValueMetricReceivedTimestamp,
				SlowConsumerAlert:                    slowConsumerAlert,
				LastSlowConsumerAlertTimestamp:       lastSlowConsumerAlertTimestamp,
			})

			internalMetrics = metricsStore.GetInternalMetrics()
		})

		It("sets the TotalEnvelopesReceived", func() {
			Expect(internalMetrics.TotalEnvelopesReceived).To(Equal(totalEnvelopesReceived))
		})

		It("sets the LastEnvelopReceivedTimestamp", func() {
			Expect(internalMetrics.LastEnvelopReceivedTimestamp).To(Equal(lastEnvelopeReceivedTimestamp))
		})

		It("sets the TotalMetricsReceived", func() {
			Expect(internalMetrics.TotalMetricsReceived).To(Equal(totalMetricsReceived))
		})

		It("sets the LastMetricReceivedTimestamp", func() {
			Expect(internalMetrics.LastMetricReceivedTimestamp).To(Equal(lastMetricReceivedTimestamp))
		})

		It("sets the TotalContainerMetricsReceived", func() {
			Expect(internalMetrics.TotalContainerMetricsReceived).To(Equal(totalContainerMetricsReceived))
		})

		It("sets the LastContainerMetricReceivedTimestamp", func() {
			Expect(internalMetrics.LastContainerMetricReceivedTimestamp).To(Equal(lastContainerMetricReceivedTimestamp))
		})

		It("sets the TotalCounterEventsReceived", func() {
			Expect(internalMetrics.TotalCounterEventsReceived).To(Equal(totalCounterEventsReceived))
		})

		It("sets the LastCounterEventReceivedTimestamp", func() {
			Expect(internalMetrics.LastCounterEventReceivedTimestamp).To(Equal(lastCounterEventReceivedTimestamp))
		})

		It("sets the TotalValueMetricsReceived", func() {
			Expect(internalMetrics.TotalValueMetricsReceived).To(Equal(totalValueMetricsReceived))
		})

		It("sets the LastValueMetricReceivedTimestamp", func() {
			Expect(internalMetrics.LastValueMetricReceivedTimestamp).To(Equal(lastValueMetricReceivedTimestamp))
		})

		It("sets the SlowConsumerAlert", func() {
			Expect(internalMetrics.SlowConsumerAlert).To(Equal(slowConsumerAlert))
		})

		It("sets the LastSlowConsumerAlertTimestamp", func() {
			Expect(internalMetrics.LastSlowConsumerAlertTimestamp).To(Equal(lastSlowConsumerAlertTimestamp))
		})
	})

	Describe("AlertSlowConsumerError", func() {
		BeforeEach(func() {
			metricsStore.AlertSlowConsumerError()

			internalMetrics = metricsStore.GetInternalMetrics()
		})

		It("sets the SlowConsumerAlert", func() {
			Expect(internalMetrics.SlowConsumerAlert).To(BeTrue())
		})

		It("sets the LastSlowConsumerAlertTimestamp", func() {
			Expect(internalMetrics.LastSlowConsumerAlertTimestamp).ToNot(Equal(int64(0)))
		})
	})

	Describe("AddMetric", func() {
		BeforeEach(func() {
			metricsStore.AddMetric(
				&events.Envelope{
					Origin:     proto.String(origin),
					EventType:  events.Envelope_Error.Enum(),
					Timestamp:  proto.Int64(metricTimestamp),
					Deployment: proto.String(boshDeployment),
					Job:        proto.String(boshJob),
					Index:      proto.String(boshIndex),
					Ip:         proto.String(boshIP),
					Error: &events.Error{
						Source:  proto.String("error-source"),
						Code:    proto.Int32(127),
						Message: proto.String("error-message"),
					},
				},
			)

			metricsStore.AddMetric(
				&events.Envelope{
					Origin:     proto.String(origin),
					EventType:  events.Envelope_ContainerMetric.Enum(),
					Timestamp:  proto.Int64(metricTimestamp),
					Deployment: proto.String(boshDeployment),
					Job:        proto.String(boshJob),
					Index:      proto.String(boshIndex),
					Ip:         proto.String(boshIP),
					ContainerMetric: &events.ContainerMetric{
						ApplicationId:    proto.String(containerMetricApplicationId),
						InstanceIndex:    proto.Int32(containerMetricInstanceIndex),
						CpuPercentage:    proto.Float64(containerMetricCpuPercentage),
						MemoryBytes:      proto.Uint64(containerMetricMemoryBytes),
						DiskBytes:        proto.Uint64(containerMetricDiskBytes),
						MemoryBytesQuota: proto.Uint64(containerMetricMemoryBytesQuota),
						DiskBytesQuota:   proto.Uint64(containerMetricDiskBytesQuota),
					},
				},
			)

			containerMetric = metrics.ContainerMetric{
				Origin:           origin,
				Timestamp:        metricTimestamp,
				Deployment:       boshDeployment,
				Job:              boshJob,
				Index:            boshIndex,
				IP:               boshIP,
				ApplicationId:    containerMetricApplicationId,
				InstanceIndex:    containerMetricInstanceIndex,
				CpuPercentage:    containerMetricCpuPercentage,
				MemoryBytes:      containerMetricMemoryBytes,
				DiskBytes:        containerMetricDiskBytes,
				MemoryBytesQuota: containerMetricMemoryBytesQuota,
				DiskBytesQuota:   containerMetricDiskBytesQuota,
			}

			metricsStore.AddMetric(
				&events.Envelope{
					Origin:     proto.String(origin),
					EventType:  events.Envelope_CounterEvent.Enum(),
					Timestamp:  proto.Int64(metricTimestamp),
					Deployment: proto.String(boshDeployment),
					Job:        proto.String(boshJob),
					Index:      proto.String(boshIndex),
					Ip:         proto.String(boshIP),
					CounterEvent: &events.CounterEvent{
						Name:  proto.String(counterEventName),
						Delta: proto.Uint64(counterEventDelta),
						Total: proto.Uint64(counterEventTotal),
					},
				},
			)

			counterEvent = metrics.CounterEvent{
				Origin:     origin,
				Timestamp:  metricTimestamp,
				Deployment: boshDeployment,
				Job:        boshJob,
				Index:      boshIndex,
				IP:         boshIP,
				Name:       counterEventName,
				Delta:      counterEventDelta,
				Total:      counterEventTotal,
			}

			metricsStore.AddMetric(
				&events.Envelope{
					Origin:     proto.String(origin),
					EventType:  events.Envelope_ValueMetric.Enum(),
					Timestamp:  proto.Int64(metricTimestamp),
					Deployment: proto.String(boshDeployment),
					Job:        proto.String(boshJob),
					Index:      proto.String(boshIndex),
					Ip:         proto.String(boshIP),
					ValueMetric: &events.ValueMetric{
						Name:  proto.String(valueMetricName),
						Value: proto.Float64(valueMetricValue),
						Unit:  proto.String(valueMetricUnit),
					},
				},
			)

			valueMetric = metrics.ValueMetric{
				Origin:     origin,
				Timestamp:  metricTimestamp,
				Deployment: boshDeployment,
				Job:        boshJob,
				Index:      boshIndex,
				IP:         boshIP,
				Name:       valueMetricName,
				Value:      valueMetricValue,
				Unit:       valueMetricUnit,
			}

			internalMetrics = metricsStore.GetInternalMetrics()
			containerMetrics = metricsStore.GetContainerMetrics()
			counterEvents = metricsStore.GetCounterEvents()
			valueMetrics = metricsStore.GetValueMetrics()
		})

		It("returns the TotalEnvelopesReceived", func() {
			Expect(internalMetrics.TotalEnvelopesReceived).To(Equal(int64(4)))
		})

		It("returns the LastEnvelopReceivedTimestamp", func() {
			Expect(internalMetrics.LastEnvelopReceivedTimestamp).ToNot(Equal(int64(0)))
		})

		It("returns the TotalMetricsReceived", func() {
			Expect(internalMetrics.TotalMetricsReceived).To(Equal(int64(3)))
		})

		It("returns the LastMetricReceivedTimestamp", func() {
			Expect(internalMetrics.LastMetricReceivedTimestamp).ToNot(Equal(int64(0)))
		})

		It("returns the TotalContainerMetricsReceived", func() {
			Expect(internalMetrics.TotalContainerMetricsReceived).To(Equal(int64(1)))
		})

		It("returns the LastContainerMetricReceivedTimestamp", func() {
			Expect(internalMetrics.LastContainerMetricReceivedTimestamp).ToNot(Equal(int64(0)))
		})

		It("returns the TotalCounterEventsReceived", func() {
			Expect(internalMetrics.TotalCounterEventsReceived).To(Equal(int64(1)))
		})

		It("returns the LastCounterEventReceivedTimestamp", func() {
			Expect(internalMetrics.LastCounterEventReceivedTimestamp).ToNot(Equal(int64(0)))
		})

		It("returns the TotalValueMetricsReceived", func() {
			Expect(internalMetrics.TotalValueMetricsReceived).To(Equal(int64(1)))
		})

		It("returns the LastValueMetricReceivedTimestamp", func() {
			Expect(internalMetrics.LastValueMetricReceivedTimestamp).ToNot(Equal(int64(0)))
		})

		It("adds a container metric", func() {
			Expect(len(containerMetrics)).To(Equal(1))
			Expect(containerMetrics[0]).To(Equal(containerMetric))
		})

		It("adds a counter event", func() {
			Expect(len(counterEvents)).To(Equal(1))
			Expect(counterEvents[0]).To(Equal(counterEvent))
		})

		It("adds a value metric", func() {
			Expect(len(valueMetrics)).To(Equal(1))
			Expect(valueMetrics[0]).To(Equal(valueMetric))
		})
	})

	Describe("GetContainerMetrics", func() {
		BeforeEach(func() {
			metricsStore.AddMetric(
				&events.Envelope{
					Origin:     proto.String(origin),
					EventType:  events.Envelope_ContainerMetric.Enum(),
					Timestamp:  proto.Int64(metricTimestamp),
					Deployment: proto.String(boshDeployment),
					Job:        proto.String(boshJob),
					Index:      proto.String(boshIndex),
					Ip:         proto.String(boshIP),
					Tags:       map[string]string{},
					ContainerMetric: &events.ContainerMetric{
						ApplicationId:    proto.String(containerMetricApplicationId),
						InstanceIndex:    proto.Int32(containerMetricInstanceIndex),
						CpuPercentage:    proto.Float64(containerMetricCpuPercentage),
						MemoryBytes:      proto.Uint64(containerMetricMemoryBytes),
						DiskBytes:        proto.Uint64(containerMetricDiskBytes),
						MemoryBytesQuota: proto.Uint64(containerMetricMemoryBytesQuota),
						DiskBytesQuota:   proto.Uint64(containerMetricDiskBytesQuota),
					},
				},
			)

			containerMetric = metrics.ContainerMetric{
				Origin:           origin,
				Timestamp:        metricTimestamp,
				Deployment:       boshDeployment,
				Job:              boshJob,
				Index:            boshIndex,
				IP:               boshIP,
				Tags:             map[string]string{},
				ApplicationId:    containerMetricApplicationId,
				InstanceIndex:    containerMetricInstanceIndex,
				CpuPercentage:    containerMetricCpuPercentage,
				MemoryBytes:      containerMetricMemoryBytes,
				DiskBytes:        containerMetricDiskBytes,
				MemoryBytesQuota: containerMetricMemoryBytesQuota,
				DiskBytesQuota:   containerMetricDiskBytesQuota,
			}

			containerMetrics = metricsStore.GetContainerMetrics()
		})

		It("returns the container metrics", func() {
			Expect(len(containerMetrics)).To(Equal(1))
			Expect(containerMetrics[0]).To(Equal(containerMetric))
		})
	})

	Describe("FlushContainerMetrics", func() {
		BeforeEach(func() {
			metricsStore.AddMetric(
				&events.Envelope{
					Origin:     proto.String(origin),
					EventType:  events.Envelope_ContainerMetric.Enum(),
					Timestamp:  proto.Int64(metricTimestamp),
					Deployment: proto.String(boshDeployment),
					Job:        proto.String(boshJob),
					Index:      proto.String(boshIndex),
					Ip:         proto.String(boshIP),
					Tags:       map[string]string{},
					ContainerMetric: &events.ContainerMetric{
						ApplicationId:    proto.String(containerMetricApplicationId),
						InstanceIndex:    proto.Int32(containerMetricInstanceIndex),
						CpuPercentage:    proto.Float64(containerMetricCpuPercentage),
						MemoryBytes:      proto.Uint64(containerMetricMemoryBytes),
						DiskBytes:        proto.Uint64(containerMetricDiskBytes),
						MemoryBytesQuota: proto.Uint64(containerMetricMemoryBytesQuota),
						DiskBytesQuota:   proto.Uint64(containerMetricDiskBytesQuota),
					},
				},
			)

			containerMetric = metrics.ContainerMetric{
				Origin:           origin,
				Timestamp:        metricTimestamp,
				Deployment:       boshDeployment,
				Job:              boshJob,
				Index:            boshIndex,
				IP:               boshIP,
				Tags:             map[string]string{},
				ApplicationId:    containerMetricApplicationId,
				InstanceIndex:    containerMetricInstanceIndex,
				CpuPercentage:    containerMetricCpuPercentage,
				MemoryBytes:      containerMetricMemoryBytes,
				DiskBytes:        containerMetricDiskBytes,
				MemoryBytesQuota: containerMetricMemoryBytesQuota,
				DiskBytesQuota:   containerMetricDiskBytesQuota,
			}

			metricsStore.FlushContainerMetrics()
			containerMetrics = metricsStore.GetContainerMetrics()
		})

		It("returns empty container metrics", func() {
			Expect(len(containerMetrics)).To(Equal(0))
		})
	})

	Describe("GetCounterEvents", func() {
		BeforeEach(func() {
			metricsStore.AddMetric(
				&events.Envelope{
					Origin:     proto.String(origin),
					EventType:  events.Envelope_CounterEvent.Enum(),
					Timestamp:  proto.Int64(metricTimestamp),
					Deployment: proto.String(boshDeployment),
					Job:        proto.String(boshJob),
					Index:      proto.String(boshIndex),
					Ip:         proto.String(boshIP),
					Tags:       map[string]string{},
					CounterEvent: &events.CounterEvent{
						Name:  proto.String(counterEventName),
						Delta: proto.Uint64(counterEventDelta),
						Total: proto.Uint64(counterEventTotal),
					},
				},
			)

			counterEvent = metrics.CounterEvent{
				Origin:     origin,
				Timestamp:  metricTimestamp,
				Deployment: boshDeployment,
				Job:        boshJob,
				Index:      boshIndex,
				IP:         boshIP,
				Tags:       map[string]string{},
				Name:       counterEventName,
				Delta:      counterEventDelta,
				Total:      counterEventTotal,
			}

			counterEvents = metricsStore.GetCounterEvents()
		})

		It("returns the counter events", func() {
			Expect(len(counterEvents)).To(Equal(1))
			Expect(counterEvents[0]).To(Equal(counterEvent))
		})
	})

	Describe("FlushCounterEvents", func() {
		BeforeEach(func() {
			metricsStore.AddMetric(
				&events.Envelope{
					Origin:     proto.String(origin),
					EventType:  events.Envelope_CounterEvent.Enum(),
					Timestamp:  proto.Int64(metricTimestamp),
					Deployment: proto.String(boshDeployment),
					Job:        proto.String(boshJob),
					Index:      proto.String(boshIndex),
					Ip:         proto.String(boshIP),
					Tags:       map[string]string{},
					CounterEvent: &events.CounterEvent{
						Name:  proto.String(counterEventName),
						Delta: proto.Uint64(counterEventDelta),
						Total: proto.Uint64(counterEventTotal),
					},
				},
			)

			counterEvent = metrics.CounterEvent{
				Origin:     origin,
				Timestamp:  metricTimestamp,
				Deployment: boshDeployment,
				Job:        boshJob,
				Index:      boshIndex,
				IP:         boshIP,
				Tags:       map[string]string{},
				Name:       counterEventName,
				Delta:      counterEventDelta,
				Total:      counterEventTotal,
			}

			metricsStore.FlushCounterEvents()
			counterEvents = metricsStore.GetCounterEvents()
		})

		It("returns empty counter events", func() {
			Expect(len(counterEvents)).To(Equal(0))
		})
	})

	Describe("GetValueMetrics", func() {
		BeforeEach(func() {
			metricsStore.AddMetric(
				&events.Envelope{
					Origin:     proto.String(origin),
					EventType:  events.Envelope_ValueMetric.Enum(),
					Timestamp:  proto.Int64(metricTimestamp),
					Deployment: proto.String(boshDeployment),
					Job:        proto.String(boshJob),
					Index:      proto.String(boshIndex),
					Ip:         proto.String(boshIP),
					Tags:       map[string]string{},
					ValueMetric: &events.ValueMetric{
						Name:  proto.String(valueMetricName),
						Value: proto.Float64(valueMetricValue),
						Unit:  proto.String(valueMetricUnit),
					},
				},
			)

			valueMetric = metrics.ValueMetric{
				Origin:     origin,
				Timestamp:  metricTimestamp,
				Deployment: boshDeployment,
				Job:        boshJob,
				Index:      boshIndex,
				IP:         boshIP,
				Tags:       map[string]string{},
				Name:       valueMetricName,
				Value:      valueMetricValue,
				Unit:       valueMetricUnit,
			}

			valueMetrics = metricsStore.GetValueMetrics()
		})

		It("returns the value metrics", func() {
			Expect(len(valueMetrics)).To(Equal(1))
			Expect(valueMetrics[0]).To(Equal(valueMetric))
		})
	})

	Describe("FlushValueMetrics", func() {
		BeforeEach(func() {
			metricsStore.AddMetric(
				&events.Envelope{
					Origin:     proto.String(origin),
					EventType:  events.Envelope_ValueMetric.Enum(),
					Timestamp:  proto.Int64(metricTimestamp),
					Deployment: proto.String(boshDeployment),
					Job:        proto.String(boshJob),
					Index:      proto.String(boshIndex),
					Ip:         proto.String(boshIP),
					Tags:       map[string]string{},
					ValueMetric: &events.ValueMetric{
						Name:  proto.String(valueMetricName),
						Value: proto.Float64(valueMetricValue),
						Unit:  proto.String(valueMetricUnit),
					},
				},
			)

			valueMetric = metrics.ValueMetric{
				Origin:     origin,
				Timestamp:  metricTimestamp,
				Deployment: boshDeployment,
				Job:        boshJob,
				Index:      boshIndex,
				IP:         boshIP,
				Tags:       map[string]string{},
				Name:       valueMetricName,
				Value:      valueMetricValue,
				Unit:       valueMetricUnit,
			}

			metricsStore.FlushValueMetrics()
			valueMetrics = metricsStore.GetValueMetrics()
		})

		It("returns empty value metrics", func() {
			Expect(len(valueMetrics)).To(Equal(0))
		})
	})
})
