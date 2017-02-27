package core

import (
	"bytes"
	"github.com/apflieger/tie/model"
	"github.com/apflieger/tie/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/libgit2/git2go.v25"
	"testing"
)

func TestTips(t *testing.T) {
	test.RunOnRepo(t, "PrepareCommit", func(t *testing.T, context model.Context, repo *git.Repository) {
		head, headCommit, tree := PrepareCommit(repo)
		assert.Equal(t, "refs/heads/master", head.Name())
		assert.Equal(t, head.Target(), headCommit.Id())
		headCommitTree, _ := headCommit.Tree()
		assert.Equal(t, headCommitTree, tree)
	})
}

func TestPushTip(t *testing.T) {
	test.RunOnRemote(t, "PushSuccess", func(t *testing.T, context model.Context, repo, remote *git.Repository) {
		// setup a tip based on origin/master
		head, _ := repo.Head()
		tip, _ := repo.References.Create(RefsTips+"test", head.Target(), false, "")
		config, _ := repo.Config()
		config.SetString("tip.test.base", "refs/remotes/origin/master")

		oid, _ := test.Commit(repo, &test.CommitParams{
			Refname: tip.Name(),
		})

		// push the tip
		PushTip(repo, "test", context)

		// local repo should have a remote tip
		rtip, err := repo.References.Lookup(RefsRemoteTips + "origin/test")
		assert.Nil(t, err)
		assert.True(t, rtip.Target().Equal(oid))

		// remote repo should have a local tip
		originTip, err := remote.References.Lookup(RefsTips + "test")
		assert.Nil(t, err)
		assert.True(t, originTip.Target().Equal(oid))
	})

	test.RunOnRepo(t, "BaseNotRemote", func(t *testing.T, context model.Context, repo *git.Repository) {
		// setup a tip based on refs/heads/master
		head, _ := repo.Head()
		repo.References.Create(RefsTips+"test", head.Target(), false, "")
		config, _ := repo.Config()
		config.SetString("tip.test.base", "refs/heads/master")

		// push the tip
		err := PushTip(repo, "test", context)

		// push should have failed
		assert.NotNil(t, err)

		// local repo should not have a remote tip
		_, err = repo.References.Lookup(RefsRemoteTips + "origin/test")
		assert.NotNil(t, err)
	})

	test.RunOnRepo(t, "RemoteDoesntExists", func(t *testing.T, context model.Context, repo *git.Repository) {
		// setup a tip based on refs/remotes/somewhere/master
		head, _ := repo.Head()
		repo.References.Create(RefsTips+"test", head.Target(), false, "")
		config, _ := repo.Config()
		config.SetString("tip.test.base", "refs/remotes/somewhere/master")

		// push the tip
		err := PushTip(repo, "test", context)

		// push should have failed
		assert.NotNil(t, err)

		// local repo should not have a remote tip
		_, err = repo.References.Lookup(RefsRemoteTips + "somewhere/test")
		assert.NotNil(t, err)
	})

	test.RunOnRepo(t, "RemoteUnreachable", func(t *testing.T, context model.Context, repo *git.Repository) {
		// setup a tip based on refs/remotes/origin/master
		head, _ := repo.Head()
		repo.References.Create(RefsTips+"test", head.Target(), false, "")
		config, _ := repo.Config()
		config.SetString("tip.test.base", "refs/remotes/origin/master")
		// create an unreachable origin remote
		repo.Remotes.Create("origin", "/dev/null")

		// push the tip
		err := PushTip(repo, "test", context)

		// push should have failed
		assert.NotNil(t, err)

		// local repo should not have a remote tip
		_, err = repo.References.Lookup(RefsRemoteTips + "origin/test")
		assert.NotNil(t, err)
	})

	test.RunOnRemote(t, "BranchCompatibilityMode", func(t *testing.T, context model.Context, repo, remote *git.Repository) {
		// configure the repo to branch compatibility mode
		config, _ := repo.Config()
		config.SetBool("tie.pushTipsAsBranches", true)

		// setup a tip based on origin/master
		head, _ := repo.Head()
		tip, _ := repo.References.Create(RefsTips+"test", head.Target(), false, "")
		config.SetString("tip.test.base", "refs/remotes/origin/master")

		oid, _ := test.Commit(repo, &test.CommitParams{
			Refname: tip.Name(),
		})

		// push the tip
		err := PushTip(repo, "test", context)
		assert.Nil(t, err)

		// local repo should have a remote branch corresponding to the tip
		rBranch, err := repo.References.Lookup("refs/remotes/origin/tips/test")
		assert.Nil(t, err)
		assert.True(t, rBranch.Target().Equal(oid))

		// remote repo should have a local branch corresponding to the tip
		originBranch, err := remote.References.Lookup("refs/heads/tips/test")
		assert.Nil(t, err)
		assert.True(t, originBranch.Target().Equal(oid))
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

func TestDeleteTip(t *testing.T) {
	test.RunOnRepo(t, "LocalTip", func(t *testing.T, context model.Context, repo *git.Repository) {
		// create a tip with his tail and base
		test.CreateTip(repo, "test", "refs/heads/master", false)

		var out *bytes.Buffer
		DeleteTip(repo, "test", test.CreateTestLogger(&out), context)

		// tip's head should be deleted
		_, err := repo.References.Lookup(RefsTips + "test")
		assert.NotNil(t, err)
		// tip's tail should be deleted
		_, err = repo.References.Lookup(RefsTails + "test")
		assert.NotNil(t, err)
		// tip's base should be deleted
		config, _ := repo.Config()
		_, err = config.LookupString("tip.test.base")
		assert.NotNil(t, err)

		// Output should be...
		assert.Equal(t, "Deleted tip 'test'\n", out.String())
	})

	test.RunOnRemote(t, "RemoteTip", func(t *testing.T, context model.Context, repo, remote *git.Repository) {
		// create a tip with his tail and base
		tipRefName := RefsTips + "test"
		head, _ := repo.Head()
		test.CreateTip(repo, "test", "refs/remotes/origin/master", false)

		// set it up on origin
		repo.References.Create(RefsRemoteTips+"origin/test", head.Target(), false, "")
		origin, _ := repo.Remotes.Lookup("origin")
		origin.Push([]string{tipRefName + ":" + tipRefName}, nil)
		// activate branch compatibility mode
		config, _ := repo.Config()
		config.SetBool("tie.pushTipsAsBranches", true)
		repo.References.Create("refs/remotes/origin/tips/test", head.Target(), false, "")
		origin.Push([]string{tipRefName + ":refs/heads/tips/test"}, nil)

		var out *bytes.Buffer
		DeleteTip(repo, "test", test.CreateTestLogger(&out), context)

		// tip's head should be deleted
		_, err := repo.References.Lookup(tipRefName)
		assert.NotNil(t, err)
		// base config should be deleted
		_, err = config.LookupString("tip.test.base")
		assert.NotNil(t, err)
		// rtip should be deleted
		_, err = repo.References.Lookup(RefsRemoteTips + "origin/test")
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

		// Output should be...
		assert.Equal(t, "Deleted tip 'test'\n", out.String())
	})

	test.RunOnRepo(t, "UnreachableRemote", func(t *testing.T, context model.Context, repo *git.Repository) {
		// create a tip with his tail and base on origin/master
		test.CreateTip(repo, "test", "refs/remotes/origin/master", false)

		head, _ := repo.Head()
		repo.References.Create(RefsRemoteTips+"origin/test", head.Target(), false, "")
		// create an unreachable origin remote
		repo.Remotes.Create("origin", "/dev/null")

		var out *bytes.Buffer
		DeleteTip(repo, "test", test.CreateTestLogger(&out), context)

		// tip's head should be deleted
		_, err := repo.References.Lookup(RefsTips + "test")
		assert.NotNil(t, err)
		// base config should be deleted
		config, _ := repo.Config()
		_, err = config.LookupString("tip.test.base")
		assert.NotNil(t, err)
		// rtip should not be deleted
		_, err = repo.References.Lookup(RefsRemoteTips + "origin/test")
		assert.Nil(t, err)

		// Output should be...
		assert.Contains(t, out.String(), "Tip 'test' has been deleted locally but not on origin.\n")
	})
}
