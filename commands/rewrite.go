package commands

import (
	"github.com/apflieger/tie/core"
	"gopkg.in/libgit2/git2go.v25"
)

func AmendCommand(repo *git.Repository) error {
	head, headCommit, tree := core.PrepareCommit(repo)
	committer, _ := repo.DefaultSignature()
	_, err := headCommit.Amend(head.Name(), headCommit.Author(), committer, headCommit.Message(), tree)

	core.PushTip(repo, head)

	return err
}
