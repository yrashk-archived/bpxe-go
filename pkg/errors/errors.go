// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package errors

import "fmt"

type InvalidArgumentError struct {
	Expected interface{}
	Actual   interface{}
}

func (e InvalidArgumentError) Error() string {
	return fmt.Sprintf("Invalid argument: expected %v, got %v", e.Expected, e.Actual)
}

type InvalidStateError struct {
	Expected interface{}
	Actual   interface{}
}

func (e InvalidStateError) Error() string {
	return fmt.Sprintf("Invalid state: expected %v, got %v", e.Expected, e.Actual)
}

type NotFoundError struct {
	Expected interface{}
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("%v not found", e.Expected)
}

type RequirementExpectationError struct {
	Expected interface{}
	Actual   interface{}
}

func (e RequirementExpectationError) Error() string {
	return fmt.Sprintf("Requirement expectation failed: expected %v, got %v", e.Expected, e.Actual)
}
