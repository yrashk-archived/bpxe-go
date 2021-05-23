// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package clock

import (
	"context"
	"errors"
	"math"
	"time"

	"golang.org/x/sys/unix"
)

func changeMonitor(ctx context.Context, changes chan time.Time) (err error) {
	fd, err := unix.TimerfdCreate(unix.CLOCK_REALTIME, 0)
	if err != nil {
		return
	}

	err = unix.TimerfdSettime(fd, unix.TFD_TIMER_ABSTIME|unix.TFD_TIMER_CANCEL_ON_SET,
		&unix.ItimerSpec{
			Value: unix.Timespec{
				Sec:  math.MaxInt64,
				Nsec: 999_999_999,
			},
		}, nil)
	if err != nil {
		_ = unix.Close(fd)
		return
	}

	err = unix.SetNonblock(fd, true)
	if err != nil {
		_ = unix.Close(fd)
		return
	}

	go func(ctx context.Context) {
		defer func(fd int) {
			_ = unix.Close(fd)
		}(fd)

		buf := make([]byte, 8)
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second * 1):
				_, err := unix.Read(fd, buf)
				if err != nil {
					if errors.Is(err, unix.ECANCELED) {
						changes <- time.Now()
					}
				}
			}
		}
	}(ctx)
	return
}
