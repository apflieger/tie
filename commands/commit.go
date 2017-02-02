package commands

import (
	"fmt"
	"github.com/apflieger/tie/core"
	"gopkg.in/libgit2/git2go.v25"
	"regexp"
	"strings"
)

func CommitCommand(repo *git.Repository, message string) error {
	head, headCommit, tree := core.PrepareCommit(repo)
	signature, _ := repo.DefaultSignature()
	repo.CreateCommit(head.Name(), signature, signature, message, tree, headCommit)

	// push the tip on the remote corresponding to its base
	config, _ := repo.Config()
	tipName := strings.Replace(head.Name(), "refs/tips/local/", "", 1)
	base, err := config.LookupString(fmt.Sprintf("tip.%v.base", tipName))

	if err == nil && strings.HasPrefix(base, "refs/remotes/") {
		exp := regexp.MustCompile("refs/remotes/([^/]*)/.*")
		remoteName := exp.FindStringSubmatch(base)[1]
		remote, _ := repo.Remotes.Lookup(remoteName)
		remote.Push([]string{fmt.Sprintf("+%v:%v", head.Name(), head.Name())}, nil)
	}

	return nil
}
