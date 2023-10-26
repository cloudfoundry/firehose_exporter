package transform_test

import (
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	"github.com/bosh-prometheus/firehose_exporter/transform"
)

var _ = ginkgo.Describe("NormalizeName", func() {
	ginkgo.It("normalizes a name", func() {
		gomega.Expect(transform.NormalizeName("This_is_-_a-MetricName.Example/-._/._with:0_-._/totals")).To(gomega.Equal("this_is___a_metric_name_example_______with_0_____totals"))
	})
})

var _ = ginkgo.Describe("NormalizeNameDesc", func() {
	ginkgo.It("normalizes a name description", func() {
		gomega.Expect(transform.NormalizeNameDesc("/p.This_is_-_a-MetricName.Example/with:0totals")).To(gomega.Equal("/p-This_is_-_a-MetricName.Example/with:0totals"))
	})
})

var _ = ginkgo.Describe("NormalizeOriginDesc", func() {
	ginkgo.It("normalizes a description", func() {
		gomega.Expect(transform.NormalizeOriginDesc("This_is-a.Desc.Example")).To(gomega.Equal("This_is-a-Desc-Example"))
	})
})
