package rollup

import (
	"github.com/bosh-prometheus/firehose_exporter/metricmaker"
	"github.com/bosh-prometheus/firehose_exporter/metrics"
	"github.com/bosh-prometheus/firehose_exporter/transform"
	"github.com/gogo/protobuf/proto"
	"sync"
	"time"
)

type counterRollup struct {
	nodeIndex          string
	rollupTags         []string
	countersInInterval *sync.Map
	counters           *sync.Map
	keyCleaningTime    *sync.Map

	metricExpireIn        time.Duration
	cleanPeriodicDuration time.Duration
}

type CounterOpt func(r *counterRollup)

func SetCounterCleaning(metricExpireIn time.Duration, cleanPeriodicDuration time.Duration) CounterOpt {
	return func(r *counterRollup) {
		r.metricExpireIn = metricExpireIn
		r.cleanPeriodicDuration = cleanPeriodicDuration
	}
}

func NewCounterRollup(nodeIndex string, rollupTags []string, opts ...CounterOpt) *counterRollup {
	cr := &counterRollup{
		nodeIndex:             nodeIndex,
		rollupTags:            rollupTags,
		countersInInterval:    &sync.Map{},
		counters:              &sync.Map{},
		metricExpireIn:        2 * time.Hour,
		cleanPeriodicDuration: 10 * time.Minute,
		keyCleaningTime:       &sync.Map{},
	}
	for _, opt := range opts {
		opt(cr)
	}
	go cr.CleanPeriodic()
	return cr
}

func (r *counterRollup) CleanPeriodic() {
	for {
		time.Sleep(r.cleanPeriodicDuration)
		now := time.Now()
		toDelete := make([]string, 0)
		r.keyCleaningTime.Range(func(key, value interface{}) bool {
			t := value.(time.Time)
			if t.Add(r.metricExpireIn).Before(now) {
				toDelete = append(toDelete, key.(string))
			}
			return true
		})
		for _, key := range toDelete {
			r.keyCleaningTime.Delete(key)
			r.counters.Delete(key)
			r.countersInInterval.Delete(key)
		}
	}
}

func (r *counterRollup) Record(sourceId string, tags map[string]string, value int64) {
	key := keyFromTags(r.rollupTags, sourceId, tags)

	r.countersInInterval.Store(key, struct{}{})

	previousValue, ok := r.counters.Load(key)
	if ok {
		value = previousValue.(int64) + value
	}
	r.counters.Store(key, value)
	r.keyCleaningTime.Store(key, time.Now())
}

func (r *counterRollup) Rollup(timestamp int64) []*PointsBatch {
	var batches []*PointsBatch

	r.countersInInterval.Range(func(k, _ interface{}) bool {
		labels, err := labelsFromKey(k.(string), r.nodeIndex, r.rollupTags)
		if err != nil {
			return true
		}
		if _, ok := labels["app_id"]; ok {
			labels["origin"] = "cf_app"
		}
		value, _ := r.counters.Load(k)
		metric := metricmaker.NewRawMetricCounter(metrics.GorouterHttpCounterMetricName, labels, float64(value.(int64)))
		metric.Metric().TimestampMs = proto.Int64(transform.NanosecondsToMilliseconds(timestamp))
		batches = append(batches, &PointsBatch{
			Points: []*metrics.RawMetric{metric},
			Size:   metric.EstimateMetricSize(),
		})
		return true
	})
	cleanSyncMap(r.countersInInterval)

	return batches
}
