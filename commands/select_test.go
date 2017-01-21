package commands

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"github.com/apflieger/tie/core"
)

func TestSelectTip(t *testing.T) {
	repo := core.CreateTestRepo()
	head, _ := repo.Head()
	assert.Equal(t, "refs/heads/master", head.Name())

	core.Commit(repo, "refs/tips/local/test")

	head, _ = repo.Head()
	assert.Equal(t, "refs/heads/master", head.Name())

	SelectCommand(repo, "refs/tips/local/test")

	head, _ = repo.Head()
	assert.Equal(t, "refs/tips/local/test", head.Name())
}
