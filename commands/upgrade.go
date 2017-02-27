package commands

import (
	"errors"
	"fmt"
	"github.com/apflieger/tie/core"
	"github.com/apflieger/tie/model"
	"gopkg.in/libgit2/git2go.v25"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// Identical to https://github.com/libgit2/libgit2/blob/master/src/rebase.c#L28
const (
	rebaseMergeDir = "rebase-merge"
	ontoNameFile   = "onto"
	headNameFile   = "head-name"
)

/**
Select the given refname. refname can be shorthand.
*/
func UpgradeCommand(repo *git.Repository, context model.Context) error {
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

	rebase, err := repo.InitRebase(annotatedHeadCommit, annotatedUpstreamCommit, annotatedOntoCommit, rebaseOpts)

	if err != nil {
		return err
	}

	err = iterate(repo, rebase)

	if err != nil {
		return err
	}

	tailRef.SetTarget(baseRef.Target(), "tie upgrade")

	if rebase.OperationCount() > 0 {
		core.PushTip(repo, tipName, context)
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

func UpgradeContinueCommand(repo *git.Repository) error {
	rebaseOpts, _ := git.DefaultRebaseOptions()
	rebase, _ := repo.OpenRebase(rebaseOpts)
	currentOperationIndex, _ := rebase.CurrentOperationIndex()
	commit(repo, rebase, rebase.OperationAt(currentOperationIndex))
	err := iterate(repo, rebase)
	rebase.Free()
	return err
}

func iterate(repo *git.Repository, rebase *git.Rebase) error {
	for operation, itErr := rebase.Next(); itErr == nil; operation, itErr = rebase.Next() {
		err := commit(repo, rebase, operation)
		if err != nil {
			return err
		}
	}

	ontoFilePath := filepath.Join(repo.Path(), rebaseMergeDir, ontoNameFile)
	headNameFilepath := filepath.Join(repo.Path(), rebaseMergeDir, headNameFile)
	bytes, _ := ioutil.ReadFile(ontoFilePath)
	onto, _ := git.NewOid(strings.Trim(string(bytes), "\n"))
	bytes, _ = ioutil.ReadFile(headNameFilepath)
	headName := strings.Trim(string(bytes), "\n")
	tipName, _ := core.TipName(headName)

	repo.References.Create(core.RefsTails+tipName, onto, true, "tie upgrade")

	return rebase.Finish()
}

func commit(repo *git.Repository, rebase *git.Rebase, operation *git.RebaseOperation) error {
	index, _ := repo.Index()
	if index.HasConflicts() {
		return errors.New("Conflict while upgrading")
	}
	commit, _ := repo.LookupCommit(operation.Id)
	committer, _ := repo.DefaultSignature()
	return rebase.Commit(operation.Id, commit.Author(), committer, commit.Message())
}
