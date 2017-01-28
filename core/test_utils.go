package core

import (
	"gopkg.in/libgit2/git2go.v25"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func CreateTestRepo(bare bool) *git.Repository {
	// Create a temp directory
	dir, _ := ioutil.TempDir("", "tie-test-")

	// git init
	repo, _ := git.InitRepository(dir, bare)
	if bare {
		return repo
	}

	config, _ := repo.Config()
	config.SetString("user.name", "tie-test")
	config.SetString("user.email", "tie@test.com")

	// create the first commit
	signature, _ := repo.DefaultSignature()
	index, _ := repo.Index()
	oid, _ := index.WriteTree()
	tree, _ := repo.LookupTree(oid)
	repo.CreateCommit("HEAD", signature, signature, "First commit", tree, []*git.Commit{}...)

	return repo
}

func Commit(repo *git.Repository, refname string) (*git.Oid, error) {
	index, _ := repo.Index()
	oid, _ := index.WriteTree()
	signature, _ := repo.DefaultSignature()
	tree, _ := repo.LookupTree(oid)
	head, _ := repo.Head()
	parent, _ := repo.LookupCommit(head.Target())
	return repo.CreateCommit(refname, signature, signature, "A new commit", tree, parent)
}

func WriteFile(repo *git.Repository, add bool, file string, lines ...string) {
	fileName := filepath.Join(repo.Path(), "..", file) // repo.Path() is the path of .git
	ioutil.WriteFile(fileName, []byte(strings.Join(lines, "\n")), 0644)

	if add {
		index, _ := repo.Index()
		index.AddByPath(file)
		index.Write()
	}
}
