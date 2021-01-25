package metrics

import (
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/gogo/protobuf/proto"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/model"
)

var separatorByteSlice = []byte{model.SeparatorByte}

type RawMetric struct {
	metric     *dto.Metric
	metricName string
	origin     string
	metricType *dto.MetricType
	id         uint64
	help       string
	expireAt   time.Time
	swept      bool
}

func NewRawMetric(metricName string, origin string, metric *dto.Metric) *RawMetric {
	metricType := dto.MetricType_COUNTER

	if metric.Histogram != nil {
		metricType = dto.MetricType_HISTOGRAM
	}
	if metric.Untyped != nil {
		metricType = dto.MetricType_UNTYPED
	}
	if metric.Summary != nil {
		metricType = dto.MetricType_SUMMARY
	}
	if metric.Gauge != nil {
		metricType = dto.MetricType_GAUGE
	}

	return &RawMetric{
		metricName: metricName,
		origin:     origin,
		metric:     metric,
		metricType: &metricType,
	}
}

func (r *RawMetric) MetricName() string {
	return r.metricName
}

func (r *RawMetric) Metric() *dto.Metric {
	return r.metric
}

func (r *RawMetric) Origin() string {
	return r.origin
}

func (r *RawMetric) MetricType() *dto.MetricType {
	return r.metricType
}

func (r *RawMetric) Help() string {
	return r.help
}

func (r *RawMetric) SetMetricName(metricName string) {
	r.metricName = metricName
}

func (r *RawMetric) SetHelp(help string) {
	r.help = help
}

func (r *RawMetric) SetOrigin(origin string) {
	r.origin = origin

	for _, label := range r.metric.Label {
		if label.GetName() == "origin" {
			label.Value = proto.String(origin)
			return
		}
	}
}

func (r *RawMetric) IsSwept() bool {
	if r.expireAt.IsZero() || r.swept {
		return r.swept
	}

	if r.expireAt.Before(time.Now()) {
		r.swept = true
	}
	return r.swept
}

func (r *RawMetric) SetSweep(sweep bool) {
	r.swept = sweep
}

func (r *RawMetric) ExpireIn(dur time.Duration) {
	r.expireAt = time.Now().Add(dur)
}

func (r *RawMetric) EstimateMetricSize() (size int) {
	// 8 bytes for value (float64)
	size += 8

	if r.metric.TimestampMs != nil {
		// 8 bytes for timestamp (int64)
		size += 8
	}

	// add the size of all label keys and values
	for _, label := range r.metric.Label {
		size += (len(label.GetName()) + len(label.GetValue()))
	}

	return size
}

func (r *RawMetric) Id() uint64 {
	if r.id != 0 {
		return r.id
	}
	labels := r.metric.GetLabel()
	xxh := xxhash.New()
	for _, label := range labels {
		if label.GetName() == model.MetricNameLabel {
			continue
		}
		xxh.WriteString("$" + label.GetName() + "$" + label.GetValue())
		xxh.Write(separatorByteSlice)
	}
	r.id = xxh.Sum64()
	return r.id
}
