package commands

import (
	"github.com/apflieger/tie/core"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSelectTip(t *testing.T) {
	repo := core.CreateTestRepo()
	head, _ := repo.Head()

	// New tip ref created on HEAD
	core.Commit(repo, "refs/tips/local/test")

	head, _ = repo.Head()
	assert.Equal(t, "refs/heads/master", head.Name())

	// Select the test tip
	SelectCommand(repo, []string{"test"})

	// We expect HEAD to be attached on the tip
	head, _ = repo.Head()
	assert.Equal(t, "refs/tips/local/test", head.Name())
}
