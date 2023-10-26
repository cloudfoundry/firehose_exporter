package metricmaker_test

import (
	"github.com/bosh-prometheus/firehose_exporter/metricmaker"
	"github.com/bosh-prometheus/firehose_exporter/transform"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Converters", func() {
	ginkgo.BeforeEach(func() {
		metricmaker.SetMetricConverters(make([]metricmaker.MetricConverter, 0))
	})

	ginkgo.Describe("NormalizeName", func() {
		ginkgo.It("should reformat name", func() {
			m := metricmaker.NewRawMetricGauge("FooBar", make(map[string]string), 0)
			metricmaker.NormalizeName(m)
			gomega.Expect(m.MetricName()).To(gomega.Equal("foo_bar"))

		})
	})

	ginkgo.Describe("AddNamespace", func() {
		ginkgo.It("should prefix with namespace given", func() {
			m := metricmaker.NewRawMetricGauge("my_metric", make(map[string]string), 0)
			metricmaker.AddNamespace("namespace")(m)
			gomega.Expect(m.MetricName()).To(gomega.Equal("namespace_my_metric"))
		})
	})

	ginkgo.Describe("FindAndReplaceByName", func() {
		ginkgo.It("should only replace name if found", func() {
			m := metricmaker.NewRawMetricGauge("my_metric", make(map[string]string), 0)
			metricmaker.FindAndReplaceByName("foo", "bar")(m)
			gomega.Expect(m.MetricName()).To(gomega.Equal("my_metric"))

			m = metricmaker.NewRawMetricGauge("foo", make(map[string]string), 0)
			metricmaker.FindAndReplaceByName("foo", "bar")(m)
			gomega.Expect(m.MetricName()).To(gomega.Equal("bar"))
		})
	})

	ginkgo.Describe("InjectMapLabel", func() {
		ginkgo.It("should inject label given", func() {
			m := metricmaker.NewRawMetricGauge("my_metric", make(map[string]string), 0)
			metricmaker.InjectMapLabel(map[string]string{
				"foo": "bar",
			})(m)
			gomega.Expect(m.Metric().Label).To(gomega.HaveLen(1))
			gomega.Expect(m.Metric().Label[0].GetName()).To(gomega.Equal("foo"))
			gomega.Expect(m.Metric().Label[0].GetValue()).To(gomega.Equal("bar"))
		})
	})

	ginkgo.Describe("PresetLabels", func() {
		ginkgo.It("should rewrite labels", func() {
			m := metricmaker.NewRawMetricGauge("my_metric", make(map[string]string), 0)
			m.Metric().Label = transform.LabelsMapToLabelPairs(map[string]string{
				"deployment": "deployment",
				"job":        "job",
				"index":      "0",
				"ip":         "127.0.0.1",
			})
			metricmaker.PresetLabels(m)
			gomega.Expect(m.Metric().Label).To(gomega.HaveLen(8))
			labelsMap := transform.LabelPairsToLabelsMap(m.Metric().Label)

			gomega.Expect(labelsMap).To(gomega.HaveKeyWithValue("bosh_deployment", "deployment"))
			gomega.Expect(labelsMap).To(gomega.HaveKeyWithValue("bosh_job_name", "job"))
			gomega.Expect(labelsMap).To(gomega.HaveKeyWithValue("bosh_job_id", "0"))
			gomega.Expect(labelsMap).To(gomega.HaveKeyWithValue("bosh_job_ip", "127.0.0.1"))
		})
	})

	ginkgo.Describe("OrderAndSanitizeLabels", func() {
		ginkgo.It("should order and sanitize labels", func() {
			m := metricmaker.NewRawMetricGauge("my_metric", make(map[string]string), 0)
			m.Metric().Label = transform.LabelsMapToLabelPairs(map[string]string{
				"job":        "job",
				"index":      "0",
				"deployment": "deployment",
				"ip":         "127.0.0.1",
				"__value":    "value",
				"label-dash": "value",
			})
			metricmaker.OrderAndSanitizeLabels(m)

			labels := m.Metric().Label
			gomega.Expect(m.Metric().Label).To(gomega.HaveLen(5))
			gomega.Expect(labels[0].GetName()).To(gomega.Equal("deployment"))
			gomega.Expect(labels[1].GetName()).To(gomega.Equal("index"))
			gomega.Expect(labels[2].GetName()).To(gomega.Equal("ip"))
			gomega.Expect(labels[3].GetName()).To(gomega.Equal("job"))
			gomega.Expect(labels[4].GetName()).To(gomega.Equal("label_dash"))
		})
	})

	ginkgo.Describe("SuffixCounterWithTotal", func() {
		ginkgo.It("should suffix only counter metrics without _total suffix with it", func() {
			m := metricmaker.NewRawMetricCounter("my_metric", make(map[string]string), 0)
			metricmaker.SuffixCounterWithTotal(m)
			gomega.Expect(m.MetricName()).To(gomega.Equal("my_metric_total"))

			m = metricmaker.NewRawMetricGauge("my_metric", make(map[string]string), 0)
			metricmaker.SuffixCounterWithTotal(m)
			gomega.Expect(m.MetricName()).To(gomega.Equal("my_metric"))

			m = metricmaker.NewRawMetricGauge("my_metric_total", make(map[string]string), 0)
			metricmaker.SuffixCounterWithTotal(m)
			gomega.Expect(m.MetricName()).To(gomega.Equal("my_metric_total"))
		})
	})
})
