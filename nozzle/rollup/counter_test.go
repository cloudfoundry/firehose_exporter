package rollup_test

import (
	"time"

	"github.com/bosh-prometheus/firehose_exporter/metrics"
	"github.com/bosh-prometheus/firehose_exporter/nozzle/rollup"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Counter Rollup", func() {
	extract := func(batches []*rollup.PointsBatch) []*metrics.RawMetric {
		var points []*metrics.RawMetric
		for _, b := range batches {
			points = append(points, b.Points...)
		}
		return points
	}

	ginkgo.It("returns counters for rolled up events", func() {
		counterRollup := rollup.NewCounterRollup(
			"0",
			nil,
		)

		counterRollup.Record(
			"source-id",
			nil,
			1,
		)

		points := extract(counterRollup.Rollup(0))
		gomega.Expect(len(points)).To(gomega.Equal(1))
		gomega.Expect(*points[0].Metric().Counter.Value).To(gomega.BeNumerically("==", 1))
	})

	ginkgo.It("returns points that track a running total of rolled up events", func() {
		counterRollup := rollup.NewCounterRollup(
			"0",
			[]string{"included-tag"},
		)

		counterRollup.Record(
			"source-id",
			map[string]string{"included-tag": "foo", "excluded-tag": "bar"},
			1,
		)

		points := extract(counterRollup.Rollup(0))
		gomega.Expect(len(points)).To(gomega.Equal(1))
		gomega.Expect(*points[0].Metric().Counter.Value).To(gomega.BeNumerically("==", 1))

		counterRollup.Record(
			"source-id",
			map[string]string{"included-tag": "foo", "excluded-tag": "bar"},
			1,
		)

		points = extract(counterRollup.Rollup(1))
		gomega.Expect(len(points)).To(gomega.Equal(1))
		gomega.Expect(*points[0].Metric().Counter.Value).To(gomega.BeNumerically("==", float64(2)))
	})

	ginkgo.It("returns separate counters for distinct source IDs", func() {
		counterRollup := rollup.NewCounterRollup(
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
		gomega.Expect(len(points)).To(gomega.Equal(2))
	})

	ginkgo.It("returns separate counters for different included tags", func() {
		counterRollup := rollup.NewCounterRollup(
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
		gomega.Expect(len(points)).To(gomega.Equal(2))
		gomega.Expect(*points[0].Metric().Counter.Value).To(gomega.BeNumerically("==", 1))
		gomega.Expect(*points[1].Metric().Counter.Value).To(gomega.BeNumerically("==", 1))
	})

	ginkgo.It("does not return separate counters for different excluded tags", func() {
		counterRollup := rollup.NewCounterRollup(
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
		gomega.Expect(len(points)).To(gomega.Equal(1))
		gomega.Expect(*points[0].Metric().Counter.Value).To(gomega.BeNumerically("==", float64(2)))
	})

	ginkgo.Context("CleanPeriodic", func() {
		ginkgo.It("should clean metrics after amount of time", func() {
			counterRollup := rollup.NewCounterRollup(
				"0",
				nil,
				rollup.SetCounterCleaning(10*time.Millisecond, 50*time.Millisecond),
			)

			counterRollup.Record(
				"source-id",
				nil,
				1,
			)
			time.Sleep(100 * time.Millisecond)
			points := extract(counterRollup.Rollup(0))
			gomega.Expect(len(points)).To(gomega.Equal(0))
		})
	})
})
