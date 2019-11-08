package metrics

import (
	"bytes"
	"regexp"
	"strconv"
	"time"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/patrickmn/go-cache"

	"github.com/bosh-prometheus/firehose_exporter/filters"
	"github.com/bosh-prometheus/firehose_exporter/utils"
)

type Store struct {
	metricsExpiration      time.Duration
	metricsCleanupInterval time.Duration
	deploymentFilter       *filters.DeploymentFilter
	eventFilter            *filters.EventFilter
	internalMetrics        *cache.Cache
	containerMetrics       *cache.Cache
	counterEvents          *cache.Cache
	httpStartStops         *cache.Cache
	valueMetrics           *cache.Cache
}

func NewStore(
	metricsExpiration time.Duration,
	metricsCleanupInterval time.Duration,
	deploymentFilter *filters.DeploymentFilter,
	eventFilter *filters.EventFilter,
) *Store {
	internalMetrics := cache.New(metricsExpiration, metricsCleanupInterval)
	containerMetrics := cache.New(metricsExpiration, metricsCleanupInterval)
	counterEvents := cache.New(metricsExpiration, metricsCleanupInterval)
	httpStartStops := cache.New(metricsExpiration, metricsCleanupInterval)
	valueMetrics := cache.New(metricsExpiration, metricsCleanupInterval)

	store := &Store{
		metricsExpiration:      metricsExpiration,
		metricsCleanupInterval: metricsCleanupInterval,
		deploymentFilter:       deploymentFilter,
		eventFilter:            eventFilter,
		internalMetrics:        internalMetrics,
		containerMetrics:       containerMetrics,
		counterEvents:          counterEvents,
		httpStartStops:         httpStartStops,
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
	if totalContainerMetricsProcessed, ok := s.internalMetrics.Get(TotalContainerMetricsProcessedKey); ok {
		internalMetrics.TotalContainerMetricsProcessed = totalContainerMetricsProcessed.(int64)
	}
	internalMetrics.TotalContainerMetricsCached = int64(s.containerMetrics.ItemCount())
	if lastContainerMetricReceivedTimestamp, ok := s.internalMetrics.Get(LastContainerMetricReceivedTimestampKey); ok {
		internalMetrics.LastContainerMetricReceivedTimestamp = lastContainerMetricReceivedTimestamp.(int64)
	}

	if totalCounterEventsReceived, ok := s.internalMetrics.Get(TotalCounterEventsReceivedKey); ok {
		internalMetrics.TotalCounterEventsReceived = totalCounterEventsReceived.(int64)
	}
	if totalCounterEventsProcessed, ok := s.internalMetrics.Get(TotalCounterEventsProcessedKey); ok {
		internalMetrics.TotalCounterEventsProcessed = totalCounterEventsProcessed.(int64)
	}
	internalMetrics.TotalCounterEventsCached = int64(s.counterEvents.ItemCount())
	if lastCounterEventReceivedTimestamp, ok := s.internalMetrics.Get(LastCounterEventReceivedTimestampKey); ok {
		internalMetrics.LastCounterEventReceivedTimestamp = lastCounterEventReceivedTimestamp.(int64)
	}

	if totalHttpStartStopReceived, ok := s.internalMetrics.Get(TotalHttpStartStopReceivedKey); ok {
		internalMetrics.TotalHttpStartStopReceived = totalHttpStartStopReceived.(int64)
	}
	if totalHttpStartStopProcessed, ok := s.internalMetrics.Get(TotalHttpStartStopProcessedKey); ok {
		internalMetrics.TotalHttpStartStopProcessed = totalHttpStartStopProcessed.(int64)
	}
	internalMetrics.TotalHttpStartStopCached = int64(s.httpStartStops.ItemCount())
	if lastHttpStartStopReceivedTimestamp, ok := s.internalMetrics.Get(LastHttpStartStopReceivedTimestampKey); ok {
		internalMetrics.LastHttpStartStopReceivedTimestamp = lastHttpStartStopReceivedTimestamp.(int64)
	}

	if totalValueMetricsReceived, ok := s.internalMetrics.Get(TotalValueMetricsReceivedKey); ok {
		internalMetrics.TotalValueMetricsReceived = totalValueMetricsReceived.(int64)
	}
	if totalValueMetricsProcessed, ok := s.internalMetrics.Get(TotalValueMetricsProcessedKey); ok {
		internalMetrics.TotalValueMetricsProcessed = totalValueMetricsProcessed.(int64)
	}
	internalMetrics.TotalValueMetricsCached = int64(s.valueMetrics.ItemCount())
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
	s.internalMetrics.Set(TotalContainerMetricsProcessedKey, int64(internalMetrics.TotalContainerMetricsProcessed), cache.NoExpiration)
	s.internalMetrics.Set(LastContainerMetricReceivedTimestampKey, int64(internalMetrics.LastContainerMetricReceivedTimestamp), cache.NoExpiration)
	s.internalMetrics.Set(TotalCounterEventsReceivedKey, int64(internalMetrics.TotalCounterEventsReceived), cache.NoExpiration)
	s.internalMetrics.Set(TotalCounterEventsProcessedKey, int64(internalMetrics.TotalCounterEventsProcessed), cache.NoExpiration)
	s.internalMetrics.Set(LastCounterEventReceivedTimestampKey, int64(internalMetrics.LastCounterEventReceivedTimestamp), cache.NoExpiration)
	s.internalMetrics.Set(TotalHttpStartStopReceivedKey, int64(internalMetrics.TotalHttpStartStopReceived), cache.NoExpiration)
	s.internalMetrics.Set(TotalHttpStartStopProcessedKey, int64(internalMetrics.TotalHttpStartStopProcessed), cache.NoExpiration)
	s.internalMetrics.Set(LastHttpStartStopReceivedTimestampKey, int64(internalMetrics.LastHttpStartStopReceivedTimestamp), cache.NoExpiration)
	s.internalMetrics.Set(TotalValueMetricsReceivedKey, int64(internalMetrics.TotalValueMetricsReceived), cache.NoExpiration)
	s.internalMetrics.Set(TotalValueMetricsProcessedKey, int64(internalMetrics.TotalValueMetricsProcessed), cache.NoExpiration)
	s.internalMetrics.Set(LastValueMetricReceivedTimestampKey, int64(internalMetrics.LastValueMetricReceivedTimestamp), cache.NoExpiration)
	s.internalMetrics.Set(SlowConsumerAlertKey, internalMetrics.SlowConsumerAlert, cache.DefaultExpiration)
	s.internalMetrics.Set(LastSlowConsumerAlertTimestampKey, int64(internalMetrics.LastSlowConsumerAlertTimestamp), cache.NoExpiration)
}

func (s *Store) AlertSlowConsumerError() {
	s.internalMetrics.Set(SlowConsumerAlertKey, true, cache.DefaultExpiration)
	s.internalMetrics.Set(LastSlowConsumerAlertTimestampKey, time.Now().Unix(), cache.NoExpiration)
}

func (s *Store) AddMetric(envelope *events.Envelope) {
	s.internalMetrics.IncrementInt64(TotalEnvelopesReceivedKey, 1)
	s.internalMetrics.Set(LastEnvelopReceivedTimestampKey, time.Now().Unix(), cache.NoExpiration)

	switch envelope.GetEventType() {
	case events.Envelope_ContainerMetric:
		s.addContainerMetric(envelope)
	case events.Envelope_CounterEvent:
		s.addCounterEvent(envelope)
	case events.Envelope_HttpStartStop:
		s.addHttpStartStop(envelope)
	case events.Envelope_ValueMetric:
		s.addValueMetric(envelope)
	}
}

func (s *Store) GetContainerMetrics() ContainerMetrics {
	containerMetrics := ContainerMetrics{}
	for _, containerMetric := range s.containerMetrics.Items() {
		if !containerMetric.Expired() {
			containerMetrics = append(containerMetrics, containerMetric.Object.(*ContainerMetric))
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
			counterEvents = append(counterEvents, counterEvent.Object.(*CounterEvent))
		}
	}
	return counterEvents
}

func (s *Store) FlushCounterEvents() {
	s.counterEvents.Flush()
}

func (s *Store) GetHttpStartStops() HttpStartStops {
	httpStartStops := HttpStartStops{}
	for _, httpStartStop := range s.httpStartStops.Items() {
		if !httpStartStop.Expired() {
			httpStartStops = append(httpStartStops, httpStartStop.Object.(*HttpStartStop))
		}
	}
	return httpStartStops
}

func (s *Store) FlushHttpStartStops() {
	s.httpStartStops.Flush()
}

func (s *Store) GetValueMetrics() ValueMetrics {
	valueMetrics := ValueMetrics{}
	for _, valueMetric := range s.valueMetrics.Items() {
		if !valueMetric.Expired() {
			valueMetrics = append(valueMetrics, valueMetric.Object.(*ValueMetric))
		}
	}
	return valueMetrics
}

func (s *Store) FlushValueMetrics() {
	s.valueMetrics.Flush()
}

func (s *Store) addContainerMetric(envelope *events.Envelope) {
	s.internalMetrics.IncrementInt64(TotalMetricsReceivedKey, 1)
	s.internalMetrics.Set(LastMetricReceivedTimestampKey, time.Now().Unix(), cache.NoExpiration)
	s.internalMetrics.IncrementInt64(TotalContainerMetricsReceivedKey, 1)
	s.internalMetrics.Set(LastContainerMetricReceivedTimestampKey, time.Now().Unix(), cache.NoExpiration)

	if s.deploymentFilter.Enabled(envelope.GetDeployment()) && s.eventFilter.Enabled(envelope) {
		s.internalMetrics.IncrementInt64(TotalContainerMetricsProcessedKey, 1)

		containerMetric := &ContainerMetric{
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
		s.containerMetrics.Set(s.metricKey(envelope), containerMetric, cache.DefaultExpiration)
	}
}

func (s *Store) addCounterEvent(envelope *events.Envelope) {
	s.internalMetrics.IncrementInt64(TotalMetricsReceivedKey, 1)
	s.internalMetrics.Set(LastMetricReceivedTimestampKey, time.Now().Unix(), cache.NoExpiration)
	s.internalMetrics.IncrementInt64(TotalCounterEventsReceivedKey, 1)
	s.internalMetrics.Set(LastCounterEventReceivedTimestampKey, time.Now().Unix(), cache.NoExpiration)

	if s.deploymentFilter.Enabled(envelope.GetDeployment()) && s.eventFilter.Enabled(envelope) {
		s.internalMetrics.IncrementInt64(TotalCounterEventsProcessedKey, 1)

		counterEvent := &CounterEvent{
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
		s.counterEvents.Set(s.metricKey(envelope), counterEvent, cache.DefaultExpiration)
	}
}

func (s *Store) addHttpStartStop(envelope *events.Envelope) {
	s.internalMetrics.IncrementInt64(TotalMetricsReceivedKey, 1)
	s.internalMetrics.Set(LastMetricReceivedTimestampKey, time.Now().Unix(), cache.NoExpiration)
	s.internalMetrics.IncrementInt64(TotalHttpStartStopReceivedKey, 1)
	s.internalMetrics.Set(LastHttpStartStopReceivedTimestampKey, time.Now().Unix(), cache.NoExpiration)

	if s.deploymentFilter.Enabled(envelope.GetDeployment()) && s.eventFilter.Enabled(envelope) {
		s.internalMetrics.IncrementInt64(TotalHttpStartStopProcessedKey, 1)

		var httpStartStop *HttpStartStop
		storeHttpStartStop, ok := s.httpStartStops.Get(s.metricKey(envelope))
		if ok {
			httpStartStop = storeHttpStartStop.(*HttpStartStop)
		} else {
			httpStartStop = &HttpStartStop{
				Origin:        envelope.GetOrigin(),
				Timestamp:     envelope.GetTimestamp(),
				Deployment:    envelope.GetDeployment(),
				Job:           envelope.GetJob(),
				Index:         envelope.GetIndex(),
				IP:            envelope.GetIp(),
				Tags:          envelope.GetTags(),
				RequestId:     utils.UUIDToString(envelope.GetHttpStartStop().GetRequestId()),
				Method:        envelope.GetHttpStartStop().GetMethod().String(),
				Uri:           envelope.GetHttpStartStop().GetUri(),
				RemoteAddress: envelope.GetHttpStartStop().GetRemoteAddress(),
				UserAgent:     envelope.GetHttpStartStop().GetUserAgent(),
				StatusCode:    envelope.GetHttpStartStop().GetStatusCode(),
				ContentLength: envelope.GetHttpStartStop().GetContentLength(),
			}
		}

		switch envelope.GetHttpStartStop().GetPeerType() {
		case events.PeerType_Client:
			httpStartStop.ApplicationId = utils.UUIDToString(envelope.GetHttpStartStop().GetApplicationId())
			httpStartStop.InstanceIndex = envelope.GetHttpStartStop().GetInstanceIndex()
			httpStartStop.InstanceId = envelope.GetHttpStartStop().GetInstanceId()
			httpStartStop.ClientStartTimestamp = envelope.GetHttpStartStop().GetStartTimestamp()
			httpStartStop.ClientStopTimestamp = envelope.GetHttpStartStop().GetStopTimestamp()
		case events.PeerType_Server:
			httpStartStop.ServerStartTimestamp = envelope.GetHttpStartStop().GetStartTimestamp()
			httpStartStop.ServerStopTimestamp = envelope.GetHttpStartStop().GetStopTimestamp()
		default:
			return
		}

		s.httpStartStops.Set(s.metricKey(envelope), httpStartStop, cache.DefaultExpiration)
	}
}

func (s *Store) addValueMetric(envelope *events.Envelope) {
	s.internalMetrics.IncrementInt64(TotalMetricsReceivedKey, 1)
	s.internalMetrics.Set(LastMetricReceivedTimestampKey, time.Now().Unix(), cache.NoExpiration)
	s.internalMetrics.IncrementInt64(TotalValueMetricsReceivedKey, 1)
	s.internalMetrics.Set(LastValueMetricReceivedTimestampKey, time.Now().Unix(), cache.NoExpiration)

	if s.deploymentFilter.Enabled(envelope.GetDeployment()) && s.eventFilter.Enabled(envelope) {
		s.internalMetrics.IncrementInt64(TotalValueMetricsProcessedKey, 1)

		valueMetric := &ValueMetric{
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
		s.valueMetrics.Set(s.metricKey(envelope), valueMetric, cache.DefaultExpiration)
	}
}

func (s *Store) metricKey(envelope *events.Envelope) string {
	var buffer bytes.Buffer

	buffer.WriteString(envelope.GetOrigin())
	buffer.WriteString(envelope.GetDeployment())
	buffer.WriteString(envelope.GetJob())
	buffer.WriteString(envelope.GetIndex())
	buffer.WriteString(envelope.GetIp())

	switch envelope.GetEventType() {
	case events.Envelope_ContainerMetric:
		buffer.WriteString(envelope.GetContainerMetric().GetApplicationId())
		buffer.WriteString(strconv.Itoa(int(envelope.GetContainerMetric().GetInstanceIndex())))
	case events.Envelope_CounterEvent:
		buffer.WriteString(generateMetricName(envelope, "CounterEvent"))
	case events.Envelope_HttpStartStop:
		buffer.WriteString(envelope.GetHttpStartStop().GetRequestId().String())
	case events.Envelope_ValueMetric:
		buffer.WriteString(generateMetricName(envelope, "ValueMetric"))
	}

	return buffer.String()
}

// Issue:
// 		https://github.com/bosh-prometheus/firehose_exporter/issues/47
// Solution:
// 1. Check if the metric being emitted comes from the container of the app
// the way we are going to do that is by checking if the origin has a uuid
// which means its from the container
// 2. Once detected that is from the app container, then add a little more extra info
// on the metric name like tags, just to differentiate the metrics and emit all of them
func generateMetricName(envelope *events.Envelope, whichType string) string {
	var name string
	if whichType == "ValueMetric" {
		name = envelope.GetValueMetric().GetName()
	} else {
		name = envelope.GetCounterEvent().GetName()
	}
	uuidRegex := "^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$"
	reg := regexp.MustCompile(uuidRegex)
	isItValidUuid := reg.MatchString(*envelope.Origin)
	if isItValidUuid {
		if envelope.Tags != nil {
			for _, tagValue := range envelope.Tags {
				name = name + utils.NormalizeName(tagValue)
			}
		}
		return name
	}
	return name
}