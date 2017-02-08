package core

import "gopkg.in/libgit2/git2go.v25"

const OptionMissing = "OPTION_MISSING"
const OptionWithoutValue = "OPTION_WITHOUT_VALUE"

type OpenEditor func(config *git.Config, file string) (string, error)
