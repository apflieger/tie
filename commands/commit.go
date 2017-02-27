package commands

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/apflieger/tie/core"
	"github.com/apflieger/tie/model"
	"gopkg.in/libgit2/git2go.v25"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

func CommitCommand(repo *git.Repository, commitMessage string, openEditor model.OpenEditor, tipName string, context model.Context) error {
	head, headCommit, tree := core.PrepareCommit(repo)

	if tipName == model.OptionMissing {
		var notTip error
		tipName, notTip = core.TipName(head.Name())

		if notTip != nil {
			return errors.New("HEAD is not on a tip. Run 'commit -t' to create a tip on the fly.")
		}
	} else {
		tipName = strings.Trim(tipName, " ")

		if tipName == "" {
			return errors.New("Name of the tip can't be empty.")
		}

		if tipName == model.OptionWithoutValue {
			tipName, _ = core.RefName(head.Name())
			tipName = tipName + "-tip"
		}

		// Create and select a new tip
		tip, _ := repo.References.Create(core.RefsTips+tipName, head.Target(), false, "tie commit -t")
		repo.References.Create(core.RefsTails+tipName, head.Target(), false, "tie commit -t")
		config, _ := repo.Config()
		config.SetString(fmt.Sprintf("tip.%v.base", tipName), head.Name())
		head, _ = repo.References.CreateSymbolic("HEAD", tip.Name(), true, "tie commit -t")
	}

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

	core.PushTip(repo, tipName, context)

	return nil
}
