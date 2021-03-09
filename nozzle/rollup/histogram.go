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

type histogramRollup struct {
	nodeIndex            string
	rollupTags           []string
	histogramsInInterval *sync.Map
	histograms           *sync.Map
	keyCleaningTime      *sync.Map

	metricExpireIn        time.Duration
	cleanPeriodicDuration time.Duration
}

type HistogramOpt func(r *histogramRollup)

func SetHistogramCleaning(metricExpireIn time.Duration, cleanPeriodicDuration time.Duration) HistogramOpt {
	return func(r *histogramRollup) {
		r.metricExpireIn = metricExpireIn
		r.cleanPeriodicDuration = cleanPeriodicDuration
	}
}

func NewHistogramRollup(nodeIndex string, rollupTags []string, opts ...HistogramOpt) *histogramRollup {
	hr := &histogramRollup{
		nodeIndex:             nodeIndex,
		rollupTags:            rollupTags,
		histogramsInInterval:  &sync.Map{},
		histograms:            &sync.Map{},
		metricExpireIn:        2 * time.Hour,
		cleanPeriodicDuration: 10 * time.Minute,
		keyCleaningTime:       &sync.Map{},
	}

	for _, opt := range opts {
		opt(hr)
	}

	go hr.CleanPeriodic()
	return hr
}

func (r *histogramRollup) CleanPeriodic() {
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
			r.histograms.Delete(key)
			r.histogramsInInterval.Delete(key)
		}
	}
}

func (r *histogramRollup) Record(sourceId string, tags map[string]string, value int64) {
	key := keyFromTags(r.rollupTags, sourceId, tags)

	histo, found := r.histograms.Load(key)
	if !found {
		histo = prometheus.NewHistogram(prometheus.HistogramOpts{
			Name: metrics.GorouterHttpHistogramMetricName,
		})
		r.histograms.Store(key, histo)
	}

	histo.(prometheus.Histogram).Observe(transform.NanosecondsToSeconds(value))

	r.histogramsInInterval.Store(key, struct{}{})
	r.keyCleaningTime.Store(key, time.Now())
}

func (r *histogramRollup) Rollup(timestamp int64) []*PointsBatch {
	var batches []*PointsBatch

	r.histogramsInInterval.Range(func(k, _ interface{}) bool {
		labels, err := labelsFromKey(k.(string), r.nodeIndex, r.rollupTags)
		if err != nil {
			return true
		}
		if _, ok := labels["app_id"]; ok {
			labels["origin"] = "cf_app"
		}
		m := &dto.Metric{}
		histo, _ := r.histograms.Load(k)
		_ = histo.(prometheus.Histogram).Write(m)
		m.Label = transform.LabelsMapToLabelPairs(labels)

		metric := metricmaker.NewRawMetricFromMetric(metrics.GorouterHttpHistogramMetricName, m)
		metric.Metric().TimestampMs = proto.Int64(transform.NanosecondsToMilliseconds(timestamp))
		batches = append(batches, &PointsBatch{
			Points: []*metrics.RawMetric{metric},
			Size:   metric.EstimateMetricSize(),
		})
		return true
	})

	r.histogramsInInterval = &sync.Map{}

	return batches
}
