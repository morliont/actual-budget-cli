package app

import (
	"fmt"

	"github.com/morliont/actual-budget-cli/internal/bridge"
	"github.com/morliont/actual-budget-cli/internal/config"
	"github.com/morliont/actual-budget-cli/internal/output"
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
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			var res map[string]any
			if err := bridge.Run(cmd.Context(), "budgets-summary", bridge.Request{Config: cfg}, &res); err != nil {
				return err
			}
			if asJSON {
				return output.PrintJSON(res)
			}
			month := fmt.Sprint(res["month"])
			budget, _ := res["budget"].(map[string]any)
			income := fmt.Sprint(budget["income"])
			budgeted := fmt.Sprint(budget["budgeted"])
			spent := fmt.Sprint(budget["spent"])
			output.PrintTable([]string{"Month", "Income", "Budgeted", "Spent"}, [][]string{{month, income, budgeted, spent}})
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output JSON")
	return cmd
}
