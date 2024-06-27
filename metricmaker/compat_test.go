package metricmaker_test

import (
	"github.com/cloudfoundry/firehose_exporter/metricmaker"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Compat", func() {
	ginkgo.BeforeEach(func() {
		metricmaker.SetMetricConverters(make([]metricmaker.MetricConverter, 0))
	})
	ginkgo.Describe("RetroCompatMetricNames", func() {
		ginkgo.Context("when have a container metric", func() {
			ginkgo.It("should rename metric name with old name", func() {
				m := metricmaker.NewRawMetricGauge("cpu", make(map[string]string), 0)
				metricmaker.RetroCompatMetricNames(m)
				gomega.Expect(m.MetricName()).To(gomega.Equal("container_metric_cpu_percentage"))

				m = metricmaker.NewRawMetricGauge("memory", make(map[string]string), 0)
				metricmaker.RetroCompatMetricNames(m)
				gomega.Expect(m.MetricName()).To(gomega.Equal("container_metric_memory_bytes"))

				m = metricmaker.NewRawMetricGauge("disk", make(map[string]string), 0)
				metricmaker.RetroCompatMetricNames(m)
				gomega.Expect(m.MetricName()).To(gomega.Equal("container_metric_disk_bytes"))

				m = metricmaker.NewRawMetricGauge("memory_quota", make(map[string]string), 0)
				metricmaker.RetroCompatMetricNames(m)
				gomega.Expect(m.MetricName()).To(gomega.Equal("container_metric_memory_bytes_quota"))

				m = metricmaker.NewRawMetricGauge("disk_quota", make(map[string]string), 0)
				metricmaker.RetroCompatMetricNames(m)
				gomega.Expect(m.MetricName()).To(gomega.Equal("container_metric_disk_bytes_quota"))
			})
		})
		ginkgo.Context("when have a counter metric which is not a container metric", func() {
			ginkgo.It("should prefix with counter_event_ add origin and suffix with total", func() {
				m := metricmaker.NewRawMetricCounter("my_metric", make(map[string]string), 0)
				m.SetOrigin("origin")
				metricmaker.RetroCompatMetricNames(m)
				gomega.Expect(m.MetricName()).To(gomega.Equal("counter_event_origin_my_metric_total"))
			})
		})
		ginkgo.Context("when have a gauge metric which is not a container metric", func() {
			ginkgo.It("should prefix with value_metric_ add origin", func() {
				m := metricmaker.NewRawMetricGauge("my_metric", make(map[string]string), 0)
				m.SetOrigin("origin")
				metricmaker.RetroCompatMetricNames(m)
				gomega.Expect(m.MetricName()).To(gomega.Equal("value_metric_origin_my_metric"))
			})
		})
	})
})
