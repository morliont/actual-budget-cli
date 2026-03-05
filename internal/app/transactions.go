package app

import (
	"fmt"

	"github.com/morliont/actual-budget-cli/internal/bridge"
	"github.com/morliont/actual-budget-cli/internal/config"
	"github.com/morliont/actual-budget-cli/internal/output"
	"github.com/spf13/cobra"
)

func newTransactionsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "transactions", Short: "Transaction commands"}
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
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if from == "" {
				from = "1900-01-01"
			}
			if to == "" {
				to = "2999-12-31"
			}
			var res struct {
				Transactions []map[string]any `json:"transactions"`
			}
			if err := bridge.Run("transactions-list", bridge.Request{Config: cfg, Args: map[string]any{"accountId": accountID, "from": from, "to": to, "limit": limit}}, &res); err != nil {
				return err
			}
			if asJSON {
				return output.PrintJSON(res.Transactions)
			}
			rows := [][]string{}
			for _, t := range res.Transactions {
				rows = append(rows, []string{fmt.Sprint(t["date"]), fmt.Sprint(t["account"]), fmt.Sprint(t["payee_name"]), fmt.Sprint(t["amount"]), fmt.Sprint(t["notes"])})
			}
			output.PrintTable([]string{"Date", "Account", "Payee", "Amount", "Notes"}, rows)
			return nil
		},
	}

	cmd.Flags().StringVar(&accountID, "account", "", "Filter by account ID")
	cmd.Flags().StringVar(&from, "from", "", "From date YYYY-MM-DD")
	cmd.Flags().StringVar(&to, "to", "", "To date YYYY-MM-DD")
	cmd.Flags().IntVar(&limit, "limit", 100, "Max rows")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output JSON")
	return cmd
}
