package nozzle_test

import (
	"time"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/bosh-prometheus/firehose_exporter/metricmaker"
	"github.com/bosh-prometheus/firehose_exporter/metrics"
	"github.com/bosh-prometheus/firehose_exporter/testing"
	"github.com/bosh-prometheus/firehose_exporter/transform"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	dto "github.com/prometheus/client_model/go"

	. "github.com/bosh-prometheus/firehose_exporter/nozzle"
)

var _ = Describe("Nozzle", func() {
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
				[]string{"tag1", "tag2", "status_code"},
				[]string{"tag1", "tag2"},
			),
			WithFilterSelector(filterSelector),
			WithFilterDeployment(filterDeployment),
		)

	})

	JustBeforeEach(func() {
		go nozzle.Start()
	})

	It("connects and reads from a logs provider server", func() {
		addEnvelope(1, "memory", "some-source-id", streamConnector)
		addEnvelope(2, "memory", "some-source-id", streamConnector)
		addEnvelope(3, "memory", "some-source-id", streamConnector)

		Eventually(streamConnector.requests).Should(HaveLen(1))
		Expect(streamConnector.requests()[0].ShardId).To(Equal("firehose_exporter"))
		Expect(streamConnector.requests()[0].UsePreferredTags).To(BeTrue())
		Expect(streamConnector.requests()[0].Selectors).To(HaveLen(3))

		Expect(streamConnector.requests()[0].Selectors).To(ConsistOf(
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

		Eventually(streamConnector.envelopes).Should(HaveLen(0))
	})

	It("writes each envelope as a point to the firehose_exporter", func() {
		addEnvelope(1, "memory", "some-source-id", streamConnector)
		addEnvelope(2, "memory", "some-source-id", streamConnector)
		addEnvelope(3, "memory", "some-source-id", streamConnector)

		Eventually(metricStore.GetPoints).Should(HaveLen(3))
		Expect(metricStore.GetPoints()).To(testing.ContainPoints([]*metrics.RawMetric{
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

	Describe("when the envelope is a Counter", func() {
		It("converts the envelope to a Point", func() {
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

			Eventually(metricStore.GetPoints).Should(HaveLen(1))
			point := metricStore.GetPoints()[0]
			Expect(point.MetricName()).To(Equal("failures"))
			Expect(point.Metric().Counter.GetValue()).To(Equal(float64(8)))
			Expect(transform.LabelPairsToLabelsMap(point.Metric().Label)).To(HaveKeyWithValue("source_id", "source-id"))
		})
	})

	It("forwards all tags", func() {
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

		Eventually(metricStore.GetPoints).Should(HaveLen(1))

		Expect(metricStore.GetPoints()).To(testing.ContainPoint(metricmaker.NewRawMetricFromMetric("counter", &dto.Metric{
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

	Context("filter selector", func() {
		BeforeEach(func() {
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
								"cpu": &loggregator_v2.GaugeValue{
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
								"a_gauge": &loggregator_v2.GaugeValue{
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
		Describe("when selector CounterEvent", func() {
			BeforeEach(func() {
				filterSelector.Filters(FilterSelectorType_COUNTER_EVENT)
			})
			It("should only take counter metric", func() {
				Eventually(metricStore.GetPoints).Should(HaveLen(1))
				point := metricStore.GetPoints()[0]
				Expect(point.MetricName()).To(Equal("failures"))
				Expect(point.Metric().Counter.GetValue()).To(Equal(float64(8)))
				Expect(transform.LabelPairsToLabelsMap(point.Metric().Label)).To(HaveKeyWithValue("source_id", "source-id"))
			})
		})
		Describe("when selector ContainerMetric", func() {
			BeforeEach(func() {
				filterSelector.Filters(FilterSelectorType_CONTAINER_METRIC)
			})
			It("should only take container metric", func() {
				Eventually(metricStore.GetPoints).Should(HaveLen(1))
				point := metricStore.GetPoints()[0]
				Expect(point.MetricName()).To(Equal("cpu"))
				Expect(point.Metric().Gauge.GetValue()).To(Equal(float64(1)))
				Expect(transform.LabelPairsToLabelsMap(point.Metric().Label)).To(HaveKeyWithValue("source_id", "source-id"))
			})
		})
		Describe("when selector ValueMetric", func() {
			BeforeEach(func() {
				filterSelector.Filters(FilterSelectorType_VALUE_METRIC)
			})
			It("should only take gauge metric which is not container metric", func() {
				Eventually(metricStore.GetPoints).Should(HaveLen(1))
				point := metricStore.GetPoints()[0]
				Expect(point.MetricName()).To(Equal("a_gauge"))
				Expect(point.Metric().Gauge.GetValue()).To(Equal(float64(1)))
				Expect(transform.LabelPairsToLabelsMap(point.Metric().Label)).To(HaveKeyWithValue("source_id", "source-id"))
			})
		})
		Describe("when selector Http", func() {
			BeforeEach(func() {
				filterSelector.Filters(FilterSelectorType_HTTP_START_STOP)
			})
			It("should only take timer metric", func() {
				Eventually(metricStore.GetPoints).Should(HaveLen(2))
				pointCounter := metricStore.GetPoints()[0]
				Expect(pointCounter.MetricName()).To(Equal("http_total"))
				Expect(pointCounter.Metric().Counter.GetValue()).To(Equal(float64(1)))
				pointHisto := metricStore.GetPoints()[1]
				Expect(pointHisto.MetricName()).To(Equal("http_duration_seconds"))
				Expect(pointHisto.Metric().Histogram).ToNot(BeNil())
			})
		})
	})

	Context("filter deployment", func() {
		BeforeEach(func() {
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
								"memory": &loggregator_v2.GaugeValue{
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
								"cpu": &loggregator_v2.GaugeValue{
									Unit:  "",
									Value: 1,
								},
							},
						},
					},
				},
			}
		})

		It("should only take metrics from these deployments", func() {
			Eventually(metricStore.GetPoints).Should(HaveLen(2))
			point1 := metricStore.GetPoints()[0]
			Expect(point1.MetricName()).To(Equal("failures"))
			Expect(point1.Metric().Counter.GetValue()).To(Equal(float64(8)))
			Expect(transform.LabelPairsToLabelsMap(point1.Metric().Label)).To(HaveKeyWithValue("source_id", "source-id"))

			point2 := metricStore.GetPoints()[1]
			Expect(point2.MetricName()).To(Equal("memory"))
			Expect(point2.Metric().Gauge.GetValue()).To(Equal(float64(1)))
			Expect(transform.LabelPairsToLabelsMap(point2.Metric().Label)).To(HaveKeyWithValue("source_id", "source-id"))
		})
	})
})
