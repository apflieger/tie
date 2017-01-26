package args

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVerbs(t *testing.T) {
	assertCommand(t, "help")
	assertCommand(t, "select")
	assertCommand(t, "commit")
}

func assertCommand(t *testing.T, verb string) {
	command, params, _ := ParseArgs([]string{"tie", verb})
	assert.NotNil(t, command)
	assert.NotNil(t, params)
	assert.Empty(t, params)
}

func TestVerbNotFound(t *testing.T) {
	_, _, err := ParseArgs([]string{"tie", "foo"})
	assert.NotNil(t, err)
}

func TestNoArg(t *testing.T) {
	_, _, err := ParseArgs([]string{"tie"})
	if assert.NotNil(t, err) {
		assert.Equal(t, NoSuchCommandError, err)
		assert.Empty(t, err.Error())
	}
}

func TestMultipleArgs(t *testing.T) {
	command, params, _ := ParseArgs([]string{"tie", "help", "--opt", "param"})
	assert.NotNil(t, command)
	if assert.NotNil(t, params) {
		assert.Equal(t, []string{"--opt", "param"}, params)
	}
}
