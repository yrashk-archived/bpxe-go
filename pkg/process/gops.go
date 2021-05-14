// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

// +build gops

package process

import (
	"log"

	"github.com/google/gops/agent"
)

// init initializes gops agent if `gops` tag is enabled
//
// it is in this package as this is one of the most central
// packages required by tests and deployable engines
func init() {
	if err := agent.Listen(agent.Options{
		ShutdownCleanup: true,
	}); err != nil {
		log.Fatal(err)
	}
}
