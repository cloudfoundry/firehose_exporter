package rollup

import (
	"strings"

	"github.com/cloudfoundry/firehose_exporter/metrics"
	log "github.com/sirupsen/logrus"
)

type PointsBatch struct {
	Points []*metrics.RawMetric
	Size   int
}

type Rollup interface {
	Record(sourceID string, tags map[string]string, value int64)
	Rollup(timestamp int64) []*PointsBatch
}

func keyFromTags(rollupTags []string, sourceID string, tags map[string]string) string {
	filteredTags := []string{sourceID}

	for _, tag := range rollupTags {
		filteredTags = append(filteredTags, tags[tag])
	}
	return strings.Join(filteredTags, "%%")
}

func labelsFromKey(key, nodeIndex string, rollupTags []string) map[string]string {
	keyParts := strings.Split(key, "%%")

	if len(keyParts) != len(rollupTags)+1 {
		log.WithField("reason", "skipping rollup metric").WithField("count", len(keyParts)).WithField("key", key).Info(
			"skipping rollup metric",
		)
		return nil
	}

	labels := make(map[string]string)
	for index, tagName := range rollupTags {
		if value := keyParts[index+1]; value != "" {
			labels[tagName] = value
		}
	}

	labels["source_id"] = keyParts[0]
	labels["node_index"] = nodeIndex

	return labels
}
