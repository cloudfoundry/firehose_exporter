package utils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/mjseid/firehose_exporter/utils"
)

var _ = Describe("NormalizeName", func() {
	It("normalizes a name", func() {
		Expect(NormalizeName("This_is_-_a-MetricName.Example/with:0totals")).To(Equal("this_is_a_metric_name_example_with_0_totals"))
	})
})

var _ = Describe("NormalizeDesc", func() {
	It("normalizes a description", func() {
		Expect(NormalizeDesc("This_is-a.Desc.Example")).To(Equal("This_is-a-Desc-Example"))
	})
})
