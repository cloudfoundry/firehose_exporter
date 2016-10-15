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
	counterMetrics         *cache.Cache
	valueMetrics           *cache.Cache
}

func NewStore(
	metricsExpiration time.Duration,
	metricsCleanupInterval time.Duration,
) *Store {
	internalMetrics := cache.New(metricsExpiration, metricsCleanupInterval)
	internalMetrics.Set(TotalEnvelopesReceivedKey, int64(0), cache.NoExpiration)
	internalMetrics.Set(LastEnvelopReceivedTimestampKey, int64(0), cache.NoExpiration)
	internalMetrics.Set(TotalMetricsReceivedKey, int64(0), cache.NoExpiration)
	internalMetrics.Set(LastMetricReceivedTimestampKey, int64(0), cache.NoExpiration)
	internalMetrics.Set(TotalContainerMetricsReceivedKey, int64(0), cache.NoExpiration)
	internalMetrics.Set(LastContainerMetricReceivedTimestampKey, int64(0), cache.NoExpiration)
	internalMetrics.Set(TotalCounterEventsReceivedKey, int64(0), cache.NoExpiration)
	internalMetrics.Set(LastCounterEventReceivedTimestampKey, int64(0), cache.NoExpiration)
	internalMetrics.Set(TotalValueMetricsReceivedKey, int64(0), cache.NoExpiration)
	internalMetrics.Set(LastValueMetricReceivedTimestampKey, int64(0), cache.NoExpiration)
	internalMetrics.Set(SlowConsumerAlertKey, false, cache.DefaultExpiration)
	internalMetrics.Set(LastSlowConsumerAlertTimestampKey, int64(0), cache.NoExpiration)

	containerMetrics := cache.New(metricsExpiration, metricsCleanupInterval)
	counterMetrics := cache.New(metricsExpiration, metricsCleanupInterval)
	valueMetrics := cache.New(metricsExpiration, metricsCleanupInterval)

	return &Store{
		metricsExpiration:      metricsExpiration,
		metricsCleanupInterval: metricsCleanupInterval,
		internalMetrics:        internalMetrics,
		containerMetrics:       containerMetrics,
		counterMetrics:         counterMetrics,
		valueMetrics:           valueMetrics,
	}
}

func (s *Store) GetInternalMetrics() InternalMetrics {
	internalMetrics := InternalMetrics{}

	totalEnvelopesReceived, ok := s.internalMetrics.Get(TotalEnvelopesReceivedKey)
	if ok {
		internalMetrics.TotalEnvelopesReceived = totalEnvelopesReceived.(int64)
	}
	lastEnvelopReceivedTimestamp, ok := s.internalMetrics.Get(LastEnvelopReceivedTimestampKey)
	if ok {
		internalMetrics.LastEnvelopReceivedTimestamp = lastEnvelopReceivedTimestamp.(int64)
	}

	totalMetricsReceived, ok := s.internalMetrics.Get(TotalMetricsReceivedKey)
	if ok {
		internalMetrics.TotalMetricsReceived = totalMetricsReceived.(int64)
	}
	lastMetricReceivedTimestamp, ok := s.internalMetrics.Get(LastMetricReceivedTimestampKey)
	if ok {
		internalMetrics.LastMetricReceivedTimestamp = lastMetricReceivedTimestamp.(int64)
	}

	totalContainerMetricsReceived, ok := s.internalMetrics.Get(TotalContainerMetricsReceivedKey)
	if ok {
		internalMetrics.TotalContainerMetricsReceived = totalContainerMetricsReceived.(int64)
	}
	lastContainerMetricReceivedTimestamp, ok := s.internalMetrics.Get(LastContainerMetricReceivedTimestampKey)
	if ok {
		internalMetrics.LastContainerMetricReceivedTimestamp = lastContainerMetricReceivedTimestamp.(int64)
	}

	totalCounterEventsReceived, ok := s.internalMetrics.Get(TotalCounterEventsReceivedKey)
	if ok {
		internalMetrics.TotalCounterEventsReceived = totalCounterEventsReceived.(int64)
	}
	lastCounterEventReceivedTimestamp, ok := s.internalMetrics.Get(LastCounterEventReceivedTimestampKey)
	if ok {
		internalMetrics.LastCounterEventReceivedTimestamp = lastCounterEventReceivedTimestamp.(int64)
	}

	totalValueMetricsReceived, ok := s.internalMetrics.Get(TotalValueMetricsReceivedKey)
	if ok {
		internalMetrics.TotalValueMetricsReceived = totalValueMetricsReceived.(int64)
	}
	lastValueMetricReceivedTimestamp, ok := s.internalMetrics.Get(LastValueMetricReceivedTimestampKey)
	if ok {
		internalMetrics.LastValueMetricReceivedTimestamp = lastValueMetricReceivedTimestamp.(int64)
	}

	slowConsumerAlert, ok := s.internalMetrics.Get(SlowConsumerAlertKey)
	if ok {
		internalMetrics.SlowConsumerAlert = slowConsumerAlert.(bool)
	} else {
		internalMetrics.SlowConsumerAlert = false
	}
	lastSlowConsumerAlertTimestamp, ok := s.internalMetrics.Get(LastSlowConsumerAlertTimestampKey)
	if ok {
		internalMetrics.LastSlowConsumerAlertTimestamp = lastSlowConsumerAlertTimestamp.(int64)
	}

	return internalMetrics
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

func (s *Store) GetCounterMetrics() CounterMetrics {
	counterMetrics := CounterMetrics{}
	for _, counterMetric := range s.counterMetrics.Items() {
		if !counterMetric.Expired() {
			counterMetrics = append(counterMetrics, counterMetric.Object.(CounterMetric))
		}
	}
	return counterMetrics
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
		s.addCounterMetric(envelope)
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

func (s *Store) addCounterMetric(envelope *events.Envelope) {
	s.internalMetrics.IncrementInt64(TotalMetricsReceivedKey, 1)
	s.internalMetrics.Set(LastMetricReceivedTimestampKey, time.Now().UnixNano(), cache.DefaultExpiration)
	s.internalMetrics.IncrementInt64(TotalCounterEventsReceivedKey, 1)
	s.internalMetrics.Set(LastCounterEventReceivedTimestampKey, time.Now().UnixNano(), cache.DefaultExpiration)

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
	counterMetricKey := envelope.GetOrigin() + envelope.GetCounterEvent().GetName()
	s.counterMetrics.Set(counterMetricKey, counterMetric, cache.DefaultExpiration)
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
