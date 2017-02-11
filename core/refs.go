package core

import (
	"errors"
	"fmt"
	"gopkg.in/libgit2/git2go.v25"
	"regexp"
	"strings"
)

const (
	RefsTips       = "refs/tips/local/"
	RefsRemoteTips = "refs/tips/"
	RefsTails      = "refs/tails/"
)

var refPatterns []string = []string{
	"%v",
	"refs/%v",
	RefsTips + "%v",
	RefsRemoteTips + "%v",
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
	if !strings.HasPrefix(refName, RefsTips) {
		return "", errors.New("")
	}
	return strings.Replace(refName, RefsTips, "", 1), nil
}

func RemoteName(ref string) (string, error) {
	regexp := regexp.MustCompile(`refs/remotes/([^/]*)/.*`)
	matches := regexp.FindStringSubmatch(ref)
	if len(matches) < 2 {
		return "", fmt.Errorf("\"%v\" is not a remote branch.", ref)
	}
	return matches[1], nil
}
