package commands

import (
	"github.com/apflieger/tie/core"
	"github.com/apflieger/tie/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/libgit2/git2go.v25"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestUpdateCommand(t *testing.T) {

	t.Run("Update", func(t *testing.T) {

		test.RunOnThreeRepos(t, "UpdateCurrentBranch", func(t *testing.T, context test.TestContext, repo, origin, another *git.Repository) {
			// HEAD is on origin/master
			// Asserting test setup
			head, _ := repo.Head()
			assert.Equal(t, "refs/remotes/origin/master", head.Name())

			// Commit on master from "another" and push it to origin
			test.WriteFile(another, true, "foo", "foobar")
			oid, _ := test.Commit(another, nil)
			remote, _ := another.Remotes.Lookup("origin")
			remote.Push([]string{"+refs/heads/master"}, nil)

			UpdateCommand(repo, context.Context)

			originMaster, _ := repo.References.Lookup("refs/remotes/origin/master")
			// Local origin/master should be on the commit
			assert.True(t, oid.Equal(originMaster.Target()))
			// File "foo" should exist
			_, err := os.Stat(filepath.Join(repo.Workdir(), "foo"))
			assert.Nil(t, err)

			test.StatusClean(t, repo)

			// Output should be...
			assert.Equal(t, "Updated refs/remotes/origin/master\n", context.OutputBuffer.String())
		})

		test.RunOnRemote(t, "OnTip", func(t *testing.T, context test.TestContext, repo, origin *git.Repository) {
			test.CreateTip(repo, "test", "refs/remotes/origin/master", true)

			err := UpdateCommand(repo, context.Context)
			assert.Nil(t, err)
		})

		test.RunOnRemote(t, "OnTipTip", func(t *testing.T, context test.TestContext, repo, origin *git.Repository) {
			test.CreateTip(repo, "test", "refs/remotes/origin/master", false)
			test.CreateTip(repo, "test2", core.RefsTips+"test", true)

			err := UpdateCommand(repo, context.Context)
			assert.Nil(t, err)
		})

		test.RunOnRemote(t, "OnTipLocalBranch", func(t *testing.T, context test.TestContext, repo, origin *git.Repository) {
			test.CreateTip(repo, "test", "refs/heads/master", false)
			test.CreateTip(repo, "test2", core.RefsTips+"test", true)

			err := UpdateCommand(repo, context.Context)
			assert.Nil(t, err)
		})

		test.RunOnThreeRepos(t, "MultipleRefs", func(t *testing.T, context test.TestContext, repo, origin, another *git.Repository) {
			remote, _ := another.Remotes.Lookup("origin")

			// Create two branches on the remote
			remote.Push([]string{
				"refs/heads/master:refs/heads/another_branch",
				"refs/heads/master:refs/heads/to_be_deleted",
			}, nil)

			UpdateCommand(repo, context.Context)

			// The two branches should have been fetched
			_, err := repo.References.Lookup("refs/remotes/origin/another_branch")
			assert.Nil(t, err)

			_, err = repo.References.Lookup("refs/remotes/origin/to_be_deleted")
			assert.Nil(t, err)

			// Output should be...
			assert.Equal(t,
				"Created refs/remotes/origin/another_branch\n"+
					"Created refs/remotes/origin/to_be_deleted\n",
				context.OutputBuffer.String())

			// Delete a branch
			remote.Push([]string{":refs/heads/to_be_deleted"}, nil)

			// Reset the output buffer and rerun update
			context.OutputBuffer.Reset()
			UpdateCommand(repo, context.Context)

			// The branch should be pruned
			_, err = repo.References.Lookup("refs/remotes/origin/to_be_deleted")
			assert.NotNil(t, err)

			// Output should be...
			assert.Equal(t,
				"Deleted refs/remotes/origin/to_be_deleted\n",
				context.OutputBuffer.String())
		})
	})

	t.Run("Upgrade", func(t *testing.T) {

		test.RunOnRepo(t, "NoTipSelected", func(t *testing.T, context test.TestContext, repo *git.Repository) {
			err := UpdateCommand(repo, context.Context)

			if assert.NotNil(t, err) {
				assert.Equal(t, "HEAD not on a tip. Only tips can be upgraded.", err.Error())
			}
		})

		test.RunOnRepo(t, "NoBase", func(t *testing.T, context test.TestContext, repo *git.Repository) {
			// Create and select a tip
			head, _ := repo.Head()
			repo.References.Create(core.RefsTips+"test", head.Target(), true, "")
			repo.References.CreateSymbolic("HEAD", core.RefsTips+"test", true, "")

			err := UpdateCommand(repo, context.Context)

			// Upgrade requires a base to be defined
			if assert.NotNil(t, err) {
				assert.Equal(t, "Config value 'tip.test.base' was not found", err.Error())
			}
		})

		test.RunOnRepo(t, "NoTail", func(t *testing.T, context test.TestContext, repo *git.Repository) {
			// Create and select a tip
			head, _ := repo.Head()
			repo.References.Create(core.RefsTips+"test", head.Target(), true, "")
			repo.References.CreateSymbolic("HEAD", core.RefsTips+"test", true, "")

			// Configure the base of the tip
			config, _ := repo.Config()
			config.SetString("tip.test.base", "refs/heads/master")

			err := UpdateCommand(repo, context.Context)

			// Upgrade requires the tip to have a tail
			if assert.NotNil(t, err) {
				assert.Equal(t, "Reference '"+core.RefsTails+"test' not found", err.Error())
			}
		})

		test.RunOnRemote(t, "UpgradeSuccess", func(t *testing.T, context test.TestContext, repo, remote *git.Repository) {
			// Create a tip based on refs/remotes/origin/master
			head, _ := repo.Head()
			test.CreateTip(repo, "test", "refs/remotes/origin/master", false)

			// Make origin/master and the tip diverge.
			// first commit on master
			masterOid, _ := test.Commit(repo, &test.CommitParams{Refname: "refs/heads/master"})
			origin, _ := repo.Remotes.Lookup("origin")
			origin.Push([]string{"refs/heads/master"}, nil)

			// Then select the tip and commit
			repo.References.CreateSymbolic("HEAD", core.RefsTips+"test", true, "")
			now := time.Now()
			signature := &git.Signature{
				Name:  "user1",
				Email: "email@example.com",
				When:  now,
			}
			test.WriteFile(repo, true, "foo", "line1")
			test.Commit(repo, nil)
			test.WriteFile(repo, true, "foo", "line1", "line2")
			test.Commit(repo, &test.CommitParams{
				Author:   signature,
				Commiter: signature,
				Message:  "last commit",
			})

			// Do the upgrade
			err := UpdateCommand(repo, context.Context)
			assert.Nil(t, err)

			// We expect the tip to be on top of origin/master
			head, _ = repo.Head()
			headCommit, _ := repo.LookupCommit(head.Target())
			assert.True(t, headCommit.Parent(0).Parent(0).Id().Equal(masterOid))
			assert.Equal(t, "last commit", headCommit.Message())
			assert.Equal(t, "user1", headCommit.Author().Name)
			assert.Equal(t, "email@example.com", headCommit.Author().Email)
			assert.Equal(t, now.Unix(), headCommit.Author().When.Unix())

			// We expect the tail to be updated on origin/master's target
			newTailRef, _ := repo.References.Lookup(core.RefsTails + "test")
			assert.True(t, newTailRef.Target().Equal(masterOid))

			// The repo state should be clean
			test.StatusClean(t, repo)

			// We expect the tip to be pushed on origin
			remoteTip, err := remote.References.Lookup(core.RefsTips + "test")
			if assert.Nil(t, err) {
				assert.True(t, remoteTip.Target().Equal(head.Target()))
			}

			// Output should be...
			assert.Equal(t, "Upgraded current tip 'test'\n", context.OutputBuffer.String())
		})

		test.RunOnRemote(t, "ConflictAbort", func(t *testing.T, context test.TestContext, repo, remote *git.Repository) {
			// Create a tip on head based on master
			test.CreateTip(repo, "test", "refs/heads/master", false)

			head, _ := repo.Head()
			tailBeforeUpgrade := head.Target()

			// Make master and the tip having a conflict.
			// First commit to head that is on master
			test.WriteFile(repo, true, "foo", "line1")
			test.Commit(repo, nil)
			// Then select the tip and commit
			firstCommit, _ := repo.LookupCommit(head.Target())
			tree, _ := firstCommit.Tree()
			repo.CheckoutTree(tree, &git.CheckoutOpts{Strategy: git.CheckoutForce})
			repo.References.CreateSymbolic("HEAD", core.RefsTips+"test", true, "")
			test.WriteFile(repo, true, "foo", "line1 bis")
			oidBeforeUpgrade, _ := test.Commit(repo, nil)

			// Do the upgrade
			err := UpdateCommand(repo, context.Context)
			if assert.NotNil(t, err) {
				assert.Equal(t, "Conflict while upgrading", err.Error())
			}

			// Abort the upgrade
			err = UpdateAbortCommand(repo)
			assert.Nil(t, err)

			// HEAD should be back to where it was
			head, _ = repo.Head()
			assert.Equal(t, core.RefsTips+"test", head.Name())
			assert.True(t, head.Target().Equal(oidBeforeUpgrade))
			// Tip's tail should be where it was
			tail, _ := repo.References.Lookup(core.RefsTails + "test")
			assert.True(t, tail.Target().Equal(tailBeforeUpgrade))
		})

		test.RunOnRemote(t, "ConflictContinue", func(t *testing.T, context test.TestContext, repo, remote *git.Repository) {
			head, _ := repo.Head()
			// Create a tip on head based on master
			test.CreateTip(repo, "test", "refs/heads/master", false)

			// Make master and the tip having a conflict.
			// First commit to head that is on master
			test.WriteFile(repo, true, "foo", "line1")
			test.Commit(repo, nil)
			// Then select the tip and commit
			firstCommit, _ := repo.LookupCommit(head.Target())
			tree, _ := firstCommit.Tree()
			repo.CheckoutTree(tree, &git.CheckoutOpts{Strategy: git.CheckoutForce})
			repo.References.CreateSymbolic("HEAD", core.RefsTips+"test", true, "")
			test.WriteFile(repo, true, "foo", "line1 bis")
			test.Commit(repo, nil)

			// Do the upgrade
			err := UpdateCommand(repo, context.Context)
			if assert.NotNil(t, err) {
				assert.Equal(t, "Conflict while upgrading", err.Error())
			}

			// File foo should be in conflict
			index, _ := repo.Index()

			_, err = index.GetConflict("foo")
			assert.Nil(t, err, "File foo should be in conflict")

			// Resolve the conflict
			test.WriteFile(repo, true, "foo", "line1 bis")

			// Continue the upgrade
			err = UpdateContinueCommand(repo)
			assert.Nil(t, err)

			// HEAD should be one commit ahead of master
			head, _ = repo.Head()
			master, _ := repo.References.Lookup("refs/heads/master")
			masterOid := master.Target()
			headCommit, _ := repo.LookupCommit(head.Target())
			assert.True(t, headCommit.Parent(0).Id().Equal(masterOid))

			// The tail should be on master's target
			newTailRef, _ := repo.References.Lookup(core.RefsTails + "test")
			assert.True(t, newTailRef.Target().Equal(masterOid))

			// The repo state should be clean
			assert.Equal(t, git.RepositoryStateNone, repo.State())

		})

		test.RunOnRemote(t, "UpgradeEmptyTip", func(t *testing.T, context test.TestContext, repo, remote *git.Repository) {
			// Create a tip based on origin/master
			test.CreateTip(repo, "test", "refs/remotes/origin/master", true)

			// Add a commit on origin/master
			masterOid, _ := test.Commit(repo, &test.CommitParams{Refname: "refs/heads/master"})
			origin, _ := repo.Remotes.Lookup("origin")
			origin.Push([]string{"refs/heads/master"}, nil)

			// Do the upgrade
			err := UpdateCommand(repo, context.Context)

			assert.Nil(t, err)

			// The tip should be on masterOid
			upgradedTip, _ := repo.References.Lookup(core.RefsTips + "test")
			assert.True(t, masterOid.Equal(upgradedTip.Target()))

			// The tail should be on masterOid too
			upgradedTail, _ := repo.References.Lookup(core.RefsTails + "test")
			assert.True(t, masterOid.Equal(upgradedTail.Target()))

			// The tip shouldn't be pushed on origin since it is empty
			_, err = remote.References.Lookup(core.RefsTips + "test")
			assert.NotNil(t, err)

			_, err = repo.References.Lookup(core.RefsRemoteTips + "origin/test")
			assert.NotNil(t, err)
		})

		test.RunOnRepo(t, "DirtyStateError", func(t *testing.T, context test.TestContext, repo *git.Repository) {
			// Create a tip on head based on refs/remotes/origin/master
			test.CreateTip(repo, "test", "refs/heads/master", true)

			test.WriteFile(repo, true, "foo", "line")
			test.Commit(repo, nil)

			test.WriteFile(repo, false, "foo", "bar")

			err := UpdateCommand(repo, context.Context)

			assert.NotNil(t, err)
		})
	})
}
