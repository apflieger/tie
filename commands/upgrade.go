package commands

import (
	"errors"
	"fmt"
	"github.com/apflieger/tie/core"
	"gopkg.in/libgit2/git2go.v25"
)

/**
Select the given refname. refname can be shorthand.
*/
func UpgradeCommand(repo *git.Repository) error {
	head, _ := repo.Head()
	tipName, err := core.TipName(head.Name())

	if err != nil {
		return errors.New("HEAD not on a tip. Only tips can be upgraded.")
	}

	config, _ := repo.Config()
	baseRefName, err := config.LookupString(fmt.Sprintf("tip.%v.base", tipName))

	if err != nil {
		return err
	}
	baseRef, _ := repo.References.Lookup(baseRefName)

	tailRef, err := repo.References.Lookup(fmt.Sprintf("refs/tails/%v", tipName))

	if err != nil {
		return err
	}

	annotatedHeadCommit, _ := repo.AnnotatedCommitFromRef(head)
	annotatedUpstreamCommit, _ := repo.AnnotatedCommitFromRef(tailRef)
	annotatedOntoCommit, _ := repo.AnnotatedCommitFromRef(baseRef)

	rebaseOpts, _ := git.DefaultRebaseOptions()

	rebase, _ := repo.InitRebase(annotatedHeadCommit, annotatedUpstreamCommit, annotatedOntoCommit, rebaseOpts)

	for operation, itErr := rebase.Next(); itErr == nil ; operation, itErr = rebase.Next() {
		index, _ := repo.Index()
		if index.HasConflicts() {
			return errors.New("Conflict while upgrading")
		}
		commit, _ := repo.LookupCommit(operation.Id)
		committer, _ := repo.DefaultSignature()
		rebase.Commit(operation.Id, commit.Author(), committer, commit.Message())
	}

	rebase.Finish()

	tailRef.SetTarget(baseRef.Target(), "tie upgrade")

	if rebase.OperationCount() > 0 {
		head, _ = repo.Head()
		core.PushTip(repo, head)
	}

	rebase.Free()

	return nil
}

func UpgradeAbortCommand(repo *git.Repository) error {
	rebaseOpts, _ := git.DefaultRebaseOptions()
	rebase, _ := repo.OpenRebase(rebaseOpts)
	rebase.Abort()
	rebase.Free()
	return nil
}