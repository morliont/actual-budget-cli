package app

import (
	"github.com/morliont/actual-budget-cli/internal/version"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "actual-cli",
		Short: "Command line interface for Actual Budget",
		Long: `actual-cli is a command-line tool for working with Actual Budget.

Use it to authenticate once, then query accounts, transactions, and budgets
from scripts or your terminal.`,
		Example: `  actual-cli auth login --server http://localhost:5006 --budget <SYNC_ID>
  actual-cli accounts list
  actual-cli transactions list --account <ACCOUNT_ID> --from 2026-01-01 --to 2026-01-31
  actual-cli budgets summary --json`,
		Version: version.String(),
	}

	cmd.SetVersionTemplate("{{.Version}}\n")
	cmd.AddCommand(newInitCmd(), newAuthCmd(), newAccountsCmd(), newTransactionsCmd(), newBudgetsCmd())
	return cmd
}
