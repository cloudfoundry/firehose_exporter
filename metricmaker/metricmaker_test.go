package metricmaker_test

import (
	"code.cloudfoundry.org/go-loggregator/v8/rpc/loggregator_v2"
	"github.com/bosh-prometheus/firehose_exporter/metricmaker"
	"github.com/bosh-prometheus/firehose_exporter/metrics"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("MetricMaker", func() {
	ginkgo.BeforeEach(func() {
		metricmaker.SetMetricConverters(make([]metricmaker.MetricConverter, 0))
	})
	ginkgo.Context("NewRawMetricCounter", func() {
		ginkgo.It("should create raw metric with counter", func() {
			m := metricmaker.NewRawMetricCounter("cpu", map[string]string{
				"origin": "an-origin",
			}, 1)

			gomega.Expect(m.MetricName()).To(gomega.Equal("cpu"))
			gomega.Expect(m.Origin()).To(gomega.Equal("an-origin"))
			gomega.Expect(m.Metric().Counter).ToNot(gomega.BeNil())
			gomega.Expect(m.Metric().Counter.GetValue()).To(gomega.Equal(1.0))
		})
	})

	ginkgo.Context("NewRawMetricCounter", func() {
		ginkgo.It("should create raw metric with counter", func() {
			m := metricmaker.NewRawMetricGauge("cpu", map[string]string{
				"origin": "an-origin",
			}, 1)

			gomega.Expect(m.MetricName()).To(gomega.Equal("cpu"))
			gomega.Expect(m.Origin()).To(gomega.Equal("an-origin"))
			gomega.Expect(m.Metric().Gauge).ToNot(gomega.BeNil())
			gomega.Expect(m.Metric().Gauge.GetValue()).To(gomega.Equal(1.0))
		})
	})

	ginkgo.Context("NewRawMetricsFromEnvelop", func() {
		ginkgo.Context("envelop is timer", func() {
			ginkgo.It("should give an empty list", func() {
				ms := metricmaker.NewRawMetricsFromEnvelop(&loggregator_v2.Envelope{
					Timestamp:      0,
					SourceId:       "",
					InstanceId:     "",
					DeprecatedTags: nil,
					Tags:           nil,
					Message:        &loggregator_v2.Envelope_Timer{},
				})

				gomega.Expect(ms).To(gomega.HaveLen(0))
			})
		})

		ginkgo.Context("envelop is counter", func() {
			ginkgo.It("should give metric associated", func() {
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

				gomega.Expect(ms).To(gomega.HaveLen(1))
				m := ms[0]
				gomega.Expect(m.MetricName()).To(gomega.Equal("my_metric"))
				gomega.Expect(m.Origin()).To(gomega.Equal("my-origin"))
				metricDto := m.Metric()
				gomega.Expect(metricDto).ToNot(gomega.BeNil())
				gomega.Expect(metricDto.Counter).ToNot(gomega.BeNil())
				gomega.Expect(metricDto.Counter.GetValue()).To(gomega.Equal(1.0))

				gomega.Expect(metricDto.Label[0].GetName()).To(gomega.Equal("instance_id"))
				gomega.Expect(metricDto.Label[0].GetValue()).To(gomega.Equal("my-instance"))
				gomega.Expect(metricDto.Label[1].GetName()).To(gomega.Equal("origin"))
				gomega.Expect(metricDto.Label[1].GetValue()).To(gomega.Equal("my-origin"))
				gomega.Expect(metricDto.Label[2].GetName()).To(gomega.Equal("source_id"))
				gomega.Expect(metricDto.Label[2].GetValue()).To(gomega.Equal("source-id"))

			})
		})

		ginkgo.Context("envelop is gauge", func() {
			ginkgo.It("should give metric associated", func() {
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
								"my_metric_1": {
									Unit:  "bytes",
									Value: 1,
								},
								"my_metric_2": {
									Unit:  "bytes",
									Value: 1,
								},
							},
						},
					},
				})

				gomega.Expect(ms).To(gomega.HaveLen(2))
				// force reorder
				if ms[0].MetricName() != "my_metric_1" {
					ms = []*metrics.RawMetric{ms[1], ms[0]}
				}
				gomega.Expect(ms[0].MetricName()).To(gomega.Equal("my_metric_1"))
				gomega.Expect(ms[0].Origin()).To(gomega.Equal("my-origin"))
				gomega.Expect(ms[1].MetricName()).To(gomega.Equal("my_metric_2"))
				gomega.Expect(ms[1].Origin()).To(gomega.Equal("my-origin"))

				m := ms[0]
				metricDto := m.Metric()
				gomega.Expect(metricDto).ToNot(gomega.BeNil())
				gomega.Expect(metricDto.Gauge).ToNot(gomega.BeNil())
				gomega.Expect(metricDto.Gauge.GetValue()).To(gomega.Equal(1.0))

				gomega.Expect(metricDto.Label[0].GetName()).To(gomega.Equal("instance_id"))
				gomega.Expect(metricDto.Label[0].GetValue()).To(gomega.Equal("my-instance"))
				gomega.Expect(metricDto.Label[1].GetName()).To(gomega.Equal("origin"))
				gomega.Expect(metricDto.Label[1].GetValue()).To(gomega.Equal("my-origin"))
				gomega.Expect(metricDto.Label[2].GetName()).To(gomega.Equal("source_id"))
				gomega.Expect(metricDto.Label[2].GetValue()).To(gomega.Equal("source-id"))
				gomega.Expect(metricDto.Label[3].GetName()).To(gomega.Equal("unit"))
				gomega.Expect(metricDto.Label[3].GetValue()).To(gomega.Equal("bytes"))

			})
		})

	})

})
