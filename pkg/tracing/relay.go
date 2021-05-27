// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package tracing

import (
	"context"
)

type Transformer func(trace Trace) []Trace

// NewRelay starts a goroutine that relays traces from `in` Tracer to `out` Tracer
// using a given Transformer
//
// This is typically used to compartmentalize tracers.
func NewRelay(ctx context.Context, in, out Tracer, transformer Transformer) {
	ch := in.Subscribe()
	handle := out.RegisterSender()
	go func() {
		for {
			select {
			case <-in.Done():
				handle.Done()
				return
			case <-ctx.Done():
				// wait until `in` Tracer is done
			case trace, ok := <-ch:
				if ok {
					traces := transformer(trace)
					for _, trace := range traces {
						out.Trace(trace)
					}
				}
			}
		}
	}()
}
