// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package task

import (
	"context"
	"sync"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/flow/flow_interface"
	"bpxe.org/pkg/flow_node"
	"bpxe.org/pkg/flow_node/activity"
)

type message interface {
	message()
}

type nextActionMessage struct {
	response chan flow_node.Action
}

func (m nextActionMessage) message() {}

type cancelMessage struct {
	response chan bool
}

func (m cancelMessage) message() {}

type Task struct {
	*flow_node.Wiring
	element        *bpmn.Task
	runnerChannel  chan message
	activeBoundary chan bool
	bodyLock       sync.RWMutex
	body           func(*Task, context.Context) flow_node.Action
	cancel         context.CancelFunc
}

// SetBody override Task's body with an arbitrary function
//
// Since Task implements Abstract Task, it does nothing by default.
// This allows to add an implementation. Primarily used for testing.
func (node *Task) SetBody(body func(*Task, context.Context) flow_node.Action) {
	node.bodyLock.Lock()
	defer node.bodyLock.Unlock()
	node.body = body
}

func NewTask(ctx context.Context, startEvent *bpmn.Task) activity.Constructor {
	return func(wiring *flow_node.Wiring) (node activity.Activity, err error) {
		ctx, cancel := context.WithCancel(ctx)
		taskNode := &Task{
			Wiring:         wiring,
			element:        startEvent,
			runnerChannel:  make(chan message, len(wiring.Incoming)*2+1),
			activeBoundary: make(chan bool),
			cancel:         cancel,
		}
		go taskNode.runner(ctx)
		node = taskNode
		return
	}
}

func (node *Task) runner(ctx context.Context) {
	for {
		select {
		case msg := <-node.runnerChannel:
			switch m := msg.(type) {
			case cancelMessage:
				node.cancel()
				m.response <- true
			case nextActionMessage:
				node.activeBoundary <- true
				go func() {
					var action flow_node.Action
					action = flow_node.FlowAction{SequenceFlows: flow_node.AllSequenceFlows(&node.Outgoing)}
					if node.body != nil {
						node.bodyLock.RLock()
						action = node.body(node, ctx)
						node.bodyLock.RUnlock()
					}
					node.activeBoundary <- false
					m.response <- action
				}()
			default:
			}
		case <-ctx.Done():
			return
		}
	}
}

func (node *Task) NextAction(flow_interface.T) chan flow_node.Action {
	response := make(chan flow_node.Action)
	node.runnerChannel <- nextActionMessage{response: response}
	return response
}

func (node *Task) Element() bpmn.FlowNodeInterface {
	return node.element
}

func (node *Task) ActiveBoundary() <-chan bool {
	return node.activeBoundary
}

func (node *Task) Cancel() <-chan bool {
	response := make(chan bool)
	node.runnerChannel <- cancelMessage{response: response}
	return response
}
