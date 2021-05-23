// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package logic

import (
	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/event"

	"github.com/bits-and-blooms/bitset"
)

// CatchEventSatisfier is an algorithm that allows to apply events to a bpmn.CatchEventInterface element
// and obtain a determination whether all conditions were satisfied.
type CatchEventSatisfier struct {
	bpmn.CatchEventInterface
	eventDefinitionInstances []event.DefinitionInstance
	len                      uint
	chains                   []*bitset.BitSet
}

func (satisfier *CatchEventSatisfier) EventDefinitionInstances() *[]event.DefinitionInstance {
	return &satisfier.eventDefinitionInstances
}

func NewCatchEventSatisfier(catchEventElement bpmn.CatchEventInterface, eventDefinitionInstanceBuilder event.DefinitionInstanceBuilder) *CatchEventSatisfier {
	satisfier := &CatchEventSatisfier{
		CatchEventInterface: catchEventElement,
		chains:              make([]*bitset.BitSet, 0, 1),
		len:                 uint(len(catchEventElement.EventDefinitions())),
	}

	satisfier.eventDefinitionInstances = make([]event.DefinitionInstance, len(catchEventElement.EventDefinitions()))
	for k := range catchEventElement.EventDefinitions() {
		satisfier.eventDefinitionInstances[k], _ = eventDefinitionInstanceBuilder.NewEventDefinitionInstance(catchEventElement.EventDefinitions()[k])
	}

	return satisfier
}

const EventDidNotMatch = -1

// Satisfy matches an event against event definitions in CatchEvent element,
// if all conditions are satisfied, it'll return true, otherwise, false.
//
// Satisfy also returns the index of the chain operated on. Chain is a partial
// receipt of a parallel multiple event sequence.
//
// * If event didn't match, the value will be equal to EventDidNotMatch
// * If event matched, `chain` will be the matching chain's index
// * If event matched and it was not a parallel multiple catch event, or
//   parallel multiple with just one event definition, `chain` will be equal
//   to `0`
//
// It is important to mention how chains get re-ordered upon their removal.
// Chain with the largest index (the last one) gets moved to the index of
// the removed chain and the array is shrunk by one element at the end.
// The knowledge of this behavior is important for being able to mirror
// changes if necessary.
//
// Please note that Satisfy is NOT goroutine-safe and if you need to use
// it from multiple goroutines, wrap its usage with appropriate level of
// synchronization.
func (satisfier *CatchEventSatisfier) Satisfy(ev event.Event) (matched bool, chain int) {
	chain = EventDidNotMatch
	for i := range satisfier.eventDefinitionInstances {
		if ev.MatchesEventInstance(satisfier.eventDefinitionInstances[i]) {
			if !satisfier.ParallelMultiple() || satisfier.len == 1 {
				chain = 0
				matched = true
				return
			} else {
				// If there are no chains of events,
				if len(satisfier.chains) == 0 {
					bitSet := bitset.New(satisfier.len)
					bitSet.Set(uint(i))
					// create the first one
					satisfier.chains = append(satisfier.chains, bitSet)
					chain = len(satisfier.chains) - 1
				} else {
					// For every existing chain
					for j := range satisfier.chains {
						// If it doesn't have this event yet,
						if !satisfier.chains[j].Test(uint(i)) {
							// Add it to the chain
							satisfier.chains[j].Set(uint(i))
							// And check if the chain has been fully satisfied
							matched = satisfier.chains[j].All()
							if matched {
								// If it has, remove the chain
								satisfier.chains[j] = satisfier.chains[len(satisfier.chains)-1]
								satisfier.chains = satisfier.chains[:len(satisfier.chains)-1]
							}
							chain = j
							return
						}
					}
					// If no existing chain had this event not processed, create a new one
					bitSet := bitset.New(satisfier.len)
					bitSet.Set(uint(i))
					// create the first one
					satisfier.chains = append(satisfier.chains, bitSet)
					chain = len(satisfier.chains) - 1
				}
			}
			break
		}
	}
	return
}
