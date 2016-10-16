package collectors_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-community/firehose_exporter/metrics"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/cloudfoundry-community/firehose_exporter/collectors"
)

var _ = Describe("ValueMetricsCollector", func() {
	var (
		namespace              string
		metricsStore           *metrics.Store
		metricsExpiration      time.Duration
		metricsCleanupInterval time.Duration
		dopplerDeployments     []string
		valueMetricsCollector  *collectors.ValueMetricsCollector

		valueMetricsCollectorDesc *prometheus.Desc
	)

	BeforeEach(func() {
		namespace = "test_exporter"
		metricsStore = metrics.NewStore(metricsExpiration, metricsCleanupInterval)
		dopplerDeployments = []string{}

		valueMetricsCollectorDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "value_metric", "collector"),
			"Cloud Foundry Firehose value metrics collector.",
			nil,
			nil,
		)
	})

	JustBeforeEach(func() {
		valueMetricsCollector = collectors.NewValueMetricsCollector(namespace, metricsStore, dopplerDeployments)
	})

	Describe("Describe", func() {
		var (
			descriptions chan *prometheus.Desc
		)

		BeforeEach(func() {
			descriptions = make(chan *prometheus.Desc)
		})

		JustBeforeEach(func() {
			go valueMetricsCollector.Describe(descriptions)
		})

		It("returns a value_metric_collector metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(valueMetricsCollectorDesc)))
		})
	})

	Describe("Collect", func() {
		var (
			origin           = "fake-origin"
			originNormalized = "fake_origin"
			boshDeployment   = "fake-deployment-name"
			boshJob          = "fake-job-name"
			boshIndex        = "0"
			boshIP           = "1.2.3.4"

			valueMetric1Name           = "FakeValueMetric1"
			valueMetric1NameNormalized = "fake_value_metric_1"
			valueMetric1Value          = float64(2000)
			valueMetric1Unit           = "kb"

			valueMetric2Name           = "FakeValueMetric2"
			valueMetric2NameNormalized = "fake_value_metric_2"
			valueMetric2Value          = float64(15)
			valueMetric2Unit           = "count"

			metrics      chan prometheus.Metric
			valueMetric1 prometheus.Metric
			valueMetric2 prometheus.Metric
		)

		BeforeEach(func() {
			metricsStore.AddMetric(
				&events.Envelope{
					Origin:     proto.String(origin),
					EventType:  events.Envelope_ValueMetric.Enum(),
					Timestamp:  proto.Int64(time.Now().Unix() * 1000),
					Deployment: proto.String(boshDeployment),
					Job:        proto.String(boshJob),
					Index:      proto.String(boshIndex),
					Ip:         proto.String(boshIP),
					ValueMetric: &events.ValueMetric{
						Name:  proto.String(valueMetric1Name),
						Value: proto.Float64(valueMetric1Value),
						Unit:  proto.String(valueMetric1Unit),
					},
				},
			)

			metricsStore.AddMetric(
				&events.Envelope{
					Origin:     proto.String(origin),
					EventType:  events.Envelope_ValueMetric.Enum(),
					Timestamp:  proto.Int64(time.Now().Unix() * 1000),
					Deployment: proto.String(boshDeployment),
					Job:        proto.String(boshJob),
					Index:      proto.String(boshIndex),
					Ip:         proto.String(boshIP),
					ValueMetric: &events.ValueMetric{
						Name:  proto.String(valueMetric2Name),
						Value: proto.Float64(valueMetric2Value),
						Unit:  proto.String(valueMetric2Unit),
					},
				},
			)

			metrics = make(chan prometheus.Metric)

			valueMetric1 = prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "value_metric", originNormalized+"_"+valueMetric1NameNormalized),
					fmt.Sprintf("Cloud Foundry Firehose '%s' value metric.", valueMetric1Name),
					[]string{"origin", "bosh_deployment", "bosh_job", "bosh_index", "bosh_ip", "unit"},
					nil,
				),
				prometheus.GaugeValue,
				valueMetric1Value,
				origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				valueMetric1Unit,
			)

			valueMetric2 = prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "value_metric", originNormalized+"_"+valueMetric2NameNormalized),
					fmt.Sprintf("Cloud Foundry Firehose '%s' value metric.", valueMetric2Name),
					[]string{"origin", "bosh_deployment", "bosh_job", "bosh_index", "bosh_ip", "unit"},
					nil,
				),
				prometheus.GaugeValue,
				valueMetric2Value,
				origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				valueMetric2Unit,
			)
		})

		JustBeforeEach(func() {
			go valueMetricsCollector.Collect(metrics)
		})

		It("returns a value_metric_fake_origin_fake_value_metric_1 metric", func() {
			Eventually(metrics).Should(Receive(Equal(valueMetric1)))
		})

		It("returns a value_metric_fake_origin_fake_value_metric_2 metric", func() {
			Eventually(metrics).Should(Receive(Equal(valueMetric2)))
		})

		Context("when there is no value metrics", func() {
			BeforeEach(func() {
				metricsStore.FlushValueMetrics()
			})

			It("does not return any metric", func() {
				Consistently(metrics).ShouldNot(Receive())
			})
		})

		Context("when there is a deployment filter", func() {
			BeforeEach(func() {
				dopplerDeployments = []string{"fake-deployment-name"}
			})

			It("returns a value_metric_fake_origin_fake_value_metric_1 metric", func() {
				Eventually(metrics).Should(Receive(Equal(valueMetric1)))
			})

			It("returns a value_metric_fake_origin_fake_value_metric_2 metric", func() {
				Eventually(metrics).Should(Receive(Equal(valueMetric2)))
			})

			Context("and the metrics deployment does not match", func() {
				BeforeEach(func() {
					dopplerDeployments = []string{"another-fake-deployment-name"}
				})

				It("does not return any metric", func() {
					Consistently(metrics).ShouldNot(Receive())
				})
			})
		})
	})
})
