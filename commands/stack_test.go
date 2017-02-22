package commands

import (
	"github.com/apflieger/tie/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/libgit2/git2go.v25"
	"testing"
)

func TestStack(t *testing.T) {
	test.RunOnRemote(t, "SimpleStack", func(t *testing.T, repo, origin *git.Repository) {
		// create a tip 1 commit ahead of master
		test.CreateTip(repo, "test", "refs/heads/master", true)
		test.WriteFile(repo, true, "foo", "line")
		oid, _ := test.Commit(repo, nil)

		err := StackCommand(repo)

		assert.Nil(t, err)

		// master should have been fast forwarded and selected
		head, _ := repo.Head()
		assert.Equal(t, "refs/heads/master", head.Name())
		assert.True(t, head.Target().Equal(oid))

		// status should be clean
		test.StatusClean(t, repo)
	})

	test.RunOnRepo(t, "NotOnTip", func(t *testing.T, repo *git.Repository) {
		// Select a branch
		head, _ := repo.Head()
		repo.References.Create("refs/heads/test", head.Target(), false, "")
		_, err := repo.References.CreateSymbolic("HEAD", "refs/heads/test", true, "")
		assert.Nil(t, err)

		err = StackCommand(repo)

		assert.NotNil(t, err)
	})
}
