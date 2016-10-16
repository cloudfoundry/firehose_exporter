package metrics

import (
	"strconv"
	"time"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/patrickmn/go-cache"
)

type Store struct {
	metricsExpiration      time.Duration
	metricsCleanupInterval time.Duration
	internalMetrics        *cache.Cache
	containerMetrics       *cache.Cache
	counterEvents          *cache.Cache
	valueMetrics           *cache.Cache
}

func NewStore(
	metricsExpiration time.Duration,
	metricsCleanupInterval time.Duration,
) *Store {
	internalMetrics := cache.New(metricsExpiration, metricsCleanupInterval)
	containerMetrics := cache.New(metricsExpiration, metricsCleanupInterval)
	counterEvents := cache.New(metricsExpiration, metricsCleanupInterval)
	valueMetrics := cache.New(metricsExpiration, metricsCleanupInterval)

	store := &Store{
		metricsExpiration:      metricsExpiration,
		metricsCleanupInterval: metricsCleanupInterval,
		internalMetrics:        internalMetrics,
		containerMetrics:       containerMetrics,
		counterEvents:          counterEvents,
		valueMetrics:           valueMetrics,
	}
	store.SetInternalMetrics(InternalMetrics{})

	return store
}

func (s *Store) GetInternalMetrics() InternalMetrics {
	internalMetrics := InternalMetrics{}

	if totalEnvelopesReceived, ok := s.internalMetrics.Get(TotalEnvelopesReceivedKey); ok {
		internalMetrics.TotalEnvelopesReceived = totalEnvelopesReceived.(int64)
	}
	if lastEnvelopReceivedTimestamp, ok := s.internalMetrics.Get(LastEnvelopReceivedTimestampKey); ok {
		internalMetrics.LastEnvelopReceivedTimestamp = lastEnvelopReceivedTimestamp.(int64)
	}

	if totalMetricsReceived, ok := s.internalMetrics.Get(TotalMetricsReceivedKey); ok {
		internalMetrics.TotalMetricsReceived = totalMetricsReceived.(int64)
	}
	if lastMetricReceivedTimestamp, ok := s.internalMetrics.Get(LastMetricReceivedTimestampKey); ok {
		internalMetrics.LastMetricReceivedTimestamp = lastMetricReceivedTimestamp.(int64)
	}

	if totalContainerMetricsReceived, ok := s.internalMetrics.Get(TotalContainerMetricsReceivedKey); ok {
		internalMetrics.TotalContainerMetricsReceived = totalContainerMetricsReceived.(int64)
	}
	if lastContainerMetricReceivedTimestamp, ok := s.internalMetrics.Get(LastContainerMetricReceivedTimestampKey); ok {
		internalMetrics.LastContainerMetricReceivedTimestamp = lastContainerMetricReceivedTimestamp.(int64)
	}

	if totalCounterEventsReceived, ok := s.internalMetrics.Get(TotalCounterEventsReceivedKey); ok {
		internalMetrics.TotalCounterEventsReceived = totalCounterEventsReceived.(int64)
	}
	if lastCounterEventReceivedTimestamp, ok := s.internalMetrics.Get(LastCounterEventReceivedTimestampKey); ok {
		internalMetrics.LastCounterEventReceivedTimestamp = lastCounterEventReceivedTimestamp.(int64)
	}

	if totalValueMetricsReceived, ok := s.internalMetrics.Get(TotalValueMetricsReceivedKey); ok {
		internalMetrics.TotalValueMetricsReceived = totalValueMetricsReceived.(int64)
	}
	if lastValueMetricReceivedTimestamp, ok := s.internalMetrics.Get(LastValueMetricReceivedTimestampKey); ok {
		internalMetrics.LastValueMetricReceivedTimestamp = lastValueMetricReceivedTimestamp.(int64)
	}

	if slowConsumerAlert, ok := s.internalMetrics.Get(SlowConsumerAlertKey); ok {
		internalMetrics.SlowConsumerAlert = slowConsumerAlert.(bool)
	} else {
		internalMetrics.SlowConsumerAlert = false
	}
	if lastSlowConsumerAlertTimestamp, ok := s.internalMetrics.Get(LastSlowConsumerAlertTimestampKey); ok {
		internalMetrics.LastSlowConsumerAlertTimestamp = lastSlowConsumerAlertTimestamp.(int64)
	}

	return internalMetrics
}

func (s *Store) SetInternalMetrics(internalMetrics InternalMetrics) {
	s.internalMetrics.Set(TotalEnvelopesReceivedKey, int64(internalMetrics.TotalEnvelopesReceived), cache.NoExpiration)
	s.internalMetrics.Set(LastEnvelopReceivedTimestampKey, int64(internalMetrics.LastEnvelopReceivedTimestamp), cache.NoExpiration)
	s.internalMetrics.Set(TotalMetricsReceivedKey, int64(internalMetrics.TotalMetricsReceived), cache.NoExpiration)
	s.internalMetrics.Set(LastMetricReceivedTimestampKey, int64(internalMetrics.LastMetricReceivedTimestamp), cache.NoExpiration)
	s.internalMetrics.Set(TotalContainerMetricsReceivedKey, int64(internalMetrics.TotalContainerMetricsReceived), cache.NoExpiration)
	s.internalMetrics.Set(LastContainerMetricReceivedTimestampKey, int64(internalMetrics.LastContainerMetricReceivedTimestamp), cache.NoExpiration)
	s.internalMetrics.Set(TotalCounterEventsReceivedKey, int64(internalMetrics.TotalCounterEventsReceived), cache.NoExpiration)
	s.internalMetrics.Set(LastCounterEventReceivedTimestampKey, int64(internalMetrics.LastCounterEventReceivedTimestamp), cache.NoExpiration)
	s.internalMetrics.Set(TotalValueMetricsReceivedKey, int64(internalMetrics.TotalValueMetricsReceived), cache.NoExpiration)
	s.internalMetrics.Set(LastValueMetricReceivedTimestampKey, int64(internalMetrics.LastValueMetricReceivedTimestamp), cache.NoExpiration)
	s.internalMetrics.Set(SlowConsumerAlertKey, internalMetrics.SlowConsumerAlert, cache.DefaultExpiration)
	s.internalMetrics.Set(LastSlowConsumerAlertTimestampKey, int64(internalMetrics.LastSlowConsumerAlertTimestamp), cache.NoExpiration)
}

func (s *Store) GetContainerMetrics() ContainerMetrics {
	containerMetrics := ContainerMetrics{}
	for _, containerMetric := range s.containerMetrics.Items() {
		if !containerMetric.Expired() {
			containerMetrics = append(containerMetrics, containerMetric.Object.(ContainerMetric))
		}
	}
	return containerMetrics
}

func (s *Store) FlushContainerMetrics() {
	s.containerMetrics.Flush()
}

func (s *Store) GetCounterEvents() CounterEvents {
	counterEvents := CounterEvents{}
	for _, counterEvent := range s.counterEvents.Items() {
		if !counterEvent.Expired() {
			counterEvents = append(counterEvents, counterEvent.Object.(CounterEvent))
		}
	}
	return counterEvents
}

func (s *Store) FlushCounterEvents() {
	s.counterEvents.Flush()
}

func (s *Store) GetValueMetrics() ValueMetrics {
	valueMetrics := ValueMetrics{}
	for _, valueMetric := range s.valueMetrics.Items() {
		if !valueMetric.Expired() {
			valueMetrics = append(valueMetrics, valueMetric.Object.(ValueMetric))
		}
	}
	return valueMetrics
}

func (s *Store) FlushValueMetrics() {
	s.valueMetrics.Flush()
}

func (s *Store) AlertSlowConsumerError() {
	s.internalMetrics.Set(SlowConsumerAlertKey, true, cache.DefaultExpiration)
	s.internalMetrics.Set(LastSlowConsumerAlertTimestampKey, time.Now().UnixNano(), cache.DefaultExpiration)
}

func (s *Store) AddMetric(envelope *events.Envelope) {
	s.internalMetrics.IncrementInt64(TotalEnvelopesReceivedKey, 1)
	s.internalMetrics.Set(LastEnvelopReceivedTimestampKey, time.Now().UnixNano(), cache.DefaultExpiration)

	switch envelope.GetEventType() {
	case events.Envelope_ContainerMetric:
		s.addContainerMetric(envelope)
	case events.Envelope_CounterEvent:
		s.addCounterEvent(envelope)
	case events.Envelope_ValueMetric:
		s.addValueMetric(envelope)
	}
}

func (s *Store) addContainerMetric(envelope *events.Envelope) {
	s.internalMetrics.IncrementInt64(TotalMetricsReceivedKey, 1)
	s.internalMetrics.Set(LastMetricReceivedTimestampKey, time.Now().UnixNano(), cache.DefaultExpiration)
	s.internalMetrics.IncrementInt64(TotalContainerMetricsReceivedKey, 1)
	s.internalMetrics.Set(LastContainerMetricReceivedTimestampKey, time.Now().UnixNano(), cache.DefaultExpiration)

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
	containerMetricKey := envelope.GetContainerMetric().GetApplicationId() + strconv.Itoa(int(containerMetric.InstanceIndex))
	s.containerMetrics.Set(containerMetricKey, containerMetric, cache.DefaultExpiration)
}

func (s *Store) addCounterEvent(envelope *events.Envelope) {
	s.internalMetrics.IncrementInt64(TotalMetricsReceivedKey, 1)
	s.internalMetrics.Set(LastMetricReceivedTimestampKey, time.Now().UnixNano(), cache.DefaultExpiration)
	s.internalMetrics.IncrementInt64(TotalCounterEventsReceivedKey, 1)
	s.internalMetrics.Set(LastCounterEventReceivedTimestampKey, time.Now().UnixNano(), cache.DefaultExpiration)

	counterEvent := CounterEvent{
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
	counterEventKey := envelope.GetOrigin() + envelope.GetCounterEvent().GetName()
	s.counterEvents.Set(counterEventKey, counterEvent, cache.DefaultExpiration)
}

func (s *Store) addValueMetric(envelope *events.Envelope) {
	s.internalMetrics.IncrementInt64(TotalMetricsReceivedKey, 1)
	s.internalMetrics.Set(LastMetricReceivedTimestampKey, time.Now().UnixNano(), cache.DefaultExpiration)
	s.internalMetrics.IncrementInt64(TotalValueMetricsReceivedKey, 1)
	s.internalMetrics.Set(LastValueMetricReceivedTimestampKey, time.Now().UnixNano(), cache.DefaultExpiration)

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
	valueMetricKey := envelope.GetOrigin() + envelope.GetValueMetric().GetName()
	s.valueMetrics.Set(valueMetricKey, valueMetric, cache.DefaultExpiration)
}
