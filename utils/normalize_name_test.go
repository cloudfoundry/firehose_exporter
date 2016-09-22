package utils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry-community/firehose_exporter/utils"
)

var _ = Describe("NormalizeName", func() {
	It("normalizes a name", func() {
		Expect(NormalizeName("This_is_a-MetricName.Example")).To(Equal("this_is_a_metric_name_example"))
	})
})
