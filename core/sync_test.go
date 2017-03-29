package core

import (
	"github.com/apflieger/tie/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/libgit2/git2go.v25"
	"os"
	"path/filepath"
	"testing"
)

func TestUpdate(t *testing.T) {
	test.RunOnThreeRepos(t, "UpdateCurrentBranch", func(t *testing.T, context test.TestContext, repo, origin, another *git.Repository) {
		// HEAD is on origin/master
		// asserting test setup
		head, _ := repo.Head()
		assert.Equal(t, "refs/remotes/origin/master", head.Name())

		// Commit on master from "another" and push it to origin
		test.WriteFile(another, true, "foo", "foobar")
		oid, _ := test.Commit(another, nil)
		remote, _ := another.Remotes.Lookup("origin")
		remote.Push([]string{"+refs/heads/master"}, nil)

		Fetch(repo, context.Context)

		originMaster, _ := repo.References.Lookup("refs/remotes/origin/master")
		// local origin/master should be on the commit
		assert.True(t, oid.Equal(originMaster.Target()))
		// file "foo" should exist
		_, err := os.Stat(filepath.Join(repo.Workdir(), "foo"))
		assert.Nil(t, err)

		test.StatusClean(t, repo)

		// Output should be...
		assert.Equal(t, "Updated refs/remotes/origin/master\n", context.OutputBuffer.String())
	})

	test.RunOnRemote(t, "OnTip", func(t *testing.T, context test.TestContext, repo, origin *git.Repository) {
		test.CreateTip(repo, "test", "refs/remotes/origin/master", true)

		err := Fetch(repo, context.Context)
		assert.Nil(t, err)
	})

	test.RunOnRemote(t, "OnTipTip", func(t *testing.T, context test.TestContext, repo, origin *git.Repository) {
		test.CreateTip(repo, "test", "refs/remotes/origin/master", false)
		test.CreateTip(repo, "test2", RefsTips+"test", true)

		err := Fetch(repo, context.Context)
		assert.Nil(t, err)
	})

	test.RunOnRemote(t, "OnTipLocalBranch", func(t *testing.T, context test.TestContext, repo, origin *git.Repository) {
		test.CreateTip(repo, "test", "refs/heads/master", false)
		test.CreateTip(repo, "test2", RefsTips+"test", true)

		err := Fetch(repo, context.Context)
		assert.Nil(t, err)
	})

	test.RunOnThreeRepos(t, "MultipleRefs", func(t *testing.T, context test.TestContext, repo, origin, another *git.Repository) {
		remote, _ := another.Remotes.Lookup("origin")

		// Create two branches on the remote
		remote.Push([]string{
			"refs/heads/master:refs/heads/another_branch",
			"refs/heads/master:refs/heads/to_be_deleted",
		}, nil)

		Fetch(repo, context.Context)

		// The two branches should have been fetched
		_, err := repo.References.Lookup("refs/remotes/origin/another_branch")
		assert.Nil(t, err)

		_, err = repo.References.Lookup("refs/remotes/origin/to_be_deleted")
		assert.Nil(t, err)

		// Output should be...
		assert.Equal(t,
			"Created refs/remotes/origin/another_branch\n"+
				"Created refs/remotes/origin/to_be_deleted\n",
			context.OutputBuffer.String())

		// Delete a branch
		remote.Push([]string{":refs/heads/to_be_deleted"}, nil)

		// Reset the output buffer and rerun update
		context.OutputBuffer.Reset()
		Fetch(repo, context.Context)

		// The branch should be pruned
		_, err = repo.References.Lookup("refs/remotes/origin/to_be_deleted")
		assert.NotNil(t, err)

		// Output should be...
		assert.Equal(t,
			"Deleted refs/remotes/origin/to_be_deleted\n",
			context.OutputBuffer.String())
	})
}
