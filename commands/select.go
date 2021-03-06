package commands

import (
	"github.com/apflieger/tie/core"
	"gopkg.in/libgit2/git2go.v25"
)

func SelectCommand(repo *git.Repository, shorthand string) error {
	// lookup the reference
	rev, err := core.Dwim(repo, shorthand)

	if err != nil {
		return err
	}

	// checkout the index and the working tree
	commit, _ := repo.LookupCommit(rev.Target())

	tree, _ := commit.Tree()

	err = repo.CheckoutTree(tree, &git.CheckoutOpts{Strategy: git.CheckoutSafe})
	if err != nil {
		return err
	}

	// set HEAD
	_, err = repo.References.CreateSymbolic("HEAD", rev.Name(), true, "Selected "+rev.Name())

	return err
}
