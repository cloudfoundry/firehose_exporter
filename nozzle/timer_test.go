package nozzle_test

import (
	"time"

	"github.com/bosh-prometheus/firehose_exporter/metricmaker"
	"github.com/bosh-prometheus/firehose_exporter/metrics"
	. "github.com/bosh-prometheus/firehose_exporter/nozzle"
	"github.com/bosh-prometheus/firehose_exporter/transform"
	"github.com/gogo/protobuf/proto"
	dto "github.com/prometheus/client_model/go"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/bosh-prometheus/firehose_exporter/testing"
)

var _ = Describe("when the envelope is a Timer", func() {
	var (
		streamConnector  *spyStreamConnector
		nozzle           *Nozzle
		pointBuffer      chan []*metrics.RawMetric
		metricStore      *MetricStoreTesting
		filterSelector   *FilterSelector
		filterDeployment *FilterDeployment
	)

	BeforeEach(func() {
		pointBuffer = make(chan []*metrics.RawMetric)
		metricStore = NewMetricStoreTesting(pointBuffer)
		filterSelector = NewFilterSelector()
		filterDeployment = NewFilterDeployment()
		streamConnector = newSpyStreamConnector()
		nozzle = NewNozzle(streamConnector, "firehose_exporter", 0,
			pointBuffer,
			internalMetric,
			WithNozzleTimerRollup(
				100*time.Millisecond,
				[]string{"tag1", "tag2", "status_code", "app_id"},
				[]string{"tag1", "tag2", "app_id"},
			),
			WithFilterSelector(filterSelector),
			WithFilterDeployment(filterDeployment),
		)
	})

	JustBeforeEach(func() {
		go nozzle.Start()
	})

	It("rolls up configured metrics", func() {

		intervalStart := time.Now().Truncate(100 * time.Millisecond).UnixNano()

		streamConnector.envelopes <- []*loggregator_v2.Envelope{
			{
				Timestamp: intervalStart + 1,
				SourceId:  "source-id",
				Message: &loggregator_v2.Envelope_Timer{
					Timer: &loggregator_v2.Timer{
						Name:  "http",
						Start: 0,
						Stop:  5 * int64(time.Millisecond),
					},
				},
				Tags: map[string]string{
					"tag1":        "t1",
					"tag2":        "t2",
					"status_code": "500",
				},
			},
			{
				Timestamp: intervalStart + 2,
				SourceId:  "source-id",
				Message: &loggregator_v2.Envelope_Timer{
					Timer: &loggregator_v2.Timer{
						Name:  "http",
						Start: 5 * int64(time.Millisecond),
						Stop:  100 * int64(time.Millisecond),
					},
				},
				Tags: map[string]string{
					"tag1":        "t1",
					"tag2":        "t2",
					"status_code": "200",
				},
			},
			{
				Timestamp: intervalStart + 3,
				SourceId:  "source-id-2",
				Message: &loggregator_v2.Envelope_Timer{
					Timer: &loggregator_v2.Timer{
						Name:  "http",
						Start: 100 * int64(time.Millisecond),
						Stop:  106 * int64(time.Millisecond),
					},
				},
				Tags: map[string]string{
					"peer_type":   "Server",
					"status_code": "200",
				},
			},
			{
				Timestamp: intervalStart + 4,
				SourceId:  "source-id-2",
				Message: &loggregator_v2.Envelope_Timer{
					Timer: &loggregator_v2.Timer{
						Name:  "http",
						Start: 96 * int64(time.Millisecond),
						Stop:  100 * int64(time.Millisecond),
					},
				},
				Tags: map[string]string{
					"peer_type":   "Server",
					"status_code": "200",
				},
			},
			{
				Timestamp: intervalStart + 5,
				SourceId:  "source-id-2",
				Message: &loggregator_v2.Envelope_Timer{
					Timer: &loggregator_v2.Timer{
						Name:  "http",
						Start: 500 * int64(time.Millisecond),
						Stop:  1000 * int64(time.Millisecond),
					},
				},
				Tags: map[string]string{
					"peer_type":   "Server",
					"status_code": "400",
				},
			},
		}

		numberOfExpectedSeriesIncludingStatusCode := 4
		numberOfExpectedSeriesExcludingStatusCode := 2
		numberOfExpectedPoints := numberOfExpectedSeriesIncludingStatusCode + numberOfExpectedSeriesExcludingStatusCode
		Eventually(metricStore.GetPoints).Should(HaveLen(numberOfExpectedPoints))

		points := metricStore.GetPoints()

		// _count points, per series including status_code
		Expect(points).To(ContainPoints([]*metrics.RawMetric{
			metricmaker.NewRawMetricCounter("http_total", map[string]string{
				"node_index":  "0",
				"source_id":   "source-id",
				"tag1":        "t1",
				"tag2":        "t2",
				"status_code": "500",
			}, 1.0),
			metricmaker.NewRawMetricCounter("http_total", map[string]string{
				"node_index":  "0",
				"source_id":   "source-id",
				"tag1":        "t1",
				"tag2":        "t2",
				"status_code": "200",
			}, 1.0),
			metricmaker.NewRawMetricCounter("http_total", map[string]string{
				"node_index":  "0",
				"source_id":   "source-id-2",
				"status_code": "200",
			}, 2.0),
			metricmaker.NewRawMetricCounter("http_total", map[string]string{
				"node_index":  "0",
				"source_id":   "source-id-2",
				"status_code": "400",
			}, 1.0),
		}))

		// _duration_seconds histogram points, per series excluding status_code
		// only testing one series

		Expect(points).To(ContainPoints([]*metrics.RawMetric{
			metricmaker.NewRawMetricFromMetric("http_duration_seconds",
				createHistogramMetric(
					map[string]string{
						"node_index": "0",
						"source_id":  "source-id",
						"tag1":       "t1",
						"tag2":       "t2",
					},
					map[float64]uint64{
						0.005: 1,
						0.01:  1,
						0.025: 1,
						0.05:  1,
						0.1:   2,
						0.25:  2,
						0.5:   2,
						1.0:   2,
						2.5:   2,
						5.0:   2,
						10.0:  2,
					}),
			),
		}))

		firstPointTimestamp := points[0].Metric().GetTimestampMs()
		firstPointTime := time.Unix(firstPointTimestamp/1000, 0)

		Expect(firstPointTime).To(BeTemporally("~", time.Unix(0, intervalStart), time.Second))
		Expect(firstPointTime).To(Equal(firstPointTime.Truncate(100 * time.Millisecond)))
	})

	It("only rolls up gorouter metrics with a peer_type of Server", func() {
		intervalStart := time.Now().Truncate(100 * time.Millisecond).UnixNano()

		streamConnector.envelopes <- []*loggregator_v2.Envelope{
			{
				Timestamp: intervalStart + 1,
				SourceId:  "gorouter",
				Message: &loggregator_v2.Envelope_Timer{
					Timer: &loggregator_v2.Timer{
						Name:  "http",
						Start: 100 * int64(time.Millisecond),
						Stop:  106 * int64(time.Millisecond),
					},
				},
				Tags: map[string]string{
					"peer_type":   "Server",
					"status_code": "200",
				},
			},
			{
				Timestamp: intervalStart + 1,
				SourceId:  "gorouter",
				Message: &loggregator_v2.Envelope_Timer{
					Timer: &loggregator_v2.Timer{
						Name:  "http",
						Start: 100 * int64(time.Millisecond),
						Stop:  106 * int64(time.Millisecond),
					},
				},
				Tags: map[string]string{
					"peer_type":   "Client",
					"status_code": "200",
				},
			},
			{
				Timestamp: intervalStart + 2,
				SourceId:  "gorouter",
				Message: &loggregator_v2.Envelope_Timer{
					Timer: &loggregator_v2.Timer{
						Name:  "http",
						Start: 96 * int64(time.Millisecond),
						Stop:  100 * int64(time.Millisecond),
					},
				},
				Tags: map[string]string{
					"peer_type":   "Server",
					"status_code": "200",
				},
			},
			{
				Timestamp: intervalStart + 2,
				SourceId:  "gorouter",
				Message: &loggregator_v2.Envelope_Timer{
					Timer: &loggregator_v2.Timer{
						Name:  "http",
						Start: 96 * int64(time.Millisecond),
						Stop:  100 * int64(time.Millisecond),
					},
				},
				Tags: map[string]string{
					"peer_type":   "Client",
					"status_code": "200",
				},
			},
			{
				Timestamp: intervalStart + 3,
				SourceId:  "gorouter",
				Message: &loggregator_v2.Envelope_Timer{
					Timer: &loggregator_v2.Timer{
						Name:  "http",
						Start: 500 * int64(time.Millisecond),
						Stop:  1000 * int64(time.Millisecond),
					},
				},
				Tags: map[string]string{
					"peer_type":   "Server",
					"status_code": "400",
				},
			},
			{
				Timestamp: intervalStart + 3,
				SourceId:  "gorouter",
				Message: &loggregator_v2.Envelope_Timer{
					Timer: &loggregator_v2.Timer{
						Name:  "http",
						Start: 500 * int64(time.Millisecond),
						Stop:  1000 * int64(time.Millisecond),
					},
				},
				Tags: map[string]string{
					"peer_type":   "Client",
					"status_code": "400",
				},
			},
		}

		numberOfExpectedSeriesIncludingStatusCode := 2
		numberOfExpectedSeriesExcludingStatusCode := 1
		numberOfExpectedPoints := numberOfExpectedSeriesIncludingStatusCode + numberOfExpectedSeriesExcludingStatusCode
		Eventually(metricStore.GetPoints).Should(HaveLen(numberOfExpectedPoints))

		points := metricStore.GetPoints()
		// _count points, per series including status_code
		Expect(points).To(ContainPoints([]*metrics.RawMetric{
			metricmaker.NewRawMetricCounter("http_total", map[string]string{
				"node_index":  "0",
				"source_id":   "gorouter",
				"status_code": "200",
			}, 2.0),
			metricmaker.NewRawMetricCounter("http_total", map[string]string{
				"node_index":  "0",
				"source_id":   "gorouter",
				"status_code": "400",
			}, 1.0),
		}))

		// _duration_seconds histogram points, per series excluding status_code
		// only testing one series

		Expect(points).To(ContainPoints([]*metrics.RawMetric{
			metricmaker.NewRawMetricFromMetric("http_duration_seconds",
				createHistogramMetric(
					map[string]string{
						"node_index": "0",
						"source_id":  "gorouter",
					},
					map[float64]uint64{
						0.005: 1,
						0.01:  2,
						0.025: 2,
						0.05:  2,
						0.1:   2,
						0.25:  2,
						0.5:   3,
						1.0:   3,
						2.5:   3,
						5.0:   3,
						10.0:  3,
					}),
			),
		}))

		firstPointTimestamp := points[0].Metric().GetTimestampMs()
		firstPointTime := time.Unix(firstPointTimestamp/1000, 0)

		Expect(firstPointTime).To(BeTemporally("~", time.Unix(0, intervalStart), time.Second))
		Expect(firstPointTime).To(Equal(firstPointTime.Truncate(100 * time.Millisecond)))

	})

	It("ignores other metrics", func() {

		streamConnector.envelopes <- []*loggregator_v2.Envelope{
			{
				SourceId: "source-id",
				// prime number for higher numerical accuracy
				Timestamp: 10000000002065383,
				Message: &loggregator_v2.Envelope_Timer{
					Timer: &loggregator_v2.Timer{
						Name:  "not_http",
						Start: 0,
						Stop:  5,
					},
				},
			},
			{
				SourceId:  "source-id",
				Timestamp: 66606660666066601,
				Message: &loggregator_v2.Envelope_Counter{
					Counter: &loggregator_v2.Counter{
						Name:  "http",
						Total: 4,
					},
				},
			},
		}

		Eventually(metricStore.GetPoints).Should(HaveLen(1))
		Consistently(metricStore.GetPoints, .5).Should(HaveLen(1))
		m := metricmaker.NewRawMetricCounter("http", map[string]string{
			"source_id": "source-id",
		}, 4.0)
		m.Metric().TimestampMs = proto.Int64(66606660666066601)
		Expect(metricStore.GetPoints()).To(ContainPoint(m))
	})

	It("keeps a total across rollupIntervals", func() {

		baseTimer := loggregator_v2.Envelope{
			SourceId: "source-id",
			Message: &loggregator_v2.Envelope_Timer{
				Timer: &loggregator_v2.Timer{
					Name:  "http",
					Start: 0,
					Stop:  0,
				},
			},
			Tags: map[string]string{
				"tag1": "t1",
				"tag2": "t2",
			},
		}

		firstTimer := baseTimer
		firstTimer.Message.(*loggregator_v2.Envelope_Timer).Timer.Stop = 5 * int64(time.Second)

		secondTimer := baseTimer
		secondTimer.Message.(*loggregator_v2.Envelope_Timer).Timer.Stop = 2 * int64(time.Second)

		thirdTimer := baseTimer
		thirdTimer.Message.(*loggregator_v2.Envelope_Timer).Timer.Stop = 4 * int64(time.Second)

		streamConnector.envelopes <- []*loggregator_v2.Envelope{&firstTimer}

		Eventually(metricStore.GetPoints).Should(HaveLen(2))

		streamConnector.envelopes <- []*loggregator_v2.Envelope{&secondTimer, &thirdTimer}

		Eventually(metricStore.GetPoints).Should(HaveLen(4))

		secondIntervalHistogram := metricmaker.NewRawMetricFromMetric("http_duration_seconds",
			createHistogramMetric(
				map[string]string{
					"node_index": "0",
					"source_id":  "source-id",
					"tag1":       "t1",
					"tag2":       "t2",
				},
				map[float64]uint64{
					0.005: 0,
					0.01:  0,
					0.025: 0,
					0.05:  0,
					0.1:   0,
					0.25:  0,
					0.5:   0,
					1.0:   0,
					2.5:   0,
					5.0:   3,
					10.0:  3,
				}),
		)

		secondIntervalTotal := metricmaker.NewRawMetricCounter("http_total", map[string]string{
			"node_index": "0",
			"source_id":  "source-id",
			"tag1":       "t1",
			"tag2":       "t2",
		}, 3.0)

		Expect(metricStore.GetPoints()).To(ContainPoints([]*metrics.RawMetric{
			secondIntervalHistogram,
			secondIntervalTotal,
		}))
	})

	It("skip metric with a source id in form of app guid", func() {
		intervalStart := time.Now().Truncate(100 * time.Millisecond).UnixNano()

		streamConnector.envelopes <- []*loggregator_v2.Envelope{
			{
				Timestamp: intervalStart + 1,
				SourceId:  "6f0b4a14-0703-442c-bc80-bea78d31d5ab",
				Message: &loggregator_v2.Envelope_Timer{
					Timer: &loggregator_v2.Timer{
						Name:  "http",
						Start: 0,
						Stop:  5 * int64(time.Millisecond),
					},
				},
				Tags: map[string]string{
					"tag1":        "t1",
					"tag2":        "t2",
					"status_code": "500",
				},
			},
			{
				Timestamp: intervalStart + 1,
				SourceId:  "6f0b4a14-0703-442c-bc80-bea78d31d5ab",
				Message: &loggregator_v2.Envelope_Timer{
					Timer: &loggregator_v2.Timer{
						Name:  "http",
						Start: 0,
						Stop:  5 * int64(time.Millisecond),
					},
				},
				Tags: map[string]string{
					"tag1":        "t1",
					"app_id":      "6f0b4a14-0703-442c-bc80-bea78d31d5ab",
					"tag2":        "t2",
					"status_code": "500",
				},
			},
		}
		Eventually(metricStore.GetPoints).Should(HaveLen(2))
		labelsFirst := transform.LabelPairsToLabelsMap(metricStore.GetPoints()[0].Metric().GetLabel())
		Expect(labelsFirst).Should(HaveKeyWithValue("app_id", "6f0b4a14-0703-442c-bc80-bea78d31d5ab"))
		labelsSecond := transform.LabelPairsToLabelsMap(metricStore.GetPoints()[1].Metric().GetLabel())
		Expect(labelsSecond).Should(HaveKeyWithValue("app_id", "6f0b4a14-0703-442c-bc80-bea78d31d5ab"))
	})
})

func createHistogramMetric(labels map[string]string, bucketValues map[float64]uint64) *dto.Metric {
	buckets := make([]*dto.Bucket, 0)
	for k, v := range bucketValues {
		buckets = append(buckets, &dto.Bucket{
			CumulativeCount: proto.Uint64(v),
			UpperBound:      proto.Float64(k),
		})
	}
	return &dto.Metric{
		Label: transform.LabelsMapToLabelPairs(labels),
		Histogram: &dto.Histogram{
			SampleCount: nil,
			SampleSum:   nil,
			Bucket:      buckets,
		},
	}
}
