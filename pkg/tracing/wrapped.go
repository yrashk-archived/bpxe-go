package tracing

// WrappedTrace is a trace that wraps another trace.
//
// The purpose of it is to allow components to produce traces that will
// be wrapped into additional context, without being aware of it.
//
// Typically this would be done by creating a NewTraceTransformingTracer tracer
// over the original one and passing it to such components.
//
// Consumers looking for individual traces should use Unwrap to retrieve
// the original trace (as opposed to the wrapped one)
type WrappedTrace interface {
	Trace
	// Unwrap returns a wrapped trace
	Unwrap() Trace
}

// Unwrap will recursively unwrap a trace if wrapped,
// or return the trace as is if it isn't wrapped.
func Unwrap(trace Trace) Trace {
	for {
		if unwrapped, ok := trace.(WrappedTrace); ok {
			trace = unwrapped.Unwrap()
		} else {
			return trace
		}
	}
}
