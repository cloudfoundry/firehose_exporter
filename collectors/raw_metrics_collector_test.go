package collectors_test

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/bosh-prometheus/firehose_exporter/metricmaker"
	"github.com/bosh-prometheus/firehose_exporter/metrics"
	"github.com/bosh-prometheus/firehose_exporter/testing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/bosh-prometheus/firehose_exporter/collectors"
)

var _ = Describe("RawMetricsCollector", func() {
	var pointBuffer chan []*metrics.RawMetric
	var collector *collectors.RawMetricsCollector
	BeforeEach(func() {
		pointBuffer = make(chan []*metrics.RawMetric)
		collector = collectors.NewRawMetricsCollector(pointBuffer, 10*time.Minute)
	})

	AfterEach(func() {
		close(pointBuffer)
	})

	Context("Collect", func() {
		It("should save in metric store collected points", func() {
			go collector.Collect()
			pointBuffer <- []*metrics.RawMetric{
				metricmaker.NewRawMetricCounter("my_metric", map[string]string{
					"origin":   "my-origin",
					"variadic": "1",
				}, 1),
				metricmaker.NewRawMetricCounter("my_second_metric", map[string]string{
					"origin":   "my-origin",
					"variadic": "1",
				}, 1),
				metricmaker.NewRawMetricCounter("my_metric", map[string]string{
					"origin":   "my-origin",
					"variadic": "2",
				}, 1),
			}
			pointBuffer <- []*metrics.RawMetric{
				metricmaker.NewRawMetricCounter("my_metric", map[string]string{
					"origin":   "my-origin",
					"variadic": "1",
				}, 2),
			}
			time.Sleep(50 * time.Millisecond)

			ms := collector.MetricStore()

			Expect(ms).To(HaveKey("my_metric"))
			Expect(ms).To(HaveKey("my_second_metric"))
			Expect(ms["my_second_metric"]).To(HaveLen(1))
			Expect(ms["my_metric"]).To(HaveLen(2))
			Expect(ms["my_metric"]).To(testing.ContainPoints([]*metrics.RawMetric{
				metricmaker.NewRawMetricCounter("my_metric", map[string]string{
					"origin":   "my-origin",
					"variadic": "2",
				}, 1),
				metricmaker.NewRawMetricCounter("my_metric", map[string]string{
					"origin":   "my-origin",
					"variadic": "1",
				}, 2),
			}))

		})

		Context("CleanPeriodic", func() {
			It("should clean swept metrics", func() {
				collector.SetCleanPeriodicDuration(70 * time.Millisecond)
				go collector.Collect()
				go collector.CleanPeriodic()
				m := metricmaker.NewRawMetricCounter("my_metric", map[string]string{
					"origin":   "my-origin",
					"variadic": "1",
				}, 1)
				pointBuffer <- []*metrics.RawMetric{m}

				time.Sleep(50 * time.Millisecond)
				m.SetSweep(true)
				ms := collector.MetricStore()
				Expect(ms).To(HaveKey("my_metric"))
				Expect(ms["my_metric"]).To(HaveLen(1))
				Expect(ms["my_metric"][0].IsSwept()).To(BeTrue())

				time.Sleep(50 * time.Millisecond)
				ms = collector.MetricStore()
				Expect(ms).To(HaveKey("my_metric"))
				Expect(ms["my_metric"]).To(HaveLen(0))
			})
		})

		Context("RenderExpFmt", func() {
			BeforeEach(func() {
				go collector.Collect()
				pointBuffer <- []*metrics.RawMetric{
					metricmaker.NewRawMetricCounter("my_metric", map[string]string{
						"origin":   "my-origin",
						"variadic": "1",
					}, 1),
					metricmaker.NewRawMetricCounter("my_second_metric", map[string]string{
						"origin":   "my-origin",
						"variadic": "1",
					}, 1),
					metricmaker.NewRawMetricCounter("my_metric", map[string]string{
						"origin":   "my-origin",
						"variadic": "2",
					}, 1),
				}
				time.Sleep(50 * time.Millisecond)
			})
			It("should show metric in expfmt in plain text from registered internal metrics", func() {
				respRec := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
				collector.RenderExpFmt(respRec, req)

				content := respRec.Body.String()

				Expect(content).To(ContainSubstring(`go_gc_duration_seconds`))
			})
			When("no gzip is asked", func() {
				It("should show metric in expfmt in plain text", func() {
					respRec := httptest.NewRecorder()
					req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
					collector.RenderExpFmt(respRec, req)

					content := respRec.Body.String()

					Expect(content).To(ContainSubstring(`my_metric{origin="my-origin",variadic="1"} 1`))
					Expect(content).To(ContainSubstring(`my_metric{origin="my-origin",variadic="2"} 1`))
					Expect(content).To(ContainSubstring(`my_second_metric{origin="my-origin",variadic="1"} 1`))
				})
			})
			When("with gzip is asked", func() {
				It("should show metric in expfmt in gzip", func() {
					respRec := httptest.NewRecorder()
					req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
					req.Header.Set("Accept-Encoding", "gzip")
					collector.RenderExpFmt(respRec, req)

					gzipReader, err := gzip.NewReader(respRec.Body)
					Expect(err).ToNot(HaveOccurred())
					defer gzipReader.Close()

					var resB bytes.Buffer
					_, err = resB.ReadFrom(gzipReader)
					if err != nil {
						return
					}

					content := resB.String()

					Expect(content).To(ContainSubstring(`my_metric{origin="my-origin",variadic="1"} 1`))
					Expect(content).To(ContainSubstring(`my_metric{origin="my-origin",variadic="2"} 1`))
					Expect(content).To(ContainSubstring(`my_second_metric{origin="my-origin",variadic="1"} 1`))
				})
			})
		})
	})
})
