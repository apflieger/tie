package commands

import (
	"errors"
	"fmt"
	"github.com/apflieger/tie/core"
	"github.com/apflieger/tie/env"
	"gopkg.in/libgit2/git2go.v25"
)

func StackCommand(repo *git.Repository) error {
	head, _ := repo.Head()
	tipName, notTip := core.TipName(head.Name())

	if notTip != nil {
		return errors.New("HEAD not on a tip. Only tips can be stacked.")
	}

	config, _ := repo.Config()
	baseRefName, _ := config.LookupString(fmt.Sprintf("tip.%v.base", tipName))

	remoteName, pushRef, notRemote := core.ExplodeRemoteRef(baseRefName)

	if notRemote == nil && !core.IsBranch(pushRef) {
		return fmt.Errorf("Cannot stack the current tip on his base '%v'. Tips can only be stacked on branches.", baseRefName)
	}

	base, _ := repo.References.Lookup(baseRefName)
	tail, _ := repo.References.Lookup(core.RefsTails + tipName)
	if !tail.Target().Equal(base.Target()) {
		return fmt.Errorf("Current tip '%v' is out of date with its base '%v'. Please upgrade\n", tipName, baseRefName)
	}

	if notRemote != nil {
		// Base and tail should have the same target for the stack to be allowed.
		// This guaranty the base to be fastforwarded
		base.SetTarget(head.Target(), "stack tip "+tipName)
	} else {
		remote, _ := repo.Remotes.Lookup(remoteName)
		pushOptions := &git.PushOptions{
			RemoteCallbacks: git.RemoteCallbacks{
				CredentialsCallback:      env.CredentialCallback,
				CertificateCheckCallback: env.CertificateCheckCallback,
			},
		}
		// There's a vulnerability in case of a reverse fast forward reset on the remote.
		// In which case push will succeed, putting commits that have been removed back to the base.
		pushErr := remote.Push([]string{head.Name() + ":" + pushRef}, pushOptions)
		gitErr, isGitErr := pushErr.(*git.GitError)
		if isGitErr && gitErr.Code == git.ErrNonFastForward {
			return fmt.Errorf("Current tip '%v' is out of date with its base '%v'. Please upgrade\n", tipName, baseRefName)
		}
	}

	repo.References.CreateSymbolic("HEAD", baseRefName, true, "stack tip "+tipName)
	return nil
}
