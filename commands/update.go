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

func fetch(repo *git.Repository, context model.Context) error {
	head, _ := repo.Head()

	_, _, err := core.ExplodeRemoteRef(head.Name())
	if err == nil {
		// prevent update if current ref is a remote
		statusList, _ := repo.StatusList(nil)
		statusCount, _ := statusList.EntryCount()
		if statusCount != 0 {
			return fmt.Errorf("Status should be clean before updating on a remote ref: %v", head.Name())
		}
	}

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
				baselineCommit, _ := repo.LookupCommit(a)
				baselineTree, _ := baselineCommit.Tree()
				checkoutCommit, _ := repo.LookupCommit(b)
				checkoutTree, _ := checkoutCommit.Tree()
				repo.CheckoutTree(checkoutTree, &git.CheckoutOpts{
					Strategy: git.CheckoutSafe,
					Baseline: baselineTree,
				})
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
	remoteName, _, err := core.ExplodeRemoteRef(refname)
	if err == nil {
		return remoteName, nil
	}

	tipName, err := core.TipName(refname)
	if err != nil {
		return "", err
	}

	base, _ := config.LookupString(fmt.Sprintf("tip.%v.base", tipName))

	return remoteOf(base, config)
}

func UpdateCommand(repo *git.Repository, context model.Context) error {
	err := fetch(repo, context)
	if err != nil {
		return err
	}

	head, _ := repo.Head()
	tipName, err := core.TipName(head.Name())

	if err != nil {
		return nil
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

	tailRef.SetTarget(baseRef.Target(), "tie update")

	if rebase.OperationCount() > 0 {
		err = core.PushTip(repo, tipName, context)
	}

	rebase.Free()

	context.Logger.Printf("Upgraded current tip '%v' on top of '%v'\n", tipName, baseRefName)

	return nil
}

func UpdateAbortCommand(repo *git.Repository) error {
	rebaseOpts, _ := git.DefaultRebaseOptions()
	rebase, _ := repo.OpenRebase(rebaseOpts)
	rebase.Abort()
	rebase.Free()
	return nil
}

func UpdateContinueCommand(repo *git.Repository) error {
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

	repo.References.Create(core.RefsTails+tipName, onto, true, "tie update")

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
