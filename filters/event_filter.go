package filters

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/sonde-go/events"
)

type EventFilter struct {
	eventsEnabled map[events.Envelope_EventType]bool
}

func NewEventFilter(filter []string) (*EventFilter, error) {
	eventsEnabled := make(map[events.Envelope_EventType]bool)

	for _, eventName := range filter {
		eventType, err := parseEventName(eventName)
		if err != nil {
			return nil, err
		}

		eventsEnabled[eventType] = true
	}

	return &EventFilter{eventsEnabled: eventsEnabled}, nil
}

func (f *EventFilter) Enabled(envelope *events.Envelope) bool {
	if len(f.eventsEnabled) > 0 {
		if f.eventsEnabled[envelope.GetEventType()] {
			return true
		}

		return false
	}

	return true
}

func parseEventName(name string) (events.Envelope_EventType, error) {
	if eventType, ok := events.Envelope_EventType_value[name]; ok {
		switch events.Envelope_EventType(eventType) {
		case events.Envelope_ContainerMetric:
		case events.Envelope_CounterEvent:
		case events.Envelope_HttpStartStop:
		case events.Envelope_ValueMetric:
		default:
			return events.Envelope_Error, errors.New(fmt.Sprintf("Event filter `%s` is not supported", name))
		}
		return events.Envelope_EventType(eventType), nil
	}

	return events.Envelope_Error, errors.New(fmt.Sprintf("Event filter `%s` is not supported", name))
}
