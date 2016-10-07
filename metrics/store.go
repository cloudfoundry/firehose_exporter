package metrics

import (
	"strconv"
	"sync"
	"time"

	"github.com/cloudfoundry/sonde-go/events"
)

type Store struct {
	metricsGarbage        time.Duration
	metricsExpiry         time.Duration
	internalMetrics       InternalMetrics
	internalMetricsMutex  sync.Mutex
	containerMetrics      ContainerMetrics
	containerMetricsMutex sync.Mutex
	counterMetrics        CounterMetrics
	counterMetricsMutex   sync.Mutex
	valueMetrics          ValueMetrics
	valueMetricsMutex     sync.Mutex
}

func NewStore(
	metricsGarbage time.Duration,
	metricsExpiry time.Duration,
) *Store {
	return &Store{
		metricsGarbage: metricsGarbage,
		metricsExpiry:  metricsExpiry,
		internalMetrics: InternalMetrics{
			TotalEnvelopesReceived:        0,
			TotalMetricsReceived:          0,
			TotalContainerMetricsReceived: 0,
			TotalCounterEventsReceived:    0,
			TotalValueMetricsReceived:     0,
			SlowConsumerAlert:             false,
			LastReceivedMetricTimestamp:   0,
		},
		containerMetrics: make(map[string]ContainerMetric),
		counterMetrics:   make(map[string]CounterMetric),
		valueMetrics:     make(map[string]ValueMetric),
	}
}

func (s *Store) Start() {
	ticker := time.NewTicker(s.metricsGarbage).C
	for {
		select {
		case <-ticker:
			s.expireInternalMetrics()
			s.expireContainerMetrics()
			s.expireCounterMetrics()
			s.expireValueMetrics()
		}
	}
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

func (s *Store) AlertSlowConsumerError() {
	s.internalMetricsMutex.Lock()
	s.internalMetrics.SlowConsumerAlert = true
	s.internalMetricsMutex.Unlock()
}

func (s *Store) expireInternalMetrics() {
	s.internalMetricsMutex.Lock()
	s.internalMetrics.SlowConsumerAlert = false
	s.internalMetricsMutex.Unlock()
}

func (s *Store) AddMetric(envelope *events.Envelope) {
	s.internalMetricsMutex.Lock()
	s.internalMetrics.TotalEnvelopesReceived++
	s.internalMetrics.LastReceivedEnvelopTimestamp = time.Now().Unix()
	s.internalMetricsMutex.Unlock()

	switch envelope.GetEventType() {
	case events.Envelope_ContainerMetric:
		s.addContainerMetric(envelope)
	case events.Envelope_CounterEvent:
		s.addCounterMetric(envelope)
	case events.Envelope_ValueMetric:
		s.addValueMetric(envelope)
	}
}

func (s *Store) addContainerMetric(envelope *events.Envelope) {
	s.internalMetricsMutex.Lock()
	s.internalMetrics.TotalMetricsReceived++
	s.internalMetrics.LastReceivedMetricTimestamp = time.Now().Unix()
	s.internalMetrics.TotalContainerMetricsReceived++
	s.internalMetrics.LastReceivedContainerMetricTimestamp = time.Now().Unix()
	s.internalMetricsMutex.Unlock()

	s.containerMetricsMutex.Lock()
	containerMetric := ContainerMetric{
		Origin:           envelope.GetOrigin(),
		Timestamp:        envelope.GetTimestamp(),
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
	s.containerMetricsMutex.Unlock()
}

func (s *Store) expireContainerMetrics() {
	s.containerMetricsMutex.Lock()
	now := time.Now()
	for k, containerMetric := range s.containerMetrics {
		validUntil := time.Unix(containerMetric.Timestamp, 0).Add(s.metricsExpiry)
		if validUntil.Before(now) {
			delete(s.containerMetrics, k)
		}
	}
	s.containerMetricsMutex.Unlock()
}

func (s *Store) addCounterMetric(envelope *events.Envelope) {
	s.internalMetricsMutex.Lock()
	s.internalMetrics.TotalMetricsReceived++
	s.internalMetrics.LastReceivedMetricTimestamp = time.Now().Unix()
	s.internalMetrics.TotalCounterEventsReceived++
	s.internalMetrics.LastReceivedCounterEventTimestamp = time.Now().Unix()
	s.internalMetricsMutex.Unlock()

	s.counterMetricsMutex.Lock()
	counterMetric := CounterMetric{
		Origin:     envelope.GetOrigin(),
		Timestamp:  envelope.GetTimestamp(),
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
	s.counterMetricsMutex.Unlock()
}

func (s *Store) expireCounterMetrics() {
	s.counterMetricsMutex.Lock()
	now := time.Now()
	for k, counterMetric := range s.counterMetrics {
		validUntil := time.Unix(counterMetric.Timestamp, 0).Add(s.metricsExpiry)
		if validUntil.Before(now) {
			delete(s.counterMetrics, k)
		}
	}
	s.counterMetricsMutex.Unlock()
}

func (s *Store) addValueMetric(envelope *events.Envelope) {
	s.internalMetricsMutex.Lock()
	s.internalMetrics.TotalMetricsReceived++
	s.internalMetrics.LastReceivedMetricTimestamp = time.Now().Unix()
	s.internalMetrics.TotalValueMetricsReceived++
	s.internalMetrics.LastReceivedValueMetricTimestamp = time.Now().Unix()
	s.internalMetricsMutex.Unlock()

	s.valueMetricsMutex.Lock()
	valueMetric := ValueMetric{
		Origin:     envelope.GetOrigin(),
		Timestamp:  envelope.GetTimestamp(),
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
	s.valueMetricsMutex.Unlock()
}

func (s *Store) expireValueMetrics() {
	s.valueMetricsMutex.Lock()
	now := time.Now()
	for k, valueMetric := range s.valueMetrics {
		validUntil := time.Unix(valueMetric.Timestamp, 0).Add(s.metricsExpiry)
		if validUntil.Before(now) {
			delete(s.valueMetrics, k)
		}
	}
	s.valueMetricsMutex.Unlock()
}
