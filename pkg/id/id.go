package id

import (
	"bpxe.org/pkg/tracing"
)

type IdGeneratorBuilder interface {
	NewIdGenerator(tracer *tracing.Tracer) (IdGenerator, error)
	RestoreIdGenerator(serialized []byte, tracer *tracing.Tracer) (IdGenerator, error)
}

type IdGenerator interface {
	Snapshot() ([]byte, error)
	New() Id
}

type Id interface {
	Bytes() []byte
	String() string
}

var DefaultIdGeneratorBuilder IdGeneratorBuilder
