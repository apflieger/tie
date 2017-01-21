package core

import (
	"io/ioutil"
	"gopkg.in/libgit2/git2go.v25"
)

func CreateTestRepo() *git.Repository {
	// Create a temp directory
	dir, _ := ioutil.TempDir("", "tie-test-")

	// git init
	repo, _ := git.InitRepository(dir, false)

	// create the first commit
	signature, _ := repo.DefaultSignature()
	index, _ := repo.Index()
	oid, _ := index.WriteTree()
	tree, _ := repo.LookupTree(oid)
	repo.CreateCommit("HEAD", signature, signature, "First commit", tree, []*git.Commit{}...)

	return repo
}

func Commit(repo *git.Repository, refname string) (*git.Oid, error) {
	signature, _ := repo.DefaultSignature()
	index, _ := repo.Index()
	oid, _ := index.WriteTree()
	tree, _ := repo.LookupTree(oid)
	head, _ := repo.Head()
	parent, _ := repo.LookupCommit(head.Target())
	return repo.CreateCommit(refname, signature, signature, "A new commit", tree, parent)
}
