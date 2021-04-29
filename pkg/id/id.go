package id

import (
	"bpxe.org/pkg/tracing"
)

type GeneratorBuilder interface {
	NewIdGenerator(tracer *tracing.Tracer) (Generator, error)
	RestoreIdGenerator(serialized []byte, tracer *tracing.Tracer) (Generator, error)
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
