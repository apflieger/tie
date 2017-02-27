package commands

import (
	"github.com/apflieger/tie/core"
	"github.com/apflieger/tie/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/libgit2/git2go.v25"
	"testing"
	"time"
)

func TestUpgradeCommand(t *testing.T) {
	test.RunOnRepo(t, "NoTipSelected", func(t *testing.T, context test.TestContext, repo *git.Repository) {
		err := UpgradeCommand(repo, context.Context)

		if assert.NotNil(t, err) {
			assert.Equal(t, "HEAD not on a tip. Only tips can be upgraded.", err.Error())
		}
	})

	test.RunOnRepo(t, "NoBase", func(t *testing.T, context test.TestContext, repo *git.Repository) {
		// create and select a tip
		head, _ := repo.Head()
		repo.References.Create(core.RefsTips+"test", head.Target(), true, "")
		repo.References.CreateSymbolic("HEAD", core.RefsTips+"test", true, "")

		err := UpgradeCommand(repo, context.Context)

		// upgrade requires a base to be defined
		if assert.NotNil(t, err) {
			assert.Equal(t, "Config value 'tip.test.base' was not found", err.Error())
		}
	})

	test.RunOnRepo(t, "NoTail", func(t *testing.T, context test.TestContext, repo *git.Repository) {
		// create and select a tip
		head, _ := repo.Head()
		repo.References.Create(core.RefsTips+"test", head.Target(), true, "")
		repo.References.CreateSymbolic("HEAD", core.RefsTips+"test", true, "")

		// configure the base of the tip
		config, _ := repo.Config()
		config.SetString("tip.test.base", "refs/remotes/origin/master")

		err := UpgradeCommand(repo, context.Context)

		// upgrade requires the tip to have a tail
		if assert.NotNil(t, err) {
			assert.Equal(t, "Reference '"+core.RefsTails+"test' not found", err.Error())
		}
	})

	test.RunOnRemote(t, "UpgradeSuccess", func(t *testing.T, context test.TestContext, repo, remote *git.Repository) {
		// create a tip on head based on refs/remotes/origin/master
		head, _ := repo.Head()
		test.CreateTip(repo, "test", "refs/remotes/origin/master", false)

		// make origin/master and the tip diverge.
		// first commit on origin/master
		masterOid, _ := test.Commit(repo, &test.CommitParams{Refname: "refs/remotes/origin/master"})

		// then select the tip and commit
		repo.References.CreateSymbolic("HEAD", core.RefsTips+"test", true, "")
		now := time.Now()
		signature := &git.Signature{
			Name:  "user1",
			Email: "email@example.com",
			When:  now,
		}
		test.WriteFile(repo, true, "foo", "line1")
		test.Commit(repo, nil)
		test.WriteFile(repo, true, "foo", "line1", "line2")
		test.Commit(repo, &test.CommitParams{
			Author:   signature,
			Commiter: signature,
			Message:  "last commit",
		})

		// do the upgrade
		err := UpgradeCommand(repo, context.Context)
		assert.Nil(t, err)

		// we expect the tip to be on top of origin/master
		head, _ = repo.Head()
		headCommit, _ := repo.LookupCommit(head.Target())
		assert.True(t, headCommit.Parent(0).Parent(0).Id().Equal(masterOid))
		assert.Equal(t, "last commit", headCommit.Message())
		assert.Equal(t, "user1", headCommit.Author().Name)
		assert.Equal(t, "email@example.com", headCommit.Author().Email)
		assert.Equal(t, now.Unix(), headCommit.Author().When.Unix())

		// we expect the tail to be updated on origin/master's target
		newTailRef, _ := repo.References.Lookup(core.RefsTails + "test")
		assert.True(t, newTailRef.Target().Equal(masterOid))

		// the repo state should be clean
		assert.Equal(t, git.RepositoryStateNone, repo.State())

		// We expect the tip to be pushed on origin
		remoteTip, err := remote.References.Lookup(core.RefsTips + "test")
		if assert.Nil(t, err) {
			assert.True(t, remoteTip.Target().Equal(head.Target()))
		}
	})

	test.RunOnRemote(t, "ConflictAbort", func(t *testing.T, context test.TestContext, repo, remote *git.Repository) {
		// create a tip on head based on master
		test.CreateTip(repo, "test", "refs/heads/master", false)

		head, _ := repo.Head()
		tailBeforeUpgrade := head.Target()

		// make master and the tip having a conflict.
		// first commit to head that is on master
		test.WriteFile(repo, true, "foo", "line1")
		test.Commit(repo, nil)
		// then select the tip and commit
		firstCommit, _ := repo.LookupCommit(head.Target())
		tree, _ := firstCommit.Tree()
		repo.CheckoutTree(tree, &git.CheckoutOpts{Strategy: git.CheckoutForce})
		repo.References.CreateSymbolic("HEAD", core.RefsTips+"test", true, "")
		test.WriteFile(repo, true, "foo", "line1 bis")
		oidBeforeUpgrade, _ := test.Commit(repo, nil)

		// do the upgrade
		err := UpgradeCommand(repo, context.Context)
		if assert.NotNil(t, err) {
			assert.Equal(t, "Conflict while upgrading", err.Error())
		}

		// abort the upgrade
		err = UpgradeAbortCommand(repo)
		assert.Nil(t, err)

		// HEAD should be back to where it was
		head, _ = repo.Head()
		assert.Equal(t, core.RefsTips+"test", head.Name())
		assert.True(t, head.Target().Equal(oidBeforeUpgrade))
		// tip's tail should be where it was
		tail, _ := repo.References.Lookup(core.RefsTails + "test")
		assert.True(t, tail.Target().Equal(tailBeforeUpgrade))
	})

	test.RunOnRemote(t, "ConflictContinue", func(t *testing.T, context test.TestContext, repo, remote *git.Repository) {
		head, _ := repo.Head()
		// create a tip on head based on master
		test.CreateTip(repo, "test", "refs/heads/master", false)

		// make master and the tip having a conflict.
		// first commit to head that is on master
		test.WriteFile(repo, true, "foo", "line1")
		test.Commit(repo, nil)
		// then select the tip and commit
		firstCommit, _ := repo.LookupCommit(head.Target())
		tree, _ := firstCommit.Tree()
		repo.CheckoutTree(tree, &git.CheckoutOpts{Strategy: git.CheckoutForce})
		repo.References.CreateSymbolic("HEAD", core.RefsTips+"test", true, "")
		test.WriteFile(repo, true, "foo", "line1 bis")
		test.Commit(repo, nil)

		// do the upgrade
		err := UpgradeCommand(repo, context.Context)
		if assert.NotNil(t, err) {
			assert.Equal(t, "Conflict while upgrading", err.Error())
		}

		// file foo should be in conflict
		index, _ := repo.Index()

		_, err = index.GetConflict("foo")
		assert.Nil(t, err, "File foo should be in conflict")

		// resolve the conflict
		test.WriteFile(repo, true, "foo", "line1 bis")

		// continue the upgrade
		err = UpgradeContinueCommand(repo)
		assert.Nil(t, err)

		// HEAD should be one commit ahead of master
		head, _ = repo.Head()
		master, _ := repo.References.Lookup("refs/heads/master")
		masterOid := master.Target()
		headCommit, _ := repo.LookupCommit(head.Target())
		assert.True(t, headCommit.Parent(0).Id().Equal(masterOid))

		// the tail should be on master's target
		newTailRef, _ := repo.References.Lookup(core.RefsTails + "test")
		assert.True(t, newTailRef.Target().Equal(masterOid))

		// the repo state should be clean
		assert.Equal(t, git.RepositoryStateNone, repo.State())

	})

	test.RunOnRemote(t, "UpgradeEmptyTip", func(t *testing.T, context test.TestContext, repo, remote *git.Repository) {
		// create a tip on head based on origin/master
		head, _ := repo.Head()
		repo.References.Create("refs/remotes/origin/master", head.Target(), true, "")
		test.CreateTip(repo, "test", "refs/remotes/origin/master", true)

		// Add a commit on origin/master
		masterOid, _ := test.Commit(repo, &test.CommitParams{Refname: "refs/remotes/origin/master"})
		// then select the tip and commit

		assert.False(t, masterOid.Equal(head.Target()))

		// do the upgrade
		err := UpgradeCommand(repo, context.Context)

		assert.Nil(t, err)

		// the tip should be on masterOid
		upgradedTip, _ := repo.References.Lookup(core.RefsTips + "test")
		assert.True(t, masterOid.Equal(upgradedTip.Target()))

		// the tail should be on masterOid too
		upgradedTail, _ := repo.References.Lookup(core.RefsTails + "test")
		assert.True(t, masterOid.Equal(upgradedTail.Target()))

		// the tip shouldn't be pushed on origin since it is empty
		_, err = remote.References.Lookup(core.RefsTips + "test")
		assert.NotNil(t, err)

		_, err = repo.References.Lookup(core.RefsRemoteTips + "origin/test")
		assert.NotNil(t, err)
	})

	test.RunOnRepo(t, "DirtyStateError", func(t *testing.T, context test.TestContext, repo *git.Repository) {
		// create a tip on head based on refs/remotes/origin/master
		test.CreateTip(repo, "test", "refs/heads/master", true)

		test.WriteFile(repo, true, "foo", "line")
		test.Commit(repo, nil)

		test.WriteFile(repo, false, "foo", "bar")

		err := UpgradeCommand(repo, context.Context)

		assert.NotNil(t, err)
	})
}
