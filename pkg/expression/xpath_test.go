package expression

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestXPath(t *testing.T) {
	var engine Engine = NewXPath()
	compiled, err := engine.CompileExpression("a > 1")
	assert.Nil(t, err)
	result, err := engine.EvaluateExpression(compiled, map[string]interface{}{
		"a": 2,
	})
	assert.Nil(t, err)
	assert.True(t, result.(bool))
}
