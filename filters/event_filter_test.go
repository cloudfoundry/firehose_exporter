package filters_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/sonde-go/events"

	. "github.com/cloudfoundry-community/firehose_exporter/filters"
)

var _ = Describe("EventFilter", func() {
	var (
		err    error
		filter []string

		eventFilter *EventFilter
	)

	JustBeforeEach(func() {
		eventFilter, err = NewEventFilter(filter)
	})

	Describe("New", func() {
		Context("when filters are supported", func() {
			BeforeEach(func() {
				filter = []string{"ContainerMetric", "CounterEvent", "HttpStartStop", "ValueMetric"}
			})

			It("does not return an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when filters are not supported", func() {
			BeforeEach(func() {
				filter = []string{"ContainerMetric", "CounterEvent", "LogMessage"}
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Event filter `LogMessage` is not supported"))
			})
		})
	})

	Describe("Enabled", func() {
		var (
			counterEventEnvelope = &events.Envelope{
				EventType: events.Envelope_CounterEvent.Enum(),
			}

			valueMetricEnvelope = &events.Envelope{
				EventType: events.Envelope_ValueMetric.Enum(),
			}
		)

		BeforeEach(func() {
			filter = []string{"ValueMetric"}
		})

		Context("when event is enabled", func() {
			It("returns true", func() {
				Expect(eventFilter.Enabled(valueMetricEnvelope)).To(BeTrue())
			})
		})

		Context("when event is not enabled", func() {
			It("returns false", func() {
				Expect(eventFilter.Enabled(counterEventEnvelope)).To(BeFalse())
			})
		})

		Context("when there is no filter", func() {
			BeforeEach(func() {
				filter = []string{}
			})

			It("returns true", func() {
				Expect(eventFilter.Enabled(counterEventEnvelope)).To(BeTrue())
			})
		})
	})
})
