// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package event

import (
	"sync"
)

// FanOut is a straightforward Consumer + Source, forwards all consumed
// messages to all subscribers registered at it.
type FanOut struct {
	eventConsumersLock sync.RWMutex
	eventConsumers     []Consumer
}

func NewFanOut() *FanOut {
	return &FanOut{}
}

func (f *FanOut) ConsumeEvent(ev Event) (result ConsumptionResult, err error) {
	f.eventConsumersLock.RLock()
	result, err = ForwardEvent(ev, &f.eventConsumers)
	f.eventConsumersLock.RUnlock()
	return
}

func (f *FanOut) RegisterEventConsumer(ev Consumer) (err error) {
	f.eventConsumersLock.Lock()
	f.eventConsumers = append(f.eventConsumers, ev)
	f.eventConsumersLock.Unlock()
	return
}
