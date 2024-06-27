package metrics_test

import (
	"time"

	"github.com/cloudfoundry/firehose_exporter/transform"
	"github.com/gogo/protobuf/proto"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	dto "github.com/prometheus/client_model/go"

	"github.com/cloudfoundry/firehose_exporter/metrics"
)

var _ = ginkgo.Describe("RawMetric", func() {
	ginkgo.Context("EstimateMetricSize", func() {
		ginkgo.It("should give an estimate metric size based on timestamp, value and label", func() {
			m := metrics.NewRawMetric("my_metric", "my-origin", &dto.Metric{
				Label: transform.LabelsMapToLabelPairs(map[string]string{
					"origin": "my-origin",
				}),
				TimestampMs: nil,
			})

			gomega.Expect(m.EstimateMetricSize()).To(gomega.Equal(23))

			m = metrics.NewRawMetric("my_metric", "my-origin", &dto.Metric{
				Label: transform.LabelsMapToLabelPairs(map[string]string{
					"origin": "my-origin",
				}),
				TimestampMs: proto.Int64(0),
			})
			gomega.Expect(m.EstimateMetricSize()).To(gomega.Equal(23 + 8))
		})
	})

	ginkgo.Context("Set ExpireIn", func() {
		ginkgo.It("should mark as swept when expire time is passed", func() {
			m := metrics.NewRawMetric("my_metric", "my-origin", &dto.Metric{
				Label: transform.LabelsMapToLabelPairs(map[string]string{
					"origin": "my-origin",
				}),
			})
			m.ExpireIn(100 * time.Millisecond)
			gomega.Expect(m.IsSwept()).To(gomega.BeFalse())
			time.Sleep(101 * time.Millisecond)
			gomega.Expect(m.IsSwept()).To(gomega.BeTrue())
		})
	})

	ginkgo.Context("Id", func() {
		ginkgo.It("should generate an id based on metric labels", func() {
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

			gomega.Expect(m1.ID()).To(gomega.Equal(m2.ID()))
			gomega.Expect(m1.ID()).ToNot(gomega.Equal(m3.ID()))
		})
	})
})
