package core

import (
	"github.com/apflieger/tie/model"
	"github.com/apflieger/tie/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/libgit2/git2go.v25"
	"testing"
)

func TestDwim(t *testing.T) {
	test.RunOnRepo(t, "SelectTip", func(t *testing.T, context model.Context, repo *git.Repository) {
		head, _ := repo.Head()
		repo.References.Create(RefsTips+"test", head.Target(), true, "")
		repo.References.Create(RefsRemoteTips+"origin/testorigin", head.Target(), true, "")
		repo.References.Create("refs/remotes/origin/master", head.Target(), true, "")
		repo.References.Create("refs/remotes/origin/testorigin", head.Target(), true, "")

		ref, err := Dwim(repo, "foo")
		assert.Nil(t, ref)
		if assert.NotNil(t, err) {
			assert.Equal(t, err.Error(), "No ref found for shorthand 'foo'")
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
			assert.Equal(t, err.Error(), "No ref found for shorthand 'testorigin'")
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

func TestShorthand(t *testing.T) {
	assert.Equal(t, "", Shorthand(""))
	// Local branches
	assert.Equal(t, "master", Shorthand("refs/heads/master"))
	assert.Equal(t, "work/local", Shorthand("refs/heads/work/local"))
	// Remote branches
	assert.Equal(t, "origin/master", Shorthand("refs/remotes/origin/master"))
	assert.Equal(t, "somewhere/master", Shorthand("refs/remotes/somewhere/master"))
	assert.Equal(t, "origin/work/other", Shorthand("refs/remotes/origin/work/other"))
	// LKocal tips
	assert.Equal(t, "a_tip", Shorthand(RefsTips+"a_tip"))
	// Remote tips
	assert.Equal(t, "origin/a_tip", Shorthand(RefsRemoteTips+"origin/a_tip"))
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

func TestRefName(t *testing.T) {
	var err error
	var refName string

	_, err = RefName("")
	assert.NotNil(t, err)

	_, err = RefName("HEAD")
	assert.NotNil(t, err)

	refName, err = RefName("refs/heads/master")
	assert.Equal(t, "master", refName)

	refName, err = RefName("refs/heads/work/mine")
	assert.Equal(t, "mine", refName)

	refName, err = RefName("refs/remotes/origin/master")
	assert.Equal(t, "master", refName)

	refName, err = RefName("refs/remotes/origin/work/yours")
	assert.Equal(t, "yours", refName)

	refName, err = RefName(RefsTips + "my_tip")
	assert.Equal(t, "my_tip", refName)

	refName, err = RefName(RefsTips + "tmp/my_tip")
	assert.Equal(t, "my_tip", refName)

	refName, err = RefName(RefsRemoteTips + "origin/my_tip")
	assert.Equal(t, "my_tip", refName)
}

func TestExplodeRemoteRef(t *testing.T) {
	_, _, err := ExplodeRemoteRef("")
	assert.NotNil(t, err)

	_, _, err = ExplodeRemoteRef("refs/heads/master")
	assert.NotNil(t, err)

	_, _, err = ExplodeRemoteRef(RefsTips + "otherTip")
	assert.NotNil(t, err)

	_, _, err = ExplodeRemoteRef("origin/master")
	assert.NotNil(t, err)

	remote, localRef, err := ExplodeRemoteRef("refs/remotes/origin/master")
	assert.Nil(t, err)
	assert.Equal(t, "origin", remote)
	assert.Equal(t, "refs/heads/master", localRef)

	remote, localRef, err = ExplodeRemoteRef("refs/remotes/origin/features/work")
	assert.Nil(t, err)
	assert.Equal(t, "origin", remote)
	assert.Equal(t, "refs/heads/features/work", localRef)

	remote, localRef, err = ExplodeRemoteRef("refs/remotes/somemplace/features/work")
	assert.Nil(t, err)
	assert.Equal(t, "somemplace", remote)
	assert.Equal(t, "refs/heads/features/work", localRef)

	remote, localRef, err = ExplodeRemoteRef(RefsRemoteTips + "origin/work")
	assert.Nil(t, err)
	assert.Equal(t, "origin", remote)
	assert.Equal(t, RefsTips+"work", localRef)

	remote, localRef, err = ExplodeRemoteRef(RefsRemoteTips + "somewhere/work/mine")
	assert.Nil(t, err)
	assert.Equal(t, "somewhere", remote)
	assert.Equal(t, RefsTips+"work/mine", localRef)
}

func TestMatchingBranchfName(t *testing.T) {
	assert.False(t, IsBranch(""))
	assert.True(t, IsBranch("refs/heads/master"))
	assert.True(t, IsBranch("refs/heads/features/work"))
	assert.False(t, IsBranch("refs/remotes/origin/master"))
	assert.False(t, IsBranch("origin/master"))
	assert.False(t, IsBranch(RefsTips+"test"))
	assert.False(t, IsBranch(RefsRemoteTips+"test"))
	assert.False(t, IsBranch(RefsTails+"test"))
}
