package commands

import (
	"github.com/apflieger/tie/core"
	"github.com/stretchr/testify/assert"
	"gopkg.in/libgit2/git2go.v25"
	"testing"
)

func TestCommit(t *testing.T) {
	core.RunRequireRepo(t, "Commit", func(t *testing.T, repo *git.Repository) {
		// Create a file and add it to the index
		core.WriteFile(repo, true, "foo", "line")

		// create/select a tip
		head, _ := repo.Head()
		repo.References.Create("refs/tips/local/test", head.Target(), true, "")
		SelectCommand(repo, []string{"test"})
		// setup origin and base the tip on origin/master
		origin := core.CreateTestRepo(true)
		defer core.CleanRepo(origin)
		repo.Remotes.Create("origin", origin.Path())
		config, _ := repo.Config()
		config.SetString("tip.test.base", "refs/remotes/origin/master")

		// tie commit
		err := CommitCommand(repo, nil)

		// We expect the target of head to have changed, status clear and HEAD still on the tip
		head2, _ := repo.Head()
		assert.NotEqual(t, head.Target(), head2.Target())
		statusList, _ := repo.StatusList(
			&git.StatusOptions{
				Show:     git.StatusShowIndexAndWorkdir,
				Flags:    git.StatusOptIncludeUntracked,
				Pathspec: nil,
			})
		statusCount, _ := statusList.EntryCount()
		assert.Equal(t, 0, statusCount)
		assert.Nil(t, err)
		// We expect the tip to be pushed on origin
		remoteTip, err := origin.References.Lookup("refs/tips/local/test")
		if assert.Nil(t, err) {
			assert.Equal(t, 0, remoteTip.Target().Cmp(head2.Target()))
		}
	})
}
