package filters_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry-community/firehose_exporter/filters"
)

var _ = Describe("DeploymentFilter", func() {
	var (
		filter           []string
		deploymentFilter *DeploymentFilter
	)

	BeforeEach(func() {
		filter = []string{"fake-deployment-1", "fake-deployment-3"}
	})

	JustBeforeEach(func() {
		deploymentFilter = NewDeploymentFilter(filter)
	})

	Describe("Enabled", func() {
		Context("when deployment is enabled", func() {
			It("returns true", func() {
				Expect(deploymentFilter.Enabled("fake-deployment-1")).To(BeTrue())
			})
		})

		Context("when deployment is not enabled", func() {
			It("returns false", func() {
				Expect(deploymentFilter.Enabled("fake-deployment-2")).To(BeFalse())
			})
		})

		Context("when there is no filter", func() {
			BeforeEach(func() {
				filter = []string{}
			})

			It("returns true", func() {
				Expect(deploymentFilter.Enabled("fake-deployment-2")).To(BeTrue())
			})
		})

		Context("when a filter has leading and/or trailing whitespaces", func() {
			BeforeEach(func() {
				filter = []string{"   fake-deployment-1  "}
			})

			It("returns true", func() {
				Expect(deploymentFilter.Enabled("fake-deployment-1")).To(BeTrue())
			})
		})
	})
})
