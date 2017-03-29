package core

import (
	"fmt"
	"github.com/apflieger/tie/model"
	"gopkg.in/libgit2/git2go.v25"
)

func Fetch(repo *git.Repository, context model.Context) error {
	head, _ := repo.Head()
	config, _ := repo.Config()
	remoteName, err := remoteOf(head.Name(), config)
	if err != nil {
		return nil
	}
	remote, _ := repo.Remotes.Lookup(remoteName)

	remoteCallbacks := git.RemoteCallbacks{
		CredentialsCallback:      context.RemoteCallbacks.CredentialsCallback,
		CertificateCheckCallback: context.RemoteCallbacks.CertificateCheckCallback,
		UpdateTipsCallback: func(refname string, a *git.Oid, b *git.Oid) git.ErrorCode {
			if refname == head.Name() {
				commit, _ := repo.LookupCommit(b)
				tree, _ := commit.Tree()
				repo.CheckoutTree(tree, &git.CheckoutOpts{Strategy: git.CheckoutSafe})
			}
			var message string
			if a.IsZero() {
				message = "Created %v\n"
			} else if b.IsZero() {
				message = "Deleted %v\n"
			} else {
				message = "Updated %v\n"
			}
			context.Logger.Printf(message, refname)
			return git.ErrOk
		},
	}

	fetchOptions := &git.FetchOptions{
		Prune:           git.FetchPruneOn,
		RemoteCallbacks: remoteCallbacks,
	}
	refspecs, _ := remote.FetchRefspecs()
	return remote.Fetch(refspecs, fetchOptions, "")
}

func remoteOf(refname string, config *git.Config) (string, error) {
	remoteName, _, err := ExplodeRemoteRef(refname)
	if err == nil {
		return remoteName, nil
	}

	tipName, err := TipName(refname)
	if err != nil {
		return "", err
	}

	base, _ := config.LookupString(fmt.Sprintf("tip.%v.base", tipName))

	return remoteOf(base, config)
}
