package commands

import (
	"github.com/apflieger/tie/core"
	"gopkg.in/libgit2/git2go.v25"
	"log"
)

func SelectCommand(repo *git.Repository, shorthand string) error {
	rev, err := core.Dwim(repo, shorthand)

	if err != nil {
		return err
	}

	commit, _ := repo.LookupCommit(rev.Target())

	tree, _ := commit.Tree()

	err = repo.CheckoutTree(tree, &git.CheckoutOpts{Strategy: git.CheckoutSafe})
	if err != nil {
		return err
	}

	_, err = repo.References.CreateSymbolic("HEAD", rev.Name(), true, "Selected "+rev.Name())

	return err
}

func ListCommand(repo *git.Repository, logger *log.Logger, tips, branches, remotes bool) error {
	glob := core.RefsTips + "*"
	if remotes {
		glob = core.RefsRemoteTips + "*"
	}
	it, _ := repo.NewReferenceIteratorGlob(glob)
	names := it.Names()
	for name, end := names.Next(); end == nil; name, end = names.Next() {
		logger.Println(name)
	}
	return nil
}
