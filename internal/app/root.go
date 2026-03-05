package app

import "github.com/spf13/cobra"

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "actual-cli",
		Short: "CLI for Actual Budget",
	}
	cmd.AddCommand(newAuthCmd(), newAccountsCmd(), newTransactionsCmd(), newBudgetsCmd())
	return cmd
}
