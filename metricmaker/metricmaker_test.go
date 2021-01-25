package metricmaker_test

import (
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/bosh-prometheus/firehose_exporter/metricmaker"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MetricMaker", func() {
	BeforeEach(func() {
		metricmaker.SetMetricConverters(make([]metricmaker.MetricConverter, 0))
	})
	Context("NewRawMetricCounter", func() {
		It("should create raw metric with counter", func() {
			m := metricmaker.NewRawMetricCounter("cpu", map[string]string{
				"origin": "an-origin",
			}, 1)

			Expect(m.MetricName()).To(Equal("cpu"))
			Expect(m.Origin()).To(Equal("an-origin"))
			Expect(m.Metric().Counter).ToNot(BeNil())
			Expect(m.Metric().Counter.GetValue()).To(Equal(1.0))
		})
	})

	Context("NewRawMetricCounter", func() {
		It("should create raw metric with counter", func() {
			m := metricmaker.NewRawMetricGauge("cpu", map[string]string{
				"origin": "an-origin",
			}, 1)

			Expect(m.MetricName()).To(Equal("cpu"))
			Expect(m.Origin()).To(Equal("an-origin"))
			Expect(m.Metric().Gauge).ToNot(BeNil())
			Expect(m.Metric().Gauge.GetValue()).To(Equal(1.0))
		})
	})

	Context("NewRawMetricsFromEnvelop", func() {
		Context("envelop is timer", func() {
			It("should give an empty list", func() {
				ms := metricmaker.NewRawMetricsFromEnvelop(&loggregator_v2.Envelope{
					Timestamp:      0,
					SourceId:       "",
					InstanceId:     "",
					DeprecatedTags: nil,
					Tags:           nil,
					Message:        &loggregator_v2.Envelope_Timer{},
				})

				Expect(ms).To(HaveLen(0))
			})
		})

		Context("envelop is counter", func() {
			It("should give metric associated", func() {
				ms := metricmaker.NewRawMetricsFromEnvelop(&loggregator_v2.Envelope{
					Timestamp:      0,
					SourceId:       "source-id",
					InstanceId:     "my-instance",
					DeprecatedTags: nil,
					Tags: map[string]string{
						"origin": "my-origin",
					},
					Message: &loggregator_v2.Envelope_Counter{
						Counter: &loggregator_v2.Counter{
							Name:  "my_metric",
							Delta: 0,
							Total: 1,
						},
					},
				})

				Expect(ms).To(HaveLen(1))
				m := ms[0]
				Expect(m.MetricName()).To(Equal("my_metric"))
				Expect(m.Origin()).To(Equal("my-origin"))
				metricDto := m.Metric()
				Expect(metricDto).ToNot(BeNil())
				Expect(metricDto.Counter).ToNot(BeNil())
				Expect(metricDto.Counter.GetValue()).To(Equal(1.0))

				Expect(metricDto.Label[0].GetName()).To(Equal("instance_id"))
				Expect(metricDto.Label[0].GetValue()).To(Equal("my-instance"))
				Expect(metricDto.Label[1].GetName()).To(Equal("origin"))
				Expect(metricDto.Label[1].GetValue()).To(Equal("my-origin"))
				Expect(metricDto.Label[2].GetName()).To(Equal("source_id"))
				Expect(metricDto.Label[2].GetValue()).To(Equal("source-id"))

			})
		})

		Context("envelop is gauge", func() {
			It("should give metric associated", func() {
				ms := metricmaker.NewRawMetricsFromEnvelop(&loggregator_v2.Envelope{
					Timestamp:      0,
					SourceId:       "source-id",
					InstanceId:     "my-instance",
					DeprecatedTags: nil,
					Tags: map[string]string{
						"origin": "my-origin",
					},
					Message: &loggregator_v2.Envelope_Gauge{
						Gauge: &loggregator_v2.Gauge{
							Metrics: map[string]*loggregator_v2.GaugeValue{
								"my_metric_1": &loggregator_v2.GaugeValue{
									Unit:  "bytes",
									Value: 1,
								},
								"my_metric_2": &loggregator_v2.GaugeValue{
									Unit:  "bytes",
									Value: 1,
								},
							},
						},
					},
				})

				Expect(ms).To(HaveLen(2))
				Expect(ms[0].MetricName()).To(Equal("my_metric_1"))
				Expect(ms[0].Origin()).To(Equal("my-origin"))
				Expect(ms[1].MetricName()).To(Equal("my_metric_2"))
				Expect(ms[1].Origin()).To(Equal("my-origin"))

				m := ms[0]
				metricDto := m.Metric()
				Expect(metricDto).ToNot(BeNil())
				Expect(metricDto.Gauge).ToNot(BeNil())
				Expect(metricDto.Gauge.GetValue()).To(Equal(1.0))

				Expect(metricDto.Label[0].GetName()).To(Equal("instance_id"))
				Expect(metricDto.Label[0].GetValue()).To(Equal("my-instance"))
				Expect(metricDto.Label[1].GetName()).To(Equal("origin"))
				Expect(metricDto.Label[1].GetValue()).To(Equal("my-origin"))
				Expect(metricDto.Label[2].GetName()).To(Equal("source_id"))
				Expect(metricDto.Label[2].GetValue()).To(Equal("source-id"))
				Expect(metricDto.Label[3].GetName()).To(Equal("unit"))
				Expect(metricDto.Label[3].GetValue()).To(Equal("bytes"))

			})
		})

	})

})
