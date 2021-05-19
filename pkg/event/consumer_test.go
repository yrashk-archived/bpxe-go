// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package event

import (
	"errors"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/assert"
)

type erroringConsumer struct{}

func (c erroringConsumer) ConsumeProcessEvent(ProcessEvent) (ConsumptionResult, error) {
	return ConsumptionError, errors.New("some error")
}

type testEvent struct{}

func (t testEvent) MatchesEventInstance(instance Instance) bool {
	return false
}

func TestForwardProcessEvent(t *testing.T) {
	someErroringConsumers := []ProcessEventConsumer{erroringConsumer{},
		VoidProcessEventConsumer{}}
	noErroringConsumers := []ProcessEventConsumer{VoidProcessEventConsumer{},
		VoidProcessEventConsumer{}}
	allErroringConsumers := []ProcessEventConsumer{erroringConsumer{},
		erroringConsumer{}}
	var result ConsumptionResult
	var multiErr *multierror.Error
	var err error
	var ok bool

	result, err = ForwardProcessEvent(testEvent{}, &someErroringConsumers)
	assert.Equal(t, PartiallyConsumed, result)
	assert.NotNil(t, err)
	ok = errors.As(err, &multiErr)
	assert.True(t, ok)
	assert.Equal(t, 1, multiErr.Len())

	result, err = ForwardProcessEvent(testEvent{}, &noErroringConsumers)
	assert.Equal(t, Consumed, result)
	assert.Nil(t, err)

	result, err = ForwardProcessEvent(testEvent{}, &allErroringConsumers)
	assert.Equal(t, ConsumptionError, result)
	assert.NotNil(t, err)
	ok = errors.As(err, &multiErr)
	assert.True(t, ok)
	assert.Equal(t, 2, multiErr.Len())

}
