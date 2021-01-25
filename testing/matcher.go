package testing

import (
	"fmt"
	"reflect"

	"github.com/bosh-prometheus/firehose_exporter/metrics"
	"github.com/bosh-prometheus/firehose_exporter/transform"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
)

func ContainPoint(expected interface{}) types.GomegaMatcher {
	return &containPointsMatcher{
		expected: []*metrics.RawMetric{expected.(*metrics.RawMetric)},
	}
}

func ContainPoints(expected interface{}) types.GomegaMatcher {
	return &containPointsMatcher{
		expected: expected,
	}
}

type containPointsMatcher struct {
	expected interface{}
}

func (matcher *containPointsMatcher) Match(actual interface{}) (success bool, err error) {
	expectedPoints := matcher.expected.([]*metrics.RawMetric)
	points := actual.([]*metrics.RawMetric)
	foundPoints := make([]bool, len(expectedPoints))

	for _, point := range points {
		for n, expectedPoint := range expectedPoints {

			matchValue := false

			if point.Metric().Counter != nil {
				matchValue = expectedPoint.Metric().Counter != nil && expectedPoint.Metric().Counter.GetValue() == point.Metric().Counter.GetValue()
			}
			if point.Metric().Gauge != nil {
				matchValue = expectedPoint.Metric().Gauge != nil && expectedPoint.Metric().Gauge.GetValue() == point.Metric().Gauge.GetValue()
			}
			if point.Metric().Summary != nil {
				matchValue = expectedPoint.Metric().Summary != nil &&
					expectedPoint.Metric().Summary.GetSampleSum() == point.Metric().Summary.GetSampleSum() &&
					expectedPoint.Metric().Summary.GetSampleCount() == point.Metric().Summary.GetSampleCount()
			}
			if point.Metric().Untyped != nil {
				matchValue = expectedPoint.Metric().Untyped != nil &&
					expectedPoint.Metric().Untyped.GetValue() == point.Metric().Untyped.GetValue()
			}
			if point.Metric().Histogram != nil {
				matchValue = true
				for _, bucket := range point.Metric().Histogram.GetBucket() {
					if expectedPoint.Metric().Histogram == nil || !matchValue {
						matchValue = false
						break
					}
					for _, expectedBucket := range expectedPoint.Metric().Histogram.GetBucket() {
						if bucket.GetUpperBound() != expectedBucket.GetUpperBound() {
							continue
						}
						matchValue = bucket.GetCumulativeCount() == expectedBucket.GetCumulativeCount()
						if !matchValue {
							break
						}
					}
				}
				gomega.ConsistOf()
			}
			if point.MetricName() == expectedPoint.MetricName() &&
				matchValue &&
				reflect.DeepEqual(transform.LabelPairsToLabelsMap(point.Metric().Label), transform.LabelPairsToLabelsMap(expectedPoint.Metric().Label)) {
				foundPoints[n] = true
				break
			}
		}
	}

	for _, found := range foundPoints {
		if !found {
			return false, nil
		}
	}

	return true, nil
}

func (matcher *containPointsMatcher) FailureMessage(actual interface{}) (message string) {
	var actualOutput string
	for _, a := range actual.([]*metrics.RawMetric) {
		actualOutput += format.Object(a, 1)
	}
	var expectedOutput string
	for _, a := range matcher.expected.([]*metrics.RawMetric) {
		expectedOutput += format.Object(a, 1)
	}
	return fmt.Sprintf("Expected\n%s\nto contain all the points \n%s", actualOutput, expectedOutput)
}

func (matcher *containPointsMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n%#v\nnot to contain all the points \n%#v", actual, matcher.expected)
}
