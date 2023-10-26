package nozzle_test

import (
	"context"
	"log"
	"reflect"
	"sync"
	"testing"

	"code.cloudfoundry.org/go-loggregator/v8"
	"code.cloudfoundry.org/go-loggregator/v8/rpc/loggregator_v2"
	"github.com/bosh-prometheus/firehose_exporter/metrics"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var internalMetric = metrics.NewInternalMetrics("firehose", "test")

func TestNozzle(t *testing.T) {
	log.SetOutput(ginkgo.GinkgoWriter)

	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Nozzle Suite")
}

// nolint:unparam
func addEnvelope(total uint64, name, sourceID string, c *spyStreamConnector) {
	c.envelopes <- []*loggregator_v2.Envelope{
		{
			SourceId: sourceID,
			Tags:     map[string]string{},
			Message: &loggregator_v2.Envelope_Counter{
				Counter: &loggregator_v2.Counter{Name: name, Total: total},
			},
		},
	}
}

type spyStreamConnector struct {
	mu               sync.Mutex
	internalRequests []*loggregator_v2.EgressBatchRequest
	envelopes        chan []*loggregator_v2.Envelope
}

func newSpyStreamConnector() *spyStreamConnector {
	return &spyStreamConnector{
		envelopes: make(chan []*loggregator_v2.Envelope, 100),
	}
}

func (s *spyStreamConnector) Stream(_ context.Context, req *loggregator_v2.EgressBatchRequest) loggregator.EnvelopeStream {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.internalRequests = append(s.internalRequests, req)

	return func() []*loggregator_v2.Envelope {
		select {
		case ee := <-s.envelopes:
			finalEnvelopes := make([]*loggregator_v2.Envelope, 0)
			for _, e := range ee {
				wantedType := reflect.TypeOf(&loggregator_v2.Selector_Counter{})
				switch e.Message.(type) {
				case *loggregator_v2.Envelope_Gauge:
					wantedType = reflect.TypeOf(&loggregator_v2.Selector_Gauge{})
				case *loggregator_v2.Envelope_Timer:
					wantedType = reflect.TypeOf(&loggregator_v2.Selector_Timer{})
				}
				for _, selector := range req.Selectors {
					if reflect.TypeOf(selector.Message).String() == wantedType.String() {
						finalEnvelopes = append(finalEnvelopes, e)
						break
					}
				}
			}
			return finalEnvelopes
		default:
			return nil
		}
	}
}

func (s *spyStreamConnector) requests() []*loggregator_v2.EgressBatchRequest {
	s.mu.Lock()
	defer s.mu.Unlock()

	reqs := make([]*loggregator_v2.EgressBatchRequest, len(s.internalRequests))
	copy(reqs, s.internalRequests)

	return reqs
}

type MetricStoreTesting struct {
	storage     []*metrics.RawMetric
	mutex       *sync.Mutex
	pointBuffer chan []*metrics.RawMetric
}

func NewMetricStoreTesting(pointBuffer chan []*metrics.RawMetric) *MetricStoreTesting {
	mst := &MetricStoreTesting{
		pointBuffer: pointBuffer,
		storage:     make([]*metrics.RawMetric, 0),
		mutex:       &sync.Mutex{},
	}
	go mst.collect()
	return mst
}

func (m *MetricStoreTesting) collect() {
	for points := range m.pointBuffer {
		m.mutex.Lock()
		m.storage = append(m.storage, points...)
		m.mutex.Unlock()
	}
}

func (m *MetricStoreTesting) GetPoints() []*metrics.RawMetric {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.storage
}
