package rollup_test

import (
	"github.com/bosh-prometheus/firehose_exporter/metrics"
	. "github.com/bosh-prometheus/firehose_exporter/nozzle/rollup"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Counter Rollup", func() {
	extract := func(batches []*PointsBatch) []*metrics.RawMetric {
		var points []*metrics.RawMetric

		for _, b := range batches {
			for _, p := range b.Points {
				points = append(points, p)
			}
		}

		return points
	}

	It("returns counters for rolled up events", func() {
		counterRollup := NewCounterRollup(
			"0",
			nil,
		)

		counterRollup.Record(
			"source-id",
			nil,
			1,
		)

		points := extract(counterRollup.Rollup(0))
		Expect(len(points)).To(Equal(1))
		Expect(*points[0].Metric().Counter.Value).To(BeNumerically("==", 1))
	})

	It("returns points that track a running total of rolled up events", func() {
		counterRollup := NewCounterRollup(
			"0",
			[]string{"included-tag"},
		)

		counterRollup.Record(
			"source-id",
			map[string]string{"included-tag": "foo", "excluded-tag": "bar"},
			1,
		)

		points := extract(counterRollup.Rollup(0))
		Expect(len(points)).To(Equal(1))
		Expect(*points[0].Metric().Counter.Value).To(BeNumerically("==", 1))

		counterRollup.Record(
			"source-id",
			map[string]string{"included-tag": "foo", "excluded-tag": "bar"},
			1,
		)

		points = extract(counterRollup.Rollup(1))
		Expect(len(points)).To(Equal(1))
		Expect(*points[0].Metric().Counter.Value).To(BeNumerically("==", 2))
	})

	It("returns separate counters for distinct source IDs", func() {
		counterRollup := NewCounterRollup(
			"0",
			[]string{"included-tag"},
		)

		counterRollup.Record(
			"source-id",
			map[string]string{"included-tag": "foo"},
			1,
		)
		counterRollup.Record(
			"other-source-id",
			map[string]string{"included-tag": "foo"},
			1,
		)

		points := extract(counterRollup.Rollup(0))
		Expect(len(points)).To(Equal(2))
	})

	It("returns separate counters for different included tags", func() {
		counterRollup := NewCounterRollup(
			"0",
			[]string{"included-tag"},
		)

		counterRollup.Record(
			"source-id",
			map[string]string{"included-tag": "foo"},
			1,
		)
		counterRollup.Record(
			"source-id",
			map[string]string{"included-tag": "other-foo"},
			1,
		)

		points := extract(counterRollup.Rollup(0))
		Expect(len(points)).To(Equal(2))
		Expect(*points[0].Metric().Counter.Value).To(BeNumerically("==", 1))
		Expect(*points[1].Metric().Counter.Value).To(BeNumerically("==", 1))
	})

	It("does not return separate counters for different excluded tags", func() {
		counterRollup := NewCounterRollup(
			"0",
			[]string{"included-tag"},
		)

		counterRollup.Record(
			"source-id",
			map[string]string{"included-tag": "foo", "excluded-tag": "bar"},
			1,
		)
		counterRollup.Record(
			"source-id",
			map[string]string{"included-tag": "foo", "excluded-tag": "other-bar"},
			1,
		)

		points := extract(counterRollup.Rollup(0))
		Expect(len(points)).To(Equal(1))
		Expect(*points[0].Metric().Counter.Value).To(BeNumerically("==", 2))
	})
})
