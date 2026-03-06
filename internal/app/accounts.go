package app

import (
	"fmt"

	"github.com/morliont/actual-budget-cli/internal/bridge"
	"github.com/morliont/actual-budget-cli/internal/config"
	"github.com/morliont/actual-budget-cli/internal/output"
	"github.com/spf13/cobra"
)

func newAccountsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "accounts", Short: "Account commands"}
	cmd.AddCommand(newAccountsListCmd())
	return cmd
}

func newAccountsListCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			var res struct {
				Accounts []map[string]any `json:"accounts"`
			}
			if err := bridge.Run(cmd.Context(), "accounts-list", bridge.Request{Config: cfg}, &res); err != nil {
				return err
			}
			if asJSON {
				return output.PrintJSON(res.Accounts)
			}
			rows := [][]string{}
			for _, a := range res.Accounts {
				rows = append(rows, []string{fmt.Sprint(a["id"]), fmt.Sprint(a["name"]), fmt.Sprint(a["type"]), fmt.Sprint(a["offbudget"]), fmt.Sprint(a["closed"])})
			}
			output.PrintTable([]string{"ID", "Name", "Type", "Off Budget", "Closed"}, rows)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output JSON")
	return cmd
}
