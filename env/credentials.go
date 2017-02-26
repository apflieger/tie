package env

import (
	"fmt"
	"gopkg.in/libgit2/git2go.v25"
)

func credentialCallback(url string, username_from_url string, allowedTypes git.CredType) (git.ErrorCode, *git.Cred) {
	// inspired by https://github.com/jwaldrip/git-get/blob/master/callbacks.go#L26

	if allowedTypes&git.CredTypeUserpassPlaintext != 0 {
		fmt.Println("Plain text user/password not implemented yet.")
		return git.ErrUser, nil
	}
	if allowedTypes&git.CredTypeSshKey != 0 {
		i, cred := git.NewCredSshKeyFromAgent(username_from_url)
		return git.ErrorCode(i), &cred
	}
	if allowedTypes&git.CredTypeSshCustom != 0 {
		fmt.Println("Custom ssh not implemented yet.")
		return git.ErrUser, nil
	}
	if allowedTypes&git.CredTypeDefault != 0 {
		i, cred := git.NewCredDefault()
		return git.ErrorCode(i), &cred
	}

	fmt.Printf("Unhandled credential types %v\n", allowedTypes)
	return git.ErrUser, nil
}

func certificateCheckCallback(cert *git.Certificate, valid bool, hostname string) git.ErrorCode {
	return git.ErrOk
}

var RemoteCallbacks = &git.RemoteCallbacks{
	CredentialsCallback:      credentialCallback,
	CertificateCheckCallback: certificateCheckCallback,
}
