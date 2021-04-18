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
