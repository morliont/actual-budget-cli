package app

import (
	"github.com/morliont/actual-budget-cli/internal/version"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	var agentJSON bool
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
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return enforceReadOnly(cmd)
		},
	}

	var nonInteractive bool
	var readOnly bool
	cmd.SetVersionTemplate("{{.Version}}\n")
	var correlationID string
	cmd.PersistentFlags().BoolVar(&agentJSON, agentJSONFlag, false, "Output stable machine-readable JSON envelope for agents")
	cmd.PersistentFlags().BoolVar(&nonInteractive, nonInteractiveFlag, false, "Disable interactive prompts; fail fast on missing required inputs")
	cmd.PersistentFlags().StringVar(&correlationID, correlationIDFlag, "", "Optional trace/correlation ID for logs and --agent-json output (or ACTUAL_CLI_CORRELATION_ID)")
	cmd.PersistentFlags().BoolVar(&readOnly, readOnlyFlag, readOnlyDefaultFromEnv(), "Block mutating commands (default from ACTUAL_CLI_READ_ONLY)")
	cmd.AddCommand(newAuthCmd(), newAccountsCmd(), newCategoriesCmd(), newTransactionsCmd(), newBudgetsCmd(), newDoctorCmd())
	return cmd
}
