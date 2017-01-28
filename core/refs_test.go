package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDwim(t *testing.T) {
	repo := CreateTestRepo(false)
	head, _ := repo.Head()
	repo.References.Create("refs/tips/local/test", head.Target(), true, "")
	repo.References.Create("refs/tips/origin/testorigin", head.Target(), true, "")
	repo.References.Create("refs/remotes/origin/master", head.Target(), true, "")
	repo.References.Create("refs/remotes/origin/testorigin", head.Target(), true, "")

	ref, err := Dwim(repo, "foo")
	assert.Nil(t, ref)
	if assert.NotNil(t, err) {
		assert.Equal(t, err.Error(), "No ref found for shorthand \"foo\"")
	}

	ref, err = Dwim(repo, "test")
	assert.Equal(t, "refs/tips/local/test", ref.Name())
	assert.Nil(t, err)

	ref, err = Dwim(repo, "local/test")
	assert.Equal(t, "refs/tips/local/test", ref.Name())
	assert.Nil(t, err)

	ref, err = Dwim(repo, "tips/local/test")
	assert.Equal(t, "refs/tips/local/test", ref.Name())
	assert.Nil(t, err)

	ref, err = Dwim(repo, "testorigin")
	assert.Nil(t, ref)
	assert.NotNil(t, err)
	if assert.NotNil(t, err) {
		assert.Equal(t, err.Error(), "No ref found for shorthand \"testorigin\"")
	}

	ref, err = Dwim(repo, "origin/testorigin")
	assert.Equal(t, "refs/tips/origin/testorigin", ref.Name())
	assert.Nil(t, err)

	ref, err = Dwim(repo, "origin/master")
	assert.Equal(t, "refs/remotes/origin/master", ref.Name())
	assert.Nil(t, err)

	ref, err = Dwim(repo, "remotes/origin/testorigin")
	assert.Equal(t, "refs/remotes/origin/testorigin", ref.Name())
	assert.Nil(t, err)
}

func TestTipName(t *testing.T) {
	name, err := TipName("refs/remotes/origin/master")
	assert.Equal(t, "", name)
	assert.NotNil(t, err)

	name, err = TipName("refs/tips/origin/test")
	assert.Equal(t, "", name)
	assert.NotNil(t, err)

	name, err = TipName("refs/tips/local/test")
	assert.Equal(t, "test", name)
	assert.Nil(t, err)
}
