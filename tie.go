package main

import (
	"github.com/apflieger/tie/args"
	"gopkg.in/libgit2/git2go.v25"
	"os"
)

func main() {
	repo, _ := git.OpenRepository(".")
	command, params, _ := args.ParseArgs(os.Args)
	command(repo, params)
}
