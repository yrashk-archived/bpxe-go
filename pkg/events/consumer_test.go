package events

import (
	"errors"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/assert"
)

type erroringConsumer struct{}

func (c erroringConsumer) ConsumeProcessEvent(ProcessEvent) (EventConsumptionResult, error) {
	return EventConsumptionError, errors.New("some error")
}

func TestForwardProcessEvent(t *testing.T) {
	someErroringConsumers := []ProcessEventConsumer{erroringConsumer{},
		VoidProcessEventConsumer{}}
	noErroringConsumers := []ProcessEventConsumer{VoidProcessEventConsumer{},
		VoidProcessEventConsumer{}}
	allErroringConsumers := []ProcessEventConsumer{erroringConsumer{},
		erroringConsumer{}}
	var result EventConsumptionResult
	var multiErr *multierror.Error
	var err error
	var ok bool

	result, err = ForwardProcessEvent(MakeStartEvent(), &someErroringConsumers)
	assert.Equal(t, EventPartiallyConsumed, result)
	assert.NotNil(t, err)
	ok = errors.As(err, &multiErr)
	assert.True(t, ok)
	assert.Equal(t, 1, multiErr.Len())

	result, err = ForwardProcessEvent(MakeStartEvent(), &noErroringConsumers)
	assert.Equal(t, EventConsumed, result)
	assert.Nil(t, err)

	result, err = ForwardProcessEvent(MakeStartEvent(), &allErroringConsumers)
	assert.Equal(t, EventConsumptionError, result)
	assert.NotNil(t, err)
	ok = errors.As(err, &multiErr)
	assert.True(t, ok)
	assert.Equal(t, 2, multiErr.Len())

}
