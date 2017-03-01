package commands

import (
	"github.com/apflieger/tie/core"
	"github.com/apflieger/tie/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/libgit2/git2go.v25"
	"testing"
)

func TestSelectCommand(t *testing.T) {
	test.RunOnRepo(t, "SelectTip", func(t *testing.T, context test.TestContext, repo *git.Repository) {
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

	// This test demonstrates a problem when using git fetch while being attached to a remote ref
	// Don't support this for now.
	/*test.RunOnRepo(t, "UpdatedRef", func(t *testing.T, context test.TestContext, repo *git.Repository) {
		// Commit a file on master
		test.WriteFile(repo, true, "foo", "a")
		test.Commit(repo, nil)

		// Create origin/master on HEAD
		head, _ := repo.Head()
		repo.References.Create("refs/remotes/origin/master", head.Target(), false, "")

		// Select origin/master
		SelectCommand(repo, "refs/remotes/origin/master")

		// Commit a change on origin/master (this simulates a fetch)
		test.WriteFile(repo, true, "foo", "b")
		test.Commit(repo, &test.CommitParams{Refname: "refs/remotes/origin/master"})

		// After the fetch, the working tree becomes dirty
		statusList, _ := repo.StatusList(
			&git.StatusOptions{
				Show:     git.StatusShowIndexAndWorkdir,
				Flags:    git.StatusOptIncludeUntracked,
				Pathspec: nil,
			})
		statusCount, _ := statusList.EntryCount()
		assert.Equal(t, 1, statusCount)

		// re-select origin/master should clean the status
		err := SelectCommand(repo, "refs/remotes/origin/master")

		assert.Nil(t, err)

		test.StatusClean(t, repo)

		// The working tree should have the last change
		foo, _ := ioutil.ReadFile(filepath.Join(repo.Workdir(), "foo"))
		assert.Equal(t, "b", string(foo))
	})*/

	test.RunOnRepo(t, "DwimFailed", func(t *testing.T, context test.TestContext, repo *git.Repository) {
		err := SelectCommand(repo, "test")

		if assert.NotNil(t, err) {
			assert.Equal(t, "No ref found for shorthand 'test'", err.Error())
		}
	})

	test.RunOnRepo(t, "DirtyState", func(t *testing.T, context test.TestContext, repo *git.Repository) {
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
