package commands

import (
	"github.com/apflieger/tie/core"
	"gopkg.in/libgit2/git2go.v25"
	"path/filepath"
)

func CommitCommand(repo *git.Repository, message string, openEditor core.OpenEditor) error {
	if message == "" {
		config, _ := repo.Config()
		message, _ = openEditor(config, filepath.Join(repo.Path(), "COMMIT_EDITMSG"))
	}

	head, headCommit, tree := core.PrepareCommit(repo)
	signature, _ := repo.DefaultSignature()
	repo.CreateCommit(head.Name(), signature, signature, message, tree, headCommit)

	// push the tip on the remote corresponding to its base
	core.PushTip(repo, head)

	return nil
}
