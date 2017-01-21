package core

import (
	"gopkg.in/libgit2/git2go.v25"
	"fmt"
)

var refPatterns []string = []string{
	"%v",
	"refs/%v",
	"refs/tips/local/%v",
	"refs/tips/%v",
	"refs/remotes/%v",
}

func Dwim(repo *git.Repository, shorthand string) (*git.Reference, error) {
	for _, refPattern := range refPatterns {
		refname := fmt.Sprintf(refPattern, shorthand)

		if git.ReferenceIsValidName(refname) {
			ref, err := repo.References.Lookup(refname)

			if ref != nil && err == nil {
				return ref, nil
			}
		}
	}
	return nil, fmt.Errorf("No ref found for shorthand \"%v\"", shorthand)
}