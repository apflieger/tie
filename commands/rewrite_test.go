package commands

import (
	"fmt"
	"github.com/apflieger/tie/core"
	"github.com/apflieger/tie/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/libgit2/git2go.v25"
	"io/ioutil"
	"testing"
)

func TestRewrite(t *testing.T) {
	test.RunOnRemote(t, "AmendHeadTree", func(t *testing.T, repo, remote *git.Repository) {
		// commit a file on a new tip
		test.WriteFile(repo, true, "foo", "line1")
		test.Commit(repo, &test.CommitParams{
			Message: "first commit",
			Refname: core.RefsTips + "test",
		})
		config, _ := repo.Config()
		config.SetString("tip.test.base", "refs/remotes/origin/master")

		// select the tip
		repo.References.CreateSymbolic("HEAD", core.RefsTips+"test", true, "")

		// change the file
		test.WriteFile(repo, true, "foo", "line1 amended")

		// amend the last commit
		err := AmendCommand(repo, core.OptionMissing, test.MockOpenEditor)

		assert.Nil(t, err)

		// the status should be clean
		statusList, _ := repo.StatusList(
			&git.StatusOptions{
				Show:     git.StatusShowIndexAndWorkdir,
				Flags:    git.StatusOptIncludeUntracked,
				Pathspec: nil,
			})
		statusCount, _ := statusList.EntryCount()
		assert.Equal(t, 0, statusCount)

		// the file should be recorded in his last state
		head, _ := repo.Head()
		headCommit, _ := repo.LookupCommit(head.Target())
		headTree, _ := headCommit.Tree()
		fooEntry := headTree.EntryByName("foo")
		blob, _ := repo.LookupBlob(fooEntry.Id)
		assert.Equal(t, "line1 amended", fmt.Sprintf("%s", blob.Contents()))

		// there should be 2 commits on the commit log
		logSize := 1
		for commit := headCommit; commit.ParentCount() > 0; logSize++ {
			commit = commit.Parent(0)
		}
		assert.Equal(t, 2, logSize)

		// We expect the tip to be pushed on origin
		remoteTip, err := remote.References.Lookup(core.RefsTips + "test")
		if assert.Nil(t, err) {
			assert.Equal(t, 0, remoteTip.Target().Cmp(head.Target()))
		}

		// The local remote tip should be set as well
		remoteTip, err = repo.References.Lookup(core.RefsRemoteTips + "origin/test")
		if assert.Nil(t, err) {
			assert.Equal(t, 0, remoteTip.Target().Cmp(head.Target()))
		}

		// The commit message should be the same but formatted
		assert.Equal(t, "first commit\n", headCommit.Message())
	})

	test.RunOnRemote(t, "AmendHeadMessage", func(t *testing.T, repo, remote *git.Repository) {
		// commit a file on a new tip
		test.WriteFile(repo, true, "foo", "line1")
		test.Commit(repo, &test.CommitParams{
			Message: "Commit message to be amended\nWith a 2nd line.",
			Refname: core.RefsTips + "test",
		})
		config, _ := repo.Config()
		config.SetString("tip.test.base", "refs/remotes/origin/master")

		// select the tip
		repo.References.CreateSymbolic("HEAD", core.RefsTips+"test", true, "")

		var presetCommitMessage string
		// amend the last commit using tie rewrite amend -m
		AmendCommand(repo, core.OptionWithoutValue, func(config *git.Config, file string) (string, error) {
			bytes, _ := ioutil.ReadFile(file)
			presetCommitMessage = string(bytes)
			return "Commit message from mocked editor", nil
		})

		assert.Equal(t, "Commit message to be amended\nWith a 2nd line.", presetCommitMessage)

		head, _ := repo.Head()
		headCommit, _ := repo.LookupCommit(head.Target())

		// The commit message should be the same but formatted
		assert.Equal(t, "Commit message from mocked editor\n", headCommit.Message())
	})
}
