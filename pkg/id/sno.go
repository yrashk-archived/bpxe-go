package id

import (
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

func (g *Sno) NewIdGenerator(tracer *tracing.Tracer) (result Generator, err error) {
	return g.RestoreIdGenerator([]byte{}, tracer)
}

func (g *Sno) RestoreIdGenerator(bytes []byte, tracer *tracing.Tracer) (result Generator, err error) {
	var snapshot *sno.GeneratorSnapshot
	if len(bytes) > 0 {
		snapshot = new(sno.GeneratorSnapshot)
		err = json.Unmarshal(bytes, snapshot)
		if err != nil {
			return
		}
	}
	sequenceOverflowNotificationChannel := make(chan *sno.SequenceOverflowNotification)
	go func() {
		for {
			notification := <-sequenceOverflowNotificationChannel
			tracer.Trace(tracing.WarningTrace{Warning: notification})
		}
	}()

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
