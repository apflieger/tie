package commands

import (
	"github.com/apflieger/tie/core"
	"github.com/apflieger/tie/model"
	"github.com/apflieger/tie/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/libgit2/git2go.v25"
	"io/ioutil"
	"testing"
)

func TestCommitCommand(t *testing.T) {
	test.RunOnRemote(t, "Commit", func(t *testing.T, context test.TestContext, repo, remote *git.Repository) {
		head, _ := repo.Head()

		// Create a file and add it to the index
		test.WriteFile(repo, true, "foo", "line")

		// create/select a tip
		test.CreateTip(repo, "test", "refs/remotes/origin/master", true)

		// tie commit -m "fix typo"
		err := CommitCommand(repo, "fix typo", model.OptionMissing, context.Context)
		assert.Nil(t, err)

		// We expect the target of head to be one commit ahead, status clear and HEAD still on the tip
		head2, _ := repo.Head()
		newCommit, _ := repo.LookupCommit(head2.Target())
		assert.True(t, newCommit.Parent(0).Id().Equal(head.Target()))
		assert.Equal(t, "fix typo\n", newCommit.Message()) // commit message has been formatted
		statusList, _ := repo.StatusList(
			&git.StatusOptions{
				Show:     git.StatusShowIndexAndWorkdir,
				Flags:    git.StatusOptIncludeUntracked,
				Pathspec: nil,
			})
		statusCount, _ := statusList.EntryCount()
		assert.Equal(t, 0, statusCount)

		// We expect the tip to be pushed on origin
		remoteTip, err := remote.References.Lookup(core.RefsTips + "test")
		if assert.Nil(t, err) {
			assert.True(t, remoteTip.Target().Equal(head2.Target()))
		}
		// rtip should be on the same oid
		rtip, err := repo.References.Lookup(core.RefsRemoteTips + "origin/test")
		if assert.Nil(t, err) {
			assert.True(t, rtip.Target().Equal(head2.Target()))
		}
	})

	test.RunOnRemote(t, "EditCommitMessage", func(t *testing.T, context test.TestContext, repo, remote *git.Repository) {

		// create/select a tip
		test.CreateTip(repo, "test", "refs/remotes/origin/master", true)

		test.Commit(repo, &test.CommitParams{
			Message: "A commit message.\nWith a second line.",
		})

		var presetCommitMessage string

		// tie commit with empty commit message
		context.OpenEditor = func(config *git.Config, file string) (string, error) {
			bytes, _ := ioutil.ReadFile(file)
			presetCommitMessage = string(bytes)
			return "Commit message from mocked editor", nil
		}
		CommitCommand(repo, "", model.OptionMissing, context.Context)

		// The commit message should have been preset with the previous one, commented
		assert.Equal(t, "#A commit message.\n#With a second line.\n", presetCommitMessage)

		// We expect the commit message to be filled by the editor
		head2, _ := repo.Head()
		newCommit, _ := repo.LookupCommit(head2.Target())
		assert.Equal(t, "Commit message from mocked editor\n", newCommit.Message()) // commit message has been formatted
	})

	test.RunOnRepo(t, "NotOnTipError", func(t *testing.T, context test.TestContext, repo *git.Repository) {
		err := CommitCommand(repo, "Commit on master", model.OptionMissing, context.Context)

		if assert.NotNil(t, err) {
			assert.Equal(t, "HEAD is not on a tip. Run 'commit -t' to create a tip on the fly.", err.Error())
		}
	})

	test.RunOnRepo(t, "OnTheFlyTipEmptyNameError", func(t *testing.T, context test.TestContext, repo *git.Repository) {
		err := CommitCommand(repo, "Commit on master", "", context.Context)
		if assert.NotNil(t, err) {
			assert.Equal(t, "Name of the tip can't be empty.", err.Error())
		}

		err = CommitCommand(repo, "Commit on master", " ", context.Context)
		if assert.NotNil(t, err) {
			assert.Equal(t, "Name of the tip can't be empty.", err.Error())
		}
	})

	test.RunOnRepo(t, "OnTheFlyNamedTip", func(t *testing.T, context test.TestContext, repo *git.Repository) {
		// Write a file
		test.WriteFile(repo, true, "foo", "bar")

		// tie commit -t new_tip -m "Added foo"
		err := CommitCommand(repo, "Added foo", "new_tip", context.Context)
		assert.Nil(t, err)

		// HEAD should be on new_tip
		head, _ := repo.Head()
		assert.Equal(t, core.RefsTips+"new_tip", head.Name())

		// Status should be clean
		test.StatusClean(t, repo)

		// new_tip should be based on master (previous HEAD)
		config, _ := repo.Config()
		base, _ := config.LookupString("tip.new_tip.base")
		assert.Equal(t, "refs/heads/master", base)

		// new_tip's tail should be on master
		tail, _ := repo.References.Lookup(core.RefsTails + "new_tip")
		master, _ := repo.References.Lookup("refs/heads/master")
		assert.True(t, tail.Target().Equal(master.Target()))
	})

	test.RunOnRepo(t, "OnTheFlyUnnamedTip", func(t *testing.T, context test.TestContext, repo *git.Repository) {
		// Write a file
		test.WriteFile(repo, true, "foo", "bar")

		// tie commit -t new_tip -m "Added foo"
		err := CommitCommand(repo, "Added foo", model.OptionWithoutValue, context.Context)
		assert.Nil(t, err)

		// HEAD should be on master-tip
		head, _ := repo.Head()
		assert.Equal(t, core.RefsTips+"master-tip", head.Name())
	})
}
