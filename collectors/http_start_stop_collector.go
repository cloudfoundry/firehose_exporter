package collectors

import (
	"net/url"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/cloudfoundry-community/firehose_exporter/metrics"
	"github.com/cloudfoundry-community/firehose_exporter/utils"
)

type HttpStartStopCollector struct {
	namespace                          string
	metricsStore                       *metrics.Store
	requestsMetric                     *prometheus.GaugeVec
	responseSizeBytesMetric            *prometheus.SummaryVec
	lastRequestTimestampMetric         *prometheus.GaugeVec
	clientRequestDurationSecondsMetric *prometheus.SummaryVec
	serverRequestDurationSecondsMetric *prometheus.SummaryVec
}

func NewHttpStartStopCollector(
	namespace string,
	metricsStore *metrics.Store,
) *HttpStartStopCollector {
	requestsMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: http_start_stop_subsystem,
			Name:      "requests",
			Help:      "Cloud Foundry Firehose http start stop requests.",
		},
		[]string{"application_id", "instance_id", "method", "scheme", "host", "status_code"},
	)

	responseSizeBytesMetric := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: namespace,
			Subsystem: http_start_stop_subsystem,
			Name:      "response_size_bytes",
			Help:      "Summary of Cloud Foundry Firehose http start stop request size in bytes.",
		},
		[]string{"application_id", "instance_id", "method", "scheme", "host"},
	)

	lastRequestTimestampMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: http_start_stop_subsystem,
			Name:      "last_request_timestamp",
			Help:      "Number of seconds since 1970 since last http start stop received from Cloud Foundry Firehose.",
		},
		[]string{"application_id", "instance_id", "method", "scheme", "host"},
	)

	clientRequestDurationSecondsMetric := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: namespace,
			Subsystem: http_start_stop_subsystem,
			Name:      "client_request_duration_seconds",
			Help:      "Summary of Cloud Foundry Firehose http start stop client request duration in seconds.",
		},
		[]string{"application_id", "instance_id", "method", "scheme", "host"},
	)

	serverRequestDurationSecondsMetric := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: namespace,
			Subsystem: http_start_stop_subsystem,
			Name:      "server_request_duration_seconds",
			Help:      "Summary of Cloud Foundry Firehose http start stop server request duration in seconds.",
		},
		[]string{"application_id", "instance_id", "method", "scheme", "host"},
	)

	return &HttpStartStopCollector{
		namespace:                          namespace,
		metricsStore:                       metricsStore,
		requestsMetric:                     requestsMetric,
		responseSizeBytesMetric:            responseSizeBytesMetric,
		lastRequestTimestampMetric:         lastRequestTimestampMetric,
		clientRequestDurationSecondsMetric: clientRequestDurationSecondsMetric,
		serverRequestDurationSecondsMetric: serverRequestDurationSecondsMetric,
	}
}

func (c HttpStartStopCollector) Collect(ch chan<- prometheus.Metric) {
	// We reset metrics here to NOT report HttpStartStop events for
	// applications that do NOT exist anymore. The trade-off is that
	// these metrics will only report events captured during the
	// slidding window defined at the `doppler.metric-expiration`
	// command flag.
	c.requestsMetric.Reset()
	c.responseSizeBytesMetric.Reset()
	c.lastRequestTimestampMetric.Reset()
	c.clientRequestDurationSecondsMetric.Reset()
	c.serverRequestDurationSecondsMetric.Reset()

	for _, httpStartStop := range c.metricsStore.GetHttpStartStops() {
		if httpStartStop.ApplicationId == "" {
			continue
		}

		var scheme, host string
		uri, err := url.Parse(httpStartStop.Uri)
		if err == nil {
			scheme = uri.Scheme
			host = uri.Host
		}

		c.requestsMetric.WithLabelValues(
			httpStartStop.ApplicationId,
			httpStartStop.InstanceId,
			httpStartStop.Method,
			scheme,
			host,
			strconv.Itoa(int(httpStartStop.StatusCode)),
		).Inc()

		c.responseSizeBytesMetric.WithLabelValues(
			httpStartStop.ApplicationId,
			httpStartStop.InstanceId,
			httpStartStop.Method,
			scheme,
			host,
		).Observe(float64(httpStartStop.ContentLength))

		var lastRequestTimestamp float64
		if httpStartStop.ClientStartTimestamp > 0 {
			lastRequestTimestamp = utils.NanosecondsToSeconds(httpStartStop.ClientStartTimestamp)
		} else {
			lastRequestTimestamp = utils.NanosecondsToSeconds(httpStartStop.ServerStartTimestamp)
		}
		c.lastRequestTimestampMetric.WithLabelValues(
			httpStartStop.ApplicationId,
			httpStartStop.InstanceId,
			httpStartStop.Method,
			scheme,
			host,
		).Set(lastRequestTimestamp)

		clientDuration := httpStartStop.ClientStopTimestamp - httpStartStop.ClientStartTimestamp
		if clientDuration > 0 {
			c.clientRequestDurationSecondsMetric.WithLabelValues(
				httpStartStop.ApplicationId,
				httpStartStop.InstanceId,
				httpStartStop.Method,
				scheme,
				host,
			).Observe(utils.NanosecondsToSeconds(clientDuration))
		}

		serverDuration := httpStartStop.ServerStopTimestamp - httpStartStop.ServerStartTimestamp
		if serverDuration > 0 {
			c.serverRequestDurationSecondsMetric.WithLabelValues(
				httpStartStop.ApplicationId,
				httpStartStop.InstanceId,
				httpStartStop.Method,
				scheme,
				host,
			).Observe(utils.NanosecondsToSeconds(serverDuration))
		}
	}

	c.requestsMetric.Collect(ch)
	c.responseSizeBytesMetric.Collect(ch)
	c.lastRequestTimestampMetric.Collect(ch)
	c.clientRequestDurationSecondsMetric.Collect(ch)
	c.serverRequestDurationSecondsMetric.Collect(ch)
}

func (c HttpStartStopCollector) Describe(ch chan<- *prometheus.Desc) {
	c.requestsMetric.Describe(ch)
	c.responseSizeBytesMetric.Describe(ch)
	c.lastRequestTimestampMetric.Describe(ch)
	c.clientRequestDurationSecondsMetric.Describe(ch)
	c.serverRequestDurationSecondsMetric.Describe(ch)
}
