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

func Shorthand(ref string) string {
	remotesRegexp := regexp.MustCompile(`refs/(heads|remotes|tips|rtips)/(.*)`)
	matches := remotesRegexp.FindStringSubmatch(ref)
	if len(matches) == 3 {
		return matches[2]
	}
	return ref
}

func TipName(refName string) (string, error) {
	if !strings.HasPrefix(refName, RefsTips) {
		return "", errors.New("")
	}
	return strings.Replace(refName, RefsTips, "", 1), nil
}

// Return the substring of refName after the last '/'
func RefName(refName string) (string, error) {
	parts := strings.Split(refName, "/")

	if len(parts) > 1 {
		return parts[len(parts)-1], nil
	}

	return "", fmt.Errorf("'%v' is not a physical ref.", refName)
}

// Extracts the remote name of the ref and the local refname that matches
func ExplodeRemoteRef(ref string) (remote, localRefName string, err error) {
	remote, localRefName, err = explodeRemoteBranch(ref)

	if err == nil {
		return
	}

	remote, localRefName, err = explodeRemoteTip(ref)

	if err == nil {
		return
	}

	return "", "", fmt.Errorf("'%v' is not a remote ref.", ref)
}

func IsBranch(ref string) bool {
	return strings.HasPrefix(ref, "refs/heads/")
}

func explodeRemoteBranch(remoteBranchRefName string) (string, string, error) {
	remotesRegexp := regexp.MustCompile(`refs/remotes/([^/]*)/(.*)`)
	matches := remotesRegexp.FindStringSubmatch(remoteBranchRefName)
	if len(matches) == 3 {
		return matches[1], "refs/heads/" + matches[2], nil
	}
	return "", "", fmt.Errorf("'%v' is not a remote branch.", remoteBranchRefName)
}

func explodeRemoteTip(remoteBranchRefName string) (string, string, error) {
	remotesRegexp := regexp.MustCompile(RefsRemoteTips + `([^/]*)/(.*)`)
	matches := remotesRegexp.FindStringSubmatch(remoteBranchRefName)
	if len(matches) == 3 {
		return matches[1], RefsTips + matches[2], nil
	}
	return "", "", fmt.Errorf("'%v' is not a remote branch.", remoteBranchRefName)
}
