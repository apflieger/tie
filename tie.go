package main

import (
	"errors"
	"fmt"
	"github.com/apflieger/tie/commands"
	"github.com/apflieger/tie/core"
	"github.com/apflieger/tie/env"
	"github.com/apflieger/tie/model"
	"github.com/spf13/cobra"
	"gopkg.in/libgit2/git2go.v25"
	"log"
	"os"
)

func main() {
	path, err := git.Discover(".", false, nil)

	if err != nil {
		fmt.Println(err.Error())
	}

	repo, err := git.OpenRepository(path)

	if err != nil {
		fmt.Println(err.Error())
	}

	var rootCmd = &cobra.Command{
		SilenceUsage: true,
	}

	context := model.Context{
		Logger: log.New(os.Stdout, "", 0),
		RemoteCallbacks: git.RemoteCallbacks{
			CredentialsCallback:      env.CredentialCallback,
			CertificateCheckCallback: env.CertificateCheckCallback,
		},
		OpenEditor: env.OpenEditor,
	}

	rootCmd.AddCommand(buildCommitCommand(repo, context))
	rootCmd.AddCommand(buildSelectCommand(repo, context))
	rootCmd.AddCommand(buildUpgradeCommand(repo, context))
	rootCmd.AddCommand(buildRewriteCommand(repo, context))
	rootCmd.AddCommand(buildCreateCommand(repo, context))
	rootCmd.AddCommand(buildListCommand(repo, context))
	rootCmd.AddCommand(buildDeleteCommand(repo, context))
	rootCmd.AddCommand(buildStackCommand(repo, context))
	rootCmd.AddCommand(buildUpdateCommand(repo, context))

	rootCmd.Execute()
}

func buildCommitCommand(repo *git.Repository, context model.Context) *cobra.Command {
	var message, tipName string

	commitCommand := &cobra.Command{
		Use:   "commit [flags]",
		Short: "Record changes in the currently selected tip",
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.CommitCommand(repo, message, tipName, context)
		},
	}

	commitCommand.Flags().StringVarP(&message, "message", "m", "", "commit message")
	commitCommand.Flags().StringVarP(&tipName, "tip", "t", model.OptionMissing, "create and select a tip on the fly")
	commitCommand.Flag("tip").NoOptDefVal = model.OptionWithoutValue

	commitCommand.Aliases = []string{"ci"}

	return commitCommand
}

func buildSelectCommand(repo *git.Repository, context model.Context) *cobra.Command {
	selectCommand := &cobra.Command{
		Use:   "select [<tip or branch>]",
		Short: "Switch the repository on the given tip or branch",
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.SelectCommand(repo, args[0])
		},
	}

	selectCommand.Aliases = []string{"sl"}

	return selectCommand
}

func buildUpgradeCommand(repo *git.Repository, context model.Context) *cobra.Command {
	upgradeCommand := &cobra.Command{
		Use:   "upgrade",
		Short: "Get the current tip up-to-date with it's base",
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.UpgradeCommand(repo, context)
		},
	}

	abortCommand := &cobra.Command{
		Use: "abort",
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.UpgradeAbortCommand(repo)
		},
	}

	continueCommand := &cobra.Command{
		Use: "continue",
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.UpgradeContinueCommand(repo)
		},
	}

	upgradeCommand.AddCommand(abortCommand)
	upgradeCommand.AddCommand(continueCommand)

	return upgradeCommand
}

func buildRewriteCommand(repo *git.Repository, context model.Context) *cobra.Command {
	rewriteCommand := &cobra.Command{
		Short: "Allow to edit, reword or reorder current tip's commits",
		Use:   "rewrite",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return env.RewriteStartCommand(repo, context)
			} else if args[0] == "continue" {
				return env.RewriteContinueCommand(repo, context)
			} else if args[0] == "abort" {
				return env.RewriteAbortCommand(repo, context)
			} else {
				return fmt.Errorf("Incurrect verb '%v'.\n", args[0])
			}
		},
	}

	var message string

	amendCommand := &cobra.Command{
		Use:   "amend [flags]",
		Short: "Meld changes into the previous commit",
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.AmendCommand(repo, message, context)
		},
	}

	amendCommand.Flags().StringVarP(&message, "message", "m", model.OptionMissing, "commit message")
	amendCommand.Flag("message").NoOptDefVal = model.OptionWithoutValue

	rewriteCommand.AddCommand(amendCommand)

	rewriteCommand.Aliases = []string{"rw"}

	return rewriteCommand
}

func buildCreateCommand(repo *git.Repository, context model.Context) *cobra.Command {
	createCommand := &cobra.Command{
		Use:   "create <tipName> [<base>]",
		Short: "Create a tip",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("Argument missing")
			}

			tipName := args[0]

			base := ""

			if len(args) > 1 {
				base = args[1]
			}

			return commands.CreateCommand(repo, tipName, base)
		},
	}

	return createCommand
}

func buildListCommand(repo *git.Repository, context model.Context) *cobra.Command {
	var listTips, listBranches, listRemotes, listAll bool

	listCommand := &cobra.Command{
		Use:   "list [flags]",
		Short: "List tips and branches",
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.ListCommand(repo, context, listTips, listBranches, listRemotes, listAll)
		},
	}

	listCommand.Flags().BoolVarP(&listTips, "tips", "t", false, "list tips")
	listCommand.Flags().BoolVarP(&listBranches, "branches", "b", false, "list branches")
	listCommand.Flags().BoolVarP(&listRemotes, "remotes", "r", false, "list remote branches or tips")
	listCommand.Flags().BoolVarP(&listAll, "all", "a", false, "list tips and branches, local and remote")

	listCommand.Aliases = []string{"ls"}

	return listCommand
}

func buildDeleteCommand(repo *git.Repository, context model.Context) *cobra.Command {
	var stacked bool

	deleteCommand := &cobra.Command{
		Use:   "delete [flags] [<tip>]",
		Short: "Delete tips",
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.DeleteCommand(repo, stacked, args, context)
		},
	}

	deleteCommand.Flags().BoolVarP(&stacked, "stacked", "", false, "delete tips that have been stacked")

	return deleteCommand
}

func buildStackCommand(repo *git.Repository, context model.Context) *cobra.Command {
	stackCommand := &cobra.Command{
		Use:   "stack",
		Short: "Put tip's commits into a branch",
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.StackCommand(repo, context)
		},
	}

	return stackCommand
}

func buildUpdateCommand(repo *git.Repository, context model.Context) *cobra.Command {
	updateCommand := &cobra.Command{
		Use:   "update",
		Short: "Synchronize remote refs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return core.Fetch(repo, context)
		},
	}

	return updateCommand
}
