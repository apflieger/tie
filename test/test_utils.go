package test

import (
	"gopkg.in/libgit2/git2go.v25"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func RunOnRepo(t *testing.T, name string, test func(t *testing.T, repo *git.Repository)) {
	t.Run(name, func(t *testing.T) {
		repo := CreateTestRepo(false)
		defer CleanRepo(repo)
		test(t, repo)
	})
}

func RunOnRemote(t *testing.T, name string, test func(t *testing.T, repo, origin *git.Repository)) {
	t.Run(name, func(t *testing.T) {
		repo := CreateTestRepo(false)
		defer CleanRepo(repo)
		origin := CreateTestRepo(true)
		defer CleanRepo(origin)
		repo.Remotes.Create("origin", origin.Path())
		test(t, repo, origin)
	})
}

func CleanRepo(repo *git.Repository) {
	var path string

	if repo.IsBare() {
		path = repo.Path()
	} else {
		path = filepath.Join(repo.Path(), "..")
	}

	if len(path) > 5 { // Avoid dramatic rm -rf .
		os.RemoveAll(path)
	}
}

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

type CommitParams struct {
	Refname  string
	Author   *git.Signature
	Commiter *git.Signature
	Message  string
}

func Commit(repo *git.Repository, params *CommitParams) (*git.Oid, error) {
	defaultSignature, _ := repo.DefaultSignature()
	head, _ := repo.Head()
	defaultParams := &CommitParams{
		Refname:  head.Name(),
		Author:   defaultSignature,
		Commiter: defaultSignature,
		Message:  "default message",
	}

	if params == nil {
		params = defaultParams
	} else {
		if len(params.Refname) == 0 {
			params.Refname = defaultParams.Refname
		}
		if params.Author == nil {
			params.Author = defaultParams.Author
		}
		if params.Commiter == nil {
			params.Commiter = defaultParams.Commiter
		}
		if len(params.Message) == 0 {
			params.Message = defaultParams.Message
		}
	}

	index, _ := repo.Index()
	oid, _ := index.WriteTree()
	tree, _ := repo.LookupTree(oid)
	ref, err := repo.References.Lookup(params.Refname)
	// if the ref doesn't exist, lazy create it
	if err != nil {
		head, _ := repo.Head()
		ref, _ = repo.References.Create(params.Refname, head.Target(), false, "")
	}
	parent, _ := repo.LookupCommit(ref.Target())
	return repo.CreateCommit(params.Refname, params.Author, params.Commiter, params.Message, tree, parent)
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
