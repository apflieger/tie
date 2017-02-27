package core

import (
	"bytes"
	"fmt"
	"github.com/apflieger/tie/model"
	"gopkg.in/libgit2/git2go.v25"
	"log"
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

func PushTip(repo *git.Repository, tipName string, context model.Context) error {
	// lookup the remote corresponding to the base of the tip
	config, _ := repo.Config()
	base, _ := config.LookupString(fmt.Sprintf("tip.%v.base", tipName))

	remoteName, _, notRemote := ExplodeRemoteRef(base)
	if notRemote != nil {
		return notRemote
	}

	remote, unknownRemote := repo.Remotes.Lookup(remoteName)
	if unknownRemote != nil {
		return unknownRemote
	}

	// push the tip on the remote

	tip, _ := repo.References.Lookup(RefsTips + tipName)
	refspecs := []string{fmt.Sprintf("+%v:%v", RefsTips+tipName, RefsTips+tipName)}

	// handle branch compatibility mode
	compat, _ := config.LookupBool("tie.pushTipsAsBranches")

	if compat {
		refspecs = append(refspecs, fmt.Sprintf("+%v:refs/heads/tips/%v", RefsTips+tipName, tipName))
	}

	pushOptions := &git.PushOptions{
		RemoteCallbacks: context.RemoteCallbacks,
	}

	pushErr := remote.Push(refspecs, pushOptions)

	if pushErr != nil {
		return pushErr
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

func DeleteTip(repo *git.Repository, tipName string, logger *log.Logger, context model.Context) {
	// Delete the tip locally
	tip, _ := repo.References.Lookup(RefsTips + tipName)
	tip.Delete()

	tail, _ := repo.References.Lookup(RefsTails + tipName)
	tail.Delete()

	config, _ := repo.Config()
	baseKey := fmt.Sprintf("tip.%v.base", tipName)
	base, _ := config.LookupString(baseKey)
	config.Delete(baseKey)

	// Delete the tip on the remote
	var pushErr error
	remoteName, _, err := ExplodeRemoteRef(base)
	if err == nil {
		remote, _ := repo.Remotes.Lookup(remoteName)
		pushOptions := &git.PushOptions{
			RemoteCallbacks: context.RemoteCallbacks,
		}
		refspecs := []string{":" + tip.Name()}

		compat, _ := config.LookupBool("tie.pushTipsAsBranches")
		if compat {
			refspecs = append(refspecs, ":refs/heads/tips/"+tipName)
		}

		pushErr = remote.Push(refspecs, pushOptions)

		if pushErr == nil {
			rtip, _ := repo.References.Lookup(RefsRemoteTips + remoteName + "/" + tipName)
			rtip.Delete()
		}
	}

	if pushErr != nil {
		logger.Println(pushErr.Error())
		logger.Printf("Tip '%v' has been deleted locally but not on %v.\n", tipName, remoteName)
	} else {
		logger.Printf("Deleted tip '%v'", tipName)
	}
}
