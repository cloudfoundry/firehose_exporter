package utils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"

	. "github.com/cloudfoundry-community/firehose_exporter/utils"
)

var _ = Describe("UUID", func() {
	var (
		id   = "aea56fa9-72a5-44bf-5f86-00d33019a593"
		uuid = &events.UUID{
			Low:  proto.Uint64(13782322671548081582),
			High: proto.Uint64(10638937392221816415),
		}
	)

	Describe("UUIDToString", func() {
		It("converts an UUID to a string", func() {
			Expect(UUIDToString(uuid)).To(Equal(id))
		})
	})

	Describe("StringToUUID", func() {
		It("converts a string to an UUID", func() {
			Expect(StringToUUID(id)).To(Equal(uuid))
		})
	})
})
