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
