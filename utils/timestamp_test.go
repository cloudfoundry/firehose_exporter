package utils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/bosh-prometheus/firehose_exporter/utils"
)

var _ = Describe("NanosecondsToSeconds", func() {
	It("converts nanoseconds to seconds", func() {
		Expect(NanosecondsToSeconds(int64(1000000000))).To(Equal(float64(1)))
	})
})
