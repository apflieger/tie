package commands

import (
	"github.com/apflieger/tie/core"
	"github.com/apflieger/tie/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/libgit2/git2go.v25"
	"testing"
)

func TestDeleteCommand(t *testing.T) {
	test.RunOnRepo(t, "LocalTip", func(t *testing.T, context test.TestContext, repo *git.Repository) {
		// create a tip with his tail and base
		test.CreateTip(repo, "test", "refs/heads/master", false)

		err := DeleteCommand(repo, false, []string{core.RefsTips + "test"}, context.Context)

		assert.Nil(t, err)

		// tip's head should be deleted
		_, err = repo.References.Lookup(core.RefsTips + "test")
		assert.NotNil(t, err)

		// output should be...
		assert.Equal(t, "Deleted tip 'test'\n", context.OutputBuffer.String())
	})

	test.RunOnRepo(t, "Branch", func(t *testing.T, context test.TestContext, repo *git.Repository) {
		// create a local branch
		head, _ := repo.Head()
		repo.References.Create("refs/heads/test", head.Target(), false, "")

		err := DeleteCommand(repo, false, []string{"refs/heads/test"}, context.Context)

		// tie delete doesn't allow to delete branches
		assert.NotNil(t, err)

		// Branch should still be here
		_, err = repo.References.Lookup("refs/heads/test")
		assert.Nil(t, err)
	})

	test.RunOnRemote(t, "RemoteTip", func(t *testing.T, context test.TestContext, repo, remote *git.Repository) {
		// create a tip with his tail and base
		tipRefName := core.RefsTips + "test"
		head, _ := repo.Head()
		test.CreateTip(repo, "test", "refs/remotes/origin/master", false)

		// set it up on origin
		repo.References.Create(core.RefsRemoteTips+"origin/test", head.Target(), false, "")
		origin, _ := repo.Remotes.Lookup("origin")
		origin.Push([]string{tipRefName + ":" + tipRefName}, nil)
		// activate branch compatibility mode
		config, _ := repo.Config()
		config.SetString(core.PushTipsAsConfigKey, "refs/heads/")
		repo.References.Create("refs/remotes/origin/tips/test", head.Target(), false, "")
		origin.Push([]string{tipRefName + ":refs/heads/tips/test"}, nil)

		err := DeleteCommand(repo, false, []string{tipRefName}, context.Context)

		assert.Nil(t, err)

		// output should be...
		assert.Equal(t, "Deleted tip 'test'\n", context.OutputBuffer.String())
	})

	test.RunOnRepo(t, "UnreachableRemote", func(t *testing.T, context test.TestContext, repo *git.Repository) {
		// create a tip with his tail and base on origin/master
		test.CreateTip(repo, "test", "refs/remotes/origin/master", false)

		head, _ := repo.Head()
		repo.References.Create(core.RefsRemoteTips+"origin/test", head.Target(), false, "")
		// create an unreachable origin remote
		repo.Remotes.Create("origin", "/dev/null")

		err := DeleteCommand(repo, false, []string{core.RefsTips + "test"}, context.Context)

		assert.Nil(t, err)

		// output should be...
		assert.Contains(t, context.OutputBuffer.String(), "Tip 'test' has been deleted locally but not on origin")
	})

	test.RunOnRepo(t, "CurrentTip", func(t *testing.T, context test.TestContext, repo *git.Repository) {
		// Create a tip
		test.CreateTip(repo, "test", "refs/heads/master", true)

		// Commit a file
		test.WriteFile(repo, true, "foo", "bar")
		test.Commit(repo, nil)

		// Delete the tip
		err := DeleteCommand(repo, false, nil, context.Context)
		assert.Nil(t, err)

		// output should be...
		assert.Contains(t, context.OutputBuffer.String(), "Deleted tip 'test'\n")

		// HEAD should be back on master
		head, _ := repo.Head()
		assert.Equal(t, "refs/heads/master", head.Name())
		// Status should be clean. Which means that the repo has been properly reset
		test.StatusClean(t, repo)

		// The tip should be deleted
		_, err = repo.References.Lookup(core.RefsTips + "test")
		assert.NotNil(t, err)
	})

	t.Run("StackedOption", func(t *testing.T) {

		test.RunOnRepo(t, "SimpleTip", func(t *testing.T, context test.TestContext, repo *git.Repository) {
			// Create a tip
			test.CreateTip(repo, "test", "refs/heads/master", false)

			// Commit a file on the tip
			test.WriteFile(repo, true, "foo", "bar")
			oid, _ := test.Commit(repo, &test.CommitParams{Refname: core.RefsTips + "test"})

			// Update the master to the last commit
			master, _ := repo.References.Lookup("refs/heads/master")
			master, _ = master.SetTarget(oid, "")

			// Add one more commit to master
			test.Commit(repo, &test.CommitParams{Refname: "refs/heads/master"})

			// At this point, master is one commit ahead of the tip

			err := DeleteCommand(repo, true, nil, context.Context)
			assert.Nil(t, err)

			// output should be...
			assert.Contains(t, context.OutputBuffer.String(), "Deleted tip 'test'\n")

			head, _ := repo.Head()
			assert.Equal(t, "refs/heads/master", head.Name())

			_, err = repo.References.Lookup(core.RefsTips + "test")
			assert.NotNil(t, err)
		})

		test.RunOnRepo(t, "EmptyTip", func(t *testing.T, context test.TestContext, repo *git.Repository) {
			// Create a tip
			test.CreateTip(repo, "test", "refs/heads/master", true)

			// Select master
			repo.SetHead("refs/heads/master")

			err := DeleteCommand(repo, true, nil, context.Context)
			assert.Nil(t, err)

			// output should be...
			assert.Contains(t, context.OutputBuffer.String(), "Deleted tip 'test'\n")

			head, _ := repo.Head()
			assert.Equal(t, "refs/heads/master", head.Name())

			_, err = repo.References.Lookup(core.RefsTips + "test")
			assert.NotNil(t, err)
		})
	})
}
