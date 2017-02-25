package env

import (
	"bytes"
	"errors"
	"gopkg.in/libgit2/git2go.v25"
	"os"
	"os/exec"
	"strings"
)

func RewriteStartCommand(repo *git.Repository) error {
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

	return runGit(cmd)
}

func RewriteContinueCommand(repo *git.Repository) error {
	if repo.State() != git.RepositoryStateRebaseInteractive {
		return errors.New("Not in a rewrite sequence.")
	}

	cmd := exec.Command("git", "-C", repo.Workdir(),
		"rebase", "--continue")

	return runGit(cmd)
}

func RewriteAbortCommand(repo *git.Repository) error {
	if repo.State() != git.RepositoryStateRebaseInteractive {
		return errors.New("Not in a rewrite sequence.")
	}

	cmd := exec.Command("git", "-C", repo.Workdir(),
		"rebase", "--abort")

	return runGit(cmd)
}

func runGit(cmd *exec.Cmd) error {
	errOut := new(bytes.Buffer)
	cmd.Stderr = errOut
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err := cmd.Run()

	if err != nil {
		return err
	}

	return nil
}