package commands

import "gopkg.in/libgit2/git2go.v25"

type Command func(repo *git.Repository, refname string) error
