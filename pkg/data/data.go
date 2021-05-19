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

// Item is an abstract interface for a piece of data
type Item interface {
}

// IteratorStopper stops Collection iterator and releases resources
// associated with it
type IteratorStopper interface {
	// Stop does the actual stopping
	Stop()
}

type channelIteratorStopper struct {
	ch chan struct{}
}

func makeChannelIteratorStopper() channelIteratorStopper {
	return channelIteratorStopper{ch: make(chan struct{})}
}

func (c channelIteratorStopper) close() {
	close(c.ch)
}

func (c channelIteratorStopper) Stop() {
	c.ch <- struct{}{}
}

type Collection interface {
	Item
	// ItemIterator returns a channel that iterates over collection's
	// items and an IteratorStopper that must be used if iterator was
	// not exhausted, otherwise there'll be a memory leak in a form
	// of a goroutine that does nothing.
	//
	// The iterator will also clean itself up and terminate upon
	// context termination.
	ItemIterator(ctx context.Context) (chan Item, IteratorStopper)
}

type SliceIterator []Item

func (s *SliceIterator) ItemIterator(ctx context.Context) (items chan Item, stop IteratorStopper) {
	items = make(chan Item)
	stopper := makeChannelIteratorStopper()
	stop = stopper
	go func() {
	loop:
		for i := range *s {
			select {
			case <-ctx.Done():
				break loop
			case <-stopper.ch:
				break loop
			case items <- (*s)[i]:
			}
		}
		close(items)
		stopper.close()
	}()
	return
}

// ItemOrCollection will return nil if no items given,
// the same item if only one item is given and SliceIterator
// if more than one item is given. SliceIterator implements
// Collection and, therefore, also implements Item.
func ItemOrCollection(items ...Item) (item Item) {
	switch len(items) {
	case 0:
	case 1:
		item = items[0]
	default:
		item = SliceIterator(items)
	}
	return
}

// ItemAware provides basic interface of accessing data items
type ItemAware interface {
	// Unavailable returns true if the data item is an unavailable state
	Unavailable() bool
	// Get returns a channel that will eventually return the data item
	//
	// If item is in an unavailable state (see Unavailable),
	// this channel will not send anything until the item becomes available
	Get() <-chan Item
	// Put sends a request to update the item
	//
	// If item is in an unavailable state (see Unavailable),
	// the data will not update until the item becomes available
	Put(Item)
}

// ItemAwareLocator interface describes a way to find ItemAware
type ItemAwareLocator interface {
	// FindItemAwareById finds ItemAware by its bpmn.Id
	FindItemAwareById(id bpmn.IdRef) (itemAware ItemAware, found bool)
	// FindItemAwareByName finds ItemAware by its name (where applicable)
	FindItemAwareByName(name string) (itemAware ItemAware, found bool)
}
