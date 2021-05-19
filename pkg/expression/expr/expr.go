// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package expr

import (
	"context"
	"reflect"

	"bpxe.org/pkg/data"
	"bpxe.org/pkg/errors"
	"bpxe.org/pkg/expression"
	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
)

// Expr language engine
//
// https://github.com/antonmedv/expr
type Expr struct {
	env              map[string]interface{}
	itemAwareLocator data.ItemAwareLocator
}

func (engine *Expr) SetItemAwareLocator(itemAwareLocator data.ItemAwareLocator) {
	engine.itemAwareLocator = itemAwareLocator
}

func New(ctx context.Context) *Expr {
	engine := &Expr{}
	engine.env = map[string]interface{}{
		"getDataObject": func(args ...string) data.Item {
			var name string
			if len(args) == 1 {
				name = args[0]
			}
			itemAware, found := engine.itemAwareLocator.FindItemAwareByName(name)
			if !found {
				return nil
			}
			ch := itemAware.Get(ctx)
			if ch == nil {
				return nil
			}
			select {
			case <-ctx.Done():
				return nil
			case item := <-ch:
				return item
			}
		},
	}
	return engine
}

func (engine *Expr) CompileExpression(source string) (result expression.CompiledExpression, err error) {
	result, err = expr.Compile(source, expr.Env(engine.env), expr.AllowUndefinedVariables())
	return
}

func (engine *Expr) EvaluateExpression(e expression.CompiledExpression,
	data interface{},
) (result expression.Result, err error) {
	actualData := data
	if data == nil {
		actualData = engine.env
	} else if e, ok := data.(map[string]interface{}); ok {
		env := engine.env
		for k, v := range e {
			env[k] = v
		}
		actualData = env
	}

	if exp, ok := e.(*vm.Program); ok {
		result, err = expr.Run(exp, actualData)
	} else {
		err = errors.InvalidArgumentError{
			Expected: "CompiledExpression to be *github.com/antonmedv/expr/vm#Program",
			Actual:   reflect.TypeOf(e),
		}
	}
	return
}

func init() {
	expression.RegisterEngine("https://github.com/antonmedv/expr", func(ctx context.Context) expression.Engine {
		return New(ctx)
	})
}
