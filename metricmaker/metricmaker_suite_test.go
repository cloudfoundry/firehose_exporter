package metricmaker_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMetricmaker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Metricate Suite")
}

