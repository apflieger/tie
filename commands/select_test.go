package commands

import (
	"github.com/apflieger/tie/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/libgit2/git2go.v25"
	"testing"
	"bytes"
)

func TestSelect(t *testing.T) {
	test.RunOnRepo(t, "SelectTip", func(t *testing.T, repo *git.Repository) {
		head, _ := repo.Head()

		// New tip ref created on HEAD
		repo.References.Create("refs/tips/local/test", head.Target(), false, "")

		head, _ = repo.Head()
		assert.Equal(t, "refs/heads/master", head.Name())

		// Select the test tip
		SelectCommand(repo, "test")

		// We expect HEAD to be attached on the tip
		head, _ = repo.Head()
		assert.Equal(t, "refs/tips/local/test", head.Name())
	})

	test.RunOnRepo(t, "DwimFailed", func(t *testing.T, repo *git.Repository) {
		err := SelectCommand(repo, "test")

		if assert.NotNil(t, err) {
			assert.Equal(t, "No ref found for shorthand \"test\"", err.Error())
		}
	})

	test.RunOnRepo(t, "DirtyState", func(t *testing.T, repo *git.Repository) {
		// Commit a file on a new tip
		test.WriteFile(repo, true, "foo", "a")
		test.Commit(repo, &test.CommitParams{Refname: "refs/tips/local/test"})

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

func TestList(t *testing.T) {
	test.RunOnRepo(t, "DefaultListing", func(t *testing.T, repo *git.Repository) {
		setupRefs(repo)
		var logBuffer *bytes.Buffer
		ListCommand(repo, test.CreateTestLogger(&logBuffer), false, false, false)
		assert.Equal(t,
			"refs/tips/local/tip1\n" +
			"refs/tips/local/tip2\n",
			logBuffer.String())
	})
}

func setupRefs(repo *git.Repository) {
	head, _ := repo.Head()
	oid := head.Target()
	repo.References.Create("refs/tips/local/tip1", oid, false, "")
	repo.References.Create("refs/tips/local/tip2", oid, false, "")
	repo.References.Create("refs/tips/origin/tip3", oid, false, "")
	repo.References.Create("refs/tips/github/tip4", oid, false, "")
	repo.References.Create("refs/heads/branch1", oid, false, "")
	repo.References.Create("refs/remotes/origin/branch2", oid, false, "")
	repo.References.Create("refs/remotes/github/branch3", oid, false, "")
}
