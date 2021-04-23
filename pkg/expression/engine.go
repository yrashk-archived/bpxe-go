package expression

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
}
