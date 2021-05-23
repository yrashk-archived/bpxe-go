// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

// Package main in pkg/clock/manual_test is meant for manual testing
// of time change observations.
//
// Build it like this:
//
// $ CGO_ENABLED=0 go build -o bin ./pkg/clock/manual_test/
//
// And run ./bin/manual_test, preferably in a container where you can change
// date:
//
// $ docker run --cap-add=SYS_TIME -v $(pwd):/test -w /test -ti ubuntu:latest
// # ./bin/manual_test &
// ...
// # date -s 10:00
// # date -s 11:00
// # date -s 12:00
//
// If everything works correctly, you should see three messages for this test,
// and it should terminate after this.
package main

import (
	"context"
	"fmt"
	"log"

	"bpxe.org/pkg/clock"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	c, err := clock.Host(ctx)
	if err != nil {
		log.Fatalf("Error: %s", err.Error())
	}
	fmt.Println("This program monitors time changes, it's advisable to run it in a container for testing.")
	fmt.Println("Try changing time and see if it'll generate any updates.")
	fmt.Println("It'll automatically shutdown when you get three time changes.")
	i := 0
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case change := <-c.Changes():
			fmt.Printf("New time %s\n", change.String())
			i++
			if i == 3 {
				cancel()
			}
		}
	}
}
