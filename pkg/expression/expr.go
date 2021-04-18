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
