package tracing

type Tracer struct {
	traces       chan Trace
	subscription chan chan Trace
	subscribers  []chan Trace
}

func NewTracer() *Tracer {
	tracer := Tracer{
		subscribers:  make([]chan Trace, 0),
		traces:       make(chan Trace),
		subscription: make(chan chan Trace),
	}
	go tracer.runner()
	return &tracer
}

func (t *Tracer) runner() {
	for {
		select {
		case subscription := <-t.subscription:
			t.subscribers = append(t.subscribers, subscription)
		case trace := <-t.traces:
			subscribers := t.subscribers[:0]
			for _, subscriber := range t.subscribers {
				select {
				case subscriber <- trace:
					// success
					subscribers = append(subscribers, subscriber)
				default:
					// unsubcribe closed channel
				}
			}
			t.subscribers = subscribers
		}
	}
}

func (t *Tracer) Subscribe() chan Trace {
	channel := make(chan Trace)
	t.subscription <- channel
	return channel
}

func (t *Tracer) Trace(trace Trace) {
	t.traces <- trace
}
