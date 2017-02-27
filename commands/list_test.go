package commands

import (
	"bytes"
	"github.com/apflieger/tie/core"
	"github.com/apflieger/tie/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/libgit2/git2go.v25"
	"testing"
)

func TestListCommand(t *testing.T) {
	setupRefs := func(repo *git.Repository) {
		head, _ := repo.Head()
		oid := head.Target()
		repo.References.Create(core.RefsTips+"tip1", oid, false, "")
		repo.References.Create(core.RefsTips+"tip2", oid, false, "")
		repo.References.Create(core.RefsRemoteTips+"origin/tip3", oid, false, "")
		repo.References.Create(core.RefsRemoteTips+"github/tip4", oid, false, "")
		repo.References.Create("refs/heads/branch1", oid, false, "")
		repo.References.Create("refs/remotes/origin/branch2", oid, false, "")
		repo.References.Create("refs/remotes/github/branch3", oid, false, "")

		config, _ := repo.Config()
		config.SetString("tip.tip1.base", "refs/remotes/origin/branch2")
		config.SetString("tip.tip2.base", "refs/remotes/origin/branch2")
	}

	assertRefsList := func(t *testing.T, repo *git.Repository, context test.TestContext, refs ...string) {
		expectedBuffer := new(bytes.Buffer)
		head, _ := repo.Head()
		directRef, _ := head.Resolve()
		for _, ref := range refs {
			if ref == directRef.Name() {
				expectedBuffer.WriteString("* ")
			} else {
				expectedBuffer.WriteString("  ")
			}
			expectedBuffer.WriteString(ref)
			expectedBuffer.WriteString("\n")
		}
		assert.Equal(t, expectedBuffer.String(), context.OutputBuffer.String())
	}

	test.RunOnRepo(t, "DefaultListing", func(t *testing.T, context test.TestContext, repo *git.Repository) {
		setupRefs(repo)
		ListCommand(repo, context.Context, false, false, false, false)
		assertRefsList(t, repo, context,
			"refs/heads/master", // HEAD
			core.RefsTips+"tip1",
			core.RefsTips+"tip2",
			"refs/remotes/origin/branch2") // configured as base of a tip
	})

	test.RunOnRepo(t, "TipsListing", func(t *testing.T, context test.TestContext, repo *git.Repository) {
		setupRefs(repo)
		ListCommand(repo, context.Context, true, false, false, false)
		assertRefsList(t, repo, context,
			core.RefsTips+"tip1",
			core.RefsTips+"tip2")
	})

	test.RunOnRepo(t, "BranchListing", func(t *testing.T, context test.TestContext, repo *git.Repository) {
		setupRefs(repo)
		ListCommand(repo, context.Context, false, true, false, false)
		assertRefsList(t, repo, context,
			"refs/heads/branch1",
			"refs/heads/master")
	})

	test.RunOnRepo(t, "RemoteListing", func(t *testing.T, context test.TestContext, repo *git.Repository) {
		setupRefs(repo)
		ListCommand(repo, context.Context, false, false, true, false)
		assertRefsList(t, repo, context,
			core.RefsRemoteTips+"github/tip4",
			core.RefsRemoteTips+"origin/tip3",
			"refs/remotes/github/branch3",
			"refs/remotes/origin/branch2")
	})

	test.RunOnRepo(t, "RemoteTipsListing", func(t *testing.T, context test.TestContext, repo *git.Repository) {
		setupRefs(repo)
		ListCommand(repo, context.Context, true, false, true, false)
		assertRefsList(t, repo, context,
			core.RefsRemoteTips+"github/tip4",
			core.RefsRemoteTips+"origin/tip3")
	})

	test.RunOnRepo(t, "RemoteBranchListing", func(t *testing.T, context test.TestContext, repo *git.Repository) {
		setupRefs(repo)
		ListCommand(repo, context.Context, false, true, true, false)
		assertRefsList(t, repo, context,
			"refs/remotes/github/branch3",
			"refs/remotes/origin/branch2")
	})

	test.RunOnRepo(t, "LocalListing", func(t *testing.T, context test.TestContext, repo *git.Repository) {
		setupRefs(repo)
		ListCommand(repo, context.Context, true, true, false, false)
		assertRefsList(t, repo, context,
			core.RefsTips+"tip1",
			core.RefsTips+"tip2",
			"refs/heads/branch1",
			"refs/heads/master")
	})

	test.RunOnRepo(t, "AllListing", func(t *testing.T, context test.TestContext, repo *git.Repository) {
		setupRefs(repo)
		ListCommand(repo, context.Context, false, false, false, true)
		assertRefsList(t, repo, context,
			core.RefsTips+"tip1",
			core.RefsTips+"tip2",
			core.RefsRemoteTips+"github/tip4",
			core.RefsRemoteTips+"origin/tip3",
			"refs/heads/branch1",
			"refs/heads/master",
			"refs/remotes/github/branch3",
			"refs/remotes/origin/branch2")
	})

	test.RunOnRepo(t, "AllTipsListing", func(t *testing.T, context test.TestContext, repo *git.Repository) {
		setupRefs(repo)
		ListCommand(repo, context.Context, true, false, false, true)
		assertRefsList(t, repo, context,
			core.RefsTips+"tip1",
			core.RefsTips+"tip2",
			core.RefsRemoteTips+"github/tip4",
			core.RefsRemoteTips+"origin/tip3")
	})

	test.RunOnRepo(t, "AllBranchesListing", func(t *testing.T, context test.TestContext, repo *git.Repository) {
		setupRefs(repo)
		ListCommand(repo, context.Context, false, true, false, true)
		assertRefsList(t, repo, context,
			"refs/heads/branch1",
			"refs/heads/master",
			"refs/remotes/github/branch3",
			"refs/remotes/origin/branch2")
	})
}
