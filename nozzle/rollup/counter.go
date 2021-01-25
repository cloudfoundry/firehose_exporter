package rollup

import (
	"sync"

	"github.com/bosh-prometheus/firehose_exporter/metricmaker"
	"github.com/bosh-prometheus/firehose_exporter/metrics"
	"github.com/bosh-prometheus/firehose_exporter/transform"
	"github.com/gogo/protobuf/proto"
)

type counterRollup struct {
	nodeIndex          string
	rollupTags         []string
	countersInInterval map[string]struct{}
	counters           map[string]int64

	mu sync.Mutex
}

func NewCounterRollup(nodeIndex string, rollupTags []string) *counterRollup {
	return &counterRollup{
		nodeIndex:          nodeIndex,
		rollupTags:         rollupTags,
		countersInInterval: make(map[string]struct{}),
		counters:           make(map[string]int64),
	}
}

func (r *counterRollup) Record(sourceId string, tags map[string]string, value int64) {
	key := keyFromTags(r.rollupTags, sourceId, tags)

	r.mu.Lock()
	defer r.mu.Unlock()

	r.countersInInterval[key] = struct{}{}
	r.counters[key] += value
}

func (r *counterRollup) Rollup(timestamp int64) []*PointsBatch {
	var batches []*PointsBatch

	r.mu.Lock()
	defer r.mu.Unlock()

	for k := range r.countersInInterval {
		labels, err := labelsFromKey(k, r.nodeIndex, r.rollupTags)
		if err != nil {
			continue
		}
		if _, ok := labels["app_id"]; ok {
			labels["origin"] = "cf_app"
		}
		metric := metricmaker.NewRawMetricCounter(metrics.GorouterHttpCounterMetricName, labels, float64(r.counters[k]))
		metric.Metric().TimestampMs = proto.Int64(transform.NanosecondsToMilliseconds(timestamp))
		batches = append(batches, &PointsBatch{
			Points: []*metrics.RawMetric{metric},
			Size:   metric.EstimateMetricSize(),
		})
	}

	r.countersInInterval = make(map[string]struct{})

	return batches
}
