package commands

import (
	"github.com/apflieger/tie/core"
	"github.com/apflieger/tie/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/libgit2/git2go.v25"
	"testing"
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

	test.RunOnRepo(t, "NotOnTipError", func(t *testing.T, repo *git.Repository) {
		// Select a branch
		head, _ := repo.Head()
		repo.References.Create("refs/heads/test", head.Target(), false, "")
		_, err := repo.References.CreateSymbolic("HEAD", "refs/heads/test", true, "")
		assert.Nil(t, err)

		err = StackCommand(repo)

		assert.NotNil(t, err)
	})

	test.RunOnRepo(t, "LocalFastForwardError", func(t *testing.T, repo *git.Repository) {
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

	test.RunOnRepo(t, "TailNotOnBaseError", func(t *testing.T, repo *git.Repository) {
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

	test.RunOnRemote(t, "OnRemoteBranch", func(t *testing.T, repo, origin *git.Repository) {
		// Create a tip on origin/master
		test.CreateTip(repo, "test", "refs/remotes/origin/master", true)

		// Write a commit
		oid, _ := test.Commit(repo, nil)

		// Stack it
		err := StackCommand(repo)
		assert.Nil(t, err)

		master, _ := origin.References.Lookup("refs/heads/master")
		assert.True(t, master.Target().Equal(oid))

		remoteMaster, _ := repo.References.Lookup("refs/remotes/origin/master")
		assert.True(t, remoteMaster.Target().Equal(oid))
	})

	test.RunOnRemote(t, "RemoteFastForwardError", func(t *testing.T, repo, origin *git.Repository) {
		// Create a tip
		test.CreateTip(repo, "test", "refs/remotes/origin/master", false)

		// Commit on a tmp branch and push it
		head, _ := repo.Head()
		headCommit, _ := repo.LookupCommit(head.Target())
		tmp, _ := repo.CreateBranch("tmp", headCommit, false)
		test.Commit(repo, &test.CommitParams{
			Refname: tmp.Reference.Name(),
		})
		remote, _ := repo.Remotes.Lookup("origin")
		remote.Push([]string{tmp.Reference.Name()}, nil)

		// On origin, reset master to tmp
		master, _ := origin.References.Lookup("refs/heads/master")
		originTmp, _ := origin.References.Lookup("refs/heads/tmp")
		master.SetTarget(originTmp.Target(), "")

		// Kinda complicated but that's the only way I found to make
		// refs/remotes/origin/master out of sync with master on origin

		// Select the tip and commit
		repo.References.CreateSymbolic("HEAD", core.RefsTips+"test", true, "")
		test.WriteFile(repo, true, "foo", "bar")
		test.Commit(repo, nil)

		// Try to stack the tip
		err := StackCommand(repo)

		// Stack should have failed because the tip doesn't fast forward his base
		assert.NotNil(t, err)
	})

	test.RunOnRepo(t, "OnRemoteTip", func(t *testing.T, repo *git.Repository) {
		// Create a tip on a remote tip
		head, _ := repo.Head()
		repo.References.Create(core.RefsRemoteTips+"origin/test2", head.Target(), false, "")
		test.CreateTip(repo, "test", core.RefsRemoteTips+"origin/test2", true)

		// Write a commit
		test.Commit(repo, nil)

		// Stack it
		err := StackCommand(repo)

		// Stack doesn't allow to stack on tips for now
		assert.NotNil(t, err)
	})

	test.RunOnRemote(t, "TwoTipStackError", func(t *testing.T, repo, origin *git.Repository) {
		// Create a tip and commit
		test.CreateTip(repo, "test1", "refs/remotes/origin/master", true)
		test.WriteFile(repo, true, "foo", "bar")
		test.Commit(repo, nil)

		// Create a second tip on top of the first one and commit
		test.CreateTip(repo, "test2", core.RefsTips+"test1", true)
		test.WriteFile(repo, true, "foo2", "bar")
		test.Commit(repo, nil)

		// Change the base of the 2nd tip to origin/master without upgrading
		config, _ := repo.Config()
		config.SetString("tip.test2.base", "refs/remotes/origin/master")

		// At this point, test2 is ff to origin/master

		// Try to stack the tip.
		err := StackCommand(repo)

		// Stacking this tip would mean to have the commit of test1 to be stacked as well
		assert.NotNil(t, err)
	})
}