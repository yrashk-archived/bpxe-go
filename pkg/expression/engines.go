// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package expression

var enginesMap = map[string]func() Engine{
	"http://www.w3.org/1999/XPath": func() Engine {
		return NewXPath()
	},
	"https://github.com/antonmedv/expr": func() Engine {
		return NewExpr()
	},
}

func GetEngine(url string) Engine {
	if engineConstructor, ok := enginesMap[url]; ok {
		return engineConstructor()
	} else {
		return GetEngine("http://www.w3.org/1999/XPath")
	}
}
