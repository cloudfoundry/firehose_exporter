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

var _ = Describe("CounterMetricsCollector", func() {
	var (
		namespace               string
		metricsStore            *metrics.Store
		metricsExpiration       time.Duration
		metricsCleanupInterval  time.Duration
		dopplerDeployments      []string
		counterMetricsCollector *collectors.CounterMetricsCollector

		counterMetricsCollectorDesc *prometheus.Desc
	)

	BeforeEach(func() {
		namespace = "test_exporter"
		metricsStore = metrics.NewStore(metricsExpiration, metricsCleanupInterval)
		dopplerDeployments = []string{}

		counterMetricsCollectorDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "counter_event", "collector"),
			"Cloud Foundry Firehose counter metrics collector.",
			nil,
			nil,
		)
	})

	JustBeforeEach(func() {
		counterMetricsCollector = collectors.NewCounterMetricsCollector(namespace, metricsStore, dopplerDeployments)
	})

	Describe("Describe", func() {
		var (
			descriptions chan *prometheus.Desc
		)

		BeforeEach(func() {
			descriptions = make(chan *prometheus.Desc)
		})

		JustBeforeEach(func() {
			go counterMetricsCollector.Describe(descriptions)
		})

		It("returns a counter_metric_collector metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(counterMetricsCollectorDesc)))
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

			counterEvent1Name           = "FakeCounterEvent1"
			counterEvent1NameNormalized = "fake_counter_event_1"
			counterEvent1Delta          = uint64(5)
			counterEvent1Total          = uint64(1000)

			counterEvent2Name           = "FakeCounterEvent2"
			counterEvent2NameNormalized = "fake_counter_event_2"
			counterEvent2Delta          = uint64(10)
			counterEvent2Total          = uint64(2000)

			metrics             chan prometheus.Metric
			totalCounterMetric1 prometheus.Metric
			deltaCounterMetric1 prometheus.Metric
			totalCounterMetric2 prometheus.Metric
			deltaCounterMetric2 prometheus.Metric
		)

		BeforeEach(func() {
			metricsStore.AddMetric(
				&events.Envelope{
					Origin:     proto.String(origin),
					EventType:  events.Envelope_CounterEvent.Enum(),
					Timestamp:  proto.Int64(time.Now().Unix() * 1000),
					Deployment: proto.String(boshDeployment),
					Job:        proto.String(boshJob),
					Index:      proto.String(boshIndex),
					Ip:         proto.String(boshIP),
					CounterEvent: &events.CounterEvent{
						Name:  proto.String(counterEvent1Name),
						Delta: proto.Uint64(counterEvent1Delta),
						Total: proto.Uint64(counterEvent1Total),
					},
				},
			)

			metricsStore.AddMetric(
				&events.Envelope{
					Origin:     proto.String(origin),
					EventType:  events.Envelope_CounterEvent.Enum(),
					Timestamp:  proto.Int64(time.Now().Unix() * 1000),
					Deployment: proto.String(boshDeployment),
					Job:        proto.String(boshJob),
					Index:      proto.String(boshIndex),
					Ip:         proto.String(boshIP),
					CounterEvent: &events.CounterEvent{
						Name:  proto.String(counterEvent2Name),
						Delta: proto.Uint64(counterEvent2Delta),
						Total: proto.Uint64(counterEvent2Total),
					},
				},
			)

			metrics = make(chan prometheus.Metric)

			totalCounterMetric1 = prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "counter_event", originNormalized+"_total_"+counterEvent1NameNormalized),
					fmt.Sprintf("Cloud Foundry Firehose '%s' total counter event.", counterEvent1Name),
					[]string{"origin", "bosh_deployment", "bosh_job", "bosh_index", "bosh_ip"},
					nil,
				),
				prometheus.CounterValue,
				float64(counterEvent1Total),
				origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
			)

			deltaCounterMetric1 = prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "counter_event", originNormalized+"_delta_"+counterEvent1NameNormalized),
					fmt.Sprintf("Cloud Foundry Firehose '%s' delta counter event.", counterEvent1Name),
					[]string{"origin", "bosh_deployment", "bosh_job", "bosh_index", "bosh_ip"},
					nil,
				),
				prometheus.GaugeValue,
				float64(counterEvent1Delta),
				origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
			)

			totalCounterMetric2 = prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "counter_event", originNormalized+"_total_"+counterEvent2NameNormalized),
					fmt.Sprintf("Cloud Foundry Firehose '%s' total counter event.", counterEvent2Name),
					[]string{"origin", "bosh_deployment", "bosh_job", "bosh_index", "bosh_ip"},
					nil,
				),
				prometheus.CounterValue,
				float64(counterEvent2Total),
				origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
			)

			deltaCounterMetric2 = prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "counter_event", originNormalized+"_delta_"+counterEvent2NameNormalized),
					fmt.Sprintf("Cloud Foundry Firehose '%s' delta counter event.", counterEvent2Name),
					[]string{"origin", "bosh_deployment", "bosh_job", "bosh_index", "bosh_ip"},
					nil,
				),
				prometheus.GaugeValue,
				float64(counterEvent2Delta),
				origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
			)
		})

		JustBeforeEach(func() {
			go counterMetricsCollector.Collect(metrics)
		})

		It("returns a counter_event_fake_origin_total_fake_counter_event_1 metric", func() {
			Eventually(metrics).Should(Receive(Equal(totalCounterMetric1)))
		})

		It("returns a counter_event_fake_origin_delta_fake_counter_event_1 metric", func() {
			Eventually(metrics).Should(Receive(Equal(deltaCounterMetric1)))
		})

		It("returns a counter_event_fake_origin_total_fake_counter_event_2 metric", func() {
			Eventually(metrics).Should(Receive(Equal(totalCounterMetric2)))
		})

		It("returns a counter_event_fake_origin_delta_fake_counter_event_2 metric", func() {
			Eventually(metrics).Should(Receive(Equal(deltaCounterMetric2)))
		})

		Context("when there is no counter metrics", func() {
			BeforeEach(func() {
				metricsStore.FlushCounterMetrics()
			})

			It("does not return any metric", func() {
				Consistently(metrics).ShouldNot(Receive())
			})
		})

		Context("when there is a deployment filter", func() {
			BeforeEach(func() {
				dopplerDeployments = []string{"fake-deployment-name"}
			})

			It("returns a counter_metric_fake_origin_total_fake_counter_event_1 metric", func() {
				Eventually(metrics).Should(Receive(Equal(totalCounterMetric1)))
			})

			It("returns a counter_event_fake_origin_delta_fake_counter_event_1 metric", func() {
				Eventually(metrics).Should(Receive(Equal(deltaCounterMetric1)))
			})

			It("returns a couter_metric_fake_origin_total_fake_counter_event_2 metric", func() {
				Eventually(metrics).Should(Receive(Equal(totalCounterMetric2)))
			})

			It("returns a counter_event_fake_origin_delta_fake_counter_event_2 metric", func() {
				Eventually(metrics).Should(Receive(Equal(deltaCounterMetric2)))
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
