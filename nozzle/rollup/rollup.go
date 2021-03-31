package rollup

import (
	"github.com/bosh-prometheus/firehose_exporter/metrics"
	log "github.com/sirupsen/logrus"
	"strings"
)

type PointsBatch struct {
	Points []*metrics.RawMetric
	Size   int
}

type Rollup interface {
	Record(sourceId string, tags map[string]string, value int64)
	Rollup(timestamp int64) []*PointsBatch
}

func keyFromTags(rollupTags []string, sourceId string, tags map[string]string) string {
	filteredTags := []string{sourceId}

	for _, tag := range rollupTags {
		filteredTags = append(filteredTags, tags[tag])
	}
	return strings.Join(filteredTags, "%%")
}

func labelsFromKey(key, nodeIndex string, rollupTags []string) (map[string]string, error) {
	keyParts := strings.Split(key, "%%")

	if len(keyParts) != len(rollupTags)+1 {
		log.WithField("reason", "skipping rollup metric").WithField("count", len(keyParts)).WithField("key", key).Info(
			"skipping rollup metric",
		)
		return nil, nil
	}

	labels := make(map[string]string)
	for index, tagName := range rollupTags {
		if value := keyParts[index+1]; value != "" {
			labels[tagName] = value
		}
	}

	labels["source_id"] = keyParts[0]
	labels["node_index"] = nodeIndex

	return labels, nil
}
