package rollup_test

import (
	"fmt"
	"time"

	"github.com/bosh-prometheus/firehose_exporter/metrics"
	"github.com/bosh-prometheus/firehose_exporter/nozzle/rollup"
	"github.com/bosh-prometheus/firehose_exporter/transform"
	dto "github.com/prometheus/client_model/go"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

const (
	HTTPDurationName = "http_duration_seconds"
)

type histogram struct {
	points []*metrics.RawMetric
}

func (h *histogram) Count() int {
	for _, p := range h.points {
		if p.MetricName() == HTTPDurationName {
			return int(*p.Metric().Histogram.SampleCount)
		}
	}
	ginkgo.Fail("No count point found in histogram")
	return 0
}

func (h *histogram) Sum() int {
	for _, p := range h.points {
		if p.MetricName() == HTTPDurationName {
			return int(*p.Metric().Histogram.SampleSum)
		}
	}
	ginkgo.Fail("No sum point found in histogram")
	return 0
}

func (h *histogram) Points() []*metrics.RawMetric {
	return h.points
}

func (h *histogram) Bucket(le string) *dto.Histogram {
	for _, p := range h.points {
		if p.MetricName() != HTTPDurationName {
			continue
		}
		for _, label := range p.Metric().Label {
			if label.GetName() == "le" && label.GetValue() == le {
				return p.Metric().Histogram
			}
		}
	}
	ginkgo.Fail(fmt.Sprintf("No bucket point found in histogram for le = '%s'", le))
	return nil
}

var _ = ginkgo.Describe("Histogram Rollup", func() {
	extract := func(batches []*rollup.PointsBatch) []*histogram {
		var histograms []*histogram

		for _, b := range batches {
			h := &histogram{}
			h.points = append(h.points, b.Points...)
			histograms = append(histograms, h)
		}

		return histograms
	}

	ginkgo.It("returns aggregate information for rolled up events", func() {
		rollup := rollup.NewHistogramRollup(
			"0",
			nil,
		)

		rollup.Record(
			"source-id",
			nil,
			10*int64(time.Second),
		)
		rollup.Record(
			"source-id",
			nil,
			5*int64(time.Second),
		)

		histograms := extract(rollup.Rollup(0))
		gomega.Expect(len(histograms)).To(gomega.Equal(1))
		gomega.Expect(histograms[0].Count()).To(gomega.Equal(2))
		gomega.Expect(histograms[0].Sum()).To(gomega.Equal(15))
	})

	ginkgo.It("returns batches which each includes a size estimate", func() {
		rollup := rollup.NewHistogramRollup(
			"0",
			nil,
		)

		rollup.Record(
			"source-id",
			nil,
			10*int64(time.Second),
		)

		pointsBatches := rollup.Rollup(0)
		gomega.Expect(len(pointsBatches)).To(gomega.Equal(1))
		gomega.Expect(pointsBatches[0].Size).To(gomega.BeNumerically(">", 0))
	})

	ginkgo.It("returns points for each bucket in the histogram", func() {
		rollup := rollup.NewHistogramRollup(
			"0",
			nil,
		)

		rollup.Record(
			"source-id",
			nil,
			2*int64(time.Second),
		)
		rollup.Record(
			"source-id",
			nil,
			7*int64(time.Second),
		)
		rollup.Record(
			"source-id",
			nil,
			8*int64(time.Second),
		)

		histograms := extract(rollup.Rollup(0))
		gomega.Expect(len(histograms)).To(gomega.Equal(1))
	})

	ginkgo.It("returns points with the timestamp given to Rollup", func() {
		rollup := rollup.NewHistogramRollup(
			"node-index",
			nil,
		)

		rollup.Record(
			"source-id",
			nil,
			1,
		)

		histograms := extract(rollup.Rollup(88))
		gomega.Expect(len(histograms)).To(gomega.Equal(1))
	})

	ginkgo.It("returns histograms with labels based on tags", func() {
		rollup := rollup.NewHistogramRollup(
			"node-index",
			[]string{"included-tag"},
		)

		rollup.Record(
			"source-id",
			map[string]string{"included-tag": "foo", "excluded-tag": "bar"},
			1,
		)

		histograms := extract(rollup.Rollup(0))
		gomega.Expect(len(histograms)).To(gomega.Equal(1))
		for _, p := range histograms[0].Points() {
			gomega.Expect(transform.LabelPairsToLabelsMap(p.Metric().Label)).To(gomega.And(
				gomega.HaveKeyWithValue("included_tag", "foo"),
				gomega.HaveKeyWithValue("source_id", "source-id"),
				gomega.HaveKeyWithValue("node_index", "node-index"),
			))
		}
	})

	ginkgo.It("returns points that track a running total of rolled up events", func() {
		rollup := rollup.NewHistogramRollup(
			"0",
			[]string{"included-tag"},
		)

		rollup.Record(
			"source-id",
			map[string]string{"included-tag": "foo", "excluded-tag": "bar"},
			1,
		)

		histograms := extract(rollup.Rollup(0))
		gomega.Expect(len(histograms)).To(gomega.Equal(1))
		gomega.Expect(histograms[0].Count()).To(gomega.Equal(1))

		rollup.Record(
			"source-id",
			map[string]string{"included-tag": "foo", "excluded-tag": "bar"},
			1,
		)

		histograms = extract(rollup.Rollup(1))
		gomega.Expect(len(histograms)).To(gomega.Equal(1))
		gomega.Expect(histograms[0].Count()).To(gomega.Equal(2))
	})

	ginkgo.It("returns separate histograms for distinct source IDs", func() {
		rollup := rollup.NewHistogramRollup(
			"0",
			[]string{"included-tag"},
		)

		rollup.Record(
			"source-id",
			map[string]string{"included-tag": "foo"},
			1,
		)
		rollup.Record(
			"other-source-id",
			map[string]string{"included-tag": "foo"},
			1,
		)

		histograms := extract(rollup.Rollup(0))
		gomega.Expect(len(histograms)).To(gomega.Equal(2))
		gomega.Expect(histograms[0].Count()).To(gomega.Equal(1))
		gomega.Expect(histograms[1].Count()).To(gomega.Equal(1))
	})

	ginkgo.It("returns separate histograms for different included tags", func() {
		rollup := rollup.NewHistogramRollup(
			"0",
			[]string{"included-tag"},
		)

		rollup.Record(
			"source-id",
			map[string]string{"included-tag": "foo"},
			1,
		)
		rollup.Record(
			"source-id",
			map[string]string{"included-tag": "other-foo"},
			1,
		)

		histograms := extract(rollup.Rollup(0))
		gomega.Expect(len(histograms)).To(gomega.Equal(2))
		gomega.Expect(histograms[0].Count()).To(gomega.Equal(1))
		gomega.Expect(histograms[1].Count()).To(gomega.Equal(1))
	})

	ginkgo.It("does not return separate histograms for different excluded tags", func() {
		rollup := rollup.NewHistogramRollup(
			"0",
			[]string{"included-tag"},
		)

		rollup.Record(
			"source-id",
			map[string]string{"excluded-tag": "bar"},
			1,
		)
		rollup.Record(
			"source-id",
			map[string]string{"excluded-tag": "other-bar"},
			1,
		)

		histograms := extract(rollup.Rollup(0))
		gomega.Expect(len(histograms)).To(gomega.Equal(1))
		gomega.Expect(histograms[0].Count()).To(gomega.Equal(2))
		gomega.Expect(transform.LabelPairsToLabelsMap(histograms[0].Points()[0].Metric().Label)).ToNot(gomega.HaveKey("excluded-tag"))
	})

	ginkgo.Context("CleanPeriodic", func() {

		ginkgo.It("should clean metrics after amount of time", func() {
			rollup := rollup.NewHistogramRollup(
				"0",
				nil,
				rollup.SetHistogramCleaning(10*time.Millisecond, 50*time.Millisecond),
			)
			rollup.Record(
				"source-id",
				nil,
				10*int64(time.Second),
			)
			time.Sleep(100 * time.Millisecond)
			histograms := extract(rollup.Rollup(0))
			gomega.Expect(len(histograms)).To(gomega.Equal(0))
		})
	})
})
