package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type InternalMetrics struct {
	namespace                            string
	environment                          string
	TotalEnvelopesReceived               prometheus.Counter
	TotalEnvelopesDropped                prometheus.Counter
	LastEnvelopeReceivedTimestamp        prometheus.Gauge
	TotalMetricsReceived                 prometheus.Counter
	TotalMetricsDropped                  prometheus.Counter
	LastMetricReceivedTimestamp          prometheus.Gauge
	TotalContainerMetricsReceived        prometheus.Counter
	LastContainerMetricReceivedTimestamp prometheus.Gauge
	TotalCounterEventsReceived           prometheus.Counter
	LastCounterEventReceivedTimestamp    prometheus.Gauge
	TotalValueMetricsReceived            prometheus.Counter
	LastValueMetricReceivedTimestamp     prometheus.Gauge
	TotalHttpMetricsReceived             prometheus.Counter
	LastHttpMetricReceivedTimestamp      prometheus.Gauge
}

func NewInternalMetrics(namespace string, environment string) *InternalMetrics {
	im := &InternalMetrics{
		namespace:   namespace,
		environment: environment,
	}

	im.TotalEnvelopesReceived = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "total_envelopes_received",
			Help:        "Total number of envelopes received from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	im.TotalEnvelopesDropped = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "total_envelopes_dropped",
			Help:        "Total number of envelopes dropped from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	im.LastEnvelopeReceivedTimestamp = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_envelope_received_timestamp",
			Help:        "Number of seconds since 1970 since last envelope received from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	im.TotalMetricsReceived = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "total_metrics_received",
			Help:        "Total number of metrics received from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	im.TotalMetricsDropped = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "total_metrics_dropped",
			Help:        "Total number of metrics dropped from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	im.LastMetricReceivedTimestamp = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_metric_received_timestamp",
			Help:        "Number of seconds since 1970 since last metric received from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	im.TotalContainerMetricsReceived = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "total_container_metrics_received",
			Help:        "Total number of container metrics received from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)
	im.LastContainerMetricReceivedTimestamp = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_container_metric_received_timestamp",
			Help:        "Number of seconds since 1970 since last container metric received from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	im.TotalHttpMetricsReceived = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "total_http_metrics_received",
			Help:        "Total number of http metrics received from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)
	im.LastHttpMetricReceivedTimestamp = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_http_metric_received_timestamp",
			Help:        "Number of seconds since 1970 since last http metrics received from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	im.TotalCounterEventsReceived = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "total_counter_events_received",
			Help:        "Total number of counter events received from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	im.LastCounterEventReceivedTimestamp = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_counter_event_received_timestamp",
			Help:        "Number of seconds since 1970 since last counter event received from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	im.TotalValueMetricsReceived = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "total_value_metrics_received",
			Help:        "Total number of value metrics received from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)

	im.LastValueMetricReceivedTimestamp = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_value_metric_received_timestamp",
			Help:        "Number of seconds since 1970 since last value metric received from Cloud Foundry Firehose.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
	)
	return im
}
