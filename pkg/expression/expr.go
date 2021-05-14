// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package expression

import (
	"reflect"

	"bpxe.org/pkg/errors"
	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
)

// Expr language engine
//
// https://github.com/antonmedv/expr
type Expr struct{}

func MakeExpr() Expr {
	return Expr{}
}

func NewExpr() *Expr {
	engine := MakeExpr()
	return &engine
}

func (engine *Expr) CompileExpression(source string) (result CompiledExpression, err error) {
	result, err = expr.Compile(source)
	return
}

func (engine *Expr) EvaluateExpression(e CompiledExpression,
	data interface{},
) (result Result, err error) {
	if expression, ok := e.(*vm.Program); ok {
		result, err = expr.Run(expression, data)
	} else {
		err = errors.InvalidArgumentError{
			Expected: "CompiledExpression to be *github.com/antonmedv/expr/vm#Program",
			Actual:   reflect.TypeOf(e),
		}
	}
	return
}
