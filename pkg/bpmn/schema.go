package bpmn

import (
	"encoding/xml"
	"reflect"
	"strings"
)

// Base types

// XML qualified name (http://books.xmlschemata.org/relaxng/ch19-77287.html)
type QName = string

// Identifier (http://books.xmlschemata.org/relaxng/ch19-77151.html)
type Id = string

// Reference to identifiers (http://books.xmlschemata.org/relaxng/ch19-77159.html)
type IdRef = string

// Corresponds normatively to the XLink href attribute (http://books.xmlschemata.org/relaxng/ch19-77009.html)
type AnyURI = string

// Element is implemented by every BPMN document element.
type Element interface {
	// FindBy recursively searches for an Element matching a predicate.
	//
	// Returns matching predicate in `result` and sets `found` to `true`
	// if any found.
	FindBy(ElementPredicate) (result Element, found bool)
}

type ElementPredicate func(Element) bool

// Returns a function that matches Element's identifier (if the `Element`
// implements BaseElementInterface against given string. If it matches, the
// function returns `true`.
//
// To be used in conjunction with FindBy (Element interface)
func ExactId(s string) ElementPredicate {
	return func(e Element) bool {
		if el, ok := e.(BaseElementInterface); ok {
			if id, present := el.Id(); present {
				return *id == s
			} else {
				return false
			}
		} else {
			return false
		}
	}
}

// Returns a function that matches Elements by types. If it matches, the
// function returns `true.
//
// To be used in conjunction with FindBy (Element interface)
func ElementType(t Element) ElementPredicate {
	return func(e Element) bool {
		return reflect.TypeOf(e) == reflect.TypeOf(t)
	}
}

// Returns a function that matches Elements by interface implementation. If it matches, the
// function returns `true.
//
// To be used in conjunction with FindBy (Element interface)
func ElementInterface(t interface{}) ElementPredicate {
	return func(e Element) bool {
		return reflect.TypeOf(e).Implements(reflect.TypeOf(t).Elem())
	}
}

// Returns a function that matches Elements if both the receiver and the given
// predicates match.
//
// To be used in conjunction with FindBy (Element interface)
func (matcher ElementPredicate) And(andMatcher ElementPredicate) ElementPredicate {
	return func(e Element) bool {
		return matcher(e) && andMatcher(e)
	}
}

// Returns a function that matches Elements if either the receiver or the given
// predicates match.
//
// To be used in conjunction with FindBy (Element interface)
func (matcher ElementPredicate) Or(andMatcher ElementPredicate) ElementPredicate {
	return func(e Element) bool {
		return matcher(e) || andMatcher(e)
	}
}

// Special case for handling expressions being substituted for formal expression
// as per "BPMN 2.0 by Example":
//
// ```
//   <conditionExpression xsi:type="tFormalExpression">
//     ${getDataObject("TicketDataObject").status == "Resolved"}
//   </conditionExpression>
// ```
//
// Technically speaking, any type can be "patched" this way, but it seems impractical to
// do so broadly. Even within the confines of expressions, we don't want to support changing
// the type to an arbitrary one. Only specifying expressions as formal expressions will be
// possible for practicality.

// For this, a special type called AnExpression will be added and schema generator will
// replace Expression type with it.

// Expression family container
//
// Expression field can be type-switched between Expression and FormalExpression
type AnExpression struct {
	Expression ExpressionInterface
}

func (e *AnExpression) UnmarshalXML(d *xml.Decoder, start xml.StartElement) (err error) {
	formal := false
	for i := range start.Attr {
		if start.Attr[i].Name.Space == "http://www.w3.org/2001/XMLSchema-instance" &&
			start.Attr[i].Name.Local == "type" &&
			// here we're check for suffix instead of equality because
			// there doesn't seem to be a way to get a mapping between
			// bpmn schema and its namespace, so `bpmn:tFormalExpression`
			// equality check will fail if a different namespace name will
			// be used.
			(strings.HasSuffix(start.Attr[i].Value, ":tFormalExpression") ||
				start.Attr[i].Value == "tFormalExpression") {
			formal = true
			break
		}
	}
	if !formal {
		expr := DefaultExpression()
		err = d.DecodeElement(&expr, &start)
		if err != nil {
			return
		}
		e.Expression = &expr
	} else {
		expr := DefaultFormalExpression()
		err = d.DecodeElement(&expr, &start)
		if err != nil {
			return
		}
		e.Expression = &expr
	}
	return
}

func (e *AnExpression) FindBy(pred ElementPredicate) (result Element, found bool) {
	if e.Expression == nil {
		found = false
		return
	}
	return e.Expression.FindBy(pred)
}

// Generate schema files:

//go:generate saxon-he ../../schemas/BPMN20.xsd ../../schema-codegen.xsl
