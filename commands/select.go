package commands

import (
	"github.com/apflieger/tie/core"
	"gopkg.in/libgit2/git2go.v25"
)

/**
Select the given refname. refname can be shorthand.
*/
func SelectCommand(repo *git.Repository, args []string) error {
	refname := args[0]
	rev, err := core.Dwim(repo, refname)

	if err != nil {
		return err
	}

	commit, err := repo.LookupCommit(rev.Target())
	if err != nil {
		return err
	}

	tree, err := commit.Tree()
	if err != nil {
		return err
	}

	err = repo.CheckoutTree(tree, &git.CheckoutOpts{Strategy: git.CheckoutSafe})
	if err != nil {
		return err
	}

	_, err = repo.References.CreateSymbolic("HEAD", rev.Name(), true, "Selected "+rev.Name())

	return err
}
