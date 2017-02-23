package core

import (
	"errors"
	"fmt"
	"gopkg.in/libgit2/git2go.v25"
	"regexp"
	"strings"
)

const (
	RefsTips       = "refs/tips/"
	RefsRemoteTips = "refs/rtips/"
	RefsTails      = "refs/tails/"
)

var dwimPatterns []string = []string{
	"%v",
	"refs/%v",
	RefsTips + "%v",
	RefsRemoteTips + "%v",
	"refs/remotes/%v",
}

func Dwim(repo *git.Repository, shorthand string) (*git.Reference, error) {
	for _, refPattern := range dwimPatterns {
		refname := fmt.Sprintf(refPattern, shorthand)

		if git.ReferenceIsValidName(refname) {
			ref, err := repo.References.Lookup(refname)

			if ref != nil && err == nil {
				return ref, nil
			}
		}
	}
	return nil, fmt.Errorf("No ref found for shorthand '%v'", shorthand)
}

func TipName(refName string) (string, error) {
	if !strings.HasPrefix(refName, RefsTips) {
		return "", errors.New("")
	}
	return strings.Replace(refName, RefsTips, "", 1), nil
}

// Extracts the remote name of the ref that matches patterns refs/remotes/ or refs/rtips/
func RemoteName(ref string) (string, error) {
	remotesRegexp := regexp.MustCompile(`refs/remotes/([^/]*)/.*`)
	matches := remotesRegexp.FindStringSubmatch(ref)
	if len(matches) == 2 {
		return matches[1], nil
	}

	rtipsRegexp := regexp.MustCompile(RefsRemoteTips + `([^/]*)/.*`)
	matches = rtipsRegexp.FindStringSubmatch(ref)
	if len(matches) == 2 {
		return matches[1], nil
	}

	return "", fmt.Errorf("'%v' is not a remote ref.", ref)
}

func MatchingBranchfName(remoteRefName string) (string, error) {
	if remoteRefName == "refs/remotes/origin/master" {
		return "refs/heads/master", nil
	}
	return "", fmt.Errorf("'%v' is not a remote branch.", remoteRefName)
}
