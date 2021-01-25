package transform

import (
	"sort"
	"strings"

	"github.com/gogo/protobuf/proto"
	dto "github.com/prometheus/client_model/go"
)

func LabelsMapToLabelPairs(labelValues map[string]string) []*dto.LabelPair {
	labels := make([]*dto.LabelPair, 0)
	labelKeys := make([]string, 0)
	for k := range labelValues {
		if strings.HasPrefix(k, "__") {
			continue
		}
		labelKeys = append(labelKeys, k)
	}
	sort.Strings(labelKeys)

	for _, k := range labelKeys {
		labels = append(labels, &dto.LabelPair{
			Name:  proto.String(k),
			Value: proto.String(labelValues[k]),
		})
	}
	return labels
}

func LabelPairsToLabelsMap(labelPair []*dto.LabelPair) map[string]string {
	m := make(map[string]string)
	for _, label := range labelPair {
		m[label.GetName()] = label.GetValue()
	}
	return m
}

func PlaceConstLabelInLabelPair(labels []*dto.LabelPair, constKey string, required bool, possibleKeysOrigin ...string) []*dto.LabelPair {
	possibleKeysOrigin = append([]string{constKey}, possibleKeysOrigin...)
	for _, keyOrigin := range possibleKeysOrigin {
		for _, label := range labels {
			if label.GetName() == keyOrigin {
				labels = append(labels, &dto.LabelPair{
					Name:  proto.String(constKey),
					Value: proto.String(label.GetValue()),
				})
				return labels
			}
		}

	}
	if required {
		labels = append(labels, &dto.LabelPair{
			Name:  proto.String(constKey),
			Value: proto.String(""),
		})
	}

	return labels
}
