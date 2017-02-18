package commands

import (
	"github.com/apflieger/tie/core"
	"github.com/apflieger/tie/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/libgit2/git2go.v25"
	"testing"
)

func TestTipCreate(t *testing.T) {
	test.RunOnRepo(t, "FromHead", func(t *testing.T, repo *git.Repository) {
		// create a tip based on HEAD
		TipCreateCommand(repo, "test", "")

		// head should be attached to the new tip
		head, _ := repo.Head()
		assert.Equal(t, core.RefsTips+"test", head.Name())

		// tip base should be on refs/heads/master
		config, _ := repo.Config()
		baseName, _ := config.LookupString("tip.test.base")
		assert.Equal(t, "refs/heads/master", baseName)

		// tail should be on head's target
		tail, _ := repo.References.Lookup(core.RefsTails + "test")
		assert.True(t, tail.Target().Equal(head.Target()))
	})

	test.RunOnRepo(t, "OnOtherBranch", func(t *testing.T, repo *git.Repository) {
		test.WriteFile(repo, true, "foo", "line")
		oid, _ := test.Commit(repo, &test.CommitParams{
			Refname: "refs/remotes/origin/master",
		})

		// just in case
		test.StatusClean(t, repo)

		// create a tip based on HEAD
		TipCreateCommand(repo, "test", "refs/remotes/origin/master")

		// status should be clean
		test.StatusClean(t, repo)

		// head should be attached to the new tip
		head, _ := repo.Head()
		assert.Equal(t, core.RefsTips+"test", head.Name())
		// head should point to origin/master's commit
		assert.True(t, oid.Equal(head.Target()))

		// tip base should be on origin/master
		config, _ := repo.Config()
		baseName, _ := config.LookupString("tip.test.base")
		assert.Equal(t, "refs/remotes/origin/master", baseName)

		// tail should be on origin/master's target
		tail, _ := repo.References.Lookup(core.RefsTails + "test")
		assert.True(t, tail.Target().Equal(oid))
	})

	test.RunOnRepo(t, "LocalTipAlreadyExists", func(t *testing.T, repo *git.Repository) {
		head, _ := repo.Head()

		repo.References.Create(core.RefsTips+"test", head.Target(), true, "")

		// create a tip based on HEAD
		err := TipCreateCommand(repo, "test", "")

		if assert.NotNil(t, err) {
			assert.Equal(t, "Failed to write reference '"+core.RefsTips+"test': a reference with that name already exists.", err.Error())
		}
	})

	test.RunOnRepo(t, "BaseDoesntExists", func(t *testing.T, repo *git.Repository) {

		err := TipCreateCommand(repo, "test", "refs/remotes/github/master")

		if assert.NotNil(t, err) {
			assert.Equal(t, "Reference '"+"refs/remotes/github/master' not found", err.Error())
		}
	})

	test.RunOnRepo(t, "RemoteTipAlreadyExists", func(t *testing.T, repo *git.Repository) {
		head, _ := repo.Head()

		repo.References.Create(core.RefsRemoteTips+"github/test", head.Target(), true, "")
		repo.References.Create("refs/remotes/github/master", head.Target(), true, "")

		// create a tip based on some branch on github
		err := TipCreateCommand(repo, "test", "refs/remotes/github/master")

		if assert.NotNil(t, err) {
			assert.Equal(t, "Failed to create tip \"test\". A tip with that name already exists on github.", err.Error())
		}
	})
}
