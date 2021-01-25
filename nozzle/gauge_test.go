package nozzle_test

import (
	"time"

	"github.com/bosh-prometheus/firehose_exporter/metricmaker"
	"github.com/bosh-prometheus/firehose_exporter/metrics"
	. "github.com/bosh-prometheus/firehose_exporter/nozzle"
	"github.com/bosh-prometheus/firehose_exporter/testing"
	"github.com/bosh-prometheus/firehose_exporter/transform"
	"github.com/gogo/protobuf/proto"
	dto "github.com/prometheus/client_model/go"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Nozzle", func() {
	var pointBuffer chan []*metrics.RawMetric
	var metricStore *MetricStoreTesting
	BeforeEach(func() {
		pointBuffer = make(chan []*metrics.RawMetric)
		metricStore = NewMetricStoreTesting(pointBuffer)
	})

	Describe("when the envelope is a Gauge", func() {
		It("converts the envelope to a Point(s)", func() {

			streamConnector := newSpyStreamConnector()

			n := NewNozzle(streamConnector, "firehose_exporter", 0,
				pointBuffer,
				internalMetric,
				WithNozzleTimerRollup(
					100*time.Millisecond,
					[]string{"tag1", "tag2", "status_code"},
					[]string{"tag1", "tag2"},
				),
			)
			go n.Start()

			streamConnector.envelopes <- []*loggregator_v2.Envelope{
				{
					Timestamp: 20,
					SourceId:  "source-id",
					Tags:      map[string]string{},
					Message: &loggregator_v2.Envelope_Gauge{
						Gauge: &loggregator_v2.Gauge{
							Metrics: map[string]*loggregator_v2.GaugeValue{
								"input": {
									Value: 50.0,
									Unit:  "mb/s",
								},
								"output": {
									Value: 25.5,
									Unit:  "kb/s",
								},
							},
						},
					},
				},
			}

			Eventually(metricStore.GetPoints).Should(HaveLen(2))

			Expect(metricStore.GetPoints()).To(testing.ContainPoints([]*metrics.RawMetric{
				metricmaker.NewRawMetricFromMetric("input", &dto.Metric{
					Label: transform.LabelsMapToLabelPairs(map[string]string{
						"unit":      "mb/s",
						"source_id": "source-id",
					}),
					Gauge: &dto.Gauge{
						Value: proto.Float64(50.0),
					},
				}),
				metricmaker.NewRawMetricFromMetric("output", &dto.Metric{
					Label: transform.LabelsMapToLabelPairs(map[string]string{
						"unit":      "kb/s",
						"source_id": "source-id",
					}),
					Gauge: &dto.Gauge{
						Value: proto.Float64(25.5),
					},
				}),
			}))
		})

		It("preserves units on tagged envelopes", func() {

			streamConnector := newSpyStreamConnector()

			n := NewNozzle(streamConnector, "firehose_exporter", 0,
				pointBuffer,
				internalMetric,
				WithNozzleTimerRollup(
					100*time.Millisecond,
					[]string{"tag1", "tag2", "status_code"},
					[]string{"tag1", "tag2"},
				),
			)
			go n.Start()

			streamConnector.envelopes <- []*loggregator_v2.Envelope{
				{
					Timestamp: 20,
					SourceId:  "source-id",
					Message: &loggregator_v2.Envelope_Gauge{
						Gauge: &loggregator_v2.Gauge{
							Metrics: map[string]*loggregator_v2.GaugeValue{
								"gauge1": {
									Unit:  "unit1",
									Value: 1,
								},
								"gauge2": {
									Unit:  "unit2",
									Value: 2,
								},
							},
						},
					},
					Tags: map[string]string{
						"deployment": "some-deployment",
					},
				},
			}

			Eventually(metricStore.GetPoints).Should(HaveLen(2))

			Expect(metricStore.GetPoints()).To(testing.ContainPoints([]*metrics.RawMetric{
				metricmaker.NewRawMetricFromMetric("gauge1", &dto.Metric{
					Label: transform.LabelsMapToLabelPairs(map[string]string{
						"unit":       "unit1",
						"source_id":  "source-id",
						"deployment": "some-deployment",
					}),
					Gauge: &dto.Gauge{
						Value: proto.Float64(1),
					},
				}),
				metricmaker.NewRawMetricFromMetric("gauge2", &dto.Metric{
					Label: transform.LabelsMapToLabelPairs(map[string]string{
						"unit":       "unit2",
						"source_id":  "source-id",
						"deployment": "some-deployment",
					}),
					Gauge: &dto.Gauge{
						Value: proto.Float64(2),
					},
				}),
			}))
		})
	})

})
