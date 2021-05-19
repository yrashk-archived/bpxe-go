// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package data

import (
	"context"

	"bpxe.org/pkg/bpmn"
)

type runnerMessage interface {
	implementsRunnerMessage()
}

type getMessage struct {
	channel chan Item
}

func (g getMessage) implementsRunnerMessage() {}

type putMessage struct {
	item Item
}

func (p putMessage) implementsRunnerMessage() {}

type Container struct {
	bpmn.ItemAwareInterface
	runnerChannel chan runnerMessage
	item          Item
}

func NewContainer(ctx context.Context, itemAware bpmn.ItemAwareInterface) *Container {
	container := &Container{
		ItemAwareInterface: itemAware,
		runnerChannel:      make(chan runnerMessage, 1),
	}
	go container.run(ctx)
	return container
}

func (c *Container) Unavailable() bool {
	return false
}

func (c *Container) run(ctx context.Context) {
	for {
		select {
		case msg := <-c.runnerChannel:
			switch msg := msg.(type) {
			case getMessage:
				msg.channel <- c.item
			case putMessage:
				c.item = msg.item
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *Container) Get() <-chan Item {
	ch := make(chan Item)
	c.runnerChannel <- getMessage{channel: ch}
	return ch
}

func (c *Container) Put(item Item) {
	c.runnerChannel <- putMessage{item: item}
}
