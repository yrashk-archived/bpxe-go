// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package expression

import (
	"bytes"
	"reflect"

	"bpxe.org/pkg/errors"
	"github.com/ChrisTrenkamp/xsel/exec"
	"github.com/ChrisTrenkamp/xsel/grammar"
	"github.com/ChrisTrenkamp/xsel/parser"
	"github.com/ChrisTrenkamp/xsel/store"
	"github.com/Chronokeeper/anyxml"
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
	compiled, err := grammar.Build(source)
	if err == nil {
		result = &compiled
	}
	return
}

func (engine *XPath) EvaluateExpression(e CompiledExpression,
	data interface{},
) (result Result, err error) {
	if expression, ok := e.(*grammar.Grammar); ok {
		// Here, in order to save some prototyping type,
		// instead of implementing `parser.Parser` for `interface{}`,
		// we use it over `interface{}` serialized as XML.
		// This is not very efficient but does the job for now.
		// Eventually, a direct implementation of `parser.Parser`
		// over `interface{}` should be developed to optimize this path.

		var serialized []byte
		serialized, err = anyxml.Xml(data)
		if err != nil {
			result = nil
			return
		}
		p := parser.ReadXml(bytes.NewBuffer(serialized))
		var cursor store.Cursor
		cursor, err = store.CreateInMemory(p)
		if err != nil {
			return
		}
		var res exec.Result
		res, err = exec.Exec(cursor, expression)
		if err != nil {
			return
		}
		switch res := res.(type) {
		case exec.String:
			result = res.String()
		case exec.Bool:
			result = res.Bool()
		case exec.Number:
			result = res.Number()
		}
	} else {
		err = errors.InvalidArgumentError{
			Expected: "CompiledExpression to be *github.com/ChrisTrenkamp/xsel/grammar#Grammar",
			Actual:   reflect.TypeOf(e),
		}
	}
	return
}
