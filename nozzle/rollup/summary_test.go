package rollup_test

import (
	"fmt"
	"time"

	"github.com/cloudfoundry/firehose_exporter/metrics"
	"github.com/cloudfoundry/firehose_exporter/nozzle/rollup"
	"github.com/cloudfoundry/firehose_exporter/transform"
	dto "github.com/prometheus/client_model/go"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

const (
	HTTPResponseName = "http_response_size_bytes"
)

type summary struct {
	points []*metrics.RawMetric
}

func (h *summary) Count() int {
	for _, p := range h.points {
		if p.MetricName() == HTTPResponseName {
			return int(*p.Metric().Summary.SampleCount)
		}
	}
	ginkgo.Fail("No count point found in summary")
	return 0
}

func (h *summary) Sum() int {
	for _, p := range h.points {
		if p.MetricName() == HTTPResponseName {
			return int(*p.Metric().Summary.SampleSum)
		}
	}
	ginkgo.Fail("No sum point found in summary")
	return 0
}

func (h *summary) Points() []*metrics.RawMetric {
	return h.points
}

func (h *summary) Bucket(le string) *dto.Summary {
	for _, p := range h.points {
		if p.MetricName() != HTTPResponseName {
			continue
		}
		for _, label := range p.Metric().Label {
			if label.GetName() == "le" && label.GetValue() == le {
				return p.Metric().Summary
			}
		}
	}
	ginkgo.Fail(fmt.Sprintf("No bucket point found in summary for le = '%s'", le))
	return nil
}

var _ = ginkgo.Describe("summary Rollup", func() {
	extract := func(batches []*rollup.PointsBatch) []*summary {
		var summaries []*summary

		for _, b := range batches {
			h := &summary{}
			h.points = append(h.points, b.Points...)
			summaries = append(summaries, h)
		}

		return summaries
	}

	ginkgo.It("returns aggregate information for rolled up events", func() {
		rollup := rollup.NewSummaryRollup(
			"0",
			nil,
		)

		rollup.Record(
			"source-id",
			nil,
			10,
		)
		rollup.Record(
			"source-id",
			nil,
			5,
		)

		summaries := extract(rollup.Rollup(0))
		gomega.Expect(len(summaries)).To(gomega.Equal(1))
		gomega.Expect(summaries[0].Count()).To(gomega.Equal(2))
		gomega.Expect(summaries[0].Sum()).To(gomega.Equal(15))
	})

	ginkgo.It("returns batches which each includes a size estimate", func() {
		rollup := rollup.NewSummaryRollup(
			"0",
			nil,
		)

		rollup.Record(
			"source-id",
			nil,
			10,
		)

		pointsBatches := rollup.Rollup(0)
		gomega.Expect(len(pointsBatches)).To(gomega.Equal(1))
		gomega.Expect(pointsBatches[0].Size).To(gomega.BeNumerically(">", 0))
	})

	ginkgo.It("returns points for each bucket in the summary", func() {
		rollup := rollup.NewSummaryRollup(
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

		summaries := extract(rollup.Rollup(0))
		gomega.Expect(len(summaries)).To(gomega.Equal(1))
	})

	ginkgo.It("returns points with the timestamp given to Rollup", func() {
		rollup := rollup.NewSummaryRollup(
			"node-index",
			nil,
		)

		rollup.Record(
			"source-id",
			nil,
			1,
		)

		summaries := extract(rollup.Rollup(88))
		gomega.Expect(len(summaries)).To(gomega.Equal(1))
	})

	ginkgo.It("returns summaries with labels based on tags", func() {
		rollup := rollup.NewSummaryRollup(
			"node-index",
			[]string{"included-tag"},
		)

		rollup.Record(
			"source-id",
			map[string]string{"included-tag": "foo", "excluded-tag": "bar"},
			1,
		)

		summaries := extract(rollup.Rollup(0))
		gomega.Expect(len(summaries)).To(gomega.Equal(1))
		for _, p := range summaries[0].Points() {
			gomega.Expect(transform.LabelPairsToLabelsMap(p.Metric().Label)).To(gomega.And(
				gomega.HaveKeyWithValue("included_tag", "foo"),
				gomega.HaveKeyWithValue("source_id", "source-id"),
				gomega.HaveKeyWithValue("node_index", "node-index"),
			))
		}
	})

	ginkgo.It("returns points that track a running total of rolled up events", func() {
		rollup := rollup.NewSummaryRollup(
			"0",
			[]string{"included-tag"},
		)

		rollup.Record(
			"source-id",
			map[string]string{"included-tag": "foo", "excluded-tag": "bar"},
			1,
		)

		summaries := extract(rollup.Rollup(0))
		gomega.Expect(len(summaries)).To(gomega.Equal(1))
		gomega.Expect(summaries[0].Count()).To(gomega.Equal(1))

		rollup.Record(
			"source-id",
			map[string]string{"included-tag": "foo", "excluded-tag": "bar"},
			1,
		)

		summaries = extract(rollup.Rollup(1))
		gomega.Expect(len(summaries)).To(gomega.Equal(1))
		gomega.Expect(summaries[0].Count()).To(gomega.Equal(2))
	})

	ginkgo.It("returns separate summaries for distinct source IDs", func() {
		rollup := rollup.NewSummaryRollup(
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

		summaries := extract(rollup.Rollup(0))
		gomega.Expect(len(summaries)).To(gomega.Equal(2))
		gomega.Expect(summaries[0].Count()).To(gomega.Equal(1))
		gomega.Expect(summaries[1].Count()).To(gomega.Equal(1))
	})

	ginkgo.It("returns separate summaries for different included tags", func() {
		rollup := rollup.NewSummaryRollup(
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

		summaries := extract(rollup.Rollup(0))
		gomega.Expect(len(summaries)).To(gomega.Equal(2))
		gomega.Expect(summaries[0].Count()).To(gomega.Equal(1))
		gomega.Expect(summaries[1].Count()).To(gomega.Equal(1))
	})

	ginkgo.It("does not return separate summaries for different excluded tags", func() {
		rollup := rollup.NewSummaryRollup(
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

		summaries := extract(rollup.Rollup(0))
		gomega.Expect(len(summaries)).To(gomega.Equal(1))
		gomega.Expect(summaries[0].Count()).To(gomega.Equal(2))
		gomega.Expect(transform.LabelPairsToLabelsMap(summaries[0].Points()[0].Metric().Label)).ToNot(gomega.HaveKey("excluded-tag"))
	})

	ginkgo.Context("CleanPeriodic", func() {
		ginkgo.It("should clean metrics after amount of time", func() {
			rollup := rollup.NewSummaryRollup(
				"0",
				nil,
				rollup.SetSummaryCleaning(10*time.Millisecond, 50*time.Millisecond),
			)

			rollup.Record(
				"source-id",
				nil,
				10,
			)
			rollup.Record(
				"source-id",
				nil,
				5,
			)

			time.Sleep(100 * time.Millisecond)

			summaries := extract(rollup.Rollup(0))
			gomega.Expect(len(summaries)).To(gomega.Equal(0))
		})
	})
})
