package commands

import (
	"gopkg.in/libgit2/git2go.v25"
)

func AmendCommand(repo *git.Repository) error {
	head, _ := repo.Head()
	index, _ := repo.Index()
	treeObj, _ := index.WriteTree()
	tree, _ := repo.LookupTree(treeObj)
	headCommit, _ := repo.LookupCommit(head.Target())
	committer, _ := repo.DefaultSignature()
	_, err := headCommit.Amend(head.Name(), headCommit.Author(), committer, headCommit.Message(), tree)

	return err
}
