package commands

import "gopkg.in/libgit2/git2go.v25"

func CommitCommand(repo *git.Repository, args []string) error {
	head, _ := repo.Head()
	signature, _ := repo.DefaultSignature()
	index, _ := repo.Index()
	treeObj, _ := index.WriteTree()
	tree, _ := repo.LookupTree(treeObj)
	parent, _ := repo.LookupCommit(head.Target())
	repo.CreateCommit(head.Name(), signature, signature, "", tree, parent)
	return nil
}
