package metrics_test

import (
	"time"

	"github.com/bosh-prometheus/firehose_exporter/transform"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	dto "github.com/prometheus/client_model/go"

	"github.com/bosh-prometheus/firehose_exporter/metrics"
)

var _ = Describe("RawMetric", func() {
	Context("EstimateMetricSize", func() {
		It("should give an estimate metric size based on timestamp, value and label", func() {
			m := metrics.NewRawMetric("my_metric", "my-origin", &dto.Metric{
				Label: transform.LabelsMapToLabelPairs(map[string]string{
					"origin": "my-origin",
				}),
				TimestampMs: nil,
			})

			Expect(m.EstimateMetricSize()).To(Equal(23))

			m = metrics.NewRawMetric("my_metric", "my-origin", &dto.Metric{
				Label: transform.LabelsMapToLabelPairs(map[string]string{
					"origin": "my-origin",
				}),
				TimestampMs: proto.Int64(0),
			})
			Expect(m.EstimateMetricSize()).To(Equal(23 + 8))
		})
	})

	Context("Set ExpireIn", func() {
		It("should mark as swept when expire time is passed", func() {
			m := metrics.NewRawMetric("my_metric", "my-origin", &dto.Metric{
				Label: transform.LabelsMapToLabelPairs(map[string]string{
					"origin": "my-origin",
				}),
			})
			m.ExpireIn(100 * time.Millisecond)
			Expect(m.IsSwept()).To(BeFalse())
			time.Sleep(101 * time.Millisecond)
			Expect(m.IsSwept()).To(BeTrue())
		})
	})

	Context("Id", func() {
		It("should generate an id based on metric labels", func() {
			m1 := metrics.NewRawMetric("my_metric", "my-origin", &dto.Metric{
				Label: transform.LabelsMapToLabelPairs(map[string]string{
					"origin":   "my-origin",
					"variadic": "1",
				}),
			})
			m2 := metrics.NewRawMetric("my_metric", "my-origin", &dto.Metric{
				Label: transform.LabelsMapToLabelPairs(map[string]string{
					"origin":   "my-origin",
					"variadic": "1",
				}),
			})
			m3 := metrics.NewRawMetric("my_metric", "my-origin", &dto.Metric{
				Label: transform.LabelsMapToLabelPairs(map[string]string{
					"origin":   "my-origin",
					"variadic": "2",
				}),
			})

			Expect(m1.Id()).To(Equal(m2.Id()))
			Expect(m1.Id()).ToNot(Equal(m3.Id()))
		})
	})
})
