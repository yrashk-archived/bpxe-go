// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package id

import (
	"context"
	"encoding/json"

	"bpxe.org/pkg/tracing"

	"github.com/muyo/sno"
)

type Sno struct {
}

func GetSno() *Sno {
	sno_ := &Sno{}
	return sno_
}

type SnoGenerator struct {
	*sno.Generator
	tracer *tracing.Tracer
}

func (g *Sno) NewIdGenerator(ctx context.Context, tracer *tracing.Tracer) (result Generator, err error) {
	return g.RestoreIdGenerator(ctx, []byte{}, tracer)
}

func (g *Sno) RestoreIdGenerator(ctx context.Context, bytes []byte, tracer *tracing.Tracer) (result Generator, err error) {
	var snapshot *sno.GeneratorSnapshot
	if len(bytes) > 0 {
		snapshot = new(sno.GeneratorSnapshot)
		err = json.Unmarshal(bytes, snapshot)
		if err != nil {
			return
		}
	}
	sequenceOverflowNotificationChannel := make(chan *sno.SequenceOverflowNotification)
	go func(ctx context.Context) {
		for {
			select {
			case notification := <-sequenceOverflowNotificationChannel:
				tracer.Trace(tracing.WarningTrace{Warning: notification})
			case <-ctx.Done():
				return
			}
		}
	}(ctx)

	var generator *sno.Generator
	generator, err = sno.NewGenerator(snapshot, sequenceOverflowNotificationChannel)
	if err != nil {
		return
	}
	result = &SnoGenerator{Generator: generator, tracer: tracer}
	return
}

func (g *SnoGenerator) Snapshot() (result []byte, err error) {
	result, err = json.Marshal(g.Generator.Snapshot())
	return
}

type SnoId struct {
	sno.ID
}

func (g *SnoGenerator) New() Id {
	return &SnoId{ID: g.Generator.New(0)}
}

func (id *SnoId) String() string {
	return id.ID.String()
}

func (id *SnoId) Bytes() []byte {
	return id.ID.Bytes()
}

func init() {
	DefaultIdGeneratorBuilder = GetSno()
}
