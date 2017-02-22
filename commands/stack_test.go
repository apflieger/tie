package commands

import (
	"github.com/apflieger/tie/core"
	"github.com/apflieger/tie/test"
	"gopkg.in/libgit2/git2go.v25"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestStack(t *testing.T) {
	test.RunOnRemote(t, "SimpleStack", func(t *testing.T, repo, origin *git.Repository) {
		// create a tip 1 commit ahead of master
		head, _ := repo.Head()
		repo.References.Create(core.RefsTips+"test", head.Target(), false, "")
		repo.References.Create(core.RefsTails+"test", head.Target(), false, "")
		config, _ := repo.Config()
		config.SetString("tip.test.base", "refs/heads/master")

		repo.References.CreateSymbolic("HEAD", core.RefsTips+"test", true, "")
		test.WriteFile(repo, true, "foo", "line")
		oid, _ := test.Commit(repo, nil)

		err := StackCommand(repo)

		assert.Nil(t, err)

		// master should have been fast forwarded and selected
		head, _ = repo.Head()
		assert.Equal(t, "refs/heads/master", head.Name())
		assert.True(t, head.Target().Equal(oid))

		// status should be clean
		test.StatusClean(t, repo)
	})
}
