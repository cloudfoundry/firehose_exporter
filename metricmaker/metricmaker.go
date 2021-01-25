package metricmaker

import (
	"sort"
	"strings"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/bosh-prometheus/firehose_exporter/metrics"
	"github.com/bosh-prometheus/firehose_exporter/transform"
	"github.com/gogo/protobuf/proto"
	dto "github.com/prometheus/client_model/go"
)

type MetricConverter func(metric *metrics.RawMetric)

var metricConverters = []MetricConverter{
	NormalizeName,
	OrderAndSanitizeLabels,
	PresetLabels,
}

var enableEnvelopCounterDelta = false

func SetEnableEnvelopCounterDelta(enable bool) {
	enableEnvelopCounterDelta = enable
}

func PrependMetricConverter(metricConverter MetricConverter) {
	metricConverters = append([]MetricConverter{metricConverter}, metricConverters...)
}

func SetMetricConverters(newMetricConverters []MetricConverter) {
	metricConverters = newMetricConverters
}

func applyConverters(metric *metrics.RawMetric) {
	for _, metricConverter := range metricConverters {
		metricConverter(metric)
	}
}

func NewRawMetricFromMetric(metricName string, metric *dto.Metric) *metrics.RawMetric {
	m := metrics.NewRawMetric(metricName, getOriginFromMetric(metric), metric)
	applyConverters(m)
	return m
}

func NewRawMetricCounter(metricName string, labelsMap map[string]string, value float64) *metrics.RawMetric {
	origin := ""
	if val, ok := labelsMap["origin"]; ok {
		origin = val
	}
	labels := transform.LabelsMapToLabelPairs(labelsMap)
	metric := &dto.Metric{
		Label: labels,
		Counter: &dto.Counter{
			Value: proto.Float64(value),
		},
	}
	m := metrics.NewRawMetric(metricName, origin, metric)
	applyConverters(m)
	return m
}

func NewRawMetricGauge(metricName string, labelsMap map[string]string, value float64) *metrics.RawMetric {
	origin := ""
	if val, ok := labelsMap["origin"]; ok {
		origin = val
	}
	labels := transform.LabelsMapToLabelPairs(labelsMap)
	metric := &dto.Metric{
		Label: labels,
		Gauge: &dto.Gauge{
			Value: proto.Float64(value),
		},
	}
	m := metrics.NewRawMetric(metricName, origin, metric)
	applyConverters(m)
	return m
}

func NewRawMetricsFromEnvelop(envelope *loggregator_v2.Envelope) []*metrics.RawMetric {
	switch envelope.Message.(type) {
	case *loggregator_v2.Envelope_Gauge:
		return newRawMetricsFromEnvelopGauge(envelope)
	case *loggregator_v2.Envelope_Timer:
		return []*metrics.RawMetric{}
	case *loggregator_v2.Envelope_Counter:
		return newRawMetricFromEnvelopCounter(envelope)
	}
	return []*metrics.RawMetric{}
}

func newRawMetricFromEnvelopCounter(envelope *loggregator_v2.Envelope) []*metrics.RawMetric {
	counter := envelope.GetCounter()
	metricName := counter.GetName()

	metric := prepareMetricFromEnvelop(envelope)
	metric.Counter = &dto.Counter{
		Value: proto.Float64(float64(counter.GetTotal())),
	}
	m := metrics.NewRawMetric(metricName, getOriginFromMetric(metric), metric)
	applyConverters(m)

	finalMetrics := []*metrics.RawMetric{m}

	if enableEnvelopCounterDelta {
		deltaMetric := prepareMetricFromEnvelop(envelope)
		deltaMetric.Counter = &dto.Counter{
			Value: proto.Float64(float64(counter.GetDelta())),
		}
		mDelta := metrics.NewRawMetric(metricName+"_delta", getOriginFromMetric(deltaMetric), deltaMetric)
		applyConverters(mDelta)
		finalMetrics = append(finalMetrics, mDelta)
	}
	return finalMetrics
}

func newRawMetricsFromEnvelopGauge(envelope *loggregator_v2.Envelope) []*metrics.RawMetric {
	var points []*metrics.RawMetric
	gauge := envelope.GetGauge()
	for name, metric := range gauge.GetMetrics() {
		point := prepareMetricFromEnvelop(envelope)
		metricName := name
		point.Label = append(point.Label, &dto.LabelPair{
			Name:  proto.String("unit"),
			Value: proto.String(metric.GetUnit()),
		})
		point.Gauge = &dto.Gauge{
			Value: proto.Float64(metric.GetValue()),
		}

		m := metrics.NewRawMetric(metricName, getOriginFromMetric(point), point)
		applyConverters(m)
		points = append(points, m)
	}
	return points
}

func getOriginFromMetric(metric *dto.Metric) string {
	for _, label := range metric.Label {
		if label.GetName() == "origin" {
			return label.GetValue()
		}
	}
	return ""
}

func prepareMetricFromEnvelop(envelope *loggregator_v2.Envelope) *dto.Metric {
	labels := make([]*dto.LabelPair, 0)
	labelValues := envelope.GetTags()
	if labelValues == nil {
		labelValues = make(map[string]string)
	}
	if _, ok := labelValues["source_id"]; !ok && envelope.GetSourceId() != "" {
		labelValues["source_id"] = envelope.GetSourceId()
	}

	if envelope.GetInstanceId() != "" {
		labelValues["instance_id"] = envelope.GetInstanceId()
	}

	labelKeys := make([]string, 0)
	for k := range labelValues {
		if strings.HasPrefix(k, "__") {
			continue
		}
		labelKeys = append(labelKeys, k)
	}

	sort.Strings(labelKeys)

	for _, k := range labelKeys {
		labels = append(labels, &dto.LabelPair{
			Name:  proto.String(k),
			Value: proto.String(labelValues[k]),
		})
	}
	return &dto.Metric{
		Label: labels,
	}
}
