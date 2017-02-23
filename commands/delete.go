package commands

import (
	"fmt"
	"github.com/apflieger/tie/core"
	"github.com/apflieger/tie/env"
	"gopkg.in/libgit2/git2go.v25"
	"log"
)

func DeleteCommand(repo *git.Repository, logger *log.Logger, refs []string) error {
	config, _ := repo.Config()
	for _, ref := range refs {
		tipName, err := core.TipName(ref)
		if err != nil {
			return err
		}

		tip, _ := repo.References.Lookup(ref)
		tip.Delete()

		tail, _ := repo.References.Lookup(core.RefsTails + tipName)
		tail.Delete()

		baseKey := fmt.Sprintf("tip.%v.base", tipName)
		base, _ := config.LookupString(baseKey)
		config.Delete(baseKey)

		remoteName, _, err := core.ExplodeRemoteRef(base)

		if err == nil {
			remote, _ := repo.Remotes.Lookup(remoteName)
			pushOptions := &git.PushOptions{
				RemoteCallbacks: git.RemoteCallbacks{
					CredentialsCallback:      env.CredentialCallback,
					CertificateCheckCallback: env.CertificateCheckCallback,
				},
			}
			refspecs := []string{":" + ref}

			compat, _ := config.LookupBool("tie.pushTipsAsBranches")

			if compat {
				refspecs = append(refspecs, ":refs/heads/tips/"+tipName)
			}

			err := remote.Push(refspecs, pushOptions)

			if err == nil {
				rtip, _ := repo.References.Lookup(core.RefsRemoteTips + remoteName + "/" + tipName)
				rtip.Delete()
			} else {
				logger.Println(err.Error())
				logger.Printf("Tip %v has been locally deleted but is still on %v\n", tipName, remoteName)
			}
		}

		logger.Printf("Deleted tip %v", tipName)
	}
	return nil
}
