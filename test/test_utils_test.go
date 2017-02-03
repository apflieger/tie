package test

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/libgit2/git2go.v25"
	"testing"
	"time"
)

func TestDumbForCodeCoverage(t *testing.T) {
	repo := CreateTestRepo(false)
	defer CleanRepo(repo)

	WriteFile(repo, false, "foo", "bar")
	WriteFile(repo, true, "foo", "bar")
}

func TestRunOnRemote(t *testing.T) {
	testRun := false
	RunOnRemote(t, "NoParams", func(t *testing.T, repo, origin *git.Repository) {
		remote, _ := repo.Remotes.Lookup("origin")
		assert.Equal(t, origin.Path(), remote.Url())
		testRun = true
	})
	assert.True(t, testRun)
}

func TestCommit(t *testing.T) {
	RunOnRepo(t, "NoParams", func(t *testing.T, repo *git.Repository) {
		WriteFile(repo, true, "foo", "bar")
		// commit without params
		oid, err := Commit(repo, nil)

		assert.Nil(t, err)

		StatusClean(t, repo)

		// HEAD should point to the new commit
		head, _ := repo.Head()
		assert.Equal(t, 0, oid.Cmp(head.Target()))

		commit, _ := repo.LookupCommit(oid)

		// the new commit should have default signatures and message
		defaultSignature, _ := repo.DefaultSignature()
		assert.Equal(t, defaultSignature, commit.Author())
		assert.Equal(t, defaultSignature, commit.Committer())
		assert.Equal(t, "default message", commit.Message())
	})

	RunOnRepo(t, "RefParam", func(t *testing.T, repo *git.Repository) {
		headBefore, _ := repo.Head()

		// commit on a non existing ref
		oid, err := Commit(repo, &CommitParams{Refname: "refs/heads/test"})

		assert.Nil(t, err)

		head, _ := repo.Head()

		// HEAD shouldn't have changed
		assert.Equal(t, 0, headBefore.Target().Cmp(head.Target()))

		commit, _ := repo.LookupCommit(oid)

		// the ref should exist and point to the new commit with HEAD as parent
		ref, _ := repo.References.Lookup("refs/heads/test")
		assert.Equal(t, 0, oid.Cmp(ref.Target()))
		assert.Equal(t, 0, commit.Parent(0).Id().Cmp(head.Target()))

		defaultSignature, _ := repo.DefaultSignature()

		// the new commit should have default signatures and message
		assert.Equal(t, defaultSignature, commit.Author())
		assert.Equal(t, defaultSignature, commit.Committer())
		assert.Equal(t, "default message", commit.Message())
	})

	RunOnRepo(t, "CommitParams", func(t *testing.T, repo *git.Repository) {
		now := time.Now()
		author := &git.Signature{
			Name:  "Bob Morane",
			Email: "bob.morane@gmail.com",
			When:  now,
		}

		committer := &git.Signature{
			Name:  "Bill Ballantine",
			Email: "bill.ballantine@gmail.com",
			When:  time.Now().AddDate(0, 0, 1),
		}

		oid, err := Commit(repo, &CommitParams{
			Message:  "custom message",
			Author:   author,
			Commiter: committer,
		})

		assert.Nil(t, err)

		commit, _ := repo.LookupCommit(oid)

		// the new commit should have the given signatures and message
		assert.Equal(t, "Bob Morane", commit.Author().Name)
		assert.Equal(t, "bob.morane@gmail.com", commit.Author().Email)
		assert.Equal(t, now.Unix(), commit.Author().When.Unix())

		assert.Equal(t, "Bill Ballantine", commit.Committer().Name)
		assert.Equal(t, "bill.ballantine@gmail.com", commit.Committer().Email)
		assert.Equal(t, now.AddDate(0, 0, 1).Unix(), commit.Committer().When.Unix())

		assert.Equal(t, "custom message", commit.Message())
	})
}
