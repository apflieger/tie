package core

import (
	"github.com/apflieger/tie/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/libgit2/git2go.v25"
	"testing"
)

func TestTips(t *testing.T) {
	test.RunOnRepo(t, "PrepareCommit", func(t *testing.T, repo *git.Repository) {
		head, headCommit, tree := PrepareCommit(repo)
		assert.Equal(t, "refs/heads/master", head.Name())
		assert.Equal(t, head.Target(), headCommit.Id())
		headCommitTree, _ := headCommit.Tree()
		assert.Equal(t, headCommitTree, tree)
	})
}

func TestPushTip(t *testing.T) {
	test.RunOnRemote(t, "PushSuccess", func(t *testing.T, repo, remote *git.Repository) {
		// setup a tip based on origin/master
		head, _ := repo.Head()
		tip, _ := repo.References.Create("refs/tips/local/test", head.Target(), false, "")
		config, _ := repo.Config()
		config.SetString("tip.test.base", "refs/remotes/origin/master")

		// push the tip
		PushTip(repo, tip)

		// local repo should have a remote tip
		_, err := repo.References.Lookup("refs/tips/origin/test")
		assert.Nil(t, err)

		// remote repo should have a local tip
		_, err = remote.References.Lookup("refs/tips/local/test")
		assert.Nil(t, err)
	})

	test.RunOnRemote(t, "BranchCompatibilityMode", func(t *testing.T, repo, remote *git.Repository) {
		// configure the repo to branch compatibility mode
		config, _ := repo.Config()
		config.SetBool("tie.pushTipsAsBranches", true)

		// setup a tip based on origin/master
		head, _ := repo.Head()
		tip, _ := repo.References.Create("refs/tips/local/test", head.Target(), false, "")
		config.SetString("tip.test.base", "refs/remotes/origin/master")

		// push the tip
		PushTip(repo, tip)

		var err error

		// local repo should have a remote branch corresponding to the tip
		_, err = repo.References.Lookup("refs/remotes/origin/tips/test")
		assert.Nil(t, err)

		// remote repo should have a local branch corresponding to the tip
		_, err = remote.References.Lookup("refs/heads/tips/test")
		assert.Nil(t, err)
	})
}
