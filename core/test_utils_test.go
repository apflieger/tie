package core

import "testing"

func TestDumbForCodeCoverage(t *testing.T) {
	CreateTestRepo(true)

	repo := CreateTestRepo(false)

	WriteFile(repo, "foo", "bar")

	Commit(repo, "refs/heads/master")
}