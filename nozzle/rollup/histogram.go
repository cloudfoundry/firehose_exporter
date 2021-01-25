package rollup

import (
	"sync"

	"github.com/bosh-prometheus/firehose_exporter/metricmaker"
	"github.com/bosh-prometheus/firehose_exporter/metrics"
	"github.com/bosh-prometheus/firehose_exporter/transform"
	"github.com/gogo/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

type histogramRollup struct {
	nodeIndex            string
	rollupTags           []string
	histogramsInInterval map[string]struct{}
	histograms           map[string]prometheus.Histogram

	mu sync.Mutex
}

func NewHistogramRollup(nodeIndex string, rollupTags []string) *histogramRollup {
	return &histogramRollup{
		nodeIndex:            nodeIndex,
		rollupTags:           rollupTags,
		histogramsInInterval: make(map[string]struct{}),
		histograms:           make(map[string]prometheus.Histogram),
	}
}

func (r *histogramRollup) Record(sourceId string, tags map[string]string, value int64) {
	key := keyFromTags(r.rollupTags, sourceId, tags)

	r.mu.Lock()
	defer r.mu.Unlock()

	_, found := r.histograms[key]
	if !found {
		r.histograms[key] = prometheus.NewHistogram(prometheus.HistogramOpts{
			Name: metrics.GorouterHttpHistogramMetricName,
		})
	}

	r.histograms[key].Observe(transform.NanosecondsToSeconds(value))
	r.histogramsInInterval[key] = struct{}{}
}

func (r *histogramRollup) Rollup(timestamp int64) []*PointsBatch {
	var batches []*PointsBatch

	r.mu.Lock()
	defer r.mu.Unlock()

	for k := range r.histogramsInInterval {

		labels, err := labelsFromKey(k, r.nodeIndex, r.rollupTags)
		if err != nil {
			continue
		}
		if _, ok := labels["app_id"]; ok {
			labels["origin"] = "cf_app"
		}
		m := &dto.Metric{}
		_ = r.histograms[k].Write(m)
		m.Label = transform.LabelsMapToLabelPairs(labels)

		metric := metricmaker.NewRawMetricFromMetric(metrics.GorouterHttpHistogramMetricName, m)
		metric.Metric().TimestampMs = proto.Int64(transform.NanosecondsToMilliseconds(timestamp))
		batches = append(batches, &PointsBatch{
			Points: []*metrics.RawMetric{metric},
			Size:   metric.EstimateMetricSize(),
		})
	}

	r.histogramsInInterval = make(map[string]struct{})

	return batches
}
