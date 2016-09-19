package metrics

import (
	"strconv"

	"github.com/cloudfoundry/sonde-go/events"
)

type Store struct {
	internalMetrics  InternalMetrics
	containerMetrics ContainerMetrics
	counterMetrics   CounterMetrics
	valueMetrics     ValueMetrics
}

func NewStore() *Store {
	return &Store{
		internalMetrics: InternalMetrics{
			TotalEnvelopesReceived:        0,
			TotalMetricsReceived:          0,
			TotalContainerMetricsReceived: 0,
			TotalCounterEventsReceived:    0,
			TotalValueMetricsReceived:     0,
			SlowConsumerAlert:             false,
		},
		containerMetrics: make(map[string]ContainerMetric),
		counterMetrics:   make(map[string]CounterMetric),
		valueMetrics:     make(map[string]ValueMetric),
	}
}

func (s *Store) AddMetric(envelope *events.Envelope) {
	s.internalMetrics.TotalEnvelopesReceived++
	switch envelope.GetEventType() {
	case events.Envelope_ContainerMetric:
		s.internalMetrics.TotalMetricsReceived++
		s.internalMetrics.TotalContainerMetricsReceived++
		containerMetric := ContainerMetric{
			Origin:           envelope.GetOrigin(),
			Deployment:       envelope.GetDeployment(),
			Job:              envelope.GetJob(),
			Index:            envelope.GetIndex(),
			IP:               envelope.GetIp(),
			Tags:             envelope.GetTags(),
			ApplicationId:    envelope.GetContainerMetric().GetApplicationId(),
			InstanceIndex:    envelope.GetContainerMetric().GetInstanceIndex(),
			CpuPercentage:    envelope.GetContainerMetric().GetCpuPercentage(),
			MemoryBytes:      envelope.GetContainerMetric().GetMemoryBytes(),
			DiskBytes:        envelope.GetContainerMetric().GetDiskBytes(),
			MemoryBytesQuota: envelope.GetContainerMetric().GetMemoryBytesQuota(),
			DiskBytesQuota:   envelope.GetContainerMetric().GetDiskBytesQuota(),
		}
		containerKey := envelope.GetContainerMetric().GetApplicationId() + strconv.Itoa(int(containerMetric.InstanceIndex))
		s.containerMetrics[containerKey] = containerMetric
	case events.Envelope_CounterEvent:
		s.internalMetrics.TotalMetricsReceived++
		s.internalMetrics.TotalCounterEventsReceived++
		counterMetric := CounterMetric{
			Origin:     envelope.GetOrigin(),
			Deployment: envelope.GetDeployment(),
			Job:        envelope.GetJob(),
			Index:      envelope.GetIndex(),
			IP:         envelope.GetIp(),
			Tags:       envelope.GetTags(),
			Name:       envelope.GetCounterEvent().GetName(),
			Delta:      envelope.GetCounterEvent().GetDelta(),
			Total:      envelope.GetCounterEvent().GetTotal(),
		}
		counterKey := envelope.GetCounterEvent().GetName()
		s.counterMetrics[counterKey] = counterMetric
	case events.Envelope_ValueMetric:
		s.internalMetrics.TotalMetricsReceived++
		s.internalMetrics.TotalValueMetricsReceived++
		valueMetric := ValueMetric{
			Origin:     envelope.GetOrigin(),
			Deployment: envelope.GetDeployment(),
			Job:        envelope.GetJob(),
			Index:      envelope.GetIndex(),
			IP:         envelope.GetIp(),
			Tags:       envelope.GetTags(),
			Name:       envelope.GetValueMetric().GetName(),
			Value:      envelope.GetValueMetric().GetValue(),
			Unit:       envelope.GetValueMetric().GetUnit(),
		}
		valueKey := envelope.GetValueMetric().GetName()
		s.valueMetrics[valueKey] = valueMetric
	}
}

func (s *Store) AlertSlowConsumerError() {
	s.internalMetrics.SlowConsumerAlert = true
}

func (s *Store) GetInternalMetrics() InternalMetrics {
	return s.internalMetrics
}

func (s *Store) GetContainerMetrics() ContainerMetrics {
	return s.containerMetrics
}

func (s *Store) GetCounterMetrics() CounterMetrics {
	return s.counterMetrics
}

func (s *Store) GetValueMetrics() ValueMetrics {
	return s.valueMetrics
}
