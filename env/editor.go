package env

import (
	"gopkg.in/libgit2/git2go.v25"
	"io/ioutil"
	"os"
	"os/exec"
)

func OpenEditor(config *git.Config, filepath string) (string, error) {
	//inspired by https://github.com/git/git/blob/master/editor.c

	editor, _ := config.LookupString("core.editor")

	if len(editor) == 0 {
		editor = os.Getenv("GIT_EDITOR")
	}

	if len(editor) == 0 {
		editor = os.Getenv("VISUAL")
	}

	if len(editor) == 0 {
		editor = os.Getenv("EDITOR")
	}

	if len(editor) == 0 {
		editor = "vi"
	}

	cmd := exec.Command(editor, filepath)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err := cmd.Run()

	if err != nil {
		return "", err
	}

	fileContent, err := ioutil.ReadFile(filepath)

	if err != nil {
		return "", err
	}

	return string(fileContent), nil
}
