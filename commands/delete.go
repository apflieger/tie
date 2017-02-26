package commands

import (
	"github.com/apflieger/tie/core"
	"gopkg.in/libgit2/git2go.v25"
	"log"
)

func DeleteCommand(repo *git.Repository, logger *log.Logger, refs []string) error {
	for _, ref := range refs {
		tipName, err := core.TipName(ref)
		if err != nil {
			return err
		}

		core.DeleteTip(repo, logger, tipName)
	}
	return nil
}
