package core

import (
	"github.com/apflieger/tie/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/libgit2/git2go.v25"
	"testing"
)

func TestDwim(t *testing.T) {
	test.RunOnRepo(t, "SelectTip", func(t *testing.T, repo *git.Repository) {
		head, _ := repo.Head()
		repo.References.Create(RefsTips+"test", head.Target(), true, "")
		repo.References.Create(RefsRemoteTips+"origin/testorigin", head.Target(), true, "")
		repo.References.Create("refs/remotes/origin/master", head.Target(), true, "")
		repo.References.Create("refs/remotes/origin/testorigin", head.Target(), true, "")

		ref, err := Dwim(repo, "foo")
		assert.Nil(t, ref)
		if assert.NotNil(t, err) {
			assert.Equal(t, err.Error(), "No ref found for shorthand \"foo\"")
		}

		ref, err = Dwim(repo, "test")
		assert.Equal(t, RefsTips+"test", ref.Name())
		assert.Nil(t, err)

		ref, err = Dwim(repo, "tips/test")
		assert.Equal(t, RefsTips+"test", ref.Name())
		assert.Nil(t, err)

		ref, err = Dwim(repo, "testorigin")
		assert.Nil(t, ref)
		assert.NotNil(t, err)
		if assert.NotNil(t, err) {
			assert.Equal(t, err.Error(), "No ref found for shorthand \"testorigin\"")
		}

		ref, err = Dwim(repo, "origin/testorigin")
		assert.Equal(t, RefsRemoteTips+"origin/testorigin", ref.Name())
		assert.Nil(t, err)

		ref, err = Dwim(repo, "rtips/origin/testorigin")
		assert.Equal(t, RefsRemoteTips+"origin/testorigin", ref.Name())
		assert.Nil(t, err)

		ref, err = Dwim(repo, "origin/master")
		assert.Equal(t, "refs/remotes/origin/master", ref.Name())
		assert.Nil(t, err)

		ref, err = Dwim(repo, "remotes/origin/testorigin")
		assert.Equal(t, "refs/remotes/origin/testorigin", ref.Name())
		assert.Nil(t, err)
	})
}

func TestTipName(t *testing.T) {
	name, err := TipName("refs/remotes/origin/master")
	assert.Equal(t, "", name)
	assert.NotNil(t, err)

	name, err = TipName(RefsRemoteTips + "origin/test")
	assert.Equal(t, "", name)
	assert.NotNil(t, err)

	name, err = TipName(RefsTips + "test")
	assert.Equal(t, "test", name)
	assert.Nil(t, err)
}

func TestRemoteName(t *testing.T) {
	_, err := RemoteName("refs/heads/master")
	assert.NotNil(t, err)

	_, err = RemoteName(RefsTips + "otherTip")
	assert.NotNil(t, err)

	remote, err := RemoteName("refs/remotes/origin/master")
	assert.Nil(t, err)
	assert.Equal(t, "origin", remote)

	remote, err = RemoteName("refs/remotes/origin/features/work")
	assert.Nil(t, err)
	assert.Equal(t, "origin", remote)

	remote, err = RemoteName(RefsRemoteTips + "origin/work")
	assert.Nil(t, err)
	assert.Equal(t, "origin", remote)
}
