package nozzle_test

import (
	"time"

	"code.cloudfoundry.org/go-loggregator/v8/rpc/loggregator_v2"
	"github.com/bosh-prometheus/firehose_exporter/metricmaker"
	"github.com/bosh-prometheus/firehose_exporter/metrics"
	"github.com/bosh-prometheus/firehose_exporter/testing"
	"github.com/bosh-prometheus/firehose_exporter/transform"
	"github.com/gogo/protobuf/proto"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	dto "github.com/prometheus/client_model/go"

	"github.com/bosh-prometheus/firehose_exporter/nozzle"
)

var _ = ginkgo.Describe("Nozzle", func() {
	var (
		streamConnector  *spyStreamConnector
		noz              *nozzle.Nozzle
		pointBuffer      chan []*metrics.RawMetric
		metricStore      *MetricStoreTesting
		filterSelector   *nozzle.FilterSelector
		filterDeployment *nozzle.FilterDeployment
	)

	ginkgo.BeforeEach(func() {
		pointBuffer = make(chan []*metrics.RawMetric)
		metricStore = NewMetricStoreTesting(pointBuffer)
		filterSelector = nozzle.NewFilterSelector()
		filterDeployment = nozzle.NewFilterDeployment()
		streamConnector = newSpyStreamConnector()

		noz = nozzle.NewNozzle(streamConnector, "firehose_exporter", 0,
			pointBuffer,
			internalMetric,
			nozzle.WithNozzleTimerRollup(
				100*time.Millisecond,
				[]string{"tag1", "tag2", "status_code"},
				[]string{"tag1", "tag2"},
			),
			nozzle.WithFilterSelector(filterSelector),
			nozzle.WithFilterDeployment(filterDeployment),
		)

	})

	ginkgo.JustBeforeEach(func() {
		go noz.Start()
	})

	ginkgo.It("connects and reads from a logs provider server", func() {
		addEnvelope(1, "memory", "some-source-id", streamConnector)
		addEnvelope(2, "memory", "some-source-id", streamConnector)
		addEnvelope(3, "memory", "some-source-id", streamConnector)

		gomega.Eventually(streamConnector.requests).Should(gomega.HaveLen(1))
		gomega.Expect(streamConnector.requests()[0].ShardId).To(gomega.Equal("firehose_exporter"))
		gomega.Expect(streamConnector.requests()[0].UsePreferredTags).To(gomega.BeTrue())
		gomega.Expect(streamConnector.requests()[0].Selectors).To(gomega.HaveLen(3))

		gomega.Expect(streamConnector.requests()[0].Selectors).To(gomega.ConsistOf(
			[]*loggregator_v2.Selector{
				{
					Message: &loggregator_v2.Selector_Gauge{
						Gauge: &loggregator_v2.GaugeSelector{},
					},
				},
				{
					Message: &loggregator_v2.Selector_Counter{
						Counter: &loggregator_v2.CounterSelector{},
					},
				},
				{
					Message: &loggregator_v2.Selector_Timer{
						Timer: &loggregator_v2.TimerSelector{},
					},
				},
			},
		))

		gomega.Eventually(streamConnector.envelopes).Should(gomega.HaveLen(0))
	})

	ginkgo.It("writes each envelope as a point to the firehose_exporter", func() {
		addEnvelope(1, "memory", "some-source-id", streamConnector)
		addEnvelope(2, "memory", "some-source-id", streamConnector)
		addEnvelope(3, "memory", "some-source-id", streamConnector)

		gomega.Eventually(metricStore.GetPoints).Should(gomega.HaveLen(3))
		gomega.Expect(metricStore.GetPoints()).To(testing.ContainPoints([]*metrics.RawMetric{
			metricmaker.NewRawMetricFromMetric("memory", &dto.Metric{
				Label: transform.LabelsMapToLabelPairs(map[string]string{
					"source_id": "some-source-id",
				}),
				Counter: &dto.Counter{
					Value: proto.Float64(1),
				},
			}),
			metricmaker.NewRawMetricFromMetric("memory", &dto.Metric{
				Label: transform.LabelsMapToLabelPairs(map[string]string{
					"source_id": "some-source-id",
				}),
				Counter: &dto.Counter{
					Value: proto.Float64(2),
				},
			}),
			metricmaker.NewRawMetricFromMetric("memory", &dto.Metric{
				Label: transform.LabelsMapToLabelPairs(map[string]string{
					"source_id": "some-source-id",
				}),
				Counter: &dto.Counter{
					Value: proto.Float64(3),
				},
			}),
		}))
	})

	ginkgo.Describe("when the envelope is a Counter", func() {
		ginkgo.It("converts the envelope to a Point", func() {
			streamConnector.envelopes <- []*loggregator_v2.Envelope{
				{
					Timestamp: 20,
					SourceId:  "source-id",
					Tags:      map[string]string{},
					Message: &loggregator_v2.Envelope_Counter{
						Counter: &loggregator_v2.Counter{
							Name:  "failures",
							Total: 8,
						},
					},
				},
			}

			gomega.Eventually(metricStore.GetPoints).Should(gomega.HaveLen(1))
			point := metricStore.GetPoints()[0]
			gomega.Expect(point.MetricName()).To(gomega.Equal("failures"))
			gomega.Expect(point.Metric().Counter.GetValue()).To(gomega.Equal(float64(8)))
			gomega.Expect(transform.LabelPairsToLabelsMap(point.Metric().Label)).To(gomega.HaveKeyWithValue("source_id", "source-id"))
		})
	})

	ginkgo.It("forwards all tags", func() {
		streamConnector.envelopes <- []*loggregator_v2.Envelope{
			{
				Timestamp: 20,
				SourceId:  "source-id",
				Message: &loggregator_v2.Envelope_Counter{
					Counter: &loggregator_v2.Counter{
						Name:  "counter",
						Total: 50,
					},
				},
				Tags: map[string]string{
					"forwarded-tag-1": "forwarded value",
					"forwarded-tag-2": "forwarded value",
				},
			},
		}

		gomega.Eventually(metricStore.GetPoints).Should(gomega.HaveLen(1))

		gomega.Expect(metricStore.GetPoints()).To(testing.ContainPoint(metricmaker.NewRawMetricFromMetric("counter", &dto.Metric{
			Label: transform.LabelsMapToLabelPairs(map[string]string{
				"forwarded-tag-1": "forwarded value",
				"forwarded-tag-2": "forwarded value",
				"source_id":       "source-id",
			}),
			Counter: &dto.Counter{
				Value: proto.Float64(50.0),
			},
		})))
	})

	ginkgo.Context("filter selector", func() {
		ginkgo.BeforeEach(func() {
			filterSelector.DisableAll()
			streamConnector.envelopes <- []*loggregator_v2.Envelope{
				{
					Timestamp: 20,
					SourceId:  "source-id",
					Tags:      map[string]string{},
					Message: &loggregator_v2.Envelope_Counter{
						Counter: &loggregator_v2.Counter{
							Name:  "failures",
							Total: 8,
						},
					},
				},
				{
					Timestamp: 20,
					SourceId:  "source-id",
					Tags:      map[string]string{},
					Message: &loggregator_v2.Envelope_Gauge{
						Gauge: &loggregator_v2.Gauge{
							Metrics: map[string]*loggregator_v2.GaugeValue{
								"cpu": {
									Unit:  "",
									Value: 1,
								},
							},
						},
					},
				},
				{
					Timestamp: 20,
					SourceId:  "source-id",
					Tags:      map[string]string{},
					Message: &loggregator_v2.Envelope_Gauge{
						Gauge: &loggregator_v2.Gauge{
							Metrics: map[string]*loggregator_v2.GaugeValue{
								"a_gauge": {
									Unit:  "",
									Value: 1,
								},
							},
						},
					},
				},
				{
					Timestamp: 20,
					SourceId:  "gorouter",
					Tags:      map[string]string{},
					Message: &loggregator_v2.Envelope_Timer{
						Timer: &loggregator_v2.Timer{
							Name:  "http",
							Start: 1,
							Stop:  2,
						},
					},
				},
			}
		})
		ginkgo.Describe("when selector CounterEvent", func() {
			ginkgo.BeforeEach(func() {
				filterSelector.Filters(nozzle.FilterSelectorTypeCounterEvent)
			})
			ginkgo.It("should only take counter metric", func() {
				gomega.Eventually(metricStore.GetPoints).Should(gomega.HaveLen(1))
				point := metricStore.GetPoints()[0]
				gomega.Expect(point.MetricName()).To(gomega.Equal("failures"))
				gomega.Expect(point.Metric().Counter.GetValue()).To(gomega.Equal(float64(8)))
				gomega.Expect(transform.LabelPairsToLabelsMap(point.Metric().Label)).To(gomega.HaveKeyWithValue("source_id", "source-id"))
			})
		})
		ginkgo.Describe("when selector ContainerMetric", func() {
			ginkgo.BeforeEach(func() {
				filterSelector.Filters(nozzle.FilterSelectorTypeContainerMetric)
			})
			ginkgo.It("should only take container metric", func() {
				gomega.Eventually(metricStore.GetPoints).Should(gomega.HaveLen(1))
				point := metricStore.GetPoints()[0]
				gomega.Expect(point.MetricName()).To(gomega.Equal("cpu"))
				gomega.Expect(point.Metric().Gauge.GetValue()).To(gomega.Equal(float64(1)))
				gomega.Expect(transform.LabelPairsToLabelsMap(point.Metric().Label)).To(gomega.HaveKeyWithValue("source_id", "source-id"))
			})
		})
		ginkgo.Describe("when selector ValueMetric", func() {
			ginkgo.BeforeEach(func() {
				filterSelector.Filters(nozzle.FilterSelectorTypeValueMetric)
			})
			ginkgo.It("should only take gauge metric which is not container metric", func() {
				gomega.Eventually(metricStore.GetPoints).Should(gomega.HaveLen(1))
				point := metricStore.GetPoints()[0]
				gomega.Expect(point.MetricName()).To(gomega.Equal("a_gauge"))
				gomega.Expect(point.Metric().Gauge.GetValue()).To(gomega.Equal(float64(1)))
				gomega.Expect(transform.LabelPairsToLabelsMap(point.Metric().Label)).To(gomega.HaveKeyWithValue("source_id", "source-id"))
			})
		})
		ginkgo.Describe("when selector Http", func() {
			ginkgo.BeforeEach(func() {
				filterSelector.Filters(nozzle.FilterSelectorTypeHTTPStartStop)
			})
			ginkgo.It("should only take timer metric", func() {
				gomega.Eventually(metricStore.GetPoints).Should(gomega.HaveLen(2))
				pointCounter := metricStore.GetPoints()[0]
				gomega.Expect(pointCounter.MetricName()).To(gomega.Equal("http_total"))
				gomega.Expect(pointCounter.Metric().Counter.GetValue()).To(gomega.Equal(float64(1)))
				pointHisto := metricStore.GetPoints()[1]
				gomega.Expect(pointHisto.MetricName()).To(gomega.Equal("http_duration_seconds"))
				gomega.Expect(pointHisto.Metric().Histogram).ToNot(gomega.BeNil())
			})
		})
	})

	ginkgo.Context("filter deployment", func() {
		ginkgo.BeforeEach(func() {
			filterDeployment.SetDeployments("cf", "bosh")
			streamConnector.envelopes <- []*loggregator_v2.Envelope{
				{
					Timestamp: 20,
					SourceId:  "source-id",
					Tags: map[string]string{
						"deployment": "cf",
					},
					Message: &loggregator_v2.Envelope_Counter{
						Counter: &loggregator_v2.Counter{
							Name:  "failures",
							Total: 8,
						},
					},
				},
				{
					Timestamp: 20,
					SourceId:  "source-id",
					Tags: map[string]string{
						"deployment": "bosh",
					},
					Message: &loggregator_v2.Envelope_Gauge{
						Gauge: &loggregator_v2.Gauge{
							Metrics: map[string]*loggregator_v2.GaugeValue{
								"memory": {
									Unit:  "",
									Value: 1,
								},
							},
						},
					},
				},
				{
					Timestamp: 20,
					SourceId:  "source-id",
					Tags: map[string]string{
						"deployment": "other",
					},
					Message: &loggregator_v2.Envelope_Gauge{
						Gauge: &loggregator_v2.Gauge{
							Metrics: map[string]*loggregator_v2.GaugeValue{
								"cpu": {
									Unit:  "",
									Value: 1,
								},
							},
						},
					},
				},
			}
		})

		ginkgo.It("should only take metrics from these deployments", func() {
			gomega.Eventually(metricStore.GetPoints).Should(gomega.HaveLen(2))
			point1 := metricStore.GetPoints()[0]
			gomega.Expect(point1.MetricName()).To(gomega.Equal("failures"))
			gomega.Expect(point1.Metric().Counter.GetValue()).To(gomega.Equal(float64(8)))
			gomega.Expect(transform.LabelPairsToLabelsMap(point1.Metric().Label)).To(gomega.HaveKeyWithValue("source_id", "source-id"))

			point2 := metricStore.GetPoints()[1]
			gomega.Expect(point2.MetricName()).To(gomega.Equal("memory"))
			gomega.Expect(point2.Metric().Gauge.GetValue()).To(gomega.Equal(float64(1)))
			gomega.Expect(transform.LabelPairsToLabelsMap(point2.Metric().Label)).To(gomega.HaveKeyWithValue("source_id", "source-id"))
		})
	})
})
