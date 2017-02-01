package commands

import (
	"fmt"
	"gopkg.in/libgit2/git2go.v25"
	"regexp"
	"strings"
)

func CommitCommand(repo *git.Repository, message string) error {
	// like git commit
	head, _ := repo.Head()
	signature, _ := repo.DefaultSignature()
	index, _ := repo.Index()
	treeObj, _ := index.WriteTree()
	tree, _ := repo.LookupTree(treeObj)
	parent, _ := repo.LookupCommit(head.Target())
	repo.CreateCommit(head.Name(), signature, signature, message, tree, parent)

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
