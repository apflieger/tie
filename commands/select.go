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
	if !tips && !branches && !remotes {
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

	if tips && !remotes {
		addGlob(core.RefsTips + "*")
	}

	if tips && remotes {
		addGlob(core.RefsRemoteTips + "*")
	}

	if branches && !remotes {
		addGlob("refs/heads/*")
	}

	if branches && remotes {
		addGlob("refs/remotes/*")
	}

	if remotes && !tips && !branches {
		addGlob(core.RefsRemoteTips + "*")
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
