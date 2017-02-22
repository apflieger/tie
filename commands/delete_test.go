package commands

import (
	"bytes"
	"github.com/apflieger/tie/core"
	"github.com/apflieger/tie/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/libgit2/git2go.v25"
	"testing"
)

func TestDeleteCommand(t *testing.T) {
	test.RunOnRepo(t, "LocalTip", func(t *testing.T, repo *git.Repository) {
		// create a tip with his tail and base
		test.CreateTip(repo, "test", "refs/heads/master", false)

		var logBuffer *bytes.Buffer
		err := DeleteCommand(repo, test.CreateTestLogger(&logBuffer), []string{core.RefsTips + "test"})

		assert.Nil(t, err)

		// tip's head should be deleted
		_, err = repo.References.Lookup(core.RefsTips + "test")
		assert.NotNil(t, err)
		// tip's tail should be deleted
		_, err = repo.References.Lookup(core.RefsTails + "test")
		assert.NotNil(t, err)
		// tip's base should be deleted
		config, _ := repo.Config()
		_, err = config.LookupString("tip.test.base")
		assert.NotNil(t, err)
	})

	test.RunOnRepo(t, "Branch", func(t *testing.T, repo *git.Repository) {
		// create a local branch
		head, _ := repo.Head()
		repo.References.Create("refs/heads/test", head.Target(), false, "")

		var logBuffer *bytes.Buffer
		err := DeleteCommand(repo, test.CreateTestLogger(&logBuffer), []string{"refs/heads/test"})

		// tie delete doesn't allow to delete branches
		assert.NotNil(t, err)

		// Branch should still be here
		_, err = repo.References.Lookup("refs/heads/test")
		assert.Nil(t, err)
	})

	test.RunOnRemote(t, "RemoteTip", func(t *testing.T, repo, remote *git.Repository) {
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
		config.SetBool("tie.pushTipsAsBranches", true)
		repo.References.Create("refs/remotes/origin/tips/test", head.Target(), false, "")
		origin.Push([]string{tipRefName + ":refs/heads/tips/test"}, nil)

		var logBuffer *bytes.Buffer
		err := DeleteCommand(repo, test.CreateTestLogger(&logBuffer), []string{tipRefName})

		assert.Nil(t, err)

		// tip's head should be deleted
		_, err = repo.References.Lookup(tipRefName)
		assert.NotNil(t, err)
		// base config should be deleted
		_, err = config.LookupString("tip.test.base")
		assert.NotNil(t, err)
		// rtip should be deleted
		_, err = repo.References.Lookup(core.RefsRemoteTips + "origin/test")
		assert.NotNil(t, err)
		// tip on remote should be deleted
		_, err = remote.References.Lookup(tipRefName)
		assert.NotNil(t, err)

		// the branch tip should be deleted on origin
		_, err = remote.References.Lookup("refs/heads/tips/test")
		assert.NotNil(t, err)
		// same for the local remote branch
		_, err = repo.References.Lookup("refs/remotes/origin/tips/test")
		assert.NotNil(t, err)
	})

	test.RunOnRepo(t, "UnreachableRemote", func(t *testing.T, repo *git.Repository) {
		// create a tip with his tail and base on origin/master
		test.CreateTip(repo, "test", "refs/remotes/origin/master", false)

		head, _ := repo.Head()
		repo.References.Create(core.RefsRemoteTips+"origin/test", head.Target(), false, "")
		// create an unreachable origin remote
		repo.Remotes.Create("origin", "/dev/null")

		var logBuffer *bytes.Buffer
		err := DeleteCommand(repo, test.CreateTestLogger(&logBuffer), []string{core.RefsTips + "test"})

		assert.Nil(t, err)

		// tip's head should be deleted
		_, err = repo.References.Lookup(core.RefsTips + "test")
		assert.NotNil(t, err)
		// base config should be deleted
		config, _ := repo.Config()
		_, err = config.LookupString("tip.test.base")
		assert.NotNil(t, err)
		// rtip should not be deleted
		_, err = repo.References.Lookup(core.RefsRemoteTips + "origin/test")
		assert.Nil(t, err)
	})
}
