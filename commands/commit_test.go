package commands

import (
	"testing"
	"github.com/apflieger/tie/core"
	"github.com/stretchr/testify/assert"
	"gopkg.in/libgit2/git2go.v25"
)

func TestCommit(t *testing.T) {
	repo := core.CreateTestRepo()

	// Create a file and add it to the index
	core.WriteFile(repo, "foo", "line")
	index, _ := repo.Index()
	index.AddByPath("foo")
	index.Write()

	// Select a tip and commit on it
	head, _ := repo.Head()
	repo.References.Create("refs/tips/local/test", head.Target(), true, "")
	SelectCommand(repo, []string{"test"})
	err := CommitCommand(repo, nil)

	// We expect the target of head to have changed, status clear and HEAD still on the tip
	head2, _ := repo.Head()
	assert.NotEqual(t, head.Target(), head2.Target())
	statusList, _ := repo.StatusList(
		&git.StatusOptions{
			Show: git.StatusShowIndexAndWorkdir,
			Flags: git.StatusOptIncludeUntracked,
			Pathspec: nil,
		})
	statusCount, _ := statusList.EntryCount()
	assert.Equal(t, 0, statusCount)
	assert.Nil(t, err)
}