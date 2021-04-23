package expression

import (
	"bytes"
	"encoding/json"
	"reflect"

	"bpxe.org/pkg/errors"
	"github.com/antchfx/jsonquery"
	"github.com/antchfx/xpath"
)

// XPath language engine
//
// Implementation details and limitations as per https://github.com/antchfx/xpath
type XPath struct{}

func MakeXPath() XPath {
	return XPath{}
}

func NewXPath() *XPath {
	engine := MakeXPath()
	return &engine
}

func (engine *XPath) CompileExpression(source string) (result CompiledExpression, err error) {
	result, err = xpath.Compile(source)
	return
}

func (engine *XPath) EvaluateExpression(e CompiledExpression,
	data interface{},
) (result Result, err error) {
	if expression, ok := e.(*xpath.Expr); ok {
		// Here, in order to save some prototyping type,
		// instead of implementing `NodeNavigator` for `interface{}`,
		// we use `jsonquery` over `interface{}` serialized as JSON.
		// This is not very efficient but does the job for now.
		// Eventually, a direct implementation of `NodeNavigator`
		// over `interface{}` should be developed to optimize this path.
		var jsonified []byte
		var doc *jsonquery.Node
		jsonified, err = json.Marshal(data)
		if err != nil {
			result = nil
			return
		}
		doc, err = jsonquery.Parse(bytes.NewReader(jsonified))
		if err != nil {
			result = nil
			return
		}
		result = expression.Evaluate(jsonquery.CreateXPathNavigator(doc))
	} else {
		err = errors.InvalidArgumentError{
			Expected: "CompiledExpression to be *github.com/antchfx/xpath#Expr",
			Actual:   reflect.TypeOf(e),
		}
	}
	return
}
