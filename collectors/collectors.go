package collectors

// Metric name parts.
const (
	// Namespace for all metrics.
	namespace = "firehose"

	// Subsystem(s).
	subsystem = "exporter"

	// Container Metrics Subsystem.
	container_metrics_subsystem = subsystem + "_container_metric"

	// Counter Events Subsystem.
	counter_events_subsystem = subsystem + "_counter_event"

	// Value Metrics Subsystem.
	value_metrics_subsystem = subsystem + "_value_metric"
)
