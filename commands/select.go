package commands

import "gopkg.in/libgit2/git2go.v25"

func SelectCommand(repo *git.Repository, refname string) error {
	rev, err := repo.Revparse(refname)

	if err != nil {
		return err
	}

	commit, err := rev.From().AsCommit()
	if err != nil {
		return err
	}

	tree, err := commit.Tree()
	if err != nil {
		return err
	}

	err = repo.CheckoutTree(tree, &git.CheckoutOpts{Strategy: git.CheckoutSafe})
	if err != nil {
		return err
	}

	_, err = repo.References.CreateSymbolic("HEAD", refname, true, "Selected "+refname)

	return err
}
