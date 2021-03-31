package rollup

import (
	"github.com/bosh-prometheus/firehose_exporter/metricmaker"
	"github.com/bosh-prometheus/firehose_exporter/metrics"
	"github.com/bosh-prometheus/firehose_exporter/transform"
	"github.com/gogo/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"sync"
	"time"
)

type summaryRollup struct {
	nodeIndex           string
	rollupTags          []string
	summariesInInterval *sync.Map
	summaries           *sync.Map
	keyCleaningTime     *sync.Map

	metricExpireIn        time.Duration
	cleanPeriodicDuration time.Duration
}

type SummaryOpt func(r *summaryRollup)

func SetSummaryCleaning(metricExpireIn time.Duration, cleanPeriodicDuration time.Duration) SummaryOpt {
	return func(r *summaryRollup) {
		r.metricExpireIn = metricExpireIn
		r.cleanPeriodicDuration = cleanPeriodicDuration
	}
}

func NewSummaryRollup(nodeIndex string, rollupTags []string, opts ...SummaryOpt) *summaryRollup {
	sr := &summaryRollup{
		nodeIndex:             nodeIndex,
		rollupTags:            rollupTags,
		summaries:             &sync.Map{},
		summariesInInterval:   &sync.Map{},
		metricExpireIn:        2 * time.Hour,
		cleanPeriodicDuration: 10 * time.Minute,
		keyCleaningTime:       &sync.Map{},
	}

	for _, opt := range opts {
		opt(sr)
	}

	go sr.CleanPeriodic()
	return sr
}

func (r *summaryRollup) Record(sourceId string, tags map[string]string, value int64) {
	key := keyFromTags(r.rollupTags, sourceId, tags)

	summary, found := r.summaries.Load(key)
	if !found {
		summary = prometheus.NewSummary(prometheus.SummaryOpts{
			Name:       metrics.GorouterHttpSummaryMetricName,
			Objectives: map[float64]float64{0.2: 0.05, 0.5: 0.05, 0.75: 0.02, 0.95: 0.01},
		})
		r.summaries.Store(key, summary)
	}
	summary.(prometheus.Summary).Observe(float64(value))

	r.summariesInInterval.Store(key, struct{}{})
	r.keyCleaningTime.Store(key, time.Now())
}

func (r *summaryRollup) CleanPeriodic() {
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
			r.summariesInInterval.Delete(key)
			r.summaries.Delete(key)
		}
	}
}

func (r *summaryRollup) Rollup(timestamp int64) []*PointsBatch {
	var batches []*PointsBatch

	r.summariesInInterval.Range(func(k, _ interface{}) bool {
		labels, err := labelsFromKey(k.(string), r.nodeIndex, r.rollupTags)
		if err != nil {
			return true
		}
		if _, ok := labels["app_id"]; ok {
			labels["origin"] = "cf_app"
		}
		m := &dto.Metric{}
		summary, _ := r.summaries.Load(k)
		_ = summary.(prometheus.Summary).Write(m)
		m.Label = transform.LabelsMapToLabelPairs(labels)

		metric := metricmaker.NewRawMetricFromMetric(metrics.GorouterHttpSummaryMetricName, m)
		metric.Metric().TimestampMs = proto.Int64(transform.NanosecondsToMilliseconds(timestamp))
		batches = append(batches, &PointsBatch{
			Points: []*metrics.RawMetric{metric},
			Size:   metric.EstimateMetricSize(),
		})
		return true
	})

	cleanSyncMap(r.summariesInInterval)

	return batches
}
