package env

import (
	"errors"
	"github.com/apflieger/tie/core"
	"github.com/apflieger/tie/model"
	"gopkg.in/libgit2/git2go.v25"
	"os"
	"os/exec"
	"strings"
)

func RewriteStartCommand(repo *git.Repository, context model.Context) error {
	head, _ := repo.Head()

	if !strings.HasPrefix(head.Name(), "refs/tips/") {
		return errors.New("Not on a tip. Only tips can be rewritten.")
	}

	tipName := strings.Replace(head.Name(), "refs/tips/", "", 1)

	tail, _ := repo.References.Lookup("refs/tails/" + tipName)

	cmd := exec.Command("git", "-C", repo.Workdir(),
		"rebase", "-i",
		"--onto", tail.Target().String(),
		tail.Target().String())

	err := runGit(cmd)

	if err == nil && repo.State() == git.RepositoryStateNone {
		core.PushTip(repo, tipName, context)
	}

	return nil
}

func RewriteContinueCommand(repo *git.Repository, context model.Context) error {
	if repo.State() != git.RepositoryStateRebaseInteractive {
		return errors.New("Not in a rewrite sequence.")
	}

	cmd := exec.Command("git", "-C", repo.Workdir(),
		"rebase", "--continue")

	err := runGit(cmd)

	if err == nil && repo.State() == git.RepositoryStateNone {
		head, _ := repo.Head()
		tipName := strings.Replace(head.Name(), "refs/tips/", "", 1)
		core.PushTip(repo, tipName, context)
	}

	return nil
}

func RewriteAbortCommand(repo *git.Repository, context model.Context) error {
	if repo.State() != git.RepositoryStateRebaseInteractive {
		return errors.New("Not in a rewrite sequence.")
	}

	cmd := exec.Command("git", "-C", repo.Workdir(),
		"rebase", "--abort")

	runGit(cmd)

	return nil
}

func runGit(cmd *exec.Cmd) error {
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err := cmd.Run()

	if err != nil {
		return err
	}

	return nil
}
