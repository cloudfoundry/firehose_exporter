package collectors

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/bosh-prometheus/firehose_exporter/metrics"
)

type InternalMetricsCollector struct {
	namespace                                  string
	environment                                string
	metricsStore                               *metrics.Store
	totalEnvelopesReceivedMetric               prometheus.Gauge
	lastEnvelopeReceivedTimestampMetric        prometheus.Gauge
	totalMetricsReceivedMetric                 prometheus.Gauge
	lastMetricReceivedTimestampMetric          prometheus.Gauge
	totalContainerMetricsReceivedMetric        prometheus.Gauge
	totalContainerMetricsProcessedMetric       prometheus.Gauge
	containerMetricsCachedMetric               prometheus.Gauge
	lastContainerMetricReceivedTimestampMetric prometheus.Gauge
	totalCounterEventsReceivedMetric           prometheus.Gauge
	totalCounterEventsProcessedMetric          prometheus.Gauge
	counterEventsCachedMetric                  prometheus.Gauge
	lastCounterEventReceivedTimestampMetric    prometheus.Gauge
	totalHttpStartStopReceivedMetric           prometheus.Gauge
	totalHttpStartStopProcessedMetric          prometheus.Gauge
	httpStartStopCachedMetric                  prometheus.Gauge
	lastHttpStartStopReceivedTimestampMetric   prometheus.Gauge
	totalValueMetricsReceivedMetric            prometheus.Gauge
	totalValueMetricsProcessedMetric           prometheus.Gauge
	valueMetricsCachedMetric                   prometheus.Gauge
	lastValueMetricReceivedTimestampMetric     prometheus.Gauge
	slowConsumerAlertMetric                    prometheus.Gauge
	lastSlowConsumerAlertTimestampMetric       prometheus.Gauge
}

func NewInternalMetricsCollector(
	namespace string,
	environment string,
	metricsStore *metrics.Store,
) *InternalMetricsCollector {
	totalEnvelopesReceivedMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "total_envelopes_received",
			Help:        "Total number of envelopes received from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	lastEnvelopeReceivedTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_envelope_received_timestamp",
			Help:        "Number of seconds since 1970 since last envelope received from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	totalMetricsReceivedMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "total_metrics_received",
			Help:        "Total number of metrics received from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	lastMetricReceivedTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_metric_received_timestamp",
			Help:        "Number of seconds since 1970 since last metric received from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	totalContainerMetricsReceivedMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "total_container_metrics_received",
			Help:        "Total number of container metrics received from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	totalContainerMetricsProcessedMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "total_container_metrics_processed",
			Help:        "Total number of container metrics processed from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	containerMetricsCachedMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "container_metrics_cached",
			Help:        "Number of container metrics cached from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	lastContainerMetricReceivedTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_container_metric_received_timestamp",
			Help:        "Number of seconds since 1970 since last container metric received from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	totalCounterEventsReceivedMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "total_counter_events_received",
			Help:        "Total number of counter events received from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	totalCounterEventsProcessedMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "total_counter_events_processed",
			Help:        "Total number of counter events processed from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	counterEventsCachedMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "counter_events_cached",
			Help:        "Number of counter events cached from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	lastCounterEventReceivedTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_counter_event_received_timestamp",
			Help:        "Number of seconds since 1970 since last counter event received from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	totalHttpStartStopReceivedMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "total_http_start_stop_received",
			Help:        "Total number of http start stop received from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	totalHttpStartStopProcessedMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "total_http_start_stop_processed",
			Help:        "Total number of http start stop processed from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	httpStartStopCachedMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "http_start_stop_cached",
			Help:        "Number of http start stop cached from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	lastHttpStartStopReceivedTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_http_start_stop_received_timestamp",
			Help:        "Number of seconds since 1970 since last http start stop received from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	totalValueMetricsReceivedMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "total_value_metrics_received",
			Help:        "Total number of value metrics received from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	totalValueMetricsProcessedMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "total_value_metrics_processed",
			Help:        "Total number of value metrics processed from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	valueMetricsCachedMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "value_metrics_cached",
			Help:        "Number of value metrics cached from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	lastValueMetricReceivedTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_value_metric_received_timestamp",
			Help:        "Number of seconds since 1970 since last value metric received from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	slowConsumerAlertMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "slow_consumer_alert",
			Help:        "Nozzle could not keep up with Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	lastSlowConsumerAlertTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_slow_consumer_alert_timestamp",
			Help:        "Number of seconds since 1970 since last slow consumer alert received from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	collector := &InternalMetricsCollector{
		namespace:                                  namespace,
		environment:                                environment,
		metricsStore:                               metricsStore,
		totalEnvelopesReceivedMetric:               totalEnvelopesReceivedMetric,
		lastEnvelopeReceivedTimestampMetric:        lastEnvelopeReceivedTimestampMetric,
		totalMetricsReceivedMetric:                 totalMetricsReceivedMetric,
		lastMetricReceivedTimestampMetric:          lastMetricReceivedTimestampMetric,
		totalContainerMetricsReceivedMetric:        totalContainerMetricsReceivedMetric,
		totalContainerMetricsProcessedMetric:       totalContainerMetricsProcessedMetric,
		containerMetricsCachedMetric:               containerMetricsCachedMetric,
		lastContainerMetricReceivedTimestampMetric: lastContainerMetricReceivedTimestampMetric,
		totalCounterEventsReceivedMetric:           totalCounterEventsReceivedMetric,
		totalCounterEventsProcessedMetric:          totalCounterEventsProcessedMetric,
		counterEventsCachedMetric:                  counterEventsCachedMetric,
		lastCounterEventReceivedTimestampMetric:    lastCounterEventReceivedTimestampMetric,
		totalHttpStartStopReceivedMetric:           totalHttpStartStopReceivedMetric,
		totalHttpStartStopProcessedMetric:          totalHttpStartStopProcessedMetric,
		httpStartStopCachedMetric:                  httpStartStopCachedMetric,
		lastHttpStartStopReceivedTimestampMetric:   lastHttpStartStopReceivedTimestampMetric,
		totalValueMetricsReceivedMetric:            totalValueMetricsReceivedMetric,
		totalValueMetricsProcessedMetric:           totalValueMetricsProcessedMetric,
		valueMetricsCachedMetric:                   valueMetricsCachedMetric,
		lastValueMetricReceivedTimestampMetric:     lastValueMetricReceivedTimestampMetric,
		slowConsumerAlertMetric:                    slowConsumerAlertMetric,
		lastSlowConsumerAlertTimestampMetric:       lastSlowConsumerAlertTimestampMetric,
	}
	return collector
}

func (c InternalMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	internalMetrics := c.metricsStore.GetInternalMetrics()

	c.totalEnvelopesReceivedMetric.Set(float64(internalMetrics.TotalEnvelopesReceived))
	c.totalEnvelopesReceivedMetric.Collect(ch)

	c.lastEnvelopeReceivedTimestampMetric.Set(float64(internalMetrics.LastEnvelopReceivedTimestamp))
	c.lastEnvelopeReceivedTimestampMetric.Collect(ch)

	c.totalMetricsReceivedMetric.Set(float64(internalMetrics.TotalMetricsReceived))
	c.totalMetricsReceivedMetric.Collect(ch)

	c.lastMetricReceivedTimestampMetric.Set(float64(internalMetrics.LastMetricReceivedTimestamp))
	c.lastMetricReceivedTimestampMetric.Collect(ch)

	c.totalContainerMetricsReceivedMetric.Set(float64(internalMetrics.TotalContainerMetricsReceived))
	c.totalContainerMetricsReceivedMetric.Collect(ch)

	c.totalContainerMetricsProcessedMetric.Set(float64(internalMetrics.TotalContainerMetricsProcessed))
	c.totalContainerMetricsProcessedMetric.Collect(ch)

	c.containerMetricsCachedMetric.Set(float64(internalMetrics.TotalContainerMetricsCached))
	c.containerMetricsCachedMetric.Collect(ch)

	c.lastContainerMetricReceivedTimestampMetric.Set(float64(internalMetrics.LastContainerMetricReceivedTimestamp))
	c.lastContainerMetricReceivedTimestampMetric.Collect(ch)

	c.totalCounterEventsReceivedMetric.Set(float64(internalMetrics.TotalCounterEventsReceived))
	c.totalCounterEventsReceivedMetric.Collect(ch)

	c.totalCounterEventsProcessedMetric.Set(float64(internalMetrics.TotalCounterEventsProcessed))
	c.totalCounterEventsProcessedMetric.Collect(ch)

	c.counterEventsCachedMetric.Set(float64(internalMetrics.TotalCounterEventsCached))
	c.counterEventsCachedMetric.Collect(ch)

	c.lastCounterEventReceivedTimestampMetric.Set(float64(internalMetrics.LastCounterEventReceivedTimestamp))
	c.lastCounterEventReceivedTimestampMetric.Collect(ch)

	c.totalHttpStartStopReceivedMetric.Set(float64(internalMetrics.TotalHttpStartStopReceived))
	c.totalHttpStartStopReceivedMetric.Collect(ch)

	c.totalHttpStartStopProcessedMetric.Set(float64(internalMetrics.TotalHttpStartStopProcessed))
	c.totalHttpStartStopProcessedMetric.Collect(ch)

	c.httpStartStopCachedMetric.Set(float64(internalMetrics.TotalHttpStartStopCached))
	c.httpStartStopCachedMetric.Collect(ch)

	c.lastHttpStartStopReceivedTimestampMetric.Set(float64(internalMetrics.LastHttpStartStopReceivedTimestamp))
	c.lastHttpStartStopReceivedTimestampMetric.Collect(ch)

	c.totalValueMetricsReceivedMetric.Set(float64(internalMetrics.TotalValueMetricsReceived))
	c.totalValueMetricsReceivedMetric.Collect(ch)

	c.totalValueMetricsProcessedMetric.Set(float64(internalMetrics.TotalValueMetricsProcessed))
	c.totalValueMetricsProcessedMetric.Collect(ch)

	c.valueMetricsCachedMetric.Set(float64(internalMetrics.TotalValueMetricsCached))
	c.valueMetricsCachedMetric.Collect(ch)

	c.lastValueMetricReceivedTimestampMetric.Set(float64(internalMetrics.LastValueMetricReceivedTimestamp))
	c.lastValueMetricReceivedTimestampMetric.Collect(ch)

	c.slowConsumerAlertMetric.Set(0)
	if internalMetrics.SlowConsumerAlert {
		c.slowConsumerAlertMetric.Set(1)
	}
	c.slowConsumerAlertMetric.Collect(ch)

	c.lastSlowConsumerAlertTimestampMetric.Set(float64(internalMetrics.LastSlowConsumerAlertTimestamp))
	c.lastSlowConsumerAlertTimestampMetric.Collect(ch)
}

func (c InternalMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	c.totalEnvelopesReceivedMetric.Describe(ch)
	c.lastEnvelopeReceivedTimestampMetric.Describe(ch)
	c.totalMetricsReceivedMetric.Describe(ch)
	c.lastMetricReceivedTimestampMetric.Describe(ch)
	c.totalContainerMetricsReceivedMetric.Describe(ch)
	c.totalContainerMetricsProcessedMetric.Describe(ch)
	c.containerMetricsCachedMetric.Describe(ch)
	c.lastContainerMetricReceivedTimestampMetric.Describe(ch)
	c.totalCounterEventsReceivedMetric.Describe(ch)
	c.totalCounterEventsProcessedMetric.Describe(ch)
	c.counterEventsCachedMetric.Describe(ch)
	c.lastCounterEventReceivedTimestampMetric.Describe(ch)
	c.totalHttpStartStopReceivedMetric.Describe(ch)
	c.totalHttpStartStopProcessedMetric.Describe(ch)
	c.httpStartStopCachedMetric.Describe(ch)
	c.lastHttpStartStopReceivedTimestampMetric.Describe(ch)
	c.totalValueMetricsReceivedMetric.Describe(ch)
	c.totalValueMetricsProcessedMetric.Describe(ch)
	c.valueMetricsCachedMetric.Describe(ch)
	c.lastValueMetricReceivedTimestampMetric.Describe(ch)
	c.slowConsumerAlertMetric.Describe(ch)
	c.lastSlowConsumerAlertTimestampMetric.Describe(ch)
}
