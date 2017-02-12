package main

import (
	"errors"
	"github.com/apflieger/tie/commands"
	"github.com/apflieger/tie/core"
	"github.com/apflieger/tie/env"
	"github.com/spf13/cobra"
	"gopkg.in/libgit2/git2go.v25"
	"log"
	"os"
)

func main() {
	repo, _ := git.OpenRepository(".")

	var rootCmd = &cobra.Command{
		SilenceUsage: true,
	}

	logger := log.New(os.Stdout, "", 0)

	rootCmd.AddCommand(buildCommitCommand(repo))
	rootCmd.AddCommand(buildSelectCommand(repo, logger))
	rootCmd.AddCommand(buildUpgradeCommand(repo))
	rootCmd.AddCommand(buildRewriteCommand(repo))
	rootCmd.AddCommand(buildTipCommand(repo))

	rootCmd.Execute()
}

func buildCommitCommand(repo *git.Repository) *cobra.Command {
	var message string

	commitCommand := &cobra.Command{
		Use: "commit",
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.CommitCommand(repo, message, env.OpenEditor)
		},
	}

	commitCommand.Flags().StringVarP(&message, "message", "m", "", "commit message")

	return commitCommand
}

func buildSelectCommand(repo *git.Repository, logger *log.Logger) *cobra.Command {

	var listTips, listBranches, listRemotes, listAll bool

	selectCommand := &cobra.Command{
		Use: "select [flags] [<tip or branch>]",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return commands.ListCommand(repo, logger, listTips, listBranches, listRemotes, listAll)
			}
			return commands.SelectCommand(repo, args[0])
		},
	}

	selectCommand.Flags().BoolVarP(&listTips, "tips", "t", false, "list tips")
	selectCommand.Flags().BoolVarP(&listBranches, "branches", "b", false, "list branches")
	selectCommand.Flags().BoolVarP(&listRemotes, "remotes", "r", false, "list remote branches or tips")
	selectCommand.Flags().BoolVarP(&listAll, "all", "a", false, "list tips and branches, local and remote")

	return selectCommand
}

func buildUpgradeCommand(repo *git.Repository) *cobra.Command {

	upgrdeCommand := &cobra.Command{
		Use: "upgrade",
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.UpgradeCommand(repo)
		},
	}

	return upgrdeCommand
}

func buildRewriteCommand(repo *git.Repository) *cobra.Command {

	rewriteCommand := &cobra.Command{
		Use: "rewrite",
	}

	var message string

	amendCommand := &cobra.Command{
		Use: "amend",
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.AmendCommand(repo, message, env.OpenEditor)
		},
	}

	amendCommand.Flags().StringVarP(&message, "message", "m", core.OptionMissing, "commit message")
	amendCommand.Flag("message").NoOptDefVal = core.OptionWithoutValue

	rewriteCommand.AddCommand(amendCommand)

	return rewriteCommand
}

func buildTipCommand(repo *git.Repository) *cobra.Command {

	tipCommand := &cobra.Command{
		Use: "tip",
	}

	createCommand := &cobra.Command{
		Use: "create <tipName> [<base>]",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("Argument missing")
			}

			tipName := args[0]

			base := ""

			if len(args) > 1 {
				base = args[1]
			}

			return commands.TipCreateCommand(repo, tipName, base)
		},
	}

	tipCommand.AddCommand(createCommand)

	return tipCommand
}
