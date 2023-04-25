package metrics

const (
	GorouterHTTPMetricName          = "http"
	GorouterHTTPCounterMetricName   = GorouterHTTPMetricName + "_total"
	GorouterHTTPHistogramMetricName = GorouterHTTPMetricName + "_duration_seconds"
	GorouterHTTPSummaryMetricName   = GorouterHTTPMetricName + "_response_size_bytes"
)
