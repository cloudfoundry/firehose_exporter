package metrics

const (
	GorouterHttpMetricName          = "http"
	GorouterHttpCounterMetricName   = GorouterHttpMetricName + "_total"
	GorouterHttpHistogramMetricName = GorouterHttpMetricName + "_duration_seconds"
	GorouterHttpSummaryMetricName   = GorouterHttpMetricName + "_response_size_bytes"
)
