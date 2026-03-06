package app

import (
	"encoding/json"
	"fmt"

	"github.com/morliont/actual-budget-cli/internal/bridge"
	"github.com/spf13/cobra"
)

func newAccountsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accounts",
		Short: "Work with accounts",
		Long:  "List and inspect account information from your Actual budget.",
	}
	cmd.AddCommand(newAccountsListCmd())
	return cmd
}

func newAccountsListCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List accounts",
		Long:  "List all accounts in the configured Actual budget.",
		Example: `  actual-cli accounts list
  actual-cli accounts list --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			var res bridge.AccountsListResponse
			if err := runBridge(cmd.Context(), "accounts-list", bridge.Request{Config: cfg}, &res); err != nil {
				return err
			}
			if asJSON {
				return printJSON(res.Accounts)
			}
			rows := make([][]string, 0, len(res.Accounts))
			for _, raw := range res.Accounts {
				var a bridge.AccountRow
				if err := json.Unmarshal(raw, &a); err != nil {
					return fmt.Errorf("invalid account payload: %w", err)
				}
				rows = append(rows, []string{a.ID, a.Name, a.Type, fmt.Sprint(a.OffBudget), fmt.Sprint(a.Closed)})
			}
			printTable([]string{"ID", "Name", "Type", "Off Budget", "Closed"}, rows)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output JSON")
	return cmd
}
