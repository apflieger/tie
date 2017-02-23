package core

import (
	"bytes"
	"fmt"
	"github.com/apflieger/tie/env"
	"gopkg.in/libgit2/git2go.v25"
	"strings"
)

func PrepareCommit(repo *git.Repository) (head *git.Reference, headCommit *git.Commit, treeToCommit *git.Tree) {
	head, _ = repo.Head()
	index, _ := repo.Index()
	treeObj, _ := index.WriteTree()
	treeToCommit, _ = repo.LookupTree(treeObj)
	headCommit, _ = repo.LookupCommit(head.Target())
	return head, headCommit, treeToCommit
}

func PushTip(repo *git.Repository, tipName string) error {
	// lookup the remote corresponding to the base of the tip
	config, _ := repo.Config()
	base, _ := config.LookupString(fmt.Sprintf("tip.%v.base", tipName))
	remoteName, _, err := ExplodeRemoteRef(base)

	if err != nil {
		return err
	}

	remote, err := repo.Remotes.Lookup(remoteName)

	if err != nil {
		return err
	}

	// push the tip on the remote
	pushOptions := &git.PushOptions{
		RemoteCallbacks: git.RemoteCallbacks{
			CredentialsCallback:      env.CredentialCallback,
			CertificateCheckCallback: env.CertificateCheckCallback,
		},
	}

	tip, _ := repo.References.Lookup(RefsTips + tipName)
	refspecs := []string{fmt.Sprintf("+%v:%v", RefsTips+tipName, RefsTips+tipName)}

	// handle branch compatibility mode
	compat, _ := config.LookupBool("tie.pushTipsAsBranches")

	if compat {
		refspecs = append(refspecs, fmt.Sprintf("+%v:refs/heads/tips/%v", RefsTips+tipName, tipName))
	}

	err = remote.Push(refspecs, pushOptions)

	if err != nil {
		return err
	}

	repo.References.Create(RefsRemoteTips+remoteName+"/"+tipName, tip.Target(), true, "push tip")

	return nil
}

// Removes comments (#) and empty lines before/after the content
func FormatCommitMessage(s string) string {
	if s == "" {
		return ""
	}
	trimmed := strings.Trim(s, " \n")
	lines := strings.Split(trimmed, "\n")
	buffer := new(bytes.Buffer)
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimLeft(line, " "), "#") {
			continue
		}
		buffer.WriteString(line + "\n")
	}
	return buffer.String()
}
