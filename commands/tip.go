package commands

import (
	"fmt"
	"gopkg.in/libgit2/git2go.v25"
)

func TipCreateCommand(repo *git.Repository, name, base string) error {
	var baseRef *git.Reference

	if len(base) == 0 {
		baseRef, _ = repo.Head()
	} else {
		baseRef, _ = repo.References.Lookup(base)
		commit, _ := repo.LookupCommit(baseRef.Target())
		tree, _ := commit.Tree()
		repo.CheckoutTree(tree, &git.CheckoutOpts{Strategy: git.CheckoutSafe})
	}

	tipRef, err := repo.References.Create(fmt.Sprintf("refs/tips/local/%v", name), baseRef.Target(), false, "tie tip create")

	if err != nil {
		return err
	}

	repo.References.Create(fmt.Sprintf("refs/tails/%v", name), baseRef.Target(), false, "tie tip create")

	repo.References.CreateSymbolic("HEAD", tipRef.Name(), true, "tie tip create")

	config, _ := repo.Config()

	config.SetString(fmt.Sprintf("tip.%v.base", name), baseRef.Name())

	return nil
}
