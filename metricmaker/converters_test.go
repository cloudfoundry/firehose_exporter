package metricmaker_test

import (
	"github.com/bosh-prometheus/firehose_exporter/metricmaker"
	"github.com/bosh-prometheus/firehose_exporter/transform"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Converters", func() {
	BeforeEach(func() {
		metricmaker.SetMetricConverters(make([]metricmaker.MetricConverter, 0))
	})

	Describe("NormalizeName", func() {
		It("should reformat name", func() {
			m := metricmaker.NewRawMetricGauge("FooBar", make(map[string]string), 0)
			metricmaker.NormalizeName(m)
			Expect(m.MetricName()).To(Equal("foo_bar"))

		})
	})

	Describe("AddNamespace", func() {
		It("should prefix with namespace given", func() {
			m := metricmaker.NewRawMetricGauge("my_metric", make(map[string]string), 0)
			metricmaker.AddNamespace("namespace")(m)
			Expect(m.MetricName()).To(Equal("namespace_my_metric"))
		})
	})

	Describe("FindAndReplaceByName", func() {
		It("should only replace name if found", func() {
			m := metricmaker.NewRawMetricGauge("my_metric", make(map[string]string), 0)
			metricmaker.FindAndReplaceByName("foo", "bar")(m)
			Expect(m.MetricName()).To(Equal("my_metric"))

			m = metricmaker.NewRawMetricGauge("foo", make(map[string]string), 0)
			metricmaker.FindAndReplaceByName("foo", "bar")(m)
			Expect(m.MetricName()).To(Equal("bar"))
		})
	})

	Describe("InjectMapLabel", func() {
		It("should inject label given", func() {
			m := metricmaker.NewRawMetricGauge("my_metric", make(map[string]string), 0)
			metricmaker.InjectMapLabel(map[string]string{
				"foo": "bar",
			})(m)
			Expect(m.Metric().Label).To(HaveLen(1))
			Expect(m.Metric().Label[0].GetName()).To(Equal("foo"))
			Expect(m.Metric().Label[0].GetValue()).To(Equal("bar"))
		})
	})

	Describe("PresetLabels", func() {
		It("should rewrite labels", func() {
			m := metricmaker.NewRawMetricGauge("my_metric", make(map[string]string), 0)
			m.Metric().Label = transform.LabelsMapToLabelPairs(map[string]string{
				"deployment": "deployment",
				"job":        "job",
				"index":      "0",
				"ip":         "127.0.0.1",
			})
			metricmaker.PresetLabels(m)
			Expect(m.Metric().Label).To(HaveLen(8))
			labelsMap := transform.LabelPairsToLabelsMap(m.Metric().Label)

			Expect(labelsMap).To(HaveKeyWithValue("bosh_deployment", "deployment"))
			Expect(labelsMap).To(HaveKeyWithValue("bosh_job_name", "job"))
			Expect(labelsMap).To(HaveKeyWithValue("bosh_job_id", "0"))
			Expect(labelsMap).To(HaveKeyWithValue("bosh_job_ip", "127.0.0.1"))
		})
	})

	Describe("OrderAndSanitizeLabels", func() {
		It("should order and sanitize labels", func() {
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
			Expect(m.Metric().Label).To(HaveLen(5))
			Expect(labels[0].GetName()).To(Equal("deployment"))
			Expect(labels[1].GetName()).To(Equal("index"))
			Expect(labels[2].GetName()).To(Equal("ip"))
			Expect(labels[3].GetName()).To(Equal("job"))
			Expect(labels[4].GetName()).To(Equal("label_dash"))
		})
	})

	Describe("SuffixCounterWithTotal", func() {
		It("should suffix only counter metrics without _total suffix with it", func() {
			m := metricmaker.NewRawMetricCounter("my_metric", make(map[string]string), 0)
			metricmaker.SuffixCounterWithTotal(m)
			Expect(m.MetricName()).To(Equal("my_metric_total"))

			m = metricmaker.NewRawMetricGauge("my_metric", make(map[string]string), 0)
			metricmaker.SuffixCounterWithTotal(m)
			Expect(m.MetricName()).To(Equal("my_metric"))

			m = metricmaker.NewRawMetricGauge("my_metric_total", make(map[string]string), 0)
			metricmaker.SuffixCounterWithTotal(m)
			Expect(m.MetricName()).To(Equal("my_metric_total"))
		})
	})
})
