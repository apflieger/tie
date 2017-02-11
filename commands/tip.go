package commands

import (
	"fmt"
	"github.com/apflieger/tie/core"
	"gopkg.in/libgit2/git2go.v25"
)

func TipCreateCommand(repo *git.Repository, name, base string) error {
	var baseRef *git.Reference

	if len(base) == 0 {
		baseRef, _ = repo.Head()
	} else {
		var err error
		baseRef, err = repo.References.Lookup(base)
		if err != nil {
			return err
		}
		commit, _ := repo.LookupCommit(baseRef.Target())
		tree, _ := commit.Tree()
		repo.CheckoutTree(tree, &git.CheckoutOpts{Strategy: git.CheckoutSafe})
	}

	if remote, err := core.RemoteName(baseRef.Name()); err == nil {
		if _, err = repo.References.Lookup(fmt.Sprintf("refs/tips/%v/%v", remote, name)); err == nil {
			return fmt.Errorf("Failed to create tip \"%v\". A tip with that name already exists on %v.", name, remote)
		}
	}

	tipRef, err := repo.References.Create(core.RefsTips+name, baseRef.Target(), false, "tie tip create")

	if err != nil {
		return err
	}

	repo.References.Create(core.RefsTails+name, baseRef.Target(), false, "tie tip create")

	repo.References.CreateSymbolic("HEAD", tipRef.Name(), true, "tie tip create")

	config, _ := repo.Config()

	config.SetString(fmt.Sprintf("tip.%v.base", name), baseRef.Name())

	return nil
}
