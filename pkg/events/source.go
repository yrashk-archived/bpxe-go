package events

type ProcessEventSource interface {
	RegisterProcessEventConsumer(ProcessEventConsumer) error
}

type VoidProcessEventSource struct{}

func (t VoidProcessEventSource) RegisterProcessEventConsumer(
	consumer ProcessEventConsumer,
) (err error) {
	return
}
