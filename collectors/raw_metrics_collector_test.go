package collectors_test

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/cloudfoundry/firehose_exporter/metricmaker"
	"github.com/cloudfoundry/firehose_exporter/metrics"
	"github.com/cloudfoundry/firehose_exporter/testing"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	"github.com/cloudfoundry/firehose_exporter/collectors"
)

var _ = ginkgo.Describe("RawMetricsCollector", func() {
	var pointBuffer chan []*metrics.RawMetric
	var collector *collectors.RawMetricsCollector
	ginkgo.BeforeEach(func() {
		pointBuffer = make(chan []*metrics.RawMetric)
		collector = collectors.NewRawMetricsCollector(pointBuffer, 10*time.Minute)
	})

	ginkgo.AfterEach(func() {
		close(pointBuffer)
	})

	ginkgo.Context("Collect", func() {
		ginkgo.It("should save in metric store collected points", func() {
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

			gomega.Expect(ms).To(gomega.HaveKey("my_metric"))
			gomega.Expect(ms).To(gomega.HaveKey("my_second_metric"))
			gomega.Expect(ms["my_second_metric"]).To(gomega.HaveLen(1))
			gomega.Expect(ms["my_metric"]).To(gomega.HaveLen(2))
			gomega.Expect(ms["my_metric"]).To(testing.ContainPoints([]*metrics.RawMetric{
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

		ginkgo.Context("CleanPeriodic", func() {
			ginkgo.It("should clean swept metrics", func() {
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
				gomega.Expect(ms).To(gomega.HaveKey("my_metric"))
				gomega.Expect(ms["my_metric"]).To(gomega.HaveLen(1))
				gomega.Expect(ms["my_metric"][0].IsSwept()).To(gomega.BeTrue())

				time.Sleep(50 * time.Millisecond)
				ms = collector.MetricStore()
				gomega.Expect(ms).To(gomega.HaveKey("my_metric"))
				gomega.Expect(ms["my_metric"]).To(gomega.HaveLen(0))
			})
		})

		ginkgo.Context("RenderExpFmt", func() {
			ginkgo.BeforeEach(func() {
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
			ginkgo.It("should show metric in expfmt in plain text from registered internal metrics", func() {
				respRec := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
				collector.RenderExpFmt(respRec, req)

				content := respRec.Body.String()

				gomega.Expect(content).To(gomega.ContainSubstring(`go_gc_duration_seconds`))
			})
			ginkgo.When("no gzip is asked", func() {
				ginkgo.It("should show metric in expfmt in plain text", func() {
					respRec := httptest.NewRecorder()
					req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
					collector.RenderExpFmt(respRec, req)

					content := respRec.Body.String()

					gomega.Expect(content).To(gomega.ContainSubstring(`my_metric{origin="my-origin",variadic="1"} 1`))
					gomega.Expect(content).To(gomega.ContainSubstring(`my_metric{origin="my-origin",variadic="2"} 1`))
					gomega.Expect(content).To(gomega.ContainSubstring(`my_second_metric{origin="my-origin",variadic="1"} 1`))
				})
			})
			ginkgo.When("with gzip is asked", func() {
				ginkgo.It("should show metric in expfmt in gzip", func() {
					respRec := httptest.NewRecorder()
					req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
					req.Header.Set("Accept-Encoding", "gzip")
					collector.RenderExpFmt(respRec, req)

					gzipReader, err := gzip.NewReader(respRec.Body)
					gomega.Expect(err).ToNot(gomega.HaveOccurred())
					defer gzipReader.Close()

					var resB bytes.Buffer
					_, err = resB.ReadFrom(gzipReader)
					if err != nil {
						return
					}

					content := resB.String()

					gomega.Expect(content).To(gomega.ContainSubstring(`my_metric{origin="my-origin",variadic="1"} 1`))
					gomega.Expect(content).To(gomega.ContainSubstring(`my_metric{origin="my-origin",variadic="2"} 1`))
					gomega.Expect(content).To(gomega.ContainSubstring(`my_second_metric{origin="my-origin",variadic="1"} 1`))
				})
			})
		})
	})
})
