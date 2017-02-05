package commands

import (
	"fmt"
	"github.com/apflieger/tie/core"
	"gopkg.in/libgit2/git2go.v25"
	"strings"
)

func CommitCommand(repo *git.Repository, message string) error {
	head, headCommit, tree := core.PrepareCommit(repo)
	signature, _ := repo.DefaultSignature()
	repo.CreateCommit(head.Name(), signature, signature, message, tree, headCommit)

	// push the tip on the remote corresponding to its base
	config, _ := repo.Config()
	tipName := strings.Replace(head.Name(), "refs/tips/local/", "", 1)
	base, _ := config.LookupString(fmt.Sprintf("tip.%v.base", tipName))

	if remoteName, err := core.RemoteName(base); err == nil {
		remote, _ := repo.Remotes.Lookup(remoteName)
		remote.Push([]string{fmt.Sprintf("+%v:%v", head.Name(), head.Name())}, nil)
	}

	return nil
}
