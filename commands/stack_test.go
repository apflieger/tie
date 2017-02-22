package commands

import (
	"github.com/apflieger/tie/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/libgit2/git2go.v25"
	"testing"
	"github.com/apflieger/tie/core"
)

func TestStack(t *testing.T) {
	test.RunOnRepo(t, "SimpleStack", func(t *testing.T, repo *git.Repository) {
		// Create a tip 1 commit ahead of master
		test.CreateTip(repo, "test", "refs/heads/master", true)
		test.WriteFile(repo, true, "foo", "line")
		oid, _ := test.Commit(repo, nil)

		err := StackCommand(repo)

		assert.Nil(t, err)

		// Master should have been fast forwarded and selected
		head, _ := repo.Head()
		assert.Equal(t, "refs/heads/master", head.Name())
		assert.True(t, head.Target().Equal(oid))

		// Status should be clean
		test.StatusClean(t, repo)
	})

	test.RunOnRepo(t, "NotOnTip", func(t *testing.T, repo *git.Repository) {
		// Select a branch
		head, _ := repo.Head()
		repo.References.Create("refs/heads/test", head.Target(), false, "")
		_, err := repo.References.CreateSymbolic("HEAD", "refs/heads/test", true, "")
		assert.Nil(t, err)

		err = StackCommand(repo)

		assert.NotNil(t, err)
	})

	test.RunOnRepo(t, "FastForward", func(t *testing.T, repo *git.Repository) {
		// Create a tip
		test.CreateTip(repo, "test", "refs/heads/master", false)

		// Commit on master
		test.Commit(repo, nil)

		// Select the tip and commit
		repo.References.CreateSymbolic("HEAD", core.RefsTips+"test", true, "")
		test.Commit(repo, nil)

		// Try to stack the tip
		err := StackCommand(repo)

		// Stack should have failed because the tip doesn't fast forward his base
		assert.NotNil(t, err)
	})

	test.RunOnRepo(t, "TailOnBase", func(t *testing.T, repo *git.Repository) {
		head, _ := repo.Head()
		firstCommit := head.Target()

		// Commit on master
		test.Commit(repo, nil)

		// Create a tip and commit (just to have something to stack)
		test.CreateTip(repo, "test", "refs/heads/master", true)
		test.Commit(repo, nil)

		// Reset master to the first commit
		master, _ := repo.References.Lookup("refs/heads/master")
		master.SetTarget(firstCommit, "")

		// Try to stack the tip
		err := StackCommand(repo)

		// Stack should have failed because the base and the tail are not on the same commit.
		// This would lead to push a commit that doesn't belong to the tip.
		assert.NotNil(t, err)
	})
}
