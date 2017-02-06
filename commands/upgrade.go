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

	tailRef, err := repo.References.Lookup(fmt.Sprintf("refs/tails/%v", tipName))

	if err != nil {
		return err
	}

	tailOid := tailRef.Target()

	cpCommitRange := []*git.Commit{}

	for loopCommit, _ := repo.LookupCommit(head.Target()); loopCommit.Id().Cmp(tailOid) != 0; loopCommit = loopCommit.Parent(0) {
		cpCommitRange = append(cpCommitRange, loopCommit)
	}

	for i := len(cpCommitRange)/2 - 1; i >= 0; i-- {
		opp := len(cpCommitRange) - 1 - i
		cpCommitRange[i], cpCommitRange[opp] = cpCommitRange[opp], cpCommitRange[i]
	}

	baseRef, _ := repo.References.Lookup(baseRefName)
	baseCommit, _ := repo.LookupCommit(baseRef.Target())
	repo.ResetToCommit(baseCommit, git.ResetHard, &git.CheckoutOpts{Strategy: git.CheckoutSafe})

	headCommit := baseCommit
	for _, cpCommit := range cpCommitRange {
		cherrypickOptions, _ := git.DefaultCherrypickOptions()
		err := repo.Cherrypick(cpCommit, cherrypickOptions)
		if err != nil {
			return err
		}

		index, _ := repo.Index()
		index.Write()
		treeObj, _ := index.WriteTree()
		tree, _ := repo.LookupTree(treeObj)
		author := cpCommit.Author()
		committer := cpCommit.Committer()
		newCommitOid, _ := repo.CreateCommit(head.Name(), author, committer, cpCommit.Message(), tree, headCommit)
		newCommit, _ := repo.LookupCommit(newCommitOid)
		headCommit = newCommit
	}

	tailRef.SetTarget(baseCommit.Id(), "tie update")

	repo.StateCleanup()

	return nil
}
