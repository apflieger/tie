package commands

import (
	"github.com/apflieger/tie/core"
	"gopkg.in/libgit2/git2go.v25"
	"log"
)

func ListCommand(repo *git.Repository, logger *log.Logger, tips, branches, remotes, all bool) error {
	list := []string{}

	add := func(s string) {
		for _, e := range list {
			if e == s {
				return
			}
		}
		list = append(list, s)
	}

	addGlob := func(glob string) {
		it, _ := repo.NewReferenceIteratorGlob(glob)
		names := it.Names()
		for name, end := names.Next(); end == nil; name, end = names.Next() {
			add(name)
		}
	}

	// default listing
	if !all && !tips && !branches && !remotes {
		// display HEAD direct ref first
		head, _ := repo.Head()
		directRef, _ := head.Resolve()
		add(directRef.Name())

		// Then list the tips
		addGlob(core.RefsTips + "*")

		// finally list the commonly used bases
		config, _ := repo.Config()
		it, _ := config.NewIteratorGlob("tip.*.base")
		for entry, end := it.Next(); end == nil; entry, end = it.Next() {
			add(entry.Value)
		}
	}

	if all || (tips && !remotes) {
		addGlob(core.RefsTips + "*")
	}

	if all || (remotes && (tips || !branches)) {
		addGlob(core.RefsRemoteTips + "*")
	}

	if all || (branches && !remotes) {
		addGlob("refs/heads/*")
	}

	if all || (remotes && (branches || !tips)) {
		addGlob("refs/remotes/*")
	}

	head, _ := repo.Head()
	directRef, _ := head.Resolve()
	for _, ref := range list {
		prefix := "  "
		if ref == directRef.Name() {
			prefix = "* "
		}
		logger.Println(prefix + ref)
	}

	return nil
}
