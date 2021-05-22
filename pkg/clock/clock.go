// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package clock

import (
	"time"
)

// Clock is a generic interface for clocks
type Clock interface {
	// Now returns current time
	Now() time.Time
	// After returns a channel that will send one and only one
	// timestamp once it waited for the given duration
	//
	// Implementation of this method needs to guarantee that
	// the returned time is either equal or greater than the given
	// one plus duration
	After(time.Duration) <-chan time.Time
	// Until returns a channel that will send one and only one
	// timestamp once the provided time has occurred
	//
	// Implementation of this method needs to guarantee that
	// the returned time is either equal or greater than the given
	// one
	Until(time.Time) <-chan time.Time
	// Changes returns a channel that will send a message when
	// the clock detects a change (in case of a real clock, can
	// be a significant drift forward, or a drift backward; for
	// mock clock, an explicit change)
	Changes() <-chan time.Time
}
