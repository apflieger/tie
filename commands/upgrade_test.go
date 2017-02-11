package commands

import (
	"github.com/apflieger/tie/core"
	"github.com/apflieger/tie/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/libgit2/git2go.v25"
	"testing"
	"time"
)

func TestUpgrade(t *testing.T) {
	test.RunOnRepo(t, "NoTipSelected", func(t *testing.T, repo *git.Repository) {
		err := UpgradeCommand(repo)

		if assert.NotNil(t, err) {
			assert.Equal(t, "HEAD not on a tip. Only tips can be upgraded.", err.Error())
		}
	})

	test.RunOnRepo(t, "NoBase", func(t *testing.T, repo *git.Repository) {
		head, _ := repo.Head()
		repo.References.Create(core.RefsTips+"test", head.Target(), true, "")
		SelectCommand(repo, core.RefsTips+"test")

		err := UpgradeCommand(repo)

		if assert.NotNil(t, err) {
			assert.Equal(t, "Config value 'tip.test.base' was not found", err.Error())
		}
	})

	test.RunOnRepo(t, "NoTail", func(t *testing.T, repo *git.Repository) {
		head, _ := repo.Head()
		repo.References.Create(core.RefsTips+"test", head.Target(), true, "")
		SelectCommand(repo, core.RefsTips+"test")
		config, _ := repo.Config()
		config.SetString("tip.test.base", "refs/remotes/origin/master")

		err := UpgradeCommand(repo)

		if assert.NotNil(t, err) {
			assert.Equal(t, "Reference '"+core.RefsTails+"test' not found", err.Error())
		}
	})

	test.RunOnRemote(t, "UpgradeSuccess", func(t *testing.T, repo, remote *git.Repository) {
		// create a tip on head based on refs/remotes/origin/master
		head, _ := repo.Head()
		config, _ := repo.Config()
		config.SetString("tip.test.base", "refs/remotes/origin/master")
		repo.References.Create(core.RefsTips+"test", head.Target(), true, "")
		repo.References.Create(core.RefsTails+"test", head.Target(), true, "")

		// make origin/master and the tip diverge.
		masterOid, _ := test.Commit(repo, &test.CommitParams{Refname: "refs/remotes/origin/master"})
		SelectCommand(repo, core.RefsTips+"test")
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
		err := UpgradeCommand(repo)
		assert.Nil(t, err)

		// we expect the tip to be on top of origin/master
		head, _ = repo.Head()
		headCommit, _ := repo.LookupCommit(head.Target())
		assert.Equal(t, 0, headCommit.Parent(0).Parent(0).Id().Cmp(masterOid))
		assert.Equal(t, "last commit", headCommit.Message())
		assert.Equal(t, "user1", headCommit.Author().Name)
		assert.Equal(t, "email@example.com", headCommit.Author().Email)
		assert.Equal(t, now.Unix(), headCommit.Author().When.Unix())

		// we expect the tail to be updated on origin/master's target
		newTailRef, _ := repo.References.Lookup(core.RefsTails + "test")
		assert.Equal(t, 0, newTailRef.Target().Cmp(masterOid))

		// the repo state should be clean
		assert.Equal(t, git.RepositoryStateNone, repo.State())

		// We expect the tip to be pushed on origin
		remoteTip, err := remote.References.Lookup(core.RefsTips + "test")
		if assert.Nil(t, err) {
			assert.Equal(t, 0, remoteTip.Target().Cmp(head.Target()))
		}
	})

	test.RunOnRemote(t, "Conflict", func(t *testing.T, repo, remote *git.Repository) {
		// create a tip on head based on master
		head, _ := repo.Head()
		config, _ := repo.Config()
		config.SetString("tip.test.base", "refs/heads/master")
		repo.References.Create(core.RefsTips+"test", head.Target(), true, "")
		repo.References.Create(core.RefsTails+"test", head.Target(), true, "")

		// make master and the tip having a conflict.
		test.WriteFile(repo, true, "foo", "line1")
		test.Commit(repo, nil)
		SelectCommand(repo, core.RefsTips+"test")
		test.WriteFile(repo, true, "foo", "line1 bis")
		test.Commit(repo, nil)

		// do the upgrade
		err := UpgradeCommand(repo)
		if assert.NotNil(t, err) {
			assert.Equal(t, "Conflict while upgrading", err.Error())
		}
	})
}
