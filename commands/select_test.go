package commands

import (
	"bytes"
	"github.com/apflieger/tie/core"
	"github.com/apflieger/tie/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/libgit2/git2go.v25"
	"testing"
)

func TestSelect(t *testing.T) {
	test.RunOnRepo(t, "SelectTip", func(t *testing.T, repo *git.Repository) {
		head, _ := repo.Head()

		// New tip ref created on HEAD
		repo.References.Create(core.RefsTips+"test", head.Target(), false, "")

		head, _ = repo.Head()
		assert.Equal(t, "refs/heads/master", head.Name())

		// Select the test tip
		SelectCommand(repo, "test")

		// We expect HEAD to be attached on the tip
		head, _ = repo.Head()
		assert.Equal(t, core.RefsTips+"test", head.Name())
	})

	test.RunOnRepo(t, "DwimFailed", func(t *testing.T, repo *git.Repository) {
		err := SelectCommand(repo, "test")

		if assert.NotNil(t, err) {
			assert.Equal(t, "No ref found for shorthand \"test\"", err.Error())
		}
	})

	test.RunOnRepo(t, "DirtyState", func(t *testing.T, repo *git.Repository) {
		// Commit a file on a new tip
		test.WriteFile(repo, true, "foo", "a")
		test.Commit(repo, &test.CommitParams{Refname: core.RefsTips + "test"})

		// write the same file on the working tree
		test.WriteFile(repo, false, "foo", "b")

		// select the tip
		err := SelectCommand(repo, "test")

		// We expect the select to fail because the checkout has a conflict
		if assert.NotNil(t, err) {
			assert.Equal(t, "1 conflict prevents checkout", err.Error())
		}
	})
}

func TestList(t *testing.T) {
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

	assertRefsList := func(t *testing.T, repo *git.Repository, buffer *bytes.Buffer, refs ...string) {
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
		assert.Equal(t, expectedBuffer.String(), buffer.String())
	}

	test.RunOnRepo(t, "DefaultListing", func(t *testing.T, repo *git.Repository) {
		setupRefs(repo)
		var logBuffer *bytes.Buffer
		ListCommand(repo, test.CreateTestLogger(&logBuffer), false, false, false)
		assertRefsList(t, repo, logBuffer,
			"refs/heads/master", // HEAD
			core.RefsTips+"tip1",
			core.RefsTips+"tip2",
			"refs/remotes/origin/branch2") // configured as base of a tip
	})

	test.RunOnRepo(t, "TipsListing", func(t *testing.T, repo *git.Repository) {
		setupRefs(repo)
		var logBuffer *bytes.Buffer
		ListCommand(repo, test.CreateTestLogger(&logBuffer), true, false, false)
		assertRefsList(t, repo, logBuffer,
			core.RefsTips+"tip1",
			core.RefsTips+"tip2")
	})

	test.RunOnRepo(t, "BranchListing", func(t *testing.T, repo *git.Repository) {
		setupRefs(repo)
		var logBuffer *bytes.Buffer
		ListCommand(repo, test.CreateTestLogger(&logBuffer), false, true, false)
		assertRefsList(t, repo, logBuffer,
			"refs/heads/branch1",
			"refs/heads/master")
	})

	test.RunOnRepo(t, "RemoteListing", func(t *testing.T, repo *git.Repository) {
		setupRefs(repo)
		var logBuffer *bytes.Buffer
		ListCommand(repo, test.CreateTestLogger(&logBuffer), false, false, true)
		assertRefsList(t, repo, logBuffer,
			core.RefsRemoteTips+"github/tip4",
			core.RefsRemoteTips+"origin/tip3",
			"refs/remotes/github/branch3",
			"refs/remotes/origin/branch2")
	})

	test.RunOnRepo(t, "RemoteTipsListing", func(t *testing.T, repo *git.Repository) {
		setupRefs(repo)
		var logBuffer *bytes.Buffer
		ListCommand(repo, test.CreateTestLogger(&logBuffer), true, false, true)
		assertRefsList(t, repo, logBuffer,
			core.RefsRemoteTips+"github/tip4",
			core.RefsRemoteTips+"origin/tip3")
	})

	test.RunOnRepo(t, "RemoteBranchListing", func(t *testing.T, repo *git.Repository) {
		setupRefs(repo)
		var logBuffer *bytes.Buffer
		ListCommand(repo, test.CreateTestLogger(&logBuffer), false, true, true)
		assertRefsList(t, repo, logBuffer,
			"refs/remotes/github/branch3",
			"refs/remotes/origin/branch2")
	})

	test.RunOnRepo(t, "LocalListing", func(t *testing.T, repo *git.Repository) {
		setupRefs(repo)
		var logBuffer *bytes.Buffer
		ListCommand(repo, test.CreateTestLogger(&logBuffer), true, true, false)
		assertRefsList(t, repo, logBuffer,
			core.RefsTips+"tip1",
			core.RefsTips+"tip2",
			"refs/heads/branch1",
			"refs/heads/master")
	})
}
