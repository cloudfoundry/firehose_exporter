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

const ()

type summaryRollup struct {
	nodeIndex           string
	rollupTags          []string
	summariesInInterval map[string]struct{}
	summaries           map[string]prometheus.Summary

	mu sync.Mutex
}

func NewSummaryRollup(nodeIndex string, rollupTags []string) *summaryRollup {
	return &summaryRollup{
		nodeIndex:           nodeIndex,
		rollupTags:          rollupTags,
		summariesInInterval: make(map[string]struct{}),
		summaries:           make(map[string]prometheus.Summary),
	}
}

func (r *summaryRollup) Record(sourceId string, tags map[string]string, value int64) {
	key := keyFromTags(r.rollupTags, sourceId, tags)

	r.mu.Lock()
	defer r.mu.Unlock()

	_, found := r.summaries[key]
	if !found {
		r.summaries[key] = prometheus.NewSummary(prometheus.SummaryOpts{
			Name:       metrics.GorouterHttpSummaryMetricName,
			Objectives: map[float64]float64{0.2: 0.05, 0.5: 0.05, 0.75: 0.02, 0.95: 0.01},
		})
	}

	r.summaries[key].Observe(float64(value))
	r.summariesInInterval[key] = struct{}{}
}

func (r *summaryRollup) Rollup(timestamp int64) []*PointsBatch {
	var batches []*PointsBatch

	r.mu.Lock()
	defer r.mu.Unlock()

	for k := range r.summariesInInterval {
		labels, err := labelsFromKey(k, r.nodeIndex, r.rollupTags)
		if err != nil {
			continue
		}
		if _, ok := labels["app_id"]; ok {
			labels["origin"] = "cf_app"
		}
		m := &dto.Metric{}
		_ = r.summaries[k].Write(m)
		m.Label = transform.LabelsMapToLabelPairs(labels)

		metric := metricmaker.NewRawMetricFromMetric(metrics.GorouterHttpSummaryMetricName, m)
		metric.Metric().TimestampMs = proto.Int64(transform.NanosecondsToMilliseconds(timestamp))
		batches = append(batches, &PointsBatch{
			Points: []*metrics.RawMetric{metric},
			Size:   metric.EstimateMetricSize(),
		})
	}

	r.summariesInInterval = make(map[string]struct{})

	return batches
}
