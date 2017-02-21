package commands

import (
	"bytes"
	"github.com/apflieger/tie/core"
	"github.com/apflieger/tie/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/libgit2/git2go.v25"
	"testing"
)

func TestDelete(t *testing.T) {
	test.RunOnRepo(t, "SingleTip", func(t *testing.T, repo *git.Repository) {
		// create a tip with his tail and base
		head, _ := repo.Head()
		repo.References.Create(core.RefsTips+"test", head.Target(), false, "")
		repo.References.Create(core.RefsTails+"test", head.Target(), false, "")
		config, _ := repo.Config()
		config.SetString("tip.test.base", "refs/heads/master")

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
}
