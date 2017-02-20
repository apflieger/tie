package commands

import (
	"bytes"
	"github.com/apflieger/tie/core"
	"gopkg.in/libgit2/git2go.v25"
	"io/ioutil"
	"path/filepath"
	"regexp"
)

func CommitCommand(repo *git.Repository, commitMessage string, openEditor core.OpenEditor) error {
	head, headCommit, tree := core.PrepareCommit(repo)
	tipName, _ := core.TipName(head.Name())

	if commitMessage == "" {
		linesRegexp := regexp.MustCompile(`(.*)`)
		lines := linesRegexp.FindAllString(headCommit.Message(), -1)
		presetCommitMessage := new(bytes.Buffer)
		for _, line := range lines {
			presetCommitMessage.WriteString("#" + line + "\n")
		}
		commitEditMsgFile := filepath.Join(repo.Path(), "COMMIT_EDITMSG")
		ioutil.WriteFile(commitEditMsgFile, presetCommitMessage.Bytes(), 0644)

		config, _ := repo.Config()
		commitMessage, _ = openEditor(config, commitEditMsgFile)
	}

	signature, _ := repo.DefaultSignature()
	repo.CreateCommit(head.Name(), signature, signature, core.FormatCommitMessage(commitMessage), tree, headCommit)

	core.PushTip(repo, tipName)

	return nil
}
