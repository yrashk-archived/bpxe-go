// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package expression

import (
	"sync"
)

var enginesLock sync.RWMutex
var enginesMap = make(map[string]func() Engine)

func RegisterEngine(url string, engine func() Engine) {
	enginesLock.Lock()
	enginesMap[url] = engine
	enginesLock.Unlock()
}

func GetEngine(url string) (engine Engine) {
	enginesLock.RLock()
	if engineConstructor, ok := enginesMap[url]; ok {
		engine = engineConstructor()
	} else {
		engine = enginesMap["http://www.w3.org/1999/XPath"]()
	}
	enginesLock.RUnlock()
	return
}
