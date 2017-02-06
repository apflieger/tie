package core

import (
	"gopkg.in/libgit2/git2go.v25"
	"fmt"
)

func PrepareCommit(repo *git.Repository) (head *git.Reference, headCommit *git.Commit, treeToCommit *git.Tree) {
	head, _ = repo.Head()
	index, _ := repo.Index()
	treeObj, _ := index.WriteTree()
	treeToCommit, _ = repo.LookupTree(treeObj)
	headCommit, _ = repo.LookupCommit(head.Target())
	return head, headCommit, treeToCommit
}

func PushTip(repo *git.Repository, tip *git.Reference) {
	// lookup the remote corresponding to the base of the tip
	tipName, _ := TipName(tip.Name())
	config, _ := repo.Config()
	base, _ := config.LookupString(fmt.Sprintf("tip.%v.base", tipName))
	remoteName, _ := RemoteName(base)
	remote, _ := repo.Remotes.Lookup(remoteName)

	// push the tip on the remote
	remote.Push([]string{fmt.Sprintf("+%v:%v", tip.Name(), tip.Name())}, nil)

	// create the local remote ref
	repo.References.Create(fmt.Sprintf("refs/tips/%v/%v", remoteName, tipName), tip.Target(), true, "push tip")
}