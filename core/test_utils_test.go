package core

import "testing"

func TestDumbForCodeCoverage(t *testing.T) {
	CreateTestRepo(true)

	repo := CreateTestRepo(false)

	WriteFile(repo, false, "foo", "bar")
	WriteFile(repo, true, "foo", "bar")

	Commit(repo, "refs/heads/master")
}
