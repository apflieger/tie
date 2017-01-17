package args

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVerb(t *testing.T) {
	verb, params, _ := ParseArgs([]string{"tie", "help"})
	assert.Equal(t, "help", verb)
	assert.NotNil(t, params)
	assert.Empty(t, params)
}

func TestNoArg(t *testing.T) {
	_, _, err := ParseArgs([]string{"tie"})
	if assert.NotNil(t, err) {
		assert.IsType(t, NoArgsError{}, err)
		assert.Empty(t, err.Error())
	}
}

func TestMultipleArgs(t *testing.T) {
	verb, params, _ := ParseArgs([]string{"tie", "verb", "--opt", "param"})
	assert.Equal(t, "verb", verb)
	if assert.NotNil(t, params) {
		assert.Equal(t, []string{"--opt", "param"}, params)
	}
}
