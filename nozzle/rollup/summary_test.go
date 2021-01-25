package rollup_test

import (
	"fmt"
	"time"

	"github.com/bosh-prometheus/firehose_exporter/metrics"
	. "github.com/bosh-prometheus/firehose_exporter/nozzle/rollup"
	"github.com/bosh-prometheus/firehose_exporter/transform"
	dto "github.com/prometheus/client_model/go"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type summary struct {
	points []*metrics.RawMetric
}

func (h *summary) Count() int {
	for _, p := range h.points {
		if p.MetricName() == "http_response_size_bytes" {
			return int(*p.Metric().Summary.SampleCount)
		}
	}
	Fail("No count point found in summary")
	return 0
}

func (h *summary) Sum() int {
	for _, p := range h.points {
		if p.MetricName() == "http_response_size_bytes" {
			return int(*p.Metric().Summary.SampleSum)
		}
	}
	Fail("No sum point found in summary")
	return 0
}

func (h *summary) Points() []*metrics.RawMetric {
	return h.points
}

func (h *summary) Bucket(le string) *dto.Summary {
	for _, p := range h.points {
		if p.MetricName() != "http_response_size_bytes" {
			continue
		}
		for _, label := range p.Metric().Label {
			if label.GetName() == "le" && label.GetValue() == le {
				return p.Metric().Summary
			}
		}
	}
	Fail(fmt.Sprintf("No bucket point found in summary for le = '%s'", le))
	return nil
}

var _ = Describe("summary Rollup", func() {
	extract := func(batches []*PointsBatch) []*summary {
		var summaries []*summary

		for _, b := range batches {
			h := &summary{}
			for _, p := range b.Points {
				h.points = append(h.points, p)
			}
			summaries = append(summaries, h)
		}

		return summaries
	}

	It("returns aggregate information for rolled up events", func() {
		rollup := NewSummaryRollup(
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
		Expect(len(summaries)).To(Equal(1))
		Expect(summaries[0].Count()).To(Equal(2))
		Expect(summaries[0].Sum()).To(Equal(15))
	})

	It("returns batches which each includes a size estimate", func() {
		rollup := NewSummaryRollup(
			"0",
			nil,
		)

		rollup.Record(
			"source-id",
			nil,
			10,
		)

		pointsBatches := rollup.Rollup(0)
		Expect(len(pointsBatches)).To(Equal(1))
		Expect(pointsBatches[0].Size).To(BeNumerically(">", 0))
	})

	It("returns points for each bucket in the summary", func() {
		rollup := NewSummaryRollup(
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
		Expect(len(summaries)).To(Equal(1))
	})

	It("returns points with the timestamp given to Rollup", func() {
		rollup := NewSummaryRollup(
			"node-index",
			nil,
		)

		rollup.Record(
			"source-id",
			nil,
			1,
		)

		summaries := extract(rollup.Rollup(88))
		Expect(len(summaries)).To(Equal(1))
	})

	It("returns summaries with labels based on tags", func() {
		rollup := NewSummaryRollup(
			"node-index",
			[]string{"included-tag"},
		)

		rollup.Record(
			"source-id",
			map[string]string{"included-tag": "foo", "excluded-tag": "bar"},
			1,
		)

		summaries := extract(rollup.Rollup(0))
		Expect(len(summaries)).To(Equal(1))
		for _, p := range summaries[0].Points() {
			Expect(transform.LabelPairsToLabelsMap(p.Metric().Label)).To(And(
				HaveKeyWithValue("included_tag", "foo"),
				HaveKeyWithValue("source_id", "source-id"),
				HaveKeyWithValue("node_index", "node-index"),
			))
		}
	})

	It("returns points that track a running total of rolled up events", func() {
		rollup := NewSummaryRollup(
			"0",
			[]string{"included-tag"},
		)

		rollup.Record(
			"source-id",
			map[string]string{"included-tag": "foo", "excluded-tag": "bar"},
			1,
		)

		summaries := extract(rollup.Rollup(0))
		Expect(len(summaries)).To(Equal(1))
		Expect(summaries[0].Count()).To(Equal(1))

		rollup.Record(
			"source-id",
			map[string]string{"included-tag": "foo", "excluded-tag": "bar"},
			1,
		)

		summaries = extract(rollup.Rollup(1))
		Expect(len(summaries)).To(Equal(1))
		Expect(summaries[0].Count()).To(Equal(2))
	})

	It("returns separate summaries for distinct source IDs", func() {
		rollup := NewSummaryRollup(
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
		Expect(len(summaries)).To(Equal(2))
		Expect(summaries[0].Count()).To(Equal(1))
		Expect(summaries[1].Count()).To(Equal(1))
	})

	It("returns separate summaries for different included tags", func() {
		rollup := NewSummaryRollup(
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
		Expect(len(summaries)).To(Equal(2))
		Expect(summaries[0].Count()).To(Equal(1))
		Expect(summaries[1].Count()).To(Equal(1))
	})

	It("does not return separate summaries for different excluded tags", func() {
		rollup := NewSummaryRollup(
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
		Expect(len(summaries)).To(Equal(1))
		Expect(summaries[0].Count()).To(Equal(2))
		Expect(transform.LabelPairsToLabelsMap(summaries[0].Points()[0].Metric().Label)).ToNot(HaveKey("excluded-tag"))
	})
})
