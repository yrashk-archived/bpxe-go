package event

import (
	"github.com/hashicorp/go-multierror"
)

// Result of event consumption
type ConsumptionResult int

const (
	// Consumer has successfully consumed the event
	Consumed ConsumptionResult = iota
	// Consumer had an unexpected error consuming the event
	//
	// The error is returned as a second value.
	ConsumptionError
	// Consumer with multiple sub-consumers was able to
	// have at least some of the sub-consumers consume the event
	// but there were some errors (the errors are returned as
	// `multierror` list in the second value)
	PartiallyConsumed
	// Event consumer should no longer receive events
	//
	// This is a mechanism for "unsubscribing" from an event source
	DropEventConsumer
)

// Process event consumer interface
type ProcessEventConsumer interface {
	ConsumeProcessEvent(ProcessEvent) (ConsumptionResult, error)
}

// Process event consumer that does nothing and returns EventConsumed result
type VoidProcessEventConsumer struct{}

func (t VoidProcessEventConsumer) ConsumeProcessEvent(
	ev ProcessEvent,
) (result ConsumptionResult, err error) {
	result = Consumed
	return
}

// Forwards a process event to a list of consumers
//
// This function handles full and partial failure of consumers to consume the event.
// If none of them errors out, the result will be EventConsumed.
// If some do, the result will be EventPartiallyConsumed and err will be *multierror.Errors
// If all do, the result will be EventConsumptionError and err will be *multierror.Errors
func ForwardProcessEvent(ev ProcessEvent, eventConsumers *[]ProcessEventConsumer) (result ConsumptionResult, err error) {
	var errors *multierror.Error
	for _, consumer := range *eventConsumers {
		result, consumerError := consumer.ConsumeProcessEvent(ev)
		if result == ConsumptionError && consumerError != nil {
			errors = multierror.Append(errors, consumerError)
		}
	}
	switch {
	case errors != nil && errors.Len() > 0 && errors.Len() < len(*eventConsumers):
		// if there were errors, but not in all deliveries
		result = PartiallyConsumed
		err = errors
	case errors != nil && errors.Len() > 0:
		// if there were errors everywhere
		result = ConsumptionError
		err = errors
	default:
		// if there were no errors
		result = Consumed
	}
	return
}
