package core

import (
	"fmt"
	"github.com/apflieger/tie/env"
	"gopkg.in/libgit2/git2go.v25"
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
	pushOptions := &git.PushOptions{
		RemoteCallbacks: git.RemoteCallbacks{
			CredentialsCallback:      env.CredentialCallback,
			CertificateCheckCallback: env.CertificateCheckCallback,
		},
	}

	remote.Push([]string{fmt.Sprintf("+%v:%v", tip.Name(), tip.Name())}, pushOptions)

	// create the local remote ref
	repo.References.Create(fmt.Sprintf("refs/tips/%v/%v", remoteName, tipName), tip.Target(), true, "push tip")

	// handle branch compatibility mode
	compat, _ := config.LookupBool("tie.pushTipsAsBranches")

	if compat {
		remote.Push([]string{fmt.Sprintf("+%v:refs/heads/tips/%v", tip.Name(), tipName)}, pushOptions)
		repo.References.Create(fmt.Sprintf("refs/remotes/%v/tips/%v", remoteName, tipName), tip.Target(), true, "push tip")
	}
}
