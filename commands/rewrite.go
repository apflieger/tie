package commands

import (
	"github.com/apflieger/tie/core"
	"gopkg.in/libgit2/git2go.v25"
	"io/ioutil"
	"path/filepath"
)

func AmendCommand(repo *git.Repository, commitMessage string, openEditor core.OpenEditor) error {
	head, headCommit, tree := core.PrepareCommit(repo)

	committer, _ := repo.DefaultSignature()

	if commitMessage == core.OptionMissing {
		commitMessage = headCommit.Message()
	}

	if commitMessage == core.OptionWithoutValue {
		config, _ := repo.Config()
		commitEditMsgFile := filepath.Join(repo.Path(), "COMMIT_EDITMSG")
		ioutil.WriteFile(commitEditMsgFile, []byte(headCommit.Message()), 0644)

		commitMessage, _ = openEditor(config, commitEditMsgFile)
	}

	_, err := headCommit.Amend(head.Name(), headCommit.Author(), committer, core.FormatCommitMessage(commitMessage), tree)

	tipName, _ := core.TipName(head.Name())
	core.PushTip(repo, tipName)

	return err
}
