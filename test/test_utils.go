package test

import (
	"bytes"
	"fmt"
	"github.com/apflieger/tie/model"
	"github.com/stretchr/testify/assert"
	"gopkg.in/libgit2/git2go.v25"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type TestContext struct {
	model.Context
	OutputBuffer *bytes.Buffer
}

/*
The given repo is attached on refs/heads/master, which has one single commit.
*/
func RunOnRepo(t *testing.T, name string, test func(t *testing.T, context TestContext, repo *git.Repository)) {
	t.Run(name, func(t *testing.T) {
		repo := CreateTestRepo(false)
		defer CleanRepo(repo)

		buffer := new(bytes.Buffer)
		context := TestContext{
			Context: model.Context{
				RemoteCallbacks: git.RemoteCallbacks{},
				Logger:          log.New(buffer, "", 0),
			},
			OutputBuffer: buffer,
		}
		test(t, context, repo)
	})
}

/*
The given repo is attached on refs/heads/master, which has one single commit.
repo has a configured origin remote which repository is passed as argument.
refs/heads/master is pushed to origin. refs/remotes/origin/master on repo is set.
*/
func RunOnRemote(t *testing.T, name string, test func(t *testing.T, context TestContext, repo, origin *git.Repository)) {
	t.Run(name, func(t *testing.T) {
		repo := CreateTestRepo(false)
		origin := CreateTestRepo(true)
		defer CleanRepo(repo)
		defer CleanRepo(origin)
		//fmt.Println("repo: " + repo.Workdir())
		//fmt.Println("remote: " + origin.Path())

		remote, _ := repo.Remotes.Create("origin", origin.Path())
		remote.Push([]string{"+refs/heads/master"}, nil)

		buffer := new(bytes.Buffer)
		context := TestContext{
			Context: model.Context{
				RemoteCallbacks: git.RemoteCallbacks{},
				Logger:          log.New(buffer, "", 0),
			},
			OutputBuffer: buffer,
		}
		test(t, context, repo, origin)
	})
}

func CleanRepo(repo *git.Repository) {
	var path string

	if repo.IsBare() {
		path = repo.Path()
	} else {
		path = repo.Workdir()
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

	reset := false

	if params == nil {
		params = defaultParams
	} else {
		if len(params.Refname) == 0 {
			params.Refname = defaultParams.Refname
		} else {
			reset = true
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
	createdOid, err := repo.CreateCommit(params.Refname, params.Author, params.Commiter, params.Message, tree, parent)

	if reset {
		repo.ResetToCommit(parent, git.ResetHard, &git.CheckoutOpts{Strategy: git.CheckoutForce})
	}

	return createdOid, err
}

func WriteFile(repo *git.Repository, add bool, file string, lines ...string) {
	fileName := filepath.Join(repo.Workdir(), file)
	ioutil.WriteFile(fileName, []byte(strings.Join(lines, "\n")), 0644)

	if add {
		index, _ := repo.Index()
		index.AddByPath(file)
		index.Write()
	}
}

func StatusClean(t *testing.T, repo *git.Repository) bool {
	statusList, _ := repo.StatusList(
		&git.StatusOptions{
			Show:     git.StatusShowIndexAndWorkdir,
			Flags:    git.StatusOptIncludeUntracked,
			Pathspec: nil,
		})
	statusCount, _ := statusList.EntryCount()
	return assert.Equal(t, 0, statusCount, "status not clean")
}

func CreateTip(repo *git.Repository, tipName, base string, slct bool) {
	head, _ := repo.Head()
	repo.References.Create("refs/tips/"+tipName, head.Target(), false, "")
	repo.References.Create("refs/tails/"+tipName, head.Target(), false, "")
	config, _ := repo.Config()
	config.SetString(fmt.Sprintf("tip.%v.base", tipName), base)

	if slct {
		repo.References.CreateSymbolic("HEAD", "refs/tips/"+tipName, true, "")
	}
}
