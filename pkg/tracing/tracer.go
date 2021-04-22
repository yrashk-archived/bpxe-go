package tracing

type unsubscription struct {
	channel chan Trace
	ok      chan bool
}

type Tracer struct {
	traces         chan Trace
	subscription   chan chan Trace
	unsubscription chan unsubscription
	subscribers    []chan Trace
}

func NewTracer() *Tracer {
	tracer := Tracer{
		subscribers:    make([]chan Trace, 0),
		traces:         make(chan Trace),
		subscription:   make(chan chan Trace),
		unsubscription: make(chan unsubscription),
	}
	go tracer.runner()
	return &tracer
}

func (t *Tracer) runner() {
	for {
		select {
		case subscription := <-t.subscription:
			t.subscribers = append(t.subscribers, subscription)
		case unsubscription := <-t.unsubscription:
			pos := 0
			for i := range t.subscribers {
				if t.subscribers[i] == unsubscription.channel {
					pos = i + 1
					break
				}
			}
			if pos > 0 {
				end := pos
				if pos == len(t.subscribers) {
					pos--
				}
				t.subscribers = append(t.subscribers[:pos], t.subscribers[end:]...)
				unsubscription.ok <- true
			}
		case trace := <-t.traces:
			for _, subscriber := range t.subscribers {
				subscriber <- trace
			}
		}
	}
}

func (t *Tracer) Subscribe() chan Trace {
	channel := make(chan Trace)
	t.subscription <- channel
	return channel
}

func (t *Tracer) Unsubscribe(c chan Trace) {
	okChan := make(chan bool)
	unsub := unsubscription{channel: c, ok: okChan}
loop:
	for {
		select {
		case <-c:
			continue loop
		case t.unsubscription <- unsub:
			continue loop
		case <-okChan:
			return
		}
	}
}

func (t *Tracer) Trace(trace Trace) {
	t.traces <- trace
}
