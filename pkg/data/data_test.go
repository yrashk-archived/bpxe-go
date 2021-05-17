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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSliceIterator_ItemIterator(t *testing.T) {
	items := SliceIterator([]Item{1, "hello"})
	ctx := context.Background()
	iterator, _ := items.ItemIterator(ctx)
	newItems := SliceIterator(make([]Item, 0, len(items)))
	for e := range iterator {
		newItems = append(newItems, e)
	}
	assert.Equal(t, items, newItems)
}

func TestSliceIterator_ItemIterator_Stopping(t *testing.T) {
	items := SliceIterator([]Item{1, "hello"})
	ctx := context.Background()
	iterator, stopper := items.ItemIterator(ctx)
	stopper.Stop()
	value, ok := <-iterator
	assert.Nil(t, value)
	assert.False(t, ok)
}

func TestSliceIterator_ItemIterator_ContextDone(t *testing.T) {
	items := SliceIterator([]Item{1, "hello"})
	ctx, cancel := context.WithCancel(context.Background())
	iterator, stop := items.ItemIterator(ctx)
	cancel()

	// Wait until stopper is closed. This is definitely not
	// black box testing, but this essentially ensures
	// the cancellation has been handled
	stopper := stop.(channelIteratorStopper)
	<-stopper.ch

	value, ok := <-iterator
	assert.Nil(t, value)
	assert.False(t, ok)
}
