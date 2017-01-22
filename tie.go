package main

import (
	"os"
	"gopkg.in/libgit2/git2go.v25"
	"github.com/apflieger/tie/args"
)

func main() {
	repo, _ := git.OpenRepository(".")
	command, params, _ := args.ParseArgs(os.Args)
	command(repo, params[0])
}
