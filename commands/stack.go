package commands

import (
	"errors"
	"fmt"
	"github.com/apflieger/tie/core"
	"github.com/apflieger/tie/model"
	"gopkg.in/libgit2/git2go.v25"
	"log"
)

func StackCommand(repo *git.Repository, context model.Context) error {
	head, _ := repo.Head()
	tipName, notTip := core.TipName(head.Name())

	if notTip != nil {
		return errors.New("HEAD not on a tip. Only tips can be stacked.")
	}

	config, _ := repo.Config()
	baseRefName, _ := config.LookupString(fmt.Sprintf("tip.%v.base", tipName))

	remoteName, pushRef, notRemote := core.ExplodeRemoteRef(baseRefName)
	base, _ := repo.References.Lookup(baseRefName)
	baseTipName, notOnLocalTip := core.TipName(baseRefName)

	// Allow to stack on local branch, remote branch or local tips only.
	if !(base.IsBranch() || core.IsBranch(pushRef) || notOnLocalTip == nil) {
		return fmt.Errorf("Cannot stack the current tip on his base '%v'. Tips can only be stacked on branches or local tips.", baseRefName)
	}

	tail, _ := repo.References.Lookup(core.RefsTails + tipName)
	if !tail.Target().Equal(base.Target()) {
		return fmt.Errorf("Current tip '%v' is out of date with its base '%v'. Please update\n", tipName, baseRefName)
	}

	if notRemote != nil {
		// Base and tail should have the same target for the stack to be allowed.
		// This guaranty the base to be fastforwarded
		base.SetTarget(head.Target(), "stack tip "+tipName) // base is not mutated, .Target will still return the previous one
		printStackInfo(repo, context.Logger, baseRefName, head.Name(), base.Target(), head.Target())
		if notOnLocalTip == nil {
			core.PushTip(repo, baseTipName, context)
		}
	} else {
		remote, _ := repo.Remotes.Lookup(remoteName)
		pushOptions := &git.PushOptions{
			RemoteCallbacks: context.RemoteCallbacks,
		}

		pushOptions.RemoteCallbacks.UpdateTipsCallback = func(refname string, a *git.Oid, b *git.Oid) git.ErrorCode {
			printStackInfo(repo, context.Logger, baseRefName, head.Name(), a, b)
			return git.ErrOk
		}

		// There's a vulnerability in case of a reverse fast forward reset on the remote.
		// In which case push will succeed, putting commits that have been removed back to the base.
		pushErr := remote.Push([]string{head.Name() + ":" + pushRef}, pushOptions)
		gitErr, isGitErr := pushErr.(*git.GitError)
		if isGitErr && gitErr.Code == git.ErrNonFastForward {
			return fmt.Errorf("Current tip '%v' is out of date with its base '%v'. Please update\n", tipName, baseRefName)
		}
	}

	repo.References.CreateSymbolic("HEAD", baseRefName, true, "stack tip "+tipName)

	// The tip has been successfully stacked. Now we can delete it.
	core.DeleteTip(repo, tipName, context)

	return nil
}

func printStackInfo(repo *git.Repository, logger *log.Logger, baseRefName, tipRefName string, baseOid, tipOid *git.Oid) {
	ahead, _, _ := repo.AheadBehind(tipOid, baseOid)
	plural := ""
	if ahead > 1 {
		plural = "s"
	}
	logger.Printf("%v <- %v (%v commit%v)\n",
		core.Shorthand(baseRefName),
		core.Shorthand(tipRefName),
		ahead,
		plural)
}
