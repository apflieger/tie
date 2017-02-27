package commands

import (
	"bytes"
	"github.com/apflieger/tie/core"
	"github.com/apflieger/tie/model"
	"github.com/apflieger/tie/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/libgit2/git2go.v25"
	"testing"
)

func TestDeleteCommand(t *testing.T) {
	test.RunOnRepo(t, "LocalTip", func(t *testing.T, context model.Context, repo *git.Repository) {
		// create a tip with his tail and base
		test.CreateTip(repo, "test", "refs/heads/master", false)

		var logBuffer *bytes.Buffer
		err := DeleteCommand(repo, test.CreateTestLogger(&logBuffer), []string{core.RefsTips + "test"}, context)

		assert.Nil(t, err)

		// tip's head should be deleted
		_, err = repo.References.Lookup(core.RefsTips + "test")
		assert.NotNil(t, err)

		// output should be...
		assert.Equal(t, "Deleted tip 'test'\n", logBuffer.String())
	})

	test.RunOnRepo(t, "Branch", func(t *testing.T, context model.Context, repo *git.Repository) {
		// create a local branch
		head, _ := repo.Head()
		repo.References.Create("refs/heads/test", head.Target(), false, "")

		var logBuffer *bytes.Buffer
		err := DeleteCommand(repo, test.CreateTestLogger(&logBuffer), []string{"refs/heads/test"}, context)

		// tie delete doesn't allow to delete branches
		assert.NotNil(t, err)

		// Branch should still be here
		_, err = repo.References.Lookup("refs/heads/test")
		assert.Nil(t, err)
	})

	test.RunOnRemote(t, "RemoteTip", func(t *testing.T, context model.Context, repo, remote *git.Repository) {
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
		err := DeleteCommand(repo, test.CreateTestLogger(&logBuffer), []string{tipRefName}, context)

		assert.Nil(t, err)

		// output should be...
		assert.Equal(t, "Deleted tip 'test'\n", logBuffer.String())
	})

	test.RunOnRepo(t, "UnreachableRemote", func(t *testing.T, context model.Context, repo *git.Repository) {
		// create a tip with his tail and base on origin/master
		test.CreateTip(repo, "test", "refs/remotes/origin/master", false)

		head, _ := repo.Head()
		repo.References.Create(core.RefsRemoteTips+"origin/test", head.Target(), false, "")
		// create an unreachable origin remote
		repo.Remotes.Create("origin", "/dev/null")

		var logBuffer *bytes.Buffer
		err := DeleteCommand(repo, test.CreateTestLogger(&logBuffer), []string{core.RefsTips + "test"}, context)

		assert.Nil(t, err)

		// output should be...
		assert.Contains(t, logBuffer.String(), "Tip 'test' has been deleted locally but not on origin")
	})
}
