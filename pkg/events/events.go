package events

// Marker trait for process events
type ProcessEvent interface {
	processEvent()
}

// Process has started
type StartEvent struct{}

func MakeStartEvent() StartEvent {
	return StartEvent{}
}

func (ev StartEvent) processEvent() {}

// Process has ended
type EndEvent struct{}

func MakeEndEvent() EndEvent {
	return EndEvent{}
}

func (ev EndEvent) processEvent() {}

// None event
type NoneEvent struct{}

func MakeNoneEvent() NoneEvent {
	return NoneEvent{}
}

func (ev NoneEvent) processEvent() {}

// Signal event
type SignalEvent struct {
	signalRef *string
}

func MakeSignalEvent(signalRef *string) SignalEvent {
	return SignalEvent{signalRef: signalRef}
}

func (ev SignalEvent) processEvent() {}

func (ev *SignalEvent) SignalRef() (result *string, present bool) {
	if ev.signalRef != nil {
		result = ev.signalRef
		present = true
	}
	return
}

// Cancellation event
type CancelEvent struct{}

func MakeCancelEvent() CancelEvent {
	return CancelEvent{}
}

func (ev CancelEvent) processEvent() {}

// Termination event
type TerminateEvent struct{}

func MakeTerminateEvent() TerminateEvent {
	return TerminateEvent{}
}

func (ev TerminateEvent) processEvent() {}

// Compensation event
type CompensationEvent struct {
	activityRef *string
}

func MakeCompensationEvent(activityRef *string) CompensationEvent {
	return CompensationEvent{activityRef: activityRef}
}

func (ev CompensationEvent) processEvent() {}

func (ev *CompensationEvent) ActivityRef() (result *string, present bool) {
	if ev.activityRef != nil {
		result = ev.activityRef
		present = true
	}
	return
}

// Message event
type MessageEvent struct {
	messageRef   *string
	operationRef *string
}

func MakeMessageEvent(messageRef, operationRef *string) MessageEvent {
	return MessageEvent{
		messageRef:   messageRef,
		operationRef: operationRef,
	}
}

func (ev MessageEvent) processEvent() {}

func (ev *MessageEvent) MessageRef() (result *string, present bool) {
	if ev.messageRef != nil {
		result = ev.messageRef
		present = true
	}
	return
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
	escalationRef *string
}

func MakeEscalationEvent(escalationRef *string) EscalationEvent {
	return EscalationEvent{escalationRef: escalationRef}
}

func (ev EscalationEvent) processEvent() {}

func (ev *EscalationEvent) EscalationRef() (result *string, present bool) {
	if ev.escalationRef != nil {
		result = ev.escalationRef
		present = true
	}
	return
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

func (ev LinkEvent) processEvent() {}

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
	errorRef *string
}

func MakeErrorEvent(errorRef *string) ErrorEvent {
	return ErrorEvent{errorRef: errorRef}
}

func (ev ErrorEvent) processEvent() {}

func (ev *ErrorEvent) ErrorRef() (result *string, present bool) {
	if ev.errorRef != nil {
		result = ev.errorRef
		present = true
	}
	return
}
