package rollup_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRollup(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Rollup Suite")
}
