// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

//+build !linux

package clock

import (
	"context"
	"time"
)

const hostForwardDriftTolerance = 3 * time.Second

func changeMonitor(ctx context.Context, changes chan time.Time) (err error) {
	go func(ctx context.Context) {
		for {
			t := time.Now()
			select {
			case <-ctx.Done():
				return
			case t1 := <-time.After(time.Second * 1):
				if t1.Before(t) {
					// backward drift
					changes <- t1
				} else if t1.Sub(t).Nanoseconds() > hostForwardDriftTolerance.Nanoseconds() {
					// forward drift
					changes <- t1
				}
			}
		}
	}(ctx)
	return
}
