package commands

import (
	"github.com/apflieger/tie/core"
	"gopkg.in/libgit2/git2go.v25"
	"log"
	"sort"
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
		sublist := []string{}
		for name, end := names.Next(); end == nil; name, end = names.Next() {
			sublist = append(sublist, name)
		}

		// sublist is a concatenation of file system based refs and packed ref.
		// Both of these listings are seperatly sorted but the concatenation
		// is not. This behaviour is not easily testable because libgit2 doesn't
		// allow to pack refs.
		sort.Strings(sublist)

		for _, name := range sublist {
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

	// These logic conditions required Karnaugh maps.
	// They don't mean to be easily understandable

	if (!remotes && tips) || (!branches && !remotes && all) {
		addGlob(core.RefsTips + "*")
	}

	if (remotes || all) && (tips || !branches) {
		addGlob(core.RefsRemoteTips + "*")
	}

	if (!remotes && branches) || (!tips && !remotes && all) {
		addGlob("refs/heads/*")
	}

	if (remotes || all) && (branches || !tips) {
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
