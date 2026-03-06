package app

import (
	"encoding/json"
	"fmt"

	"github.com/morliont/actual-budget-cli/internal/bridge"
	"github.com/spf13/cobra"
)

func newBudgetsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "budgets",
		Short: "Work with budget summaries",
		Long:  "Inspect high-level budget information from the current month.",
	}
	cmd.AddCommand(newBudgetsSummaryCmd())
	return cmd
}

func newBudgetsSummaryCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Show current budget month summary",
		Long:  "Show current-month budget totals (income, budgeted, spent).",
		Example: `  actual-cli budgets summary
  actual-cli budgets summary --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			var res bridge.BudgetSummaryResponse
			if err := runBridge(cmd.Context(), "budgets-summary", bridge.Request{Config: cfg}, &res); err != nil {
				return err
			}
			if asJSON {
				return printJSON(res)
			}
			var budget bridge.BudgetSummaryRow
			if err := json.Unmarshal(res.Budget, &budget); err != nil {
				return fmt.Errorf("invalid budget payload: %w", err)
			}
			printTable([]string{"Month", "Income", "Budgeted", "Spent"}, [][]string{{res.Month, fmt.Sprint(budget.Income), fmt.Sprint(budget.Budgeted), fmt.Sprint(budget.Spent)}})
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output JSON")
	return cmd
}
