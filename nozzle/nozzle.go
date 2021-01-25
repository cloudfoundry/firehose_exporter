package nozzle

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"code.cloudfoundry.org/go-diodes"
	"code.cloudfoundry.org/go-loggregator"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/bosh-prometheus/firehose_exporter/metricmaker"
	"github.com/bosh-prometheus/firehose_exporter/metrics"
	"github.com/bosh-prometheus/firehose_exporter/nozzle/rollup"
	"github.com/bosh-prometheus/firehose_exporter/utils"
	dto "github.com/prometheus/client_model/go"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	MaxBatchSizeInBytes = 32 * 1024
	lenGuid             = 36
)

var regexGuid = regexp.MustCompile(`(\{){0,1}[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}(\}){0,1}`)

// Nozzle reads envelopes and writes points to firehose_exporter.
type Nozzle struct {
	internalMetrics *metrics.InternalMetrics

	s             StreamConnector
	shardId       string
	nodeIndex     int
	ingressBuffer *diodes.OneToOne

	timerBuffer           *diodes.OneToOne
	timerRollupBufferSize uint
	rollupInterval        time.Duration
	totalRollup           rollup.Rollup
	durationRollup        rollup.Rollup
	responseSizeRollup    rollup.Rollup

	filterSelector   *FilterSelector
	filterDeployment *FilterDeployment

	pointBuffer chan []*metrics.RawMetric
}

// StreamConnector reads envelopes from the the logs provider.
type StreamConnector interface {
	// Stream creates a EnvelopeStream for the given request.
	Stream(ctx context.Context, req *loggregator_v2.EgressBatchRequest) loggregator.EnvelopeStream
}

const (
	BATCH_FLUSH_INTERVAL = 500 * time.Millisecond
)

func NewNozzle(c StreamConnector,
	shardId string,
	nodeIndex int,
	pointBuffer chan []*metrics.RawMetric,
	internalMetrics *metrics.InternalMetrics,
	opts ...Option) *Nozzle {
	n := &Nozzle{
		internalMetrics:       internalMetrics,
		s:                     c,
		shardId:               shardId,
		nodeIndex:             nodeIndex,
		timerRollupBufferSize: 4096,
		totalRollup:           rollup.NewNullRollup(),
		durationRollup:        rollup.NewNullRollup(),
		responseSizeRollup:    rollup.NewNullRollup(),
		pointBuffer:           pointBuffer,
		filterSelector:        NewFilterSelector(),
		filterDeployment:      NewFilterDeployment(),
	}

	for _, o := range opts {
		o(n)
	}

	n.timerBuffer = diodes.NewOneToOne(int(n.timerRollupBufferSize), diodes.AlertFunc(func(missed int) {
		n.internalMetrics.TotalEnvelopesDropped.Add(float64(missed))
		log.WithField("count", missed).Info("timer buffer dropped points")
	}))

	n.ingressBuffer = diodes.NewOneToOne(100000, diodes.AlertFunc(func(missed int) {
		n.internalMetrics.TotalEnvelopesDropped.Add(float64(missed))
		log.WithField("count", missed).Info("ingress buffer dropped envelopes")
	}))

	return n
}

type Option func(*Nozzle)

func WithNozzleTimerRollupBufferSize(size uint) Option {
	return func(n *Nozzle) {
		n.timerRollupBufferSize = size
	}
}

func WithFilterSelector(filterSelector *FilterSelector) Option {
	return func(n *Nozzle) {
		n.filterSelector = filterSelector
	}
}

func WithFilterDeployment(filterDeployment *FilterDeployment) Option {
	return func(n *Nozzle) {
		n.filterDeployment = filterDeployment
	}
}

func WithNozzleTimerRollup(interval time.Duration, totalResponseSizeRollupTags, durationRollupTags []string) Option {
	return func(n *Nozzle) {
		n.rollupInterval = interval

		nodeIndex := strconv.Itoa(n.nodeIndex)
		n.totalRollup = rollup.NewCounterRollup(nodeIndex, totalResponseSizeRollupTags)
		n.responseSizeRollup = rollup.NewSummaryRollup(nodeIndex, totalResponseSizeRollupTags)
		// TODO: rename HistogramRollup
		n.durationRollup = rollup.NewHistogramRollup(nodeIndex, durationRollupTags)
	}
}

// Start() starts reading envelopes from the logs provider and writes them to
// firehose_exporter.
func (n *Nozzle) Start() {

	rx := n.s.Stream(context.Background(), n.buildBatchReq())

	go n.timerProcessor()
	go n.timerEmitter()
	go n.envelopeReader(rx)
	go n.pointBatcher()
}

func (n *Nozzle) pointBatcher() {
	var size int

	poller := diodes.NewPoller(n.ingressBuffer)
	points := make([]*metrics.RawMetric, 0)

	t := time.NewTimer(BATCH_FLUSH_INTERVAL)
	for {
		data, found := poller.TryNext()

		if found {
			for _, point := range n.convertEnvelopeToPoints((*loggregator_v2.Envelope)(data)) {
				size += point.EstimateMetricSize()
				points = append(points, point)
			}
		}

		select {
		case <-t.C:
			if len(points) > 0 {
				points = n.writeToChannelOrDiscard(points)
			}
			t.Reset(BATCH_FLUSH_INTERVAL)
			size = 0
		default:
			// Do we care if one envelope produces multiple points, in which a
			// subset crosses the threshold?

			// if len(points) >= BATCH_CHANNEL_SIZE {
			if size >= MaxBatchSizeInBytes {
				points = n.writeToChannelOrDiscard(points)
				t.Reset(BATCH_FLUSH_INTERVAL)
				size = 0
			}

			// this sleep keeps us from hammering an empty channel, which
			// would otherwise cause us to peg the cpu when there's no work
			// to be done.
			if !found {
				time.Sleep(time.Millisecond)
			}
		}
	}
}

func (n *Nozzle) writeToChannelOrDiscard(points []*metrics.RawMetric) []*metrics.RawMetric {
	select {
	case n.pointBuffer <- points:
		n.internalMetrics.TotalMetricsReceived.Add(float64(len(points)))
		n.internalMetrics.LastMetricReceivedTimestamp.Set(float64(time.Now().Unix()))
		for _, point := range points {
			if utils.MetricIsContainerMetric(point) {
				n.internalMetrics.TotalContainerMetricsReceived.Inc()
				n.internalMetrics.LastContainerMetricReceivedTimestamp.Set(float64(time.Now().Unix()))
				continue
			}
			if utils.MetricIsHttpMetric(point) {
				n.internalMetrics.TotalHttpMetricsReceived.Inc()
				n.internalMetrics.LastHttpMetricReceivedTimestamp.Set(float64(time.Now().Unix()))
				continue
			}
			if *point.MetricType() == dto.MetricType_GAUGE {
				n.internalMetrics.TotalValueMetricsReceived.Inc()
				n.internalMetrics.LastValueMetricReceivedTimestamp.Set(float64(time.Now().Unix()))
				continue
			}
			if *point.MetricType() == dto.MetricType_COUNTER {
				n.internalMetrics.TotalCounterEventsReceived.Inc()
				n.internalMetrics.LastCounterEventReceivedTimestamp.Set(float64(time.Now().Unix()))
				continue
			}
		}
		return make([]*metrics.RawMetric, 0)
	default:
		// if we can't write into the channel, it must be full, so
		// we probably need to drop these envelopes on the floor
		n.internalMetrics.TotalMetricsDropped.Add(float64(len(points)))
		return points[:0]
	}
}

func (n *Nozzle) envelopeReader(rx loggregator.EnvelopeStream) {

	for {
		envelopeBatch := rx()
		for _, envelope := range envelopeBatch {
			n.ingressBuffer.Set(diodes.GenericDataType(envelope))
			n.internalMetrics.TotalEnvelopesReceived.Inc()
			n.internalMetrics.LastEnvelopeReceivedTimestamp.Set(float64(time.Now().Unix()))
		}
	}
}

func (n *Nozzle) timerProcessor() {
	poller := diodes.NewPoller(n.timerBuffer)

	for {
		data := poller.Next()
		envelope := *(*loggregator_v2.Envelope)(data)
		timer := envelope.GetTimer()
		tags := envelope.Tags
		tags["scheme"] = ""
		tags["host"] = ""
		if uri, ok := envelope.GetTags()["uri"]; ok && uri != "" {
			uri, err := url.Parse(uri)
			if err == nil {
				tags["scheme"] = uri.Scheme
				tags["host"] = uri.Host
			}
		}
		n.totalRollup.Record(envelope.SourceId, tags, 1)
		if contentLength, ok := envelope.GetTags()["content_length"]; ok && contentLength != "" {
			responseSize, err := strconv.Atoi(contentLength)
			if err == nil {
				n.responseSizeRollup.Record(envelope.SourceId, tags, int64(responseSize))
			}
		}
		n.durationRollup.Record(envelope.SourceId, tags, timer.GetStop()-timer.GetStart())
	}
}

func (n *Nozzle) timerEmitter() {
	ticker := time.NewTicker(n.rollupInterval)

	for t := range ticker.C {
		timestampNano := t.Truncate(n.rollupInterval).UnixNano()
		var size int
		var points []*metrics.RawMetric

		for _, pointsBatch := range n.totalRollup.Rollup(timestampNano) {
			points = append(points, pointsBatch.Points...)
			size += pointsBatch.Size

			if size >= MaxBatchSizeInBytes {
				points = n.writeToChannelOrDiscard(points)
				size = 0
			}
		}

		for _, pointsBatch := range n.durationRollup.Rollup(timestampNano) {
			points = append(points, pointsBatch.Points...)
			size += pointsBatch.Size

			if size >= MaxBatchSizeInBytes {
				points = n.writeToChannelOrDiscard(points)
				size = 0
			}
		}

		for _, pointsBatch := range n.responseSizeRollup.Rollup(timestampNano) {
			points = append(points, pointsBatch.Points...)
			size += pointsBatch.Size

			if size >= MaxBatchSizeInBytes {
				points = n.writeToChannelOrDiscard(points)
				size = 0
			}
		}

		if len(points) > 0 {
			points = n.writeToChannelOrDiscard(points)
		}
	}
}

func (n *Nozzle) captureGorouterHttpTimerMetricsForRollup(envelope *loggregator_v2.Envelope) {
	timer := envelope.GetTimer()

	if timer.GetName() != metrics.GorouterHttpMetricName {
		return
	}

	if envelope.GetSourceId() == "gorouter" && strings.ToLower(envelope.Tags["peer_type"]) == "client" {
		// gorouter reports both client and server timers for each request,
		// only record server types
		return
	}

	_, hasAppID := envelope.GetTags()["app_id"]
	// we skip metric with source_id with a guid (guid means an app) to avoid duplicate with metric from cf_app
	if !hasAppID && len(envelope.GetSourceId()) == lenGuid && regexGuid.MatchString(envelope.GetSourceId()) {
		return
	}

	n.timerBuffer.Set(diodes.GenericDataType(envelope))
}

func (n *Nozzle) convertEnvelopeToPoints(envelope *loggregator_v2.Envelope) []*metrics.RawMetric {
	if n.filterDeployment.IsFiltered(envelope) {
		return []*metrics.RawMetric{}
	}
	switch envelope.Message.(type) {
	case *loggregator_v2.Envelope_Gauge:
		if !n.filterSelector.ValueMetricDisabled() && !n.filterSelector.ContainerMetricDisabled() {
			break
		}
		metricsGauge := make(map[string]*loggregator_v2.GaugeValue)
		for name, m := range envelope.GetGauge().Metrics {
			if n.filterSelector.ValueMetricDisabled() && !utils.MetricNameIsContainerMetric(name) {
				continue
			}
			if n.filterSelector.ContainerMetricDisabled() && utils.MetricNameIsContainerMetric(name) {
				continue
			}
			metricsGauge[name] = m
		}
		envelope.GetGauge().Metrics = metricsGauge

	case *loggregator_v2.Envelope_Timer:
		n.captureGorouterHttpTimerMetricsForRollup(envelope)
		return []*metrics.RawMetric{}
	}
	return metricmaker.NewRawMetricsFromEnvelop(envelope)
}

func (n *Nozzle) buildBatchReq() *loggregator_v2.EgressBatchRequest {
	return &loggregator_v2.EgressBatchRequest{
		ShardId:          n.shardId,
		UsePreferredTags: true,
		Selectors:        n.filterSelector.ToSelectorTypes(),
	}
}
