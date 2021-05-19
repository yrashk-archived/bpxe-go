// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package expression

import (
	"bpxe.org/pkg/data"
)

type Compiler interface {
	CompileExpression(source string) (CompiledExpression, error)
}

type CompiledExpression interface{}

type Evaluator interface {
	EvaluateExpression(expr CompiledExpression, data interface{}) (Result, error)
}

type Result interface{}

type Engine interface {
	Compiler
	Evaluator
	SetItemAwareLocator(itemAwareLocator data.ItemAwareLocator)
}
