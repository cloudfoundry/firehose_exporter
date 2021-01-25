package transform_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/bosh-prometheus/firehose_exporter/transform"
)

var _ = Describe("NormalizeName", func() {
	It("normalizes a name", func() {
		Expect(NormalizeName("This_is_-_a-MetricName.Example/-._/._with:0_-._/totals")).To(Equal("this_is_a_metric_name_example_with_0_totals"))
	})
})

var _ = Describe("NormalizeNameDesc", func() {
	It("normalizes a name description", func() {
		Expect(NormalizeNameDesc("/p.This_is_-_a-MetricName.Example/with:0totals")).To(Equal("/p-This_is_-_a-MetricName.Example/with:0totals"))
	})
})

var _ = Describe("NormalizeOriginDesc", func() {
	It("normalizes a description", func() {
		Expect(NormalizeOriginDesc("This_is-a.Desc.Example")).To(Equal("This_is-a-Desc-Example"))
	})
})
