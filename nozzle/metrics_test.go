package nozzle_test

import (
	"time"

	"github.com/bosh-prometheus/firehose_exporter/metrics"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	"github.com/bosh-prometheus/firehose_exporter/nozzle"
)

var _ = ginkgo.Describe("collect nozzle metrics", func() {
	var pointBuffer chan []*metrics.RawMetric
	var metricStore *MetricStoreTesting
	ginkgo.BeforeEach(func() {
		pointBuffer = make(chan []*metrics.RawMetric)
		metricStore = NewMetricStoreTesting(pointBuffer)
	})
	ginkgo.It("writes Ingress, Egress and Err metrics", func() {
		streamConnector := newSpyStreamConnector()
		n := nozzle.NewNozzle(streamConnector, "firehose_exporter", 0,
			pointBuffer,
			internalMetric,
			nozzle.WithNozzleTimerRollup(
				100*time.Millisecond,
				[]string{"tag1", "tag2", "status_code"},
				[]string{"tag1", "tag2"},
			),
		)
		go n.Start()

		addEnvelope(1, "memory", "some-source-id", streamConnector)
		addEnvelope(2, "memory", "some-source-id", streamConnector)
		addEnvelope(3, "memory", "some-source-id", streamConnector)
		gomega.Eventually(metricStore.GetPoints).Should(gomega.HaveLen(3))
	})

	ginkgo.It("writes duration seconds histogram metrics", func() {
		streamConnector := newSpyStreamConnector()

		n := nozzle.NewNozzle(streamConnector, "firehose_exporter", 0,
			pointBuffer,
			internalMetric,
			nozzle.WithNozzleTimerRollup(
				100*time.Millisecond,
				[]string{"tag1", "tag2", "status_code"},
				[]string{"tag1", "tag2"},
			),
		)
		go n.Start()

		addEnvelope(1, "memory", "some-source-id", streamConnector)
	})
})
