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

	"bpxe.org/pkg/tracing"
)

type GeneratorBuilder interface {
	NewIdGenerator(ctx context.Context, tracer *tracing.Tracer) (Generator, error)
	RestoreIdGenerator(ctx context.Context, serialized []byte, tracer *tracing.Tracer) (Generator, error)
}

type Generator interface {
	Snapshot() ([]byte, error)
	New() Id
}

type Id interface {
	Bytes() []byte
	String() string
}

var DefaultIdGeneratorBuilder GeneratorBuilder
