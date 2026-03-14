package app

import (
	"encoding/json"
	"fmt"

	"github.com/morliont/actual-budget-cli/internal/bridge"
	"github.com/spf13/cobra"
)

func newTransactionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transactions",
		Short: "Work with transactions",
		Long:  "Query transactions with optional account and date filters.",
	}
	cmd.AddCommand(newTransactionsListCmd())
	return cmd
}

func newTransactionsListCmd() *cobra.Command {
	var accountID, from, to string
	var limit int
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List transactions",
		Long:  "List transactions from the configured Actual budget.",
		Example: `  actual-cli transactions list
  actual-cli transactions list --account <ACCOUNT_ID> --from 2026-01-01 --to 2026-01-31 --limit 50
  actual-cli transactions list --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			if from == "" {
				from = "1900-01-01"
			}
			if to == "" {
				to = "2999-12-31"
			}
			if err := validateDate(from, "from"); err != nil {
				return err
			}
			if err := validateDate(to, "to"); err != nil {
				return err
			}
			if err := validateDateRange(from, to); err != nil {
				return err
			}
			if err := validateLimit(limit); err != nil {
				return err
			}
			var res bridge.TransactionsListResponse
			if err := runBridge(cmd.Context(), "transactions-list", bridge.Request{Config: cfg, Args: bridge.TransactionsListArgs{AccountID: accountID, From: from, To: to, Limit: limit}}, &res); err != nil {
				return err
			}
			if useAgentJSON(cmd) {
				return printJSON(successEnvelope(cmd, map[string]any{"transactions": res.Transactions}))
			}
			if asJSON {
				return printJSON(res.Transactions)
			}
			rows := make([][]string, 0, len(res.Transactions))
			for _, raw := range res.Transactions {
				var t bridge.TransactionRow
				if err := json.Unmarshal(raw, &t); err != nil {
					return fmt.Errorf("invalid transaction payload: %w", err)
				}
				rows = append(rows, []string{t.Date, t.Account, t.PayeeName, fmt.Sprint(t.Amount), t.Notes})
			}
			printTable([]string{"Date", "Account", "Payee", "Amount", "Notes"}, rows)
			return nil
		},
	}

	cmd.Flags().StringVar(&accountID, "account", "", "Filter by account ID")
	cmd.Flags().StringVar(&from, "from", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&to, "to", "", "End date (YYYY-MM-DD)")
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum number of rows (>0)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output JSON")
	return cmd
}
