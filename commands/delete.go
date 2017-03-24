package commands

import (
	"fmt"
	"github.com/apflieger/tie/core"
	"github.com/apflieger/tie/model"
	"gopkg.in/libgit2/git2go.v25"
)

func DeleteCommand(repo *git.Repository, stacked bool, refs []string, context model.Context) error {
	if stacked {
		it, _ := repo.NewReferenceIteratorGlob(core.RefsTips + "*")
		refs = []string{}
		for tip, end := it.Next(); end == nil; tip, end = it.Next() {
			tipName, _ := core.TipName(tip.Name())
			config, _ := repo.Config()
			base, _ := config.LookupString(fmt.Sprintf("tip.%v.base", tipName))
			baseRef, _ := repo.References.Lookup(base)
			isDescendant, _ := repo.DescendantOf(baseRef.Target(), tip.Target())
			if isDescendant || baseRef.Target().Equal(tip.Target()) {
				refs = append(refs, tip.Name())
			}
		}
	}

	for _, ref := range refs {
		tipName, err := core.TipName(ref)
		if err != nil {
			return err
		}

		// If we are deleting the tip that is currently selected
		// select the base before deletion.
		head, _ := repo.Head()
		if head.Name() == ref {
			tipName, _ := core.TipName(ref)
			config, _ := repo.Config()
			base, _ := config.LookupString(fmt.Sprintf("tip.%v.base", tipName))
			baseRef, _ := repo.References.Lookup(base)

			// checkout the index and the working tree
			commit, _ := repo.LookupCommit(baseRef.Target())

			tree, _ := commit.Tree()

			repo.CheckoutTree(tree, &git.CheckoutOpts{Strategy: git.CheckoutSafe})

			// set HEAD
			_, err = repo.References.CreateSymbolic("HEAD", base, true,
				fmt.Sprintf("Select %v caused by deleting currently selected %v", base, ref))
		}

		core.DeleteTip(repo, tipName, context)
	}
	return nil
}
