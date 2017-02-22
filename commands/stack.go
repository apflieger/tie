package commands

import (
	"gopkg.in/libgit2/git2go.v25"
	"fmt"
	"github.com/apflieger/tie/core"
)

func StackCommand(repo *git.Repository) error {
	head, _ := repo.Head()
	tipName, _ := core.TipName(head.Name())
	config, _ := repo.Config()
	baseRefName, _ := config.LookupString(fmt.Sprintf("tip.%v.base", tipName))

	base, _ := repo.References.Lookup(baseRefName)

	base.SetTarget(head.Target(), "stack tip "+tipName)

	repo.References.CreateSymbolic("HEAD", baseRefName, true, "stack tip "+tipName)
	return nil
}
