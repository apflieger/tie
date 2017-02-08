package commands

import (
	"github.com/apflieger/tie/core"
	"gopkg.in/libgit2/git2go.v25"
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
		commitMessage, _ = openEditor(config, filepath.Join(repo.Path(), "COMMIT_EDITMSG"))
	}

	_, err := headCommit.Amend(head.Name(), headCommit.Author(), committer, commitMessage, tree)

	head, _ = repo.Head()
	core.PushTip(repo, head)

	return err
}
