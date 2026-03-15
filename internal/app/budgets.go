package app

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

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

type budgetTotals struct {
	Income   float64
	Budgeted float64
	Spent    float64
}

func budgetSummaryTotals(raw map[string]any) budgetTotals {
	income, incomeOK := numberFromAny(raw["income"])
	budgeted, budgetedOK := numberFromAny(raw["budgeted"])
	spent, spentOK := numberFromAny(raw["spent"])

	if !incomeOK {
		income = firstNumeric(raw,
			"totalIncome",
			"incomeTotal",
			"incomeAvailable",
			"toBudget",
		)
	}

	if !budgetedOK || !spentOK {
		fallbackBudgeted, fallbackSpent := totalsFromCategoryStructures(raw)
		if !budgetedOK {
			budgeted = fallbackBudgeted
		}
		if !spentOK {
			spent = fallbackSpent
		}
	}

	return budgetTotals{
		Income:   sanitizeFinite(income),
		Budgeted: sanitizeFinite(budgeted),
		Spent:    sanitizeFinite(spent),
	}
}

func budgetSummaryAgentPayload(res bridge.BudgetSummaryResponse) (map[string]any, error) {
	var raw map[string]any
	if err := json.Unmarshal(res.Budget, &raw); err != nil {
		return nil, fmt.Errorf("invalid budget payload: %w", err)
	}

	totals := budgetSummaryTotals(raw)
	extra := map[string]any{}
	for k, v := range raw {
		if k == "income" || k == "budgeted" || k == "spent" {
			continue
		}
		extra[k] = v
	}

	return map[string]any{
		"month":    res.Month,
		"income":   totals.Income,
		"budgeted": totals.Budgeted,
		"spent":    totals.Spent,
		"extra":    extra,
	}, nil
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
			if useAgentJSON(cmd) {
				agentBudget, err := budgetSummaryAgentPayload(res)
				if err != nil {
					return err
				}
				return printJSON(successEnvelope(cmd, map[string]any{"budget": agentBudget}))
			}
			if asJSON {
				return printJSON(res)
			}

			var raw map[string]any
			if err := json.Unmarshal(res.Budget, &raw); err != nil {
				return fmt.Errorf("invalid budget payload: %w", err)
			}
			totals := budgetSummaryTotals(raw)
			printTable([]string{"Month", "Income", "Budgeted", "Spent"}, [][]string{{res.Month, fmt.Sprint(totals.Income), fmt.Sprint(totals.Budgeted), fmt.Sprint(totals.Spent)}})
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output JSON")
	return cmd
}

func totalsFromCategoryStructures(raw map[string]any) (budgeted float64, spent float64) {
	groups, ok := raw["categoryGroups"].([]any)
	if !ok {
		groups, ok = raw["groups"].([]any)
		if !ok {
			return 0, 0
		}
	}

	for _, group := range groups {
		gm, ok := group.(map[string]any)
		if !ok {
			continue
		}
		cats, ok := gm["categories"].([]any)
		if !ok {
			continue
		}
		for _, cat := range cats {
			cm, ok := cat.(map[string]any)
			if !ok {
				continue
			}
			budgeted += firstNumeric(cm, "budgeted", "budget", "budgetedAmount")
			spent += firstNumeric(cm, "spent", "activity", "spentAmount")
		}
	}

	return sanitizeFinite(budgeted), sanitizeFinite(spent)
}

func firstNumeric(m map[string]any, keys ...string) float64 {
	for _, key := range keys {
		if v, ok := numberFromAny(m[key]); ok {
			return v
		}
	}
	return 0
}

func numberFromAny(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case int32:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		if err == nil {
			return f, true
		}
	case string:
		t := strings.TrimSpace(n)
		if t == "" {
			return 0, false
		}
		var parsed json.Number = json.Number(t)
		f, err := parsed.Float64()
		if err == nil {
			return f, true
		}
	}
	return 0, false
}

func sanitizeFinite(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return v
}
