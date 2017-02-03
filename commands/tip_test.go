package commands

import (
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
		assert.Equal(t, "refs/tips/local/test", head.Name())

		// tip base should be on refs/heads/master
		config, _ := repo.Config()
		baseName, _ := config.LookupString("tip.test.base")
		assert.Equal(t, "refs/heads/master", baseName)

		// tail should be on head's target
		tail, _ := repo.References.Lookup("refs/tails/test")
		assert.Equal(t, 0, tail.Target().Cmp(head.Target()))
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
		assert.Equal(t, "refs/tips/local/test", head.Name())
		// head should point to origin/master's commit
		assert.Equal(t, 0, oid.Cmp(head.Target()))

		// tip base should be on origin/master
		config, _ := repo.Config()
		baseName, _ := config.LookupString("tip.test.base")
		assert.Equal(t, "refs/remotes/origin/master", baseName)

		// tail should be on origin/master's target
		tail, _ := repo.References.Lookup("refs/tails/test")
		assert.Equal(t, 0, tail.Target().Cmp(oid))
	})

	test.RunOnRepo(t, "LocalTipAlreadyExists", func(t *testing.T, repo *git.Repository) {
		head, _ := repo.Head()

		repo.References.Create("refs/tips/local/test", head.Target(), true, "")

		// create a tip based on HEAD
		err := TipCreateCommand(repo, "test", "")

		if assert.NotNil(t, err) {
			assert.Equal(t, "Failed to write reference 'refs/tips/local/test': a reference with that name already exists.", err.Error())
		}
	})
}
