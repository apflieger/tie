package commands

import (
	"github.com/apflieger/tie/core"
	"github.com/stretchr/testify/assert"
	"gopkg.in/libgit2/git2go.v25"
	"testing"
)

func TestSelect(t *testing.T) {
	core.RunRequireRepo(t, "SelectTip", func(t *testing.T, repo *git.Repository) {
		head, _ := repo.Head()

		// New tip ref created on HEAD
		repo.References.Create("refs/tips/local/test", head.Target(), false, "")

		head, _ = repo.Head()
		assert.Equal(t, "refs/heads/master", head.Name())

		// Select the test tip
		SelectCommand(repo, []string{"test"})

		// We expect HEAD to be attached on the tip
		head, _ = repo.Head()
		assert.Equal(t, "refs/tips/local/test", head.Name())
	})

	core.RunRequireRepo(t, "DwimFailed", func(t *testing.T, repo *git.Repository) {
		err := SelectCommand(repo, []string{"test"})

		if assert.NotNil(t, err) {
			assert.Equal(t, "No ref found for shorthand \"test\"", err.Error())
		}
	})

	core.RunRequireRepo(t, "DirtyState", func(t *testing.T, repo *git.Repository) {
		// Commit a file on a new tip
		core.WriteFile(repo, true, "foo", "a")
		core.Commit(repo, &core.CommitParams{Refname: "refs/tips/local/test"})

		// write the same file on the working tree
		core.WriteFile(repo, false, "foo", "b")

		// select the tip
		err := SelectCommand(repo, []string{"test"})

		// We expect the select to fail because the checkout has a conflict
		if assert.NotNil(t, err) {
			assert.Equal(t, "1 conflict prevents checkout", err.Error())
		}
	})
}
