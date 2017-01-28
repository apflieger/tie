package core

import (
	"errors"
	"fmt"
	"gopkg.in/libgit2/git2go.v25"
	"strings"
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

func TipName(refName string) (string, error) {
	if !strings.HasPrefix(refName, "refs/tips/local/") {
		return "", errors.New("")
	}
	return strings.Replace(refName, "refs/tips/local/", "", 1), nil
}
