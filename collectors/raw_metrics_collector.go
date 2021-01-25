package collectors

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bosh-prometheus/firehose_exporter/metrics"
	"github.com/gogo/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	log "github.com/sirupsen/logrus"
)

var gzipPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(nil)
	},
}

type RawMetricsCollector struct {
	pointBuffer           chan []*metrics.RawMetric
	metricStore           *sync.Map
	metricExpireIn        time.Duration
	cleanPeriodicDuration time.Duration
}

func NewRawMetricsCollector(
	pointBuffer chan []*metrics.RawMetric,
	metricExpireIn time.Duration,
) *RawMetricsCollector {

	return &RawMetricsCollector{
		pointBuffer:           pointBuffer,
		metricStore:           &sync.Map{},
		metricExpireIn:        metricExpireIn,
		cleanPeriodicDuration: 30 * time.Second,
	}
}

func (c *RawMetricsCollector) Collect() {
	for points := range c.pointBuffer {
		for _, point := range points {
			smapMetric, _ := c.metricStore.LoadOrStore(point.MetricName(), &sync.Map{})
			point.ExpireIn(c.metricExpireIn)
			smapMetric.(*sync.Map).Store(point.Id(), point)
		}
	}
}

func (c *RawMetricsCollector) Start() {
	for i := 0; i < 10; i++ {
		go c.Collect()
	}
	go c.CleanPeriodic()
}

func (c *RawMetricsCollector) SetCleanPeriodicDuration(cleanPeriodicDuration time.Duration) {
	c.cleanPeriodicDuration = cleanPeriodicDuration
}

func (c *RawMetricsCollector) SetMetricExpireIn(metricExpireIn time.Duration) {
	c.metricExpireIn = metricExpireIn
}

func (c *RawMetricsCollector) CleanPeriodic() {
	for {
		time.Sleep(c.cleanPeriodicDuration)
		nbJob := 0
		c.metricStore.Range(func(_, _ interface{}) bool {
			nbJob++
			return true
		})
		nbWorker := 5
		wg := &sync.WaitGroup{}
		wg.Add(nbWorker)

		chanSmap := make(chan *sync.Map, nbJob)
		for w := 0; w < nbWorker; w++ {
			go c.cleanWorker(wg, chanSmap)
		}
		c.metricStore.Range(func(_, value interface{}) bool {
			chanSmap <- value.(*sync.Map)
			return true
		})
		close(chanSmap)
		wg.Wait()
	}
}

func (c *RawMetricsCollector) cleanWorker(wg *sync.WaitGroup, chanSmap <-chan *sync.Map) {
	defer wg.Done()
	for smapMetric := range chanSmap {
		toDelete := make([]uint64, 0)
		smapMetric.Range(func(key, value interface{}) bool {
			rawMetric := value.(*metrics.RawMetric)
			if rawMetric.IsSwept() {
				toDelete = append(toDelete, key.(uint64))
			}
			return true
		})
		for _, key := range toDelete {
			smapMetric.Delete(key)
		}
	}
}

func (c *RawMetricsCollector) RenderExpFmt(rsp http.ResponseWriter, req *http.Request) {
	format := expfmt.Negotiate(req.Header)
	header := rsp.Header()
	header.Set("Content-Type", string(format))
	w := io.Writer(rsp)

	if gzipAccepted(req.Header) {
		header.Set("Content-Encoding", "gzip")
		gz := gzipPool.Get().(*gzip.Writer)
		defer gzipPool.Put(gz)

		gz.Reset(w)
		defer gz.Close()

		w = gz
	}

	enc := expfmt.NewEncoder(w, format)

	c.metricStore.Range(func(_, value interface{}) bool {
		smapMetric := value.(*sync.Map)
		var oneRawMetric *metrics.RawMetric
		finalMetrics := make([]*dto.Metric, 0)

		smapMetric.Range(func(_, value interface{}) bool {
			rawMetric := value.(*metrics.RawMetric)
			if rawMetric.IsSwept() {
				return true
			}
			oneRawMetric = rawMetric
			finalMetrics = append(finalMetrics, rawMetric.Metric())
			return true
		})
		if oneRawMetric == nil {
			return true
		}
		metricFamily := &dto.MetricFamily{
			Name:   proto.String(oneRawMetric.MetricName()),
			Help:   proto.String(oneRawMetric.Help()),
			Type:   oneRawMetric.MetricType(),
			Metric: finalMetrics,
		}
		if err := enc.Encode(metricFamily); err != nil && !strings.Contains(err.Error(), "broken pipe") {
			log.Warningf("Error when encoding exp fmt: %s", err.Error())
		}
		return true
	})

	reg := prometheus.DefaultGatherer
	mfs, err := reg.Gather()
	if err != nil {
		log.Warnf("Could not gather prometheus collectors: %s", err.Error())
		return
	}
	for _, mf := range mfs {
		if err := enc.Encode(mf); err != nil && !strings.Contains(err.Error(), "broken pipe") {
			log.Warningf("Error when encoding exp fmt from gathered collectors: %s", err.Error())
		}
	}
	if closer, ok := enc.(expfmt.Closer); ok {
		// This in particular takes care of the final "# EOF\n" line for OpenMetrics.
		closer.Close()
	}
}

func (c *RawMetricsCollector) MetricStore() map[string][]*metrics.RawMetric {
	metricStoreMap := make(map[string][]*metrics.RawMetric)
	c.metricStore.Range(func(metricName, value interface{}) bool {
		smapMetric := value.(*sync.Map)
		finalMetrics := make([]*metrics.RawMetric, 0)
		smapMetric.Range(func(_, value interface{}) bool {
			rawMetric := value.(*metrics.RawMetric)
			finalMetrics = append(finalMetrics, rawMetric)
			return true
		})
		metricStoreMap[metricName.(string)] = finalMetrics
		return true
	})
	return metricStoreMap
}

func gzipAccepted(header http.Header) bool {
	a := header.Get("Accept-Encoding")
	parts := strings.Split(a, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "gzip" || strings.HasPrefix(part, "gzip;") {
			return true
		}
	}
	return false
}
