package commands

import (
	"github.com/apflieger/tie/core"
	"github.com/stretchr/testify/assert"
	"gopkg.in/libgit2/git2go.v25"
	"testing"
	"time"
)

func TestUpgrade(t *testing.T) {
	core.RunRequireRepo(t, "NoTipSelected", func(t *testing.T, repo *git.Repository) {
		err := UpgradeCommand(repo, nil)

		if assert.NotNil(t, err) {
			assert.Equal(t, "HEAD not on a tip. Only tips can be upgraded.", err.Error())
		}
	})

	core.RunRequireRepo(t, "NoBase", func(t *testing.T, repo *git.Repository) {
		head, _ := repo.Head()
		repo.References.Create("refs/tips/local/test", head.Target(), true, "")
		SelectCommand(repo, []string{"refs/tips/local/test"})

		err := UpgradeCommand(repo, nil)

		if assert.NotNil(t, err) {
			assert.Equal(t, "Config value 'tip.test.base' was not found", err.Error())
		}
	})

	core.RunRequireRepo(t, "NoTail", func(t *testing.T, repo *git.Repository) {
		head, _ := repo.Head()
		repo.References.Create("refs/tips/local/test", head.Target(), true, "")
		SelectCommand(repo, []string{"refs/tips/local/test"})
		config, _ := repo.Config()
		config.SetString("tip.test.base", "refs/remotes/origin/master")

		err := UpgradeCommand(repo, nil)

		if assert.NotNil(t, err) {
			assert.Equal(t, "Reference 'refs/tails/test' not found", err.Error())
		}
	})

	core.RunRequireRepo(t, "UpgradeSuccess", func(t *testing.T, repo *git.Repository) {
		head, _ := repo.Head()
		config, _ := repo.Config()
		// create a tip on head based on refs/remotes/origin/master
		config.SetString("tip.test.base", "refs/remotes/origin/master")
		repo.References.Create("refs/tips/local/test", head.Target(), true, "")
		repo.References.Create("refs/tails/test", head.Target(), true, "")

		// make origin/master and the tip diverge.
		masterOid, _ := core.Commit(repo, &core.CommitParams{Refname: "refs/remotes/origin/master"})
		SelectCommand(repo, []string{"refs/tips/local/test"})
		signature := &git.Signature{
			Name: "user1",
			Email: "email@example.com",
			When: time.Now(),
		}
		core.WriteFile(repo, true, "foo", "line1")
		core.Commit(repo, nil)
		core.WriteFile(repo, true, "foo", "line1", "line2")
		core.Commit(repo, &core.CommitParams{
			Author: signature,
			Commiter: signature,
			Message: "last commit",
		})

		// do the upgrade
		err := UpgradeCommand(repo, nil)
		assert.Nil(t, err)

		// we expect the tip to be on top of origin/master
		head, _ = repo.Head()
		headCommit, _ := repo.LookupCommit(head.Target())
		assert.Equal(t, 0, headCommit.Parent(0).Parent(0).Id().Cmp(masterOid))
		assert.Equal(t, "last commit", headCommit.Message())
		assert.Equal(t, "user1", headCommit.Author().Name)
		assert.Equal(t, "email@example.com", headCommit.Author().Email)
		assert.Equal(t, time.Now().Unix(), headCommit.Author().When.Unix())

		// we expect the tail to be updated on origin/master's target
		newTailRef, _ := repo.References.Lookup("refs/tails/test")
		assert.Equal(t, 0, newTailRef.Target().Cmp(masterOid))
	})
}
