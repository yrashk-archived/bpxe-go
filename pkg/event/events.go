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
)

type ProcessEvent interface {
	MatchesEventInstance(Instance) bool
}

// Process has started
type StartEvent struct{}

func MakeStartEvent() StartEvent {
	return StartEvent{}
}

func (ev StartEvent) MatchesEventInstance(instance Instance) bool {
	// always false because there's no event definition that matches
	return false
}

// Process has ended
type EndEvent struct {
	Element *bpmn.EndEvent
}

func MakeEndEvent(element *bpmn.EndEvent) EndEvent {
	return EndEvent{Element: element}
}

func (ev EndEvent) MatchesEventInstance(instance Instance) bool {
	// always false because there's no event definition that matches
	return false
}

// None event
type NoneEvent struct{}

func MakeNoneEvent() NoneEvent {
	return NoneEvent{}
}

func (ev NoneEvent) MatchesEventInstance(instance Instance) bool {
	// always false because there's no event definition that matches
	return false
}

// Signal event
type SignalEvent struct {
	signalRef string
}

func MakeSignalEvent(signalRef string) SignalEvent {
	return SignalEvent{signalRef: signalRef}
}

func NewSignalEvent(signalRef string) *SignalEvent {
	event := MakeSignalEvent(signalRef)
	return &event
}

func (ev *SignalEvent) MatchesEventInstance(instance Instance) bool {
	def, ok := instance.(*definitionInstance)
	if !ok {
		return false
	}
	definition, ok := def.EventDefinitionInterface.(*bpmn.SignalEventDefinition)
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

func (ev CancelEvent) MatchesEventInstance(instance Instance) bool {
	def, ok := instance.(*definitionInstance)
	if !ok {
		return false
	}
	_, ok = def.EventDefinitionInterface.(*bpmn.CancelEventDefinition)
	return ok
}

// Termination event
type TerminateEvent struct{}

func MakeTerminateEvent() TerminateEvent {
	return TerminateEvent{}
}

func (ev TerminateEvent) MatchesEventInstance(instance Instance) bool {
	def, ok := instance.(*definitionInstance)
	if !ok {
		return false
	}
	_, ok = def.EventDefinitionInterface.(*bpmn.TerminateEventDefinition)
	return ok
}

// Compensation event
type CompensationEvent struct {
	activityRef string
}

func MakeCompensationEvent(activityRef string) CompensationEvent {
	return CompensationEvent{activityRef: activityRef}
}

func (ev *CompensationEvent) MatchesEventInstance(instance Instance) bool {
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
}

func MakeMessageEvent(messageRef string, operationRef *string) MessageEvent {
	return MessageEvent{
		messageRef:   messageRef,
		operationRef: operationRef,
	}
}

func NewMessageEvent(messageRef string, operationRef *string) *MessageEvent {
	event := MakeMessageEvent(messageRef, operationRef)
	return &event
}

func (ev *MessageEvent) MatchesEventInstance(instance Instance) bool {
	def, ok := instance.(*definitionInstance)
	if !ok {
		return false
	}
	definition, ok := def.EventDefinitionInterface.(*bpmn.MessageEventDefinition)
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
}

func MakeEscalationEvent(escalationRef string) EscalationEvent {
	return EscalationEvent{escalationRef: escalationRef}
}

func (ev *EscalationEvent) MatchesEventInstance(instance Instance) bool {
	def, ok := instance.(*definitionInstance)
	if !ok {
		return false
	}
	definition, ok := def.EventDefinitionInterface.(*bpmn.EscalationEventDefinition)
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

func (ev *LinkEvent) MatchesEventInstance(instance Instance) bool {
	def, ok := instance.(*definitionInstance)
	if !ok {
		return false
	}
	definition, ok := def.EventDefinitionInterface.(*bpmn.LinkEventDefinition)
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
}

func MakeErrorEvent(errorRef string) ErrorEvent {
	return ErrorEvent{errorRef: errorRef}
}

func (ev *ErrorEvent) MatchesEventInstance(instance Instance) bool {
	def, ok := instance.(*definitionInstance)
	if !ok {
		return false
	}
	definition, ok := def.EventDefinitionInterface.(*bpmn.ErrorEventDefinition)
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
	instance Instance
}

func MakeTimerEvent(instance Instance) TimerEvent {
	return TimerEvent{instance: instance}
}

func (ev TimerEvent) MatchesEventInstance(instance Instance) bool {
	return instance == ev.instance
}

func (ev *TimerEvent) Instance() Instance {
	return ev.instance
}

// ConditionalEvent represents an event that occurs when a certain timer
// is triggered.
type ConditionalEvent struct {
	instance Instance
}

func MakeConditionalEvent(instance Instance) ConditionalEvent {
	return ConditionalEvent{instance: instance}
}

func (ev ConditionalEvent) MatchesEventInstance(instance Instance) bool {
	return instance == ev.instance
}

func (ev *ConditionalEvent) Instance() Instance {
	return ev.instance
}
