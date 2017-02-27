package commands

import (
	"github.com/apflieger/tie/core"
	"github.com/apflieger/tie/model"
	"github.com/apflieger/tie/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/libgit2/git2go.v25"
	"testing"
)

func TestSelectCommand(t *testing.T) {
	test.RunOnRepo(t, "SelectTip", func(t *testing.T, context model.Context, repo *git.Repository) {
		head, _ := repo.Head()

		// New tip ref created on HEAD
		repo.References.Create(core.RefsTips+"test", head.Target(), false, "")

		head, _ = repo.Head()
		assert.Equal(t, "refs/heads/master", head.Name())

		// Select the test tip
		SelectCommand(repo, "test")

		// We expect HEAD to be attached on the tip
		head, _ = repo.Head()
		assert.Equal(t, core.RefsTips+"test", head.Name())
	})

	test.RunOnRepo(t, "DwimFailed", func(t *testing.T, context model.Context, repo *git.Repository) {
		err := SelectCommand(repo, "test")

		if assert.NotNil(t, err) {
			assert.Equal(t, "No ref found for shorthand 'test'", err.Error())
		}
	})

	test.RunOnRepo(t, "DirtyState", func(t *testing.T, context model.Context, repo *git.Repository) {
		// Commit a file on a new tip
		test.WriteFile(repo, true, "foo", "a")
		test.Commit(repo, &test.CommitParams{Refname: core.RefsTips + "test"})

		// write the same file on the working tree
		test.WriteFile(repo, false, "foo", "b")

		// select the tip
		err := SelectCommand(repo, "test")

		// We expect the select to fail because the checkout has a conflict
		if assert.NotNil(t, err) {
			assert.Equal(t, "1 conflict prevents checkout", err.Error())
		}
	})
}
