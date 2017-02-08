package main

import (
	"github.com/apflieger/tie/commands"
	"github.com/spf13/cobra"
	"gopkg.in/libgit2/git2go.v25"
)

func main() {
	repo, _ := git.OpenRepository(".")

	var rootCmd = &cobra.Command{
		SilenceUsage: true,
	}

	rootCmd.AddCommand(buildCommitCommand(repo))
	rootCmd.AddCommand(buildSelectCommand(repo))
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
			return commands.CommitCommand(repo, message)
		},
	}

	commitCommand.Flags().StringVarP(&message, "message", "m", "", "commit message")

	return commitCommand
}

func buildSelectCommand(repo *git.Repository) *cobra.Command {

	selectCommand := &cobra.Command{
		Use: "select <tip or branch>",
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.SelectCommand(repo, args[0])
		},
	}

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

	amendCommand := &cobra.Command{
		Use: "amend",
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.AmendCommand(repo)
		},
	}

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
			return commands.TipCreateCommand(repo, args[0], args[1])
		},
	}

	tipCommand.AddCommand(createCommand)

	return tipCommand
}
