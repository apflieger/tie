package commands

import (
	"fmt"
	"github.com/apflieger/tie/core"
	"gopkg.in/libgit2/git2go.v25"
	"log"
)

func DeleteCommand(repo *git.Repository, logger *log.Logger, refs []string) error {
	config, _ := repo.Config()
	for _, ref := range refs {
		tipName, err := core.TipName(ref)
		if err != nil {
			return err
		}

		tip, _ := repo.References.Lookup(ref)
		tip.Delete()

		tail, _ := repo.References.Lookup(core.RefsTails + tipName)
		tail.Delete()

		baseKey := fmt.Sprintf("tip.%v.base", tipName)
		base, _ := config.LookupString(baseKey)
		config.Delete(baseKey)

		git.ShortenOids([]*git.Oid{tip.Target(), tail.Target()}, 8)
		logger.Printf("Deleted tip %v (%v..%v) based on %v", tipName, tail.Target().String(), tip.Target().String(), base)
	}
	return nil
}
