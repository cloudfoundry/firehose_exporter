package metricmaker

import (
	"sort"
	"strings"

	"github.com/bosh-prometheus/firehose_exporter/metrics"
	"github.com/bosh-prometheus/firehose_exporter/transform"
	"github.com/bosh-prometheus/firehose_exporter/utils"
	"github.com/gogo/protobuf/proto"
	dto "github.com/prometheus/client_model/go"
)

func NormalizeName(metric *metrics.RawMetric) {
	metric.SetMetricName(transform.NormalizeName(metric.MetricName()))
}

func SuffixCounterWithTotal(metric *metrics.RawMetric) {
	if *metric.MetricType() != dto.MetricType_COUNTER ||
		strings.HasSuffix(metric.MetricName(), "_total") ||
		strings.HasSuffix(metric.MetricName(), "_delta") ||
		utils.MetricIsContainerMetric(metric) {
		return
	}
	metric.SetMetricName(metric.MetricName() + "_total")
}

func OrderAndSanitizeLabels(metric *metrics.RawMetric) {
	metricDto := metric.Metric()
	labels := make([]*dto.LabelPair, 0)
	for _, label := range metricDto.Label {
		if strings.HasPrefix(label.GetName(), "__") {
			continue
		}
		if strings.Contains(label.GetName(), "-") {
			label.Name = proto.String(strings.Replace(label.GetName(), "-", "_", -1))
		}
		labels = append(labels, label)
	}

	sort.Slice(labels, func(i, j int) bool {
		return labels[i].GetName() < labels[j].GetName()
	})
	metricDto.Label = labels
}

func PresetLabels(metric *metrics.RawMetric) {
	metricDto := metric.Metric()
	labels := metricDto.Label
	labels = transform.PlaceConstLabelInLabelPair(labels, "bosh_deployment", false, "deployment")
	labels = transform.PlaceConstLabelInLabelPair(labels, "bosh_job_name", false, "job")
	labels = transform.PlaceConstLabelInLabelPair(labels, "bosh_job_id", false, "index")
	labels = transform.PlaceConstLabelInLabelPair(labels, "bosh_job_ip", false, "ip")
	labels = transform.PlaceConstLabelInLabelPair(labels, "application_id", false, "app_id")
	labels = transform.PlaceConstLabelInLabelPair(labels, "application_name", false, "app_name")
	metricDto.Label = labels
}

func AddNamespace(namespace string) func(metric *metrics.RawMetric) {
	return func(metric *metrics.RawMetric) {
		metric.SetMetricName(namespace + "_" + metric.MetricName())
	}
}

func InjectMapLabel(m map[string]string) func(metric *metrics.RawMetric) {
	return func(metric *metrics.RawMetric) {
		metricDto := metric.Metric()
		for k, v := range m {
			metricDto.Label = append(metricDto.Label, &dto.LabelPair{
				Name:  proto.String(k),
				Value: proto.String(v),
			})
		}
	}
}

func FindAndReplaceByName(oldMetricName, newMetricName string) func(metric *metrics.RawMetric) {
	return func(metric *metrics.RawMetric) {
		if metric.MetricName() != oldMetricName {
			return
		}
		metric.SetMetricName(newMetricName)
	}
}
