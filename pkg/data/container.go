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
	item    Item
	channel chan struct{}
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
				close(msg.channel)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *Container) Get(ctx context.Context) <-chan Item {
	ch := make(chan Item)
	select {
	case c.runnerChannel <- getMessage{channel: ch}:
		return ch
	case <-ctx.Done():
		return nil
	}
}

func (c *Container) Put(ctx context.Context, item Item) <-chan struct{} {
	ch := make(chan struct{})
	select {
	case c.runnerChannel <- putMessage{item: item, channel: ch}:
		return ch
	case <-ctx.Done():
		return nil
	}
}
