// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package event

import (
	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/data"
)

type Event interface {
	MatchesEventInstance(DefinitionInstance) bool
}

// Process has ended
type EndEvent struct {
	Element *bpmn.EndEvent
}

func MakeEndEvent(element *bpmn.EndEvent) EndEvent {
	return EndEvent{Element: element}
}

func (ev EndEvent) MatchesEventInstance(instance DefinitionInstance) bool {
	// always false because there's no event definition that matches
	return false
}

// None event
type NoneEvent struct{}

func MakeNoneEvent() NoneEvent {
	return NoneEvent{}
}

func (ev NoneEvent) MatchesEventInstance(instance DefinitionInstance) bool {
	// always false because there's no event definition that matches
	return false
}

// Signal event
type SignalEvent struct {
	signalRef string
	item      data.Item
}

func MakeSignalEvent(signalRef string, items ...data.Item) SignalEvent {
	return SignalEvent{signalRef: signalRef, item: data.ItemOrCollection(items)}
}

func NewSignalEvent(signalRef string, items ...data.Item) *SignalEvent {
	event := MakeSignalEvent(signalRef, items)
	return &event
}

func (ev *SignalEvent) MatchesEventInstance(instance DefinitionInstance) bool {
	definition, ok := instance.EventDefinition().(*bpmn.SignalEventDefinition)
	if !ok {
		return false
	}
	signalRef, present := definition.SignalRef()
	if !present {
		return false
	}
	return *signalRef == ev.signalRef
}

func (ev *SignalEvent) SignalRef() *string {
	return &ev.signalRef
}

// Cancellation event
type CancelEvent struct{}

func MakeCancelEvent() CancelEvent {
	return CancelEvent{}
}

func (ev CancelEvent) MatchesEventInstance(instance DefinitionInstance) bool {
	_, ok := instance.EventDefinition().(*bpmn.CancelEventDefinition)
	return ok
}

// Termination event
type TerminateEvent struct{}

func MakeTerminateEvent() TerminateEvent {
	return TerminateEvent{}
}

func (ev TerminateEvent) MatchesEventInstance(instance DefinitionInstance) bool {
	_, ok := instance.EventDefinition().(*bpmn.TerminateEventDefinition)
	return ok
}

// Compensation event
type CompensationEvent struct {
	activityRef string
}

func MakeCompensationEvent(activityRef string) CompensationEvent {
	return CompensationEvent{activityRef: activityRef}
}

func (ev *CompensationEvent) MatchesEventInstance(instance DefinitionInstance) bool {
	// always false because there's no event definition that matches
	return false
}

func (ev *CompensationEvent) ActivityRef() *string {
	return &ev.activityRef
}

// Message event
type MessageEvent struct {
	messageRef   string
	operationRef *string
	item         data.Item
}

func MakeMessageEvent(messageRef string, operationRef *string, items ...data.Item) MessageEvent {
	return MessageEvent{
		messageRef:   messageRef,
		operationRef: operationRef,
		item:         data.ItemOrCollection(items),
	}
}

func NewMessageEvent(messageRef string, operationRef *string, items ...data.Item) *MessageEvent {
	event := MakeMessageEvent(messageRef, operationRef, items...)
	return &event
}

func (ev *MessageEvent) MatchesEventInstance(instance DefinitionInstance) bool {
	definition, ok := instance.EventDefinition().(*bpmn.MessageEventDefinition)
	if !ok {
		return false
	}
	messageRef, present := definition.MessageRef()
	if !present {
		return false
	}
	if *messageRef != ev.messageRef {
		return false
	}
	if ev.operationRef == nil {
		if _, present := definition.OperationRef(); present {
			return false
		}
		return true
	} else {
		operationRef, present := definition.OperationRef()
		if !present {
			return false
		}
		return *operationRef == *ev.operationRef
	}
}

func (ev *MessageEvent) MessageRef() *string {
	return &ev.messageRef
}

func (ev *MessageEvent) OperationRef() (result *string, present bool) {
	if ev.operationRef != nil {
		result = ev.operationRef
		present = true
	}
	return
}

// Escalation event
type EscalationEvent struct {
	escalationRef string
	item          data.Item
}

func MakeEscalationEvent(escalationRef string, items ...data.Item) EscalationEvent {
	return EscalationEvent{escalationRef: escalationRef, item: data.ItemOrCollection(items)}
}

func (ev *EscalationEvent) MatchesEventInstance(instance DefinitionInstance) bool {
	definition, ok := instance.EventDefinition().(*bpmn.EscalationEventDefinition)
	if !ok {
		return false
	}
	escalationRef, present := definition.EscalationRef()
	if !present {
		return false
	}
	return ev.escalationRef == *escalationRef
}

func (ev *EscalationEvent) EscalationRef() *string {
	return &ev.escalationRef
}

// Link event
type LinkEvent struct {
	sources []string
	target  *string
}

func MakeLinkEvent(sources []string, target *string) LinkEvent {
	return LinkEvent{
		sources: sources,
		target:  target,
	}
}

func (ev *LinkEvent) MatchesEventInstance(instance DefinitionInstance) bool {
	definition, ok := instance.EventDefinition().(*bpmn.LinkEventDefinition)
	if !ok {
		return false
	}

	if ev.target == nil {
		if _, present := definition.Target(); present {
			return false
		}
	} else {
		target, present := definition.Target()
		if !present {
			return false
		}

		if *target != *ev.target {
			return false
		}

	}

	if definition.Sources() == nil {
		return false
	}

	sources := definition.Sources()

	if len(ev.sources) != len(*sources) {
		return false
	}

	for i := range ev.sources {
		if ev.sources[i] != (*sources)[i] {
			return false
		}
	}

	return true
}

func (ev *LinkEvent) Sources() *[]string {
	return &ev.sources
}

func (ev *LinkEvent) Target() (result *string, present bool) {
	if ev.target != nil {
		result = ev.target
		present = true
	}
	return
}

// Error event
type ErrorEvent struct {
	errorRef string
	item     data.Item
}

func MakeErrorEvent(errorRef string, items ...data.Item) ErrorEvent {
	return ErrorEvent{errorRef: errorRef, item: data.ItemOrCollection(items)}
}

func (ev *ErrorEvent) MatchesEventInstance(instance DefinitionInstance) bool {
	definition, ok := instance.EventDefinition().(*bpmn.ErrorEventDefinition)
	if !ok {
		return false
	}
	errorRef, present := definition.ErrorRef()
	if !present {
		return false
	}
	return *errorRef == ev.errorRef
}

func (ev *ErrorEvent) ErrorRef() *string {
	return &ev.errorRef
}

// TimerEvent represents an event that occurs when a certain timer
// is triggered.
type TimerEvent struct {
	instance DefinitionInstance
}

func MakeTimerEvent(instance DefinitionInstance) TimerEvent {
	return TimerEvent{instance: instance}
}

func (ev TimerEvent) MatchesEventInstance(instance DefinitionInstance) bool {
	return instance == ev.instance
}

func (ev *TimerEvent) Instance() DefinitionInstance {
	return ev.instance
}

// ConditionalEvent represents an event that occurs when a certain timer
// is triggered.
type ConditionalEvent struct {
	instance DefinitionInstance
}

func MakeConditionalEvent(instance DefinitionInstance) ConditionalEvent {
	return ConditionalEvent{instance: instance}
}

func (ev ConditionalEvent) MatchesEventInstance(instance DefinitionInstance) bool {
	return instance == ev.instance
}

func (ev *ConditionalEvent) Instance() DefinitionInstance {
	return ev.instance
}
