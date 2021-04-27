// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package flow_node

import (
	"sync"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/errors"
)

type FlowNodeMapping struct {
	mapping map[string]FlowNodeInterface
	lock    sync.RWMutex
}

func NewLockedFlowNodeMapping() *FlowNodeMapping {
	mapping := &FlowNodeMapping{
		mapping: make(map[string]FlowNodeInterface),
		lock:    sync.RWMutex{},
	}
	mapping.lock.Lock()
	return mapping
}

func (mapping *FlowNodeMapping) RegisterElementToFlowNode(element bpmn.FlowNodeInterface,
	flowNode FlowNodeInterface) (err error) {
	if id, present := element.Id(); present {
		mapping.mapping[*id] = flowNode
	} else {
		err = errors.RequirementExpectationError{
			Expected: "All flow nodes must have an ID",
			Actual:   element,
		}
	}
	return
}

func (mapping *FlowNodeMapping) Finalize() {
	mapping.lock.Unlock()
}

func (mapping *FlowNodeMapping) ResolveElementToFlowNode(
	element bpmn.FlowNodeInterface,
) (flowNode FlowNodeInterface, found bool) {
	mapping.lock.RLock()
	if id, present := element.Id(); present {
		flowNode, found = mapping.mapping[*id]
	}
	mapping.lock.RUnlock()
	return
}
