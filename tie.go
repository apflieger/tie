package main

import (
	"errors"
	"fmt"
	"github.com/apflieger/tie/commands"
	"github.com/apflieger/tie/core"
	"github.com/apflieger/tie/env"
	"github.com/spf13/cobra"
	"gopkg.in/libgit2/git2go.v25"
	"log"
	"os"
)

func main() {
	logger := log.New(os.Stdout, "", 0)

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

	rootCmd.AddCommand(buildCommitCommand(repo))
	rootCmd.AddCommand(buildSelectCommand(repo))
	rootCmd.AddCommand(buildUpgradeCommand(repo))
	rootCmd.AddCommand(buildRewriteCommand(repo))
	rootCmd.AddCommand(buildCreateCommand(repo))
	rootCmd.AddCommand(buildListCommand(repo, logger))
	rootCmd.AddCommand(buildDeleteCommand(repo, logger))

	rootCmd.Execute()
}

func buildCommitCommand(repo *git.Repository) *cobra.Command {
	var message, tipName string

	commitCommand := &cobra.Command{
		Use:   "commit",
		Short: "Record changes in the currently selected tip",
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.CommitCommand(repo, message, env.OpenEditor, tipName)
		},
	}

	commitCommand.Flags().StringVarP(&message, "message", "m", "", "commit message")
	commitCommand.Flags().StringVarP(&tipName, "tip", "t", core.OptionMissing, "create a tip on the fly")
	commitCommand.Flag("tip").NoOptDefVal = core.OptionWithoutValue

	commitCommand.Aliases = []string{"ci"}

	return commitCommand
}

func buildSelectCommand(repo *git.Repository) *cobra.Command {
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

func buildUpgradeCommand(repo *git.Repository) *cobra.Command {
	upgradeCommand := &cobra.Command{
		Use:   "upgrade",
		Short: "Get the current tip up-to-date with it's base",
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.UpgradeCommand(repo)
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

func buildRewriteCommand(repo *git.Repository) *cobra.Command {
	rewriteCommand := &cobra.Command{
		Short: "Allow to edit, reword or reorder current tip's commits",
		Use:   "rewrite",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return env.RewriteStartCommand(repo)
			} else if args[0] == "continue" {
				return env.RewriteContinueCommand(repo)
			} else if args[0] == "abort" {
				return env.RewriteAbortCommand(repo)
			} else {
				return fmt.Errorf("Incurrect verb '%v'.\n", args[0])
			}
		},
	}

	var message string

	amendCommand := &cobra.Command{
		Use:   "amend",
		Short: "Meld changes into the previous commit",
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.AmendCommand(repo, message, env.OpenEditor)
		},
	}

	amendCommand.Flags().StringVarP(&message, "message", "m", core.OptionMissing, "commit message")
	amendCommand.Flag("message").NoOptDefVal = core.OptionWithoutValue

	rewriteCommand.AddCommand(amendCommand)

	rewriteCommand.Aliases = []string{"rw"}

	return rewriteCommand
}

func buildCreateCommand(repo *git.Repository) *cobra.Command {
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

func buildListCommand(repo *git.Repository, logger *log.Logger) *cobra.Command {
	var listTips, listBranches, listRemotes, listAll bool

	listCommand := &cobra.Command{
		Use:   "list",
		Short: "List tips and branches",
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.ListCommand(repo, logger, listTips, listBranches, listRemotes, listAll)
		},
	}

	listCommand.Flags().BoolVarP(&listTips, "tips", "t", false, "list tips")
	listCommand.Flags().BoolVarP(&listBranches, "branches", "b", false, "list branches")
	listCommand.Flags().BoolVarP(&listRemotes, "remotes", "r", false, "list remote branches or tips")
	listCommand.Flags().BoolVarP(&listAll, "all", "a", false, "list tips and branches, local and remote")

	listCommand.Aliases = []string{"ls"}

	return listCommand
}

func buildDeleteCommand(repo *git.Repository, logger *log.Logger) *cobra.Command {
	deleteCommand := &cobra.Command{
		Use:   "delete",
		Short: "Delete tips",
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.DeleteCommand(repo, logger, args)
		},
	}

	return deleteCommand
}
