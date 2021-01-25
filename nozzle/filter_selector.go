package nozzle

import (
	"strings"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
)

type FilterSelectorType int32

const (
	FilterSelectorType_CONTAINER_METRIC FilterSelectorType = 0
	FilterSelectorType_COUNTER_EVENT    FilterSelectorType = 1
	FilterSelectorType_HTTP_START_STOP  FilterSelectorType = 2
	FilterSelectorType_VALUE_METRIC     FilterSelectorType = 3
)

var FilterSelectorType_value = map[string]int32{
	"containermetric": 0,
	"counterevent":    1,
	"httpstartstop":   2,
	"http":            2,
	"valuemetric":     3,
}

type FilterSelector struct {
	containerMetricDisabled bool
	counterEventDisabled    bool
	httpStartStopDisabled   bool
	valueMetricDisabled     bool
}

func NewFilterSelector(filterSelectorNames ...string) *FilterSelector {
	if len(filterSelectorNames) == 0 {
		return &FilterSelector{}
	}

	return NewFilterSelectorForced(filterSelectorNames...)
}

func NewFilterSelectorForced(filterSelectorNames ...string) *FilterSelector {
	fs := &FilterSelector{
		containerMetricDisabled: true,
		counterEventDisabled:    true,
		httpStartStopDisabled:   true,
		valueMetricDisabled:     true,
	}
	fs.FiltersByNames(filterSelectorNames...)
	return fs
}

func (f *FilterSelector) DisableAll() {
	f.containerMetricDisabled = true
	f.counterEventDisabled = true
	f.httpStartStopDisabled = true
	f.valueMetricDisabled = true
}

func (f FilterSelector) ValueMetricDisabled() bool {
	return f.valueMetricDisabled
}

func (f FilterSelector) HttpStartStopDisabled() bool {
	return f.httpStartStopDisabled
}

func (f FilterSelector) ContainerMetricDisabled() bool {
	return f.containerMetricDisabled
}

func (f FilterSelector) CounterEventDisabled() bool {
	return f.counterEventDisabled
}

func (f FilterSelector) AllGaugeDisabled() bool {
	return f.containerMetricDisabled && f.valueMetricDisabled
}

func (f *FilterSelector) Filters(filterSelectorTypes ...FilterSelectorType) {
	for _, filterSelectorType := range filterSelectorTypes {
		switch filterSelectorType {
		case FilterSelectorType_CONTAINER_METRIC:
			f.containerMetricDisabled = false
		case FilterSelectorType_COUNTER_EVENT:
			f.counterEventDisabled = false
		case FilterSelectorType_HTTP_START_STOP:
			f.httpStartStopDisabled = false
		case FilterSelectorType_VALUE_METRIC:
			f.valueMetricDisabled = false
		}
	}
}

func (f *FilterSelector) FiltersByNames(filterSelectorNames ...string) {
	filterSelectorTypes := make([]FilterSelectorType, 0)
	for _, filterSelectorName := range filterSelectorNames {
		if selectorType, ok := FilterSelectorType_value[strings.ToLower(filterSelectorName)]; ok {
			filterSelectorTypes = append(filterSelectorTypes, FilterSelectorType(selectorType))
		}
	}
	f.Filters(filterSelectorTypes...)
}

func (f *FilterSelector) ToSelectorTypes() []*loggregator_v2.Selector {
	selectors := make([]*loggregator_v2.Selector, 0)
	if !f.AllGaugeDisabled() {
		selectors = append(selectors, &loggregator_v2.Selector{
			Message: &loggregator_v2.Selector_Gauge{
				Gauge: &loggregator_v2.GaugeSelector{},
			},
		})
	}
	if !f.CounterEventDisabled() {
		selectors = append(selectors, &loggregator_v2.Selector{
			Message: &loggregator_v2.Selector_Counter{
				Counter: &loggregator_v2.CounterSelector{},
			},
		})
	}
	if !f.HttpStartStopDisabled() {
		selectors = append(selectors, &loggregator_v2.Selector{
			Message: &loggregator_v2.Selector_Timer{
				Timer: &loggregator_v2.TimerSelector{},
			},
		})
	}
	return selectors
}
