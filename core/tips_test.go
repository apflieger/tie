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
		tip, _ := repo.References.Create(RefsTips+"test", head.Target(), false, "")
		config, _ := repo.Config()
		config.SetString("tip.test.base", "refs/remotes/origin/master")

		// push the tip
		PushTip(repo, tip)

		// local repo should have a remote tip
		_, err := repo.References.Lookup(RefsRemoteTips + "origin/test")
		assert.Nil(t, err)

		// remote repo should have a local tip
		_, err = remote.References.Lookup(RefsTips + "test")
		assert.Nil(t, err)
	})

	test.RunOnRepo(t, "NoRemote", func(t *testing.T, repo *git.Repository) {
		// setup a tip based on refs/heads/master
		head, _ := repo.Head()
		tip, _ := repo.References.Create(RefsTips+"test", head.Target(), false, "")
		config, _ := repo.Config()
		config.SetString("tip.test.base", "refs/heads/master")

		// push the tip
		err := PushTip(repo, tip)

		// push should have failed
		assert.NotNil(t, err)

		// local repo should not have a remote tip
		_, err = repo.References.Lookup(RefsRemoteTips + "origin/test")
		assert.NotNil(t, err)
	})

	test.RunOnRemote(t, "BranchCompatibilityMode", func(t *testing.T, repo, remote *git.Repository) {
		// configure the repo to branch compatibility mode
		config, _ := repo.Config()
		config.SetBool("tie.pushTipsAsBranches", true)

		// setup a tip based on origin/master
		head, _ := repo.Head()
		tip, _ := repo.References.Create(RefsTips+"test", head.Target(), false, "")
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

func TestFormatCommitMessage(t *testing.T) {
	assert.Equal(t, "", FormatCommitMessage(""))
	assert.Equal(t, "test\n", FormatCommitMessage("test"))
	assert.Equal(t, "test\n", FormatCommitMessage("test\n"))
	assert.Equal(t, "test\n", FormatCommitMessage("test\n\n"))
	assert.Equal(t, "test\n", FormatCommitMessage("\ntest"))
	assert.Equal(t, "test\n", FormatCommitMessage("  \ntest"))
	assert.Equal(t, "test\n\ntext bellow\n", FormatCommitMessage("test\n\ntext bellow"))
	assert.Equal(t, "test\n  text bellow\n", FormatCommitMessage("test\n  text bellow")) // keep indentation
	assert.Equal(t, "test test\nline2.\n", FormatCommitMessage("test test\nline2."))
	assert.Equal(t, "", FormatCommitMessage("#comment"))
	assert.Equal(t, "", FormatCommitMessage("#comment\n"))
	assert.Equal(t, "test\n", FormatCommitMessage("test\n#comment"))
	assert.Equal(t, "test\n", FormatCommitMessage("test\n #comment"))
	assert.Equal(t, "test #not a comment\n", FormatCommitMessage("test #not a comment"))
}
