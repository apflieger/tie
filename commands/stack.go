package commands

import (
	"errors"
	"fmt"
	"github.com/apflieger/tie/core"
	"gopkg.in/libgit2/git2go.v25"
)

func StackCommand(repo *git.Repository) error {
	head, _ := repo.Head()
	tipName, notTip := core.TipName(head.Name())

	if notTip != nil {
		return errors.New("HEAD not on a tip. Only tips can be stacked.")
	}

	config, _ := repo.Config()
	baseRefName, _ := config.LookupString(fmt.Sprintf("tip.%v.base", tipName))

	base, _ := repo.References.Lookup(baseRefName)

	tail, _ := repo.References.Lookup(core.RefsTails+tipName)
	if !tail.Target().Equal(base.Target()) {
		return fmt.Errorf("Current tip '%v' is out of date with its base '%v'. Please upgrade\n", tipName, baseRefName)
	}

	base.SetTarget(head.Target(), "stack tip "+tipName)

	repo.References.CreateSymbolic("HEAD", baseRefName, true, "stack tip "+tipName)
	return nil
}