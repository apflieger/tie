package commands

import (
	"github.com/apflieger/tie/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/libgit2/git2go.v25"
	"testing"
)

func TestCommit(t *testing.T) {
	test.RunRequireRepo(t, "Commit", func(t *testing.T, repo *git.Repository) {
		// Create a file and add it to the index
		test.WriteFile(repo, true, "foo", "line")

		// create/select a tip
		head, _ := repo.Head()
		repo.References.Create("refs/tips/local/test", head.Target(), true, "")
		SelectCommand(repo, []string{"test"})
		// setup origin and base the tip on origin/master
		origin := test.CreateTestRepo(true)
		defer test.CleanRepo(origin)
		repo.Remotes.Create("origin", origin.Path())
		config, _ := repo.Config()
		config.SetString("tip.test.base", "refs/remotes/origin/master")

		// tie commit
		err := CommitCommand(repo, []string{"fix typo"})

		// We expect the target of head to be one commit ahead, status clear and HEAD still on the tip
		head2, _ := repo.Head()
		newCommit, _ := repo.LookupCommit(head2.Target())
		assert.Equal(t, 0, newCommit.Parent(0).Id().Cmp(head.Target()))
		assert.Equal(t, "fix typo", newCommit.Message())
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
