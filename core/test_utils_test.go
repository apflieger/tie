package core

import "testing"

func TestDumbForCodeCoverage(t *testing.T) {
	bareRepo := CreateTestRepo(true)
	defer CleanRepo(bareRepo)

	repo := CreateTestRepo(false)
	defer CleanRepo(repo)

	WriteFile(repo, false, "foo", "bar")
	WriteFile(repo, true, "foo", "bar")

	Commit(repo, "refs/heads/master")
}
