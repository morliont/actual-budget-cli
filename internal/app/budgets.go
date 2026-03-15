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
	cmd.AddCommand(newBudgetsSummaryCmd(), newBudgetsCategoriesCmd())
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
			printTable([]string{"Month", "Income", "Budgeted", "Spent"}, [][]string{{
				res.Month,
				formatCurrencyCentsBE(totals.Income),
				formatCurrencyCentsBE(totals.Budgeted),
				formatCurrencyCentsBE(totals.Spent),
			}})
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output JSON")
	return cmd
}

func newBudgetsCategoriesCmd() *cobra.Command {
	var asJSON bool
	var month string

	cmd := &cobra.Command{
		Use:   "categories",
		Short: "Show per-category budget values for a month",
		Long:  "Show budgeted, spent, remaining and carryover-aware values per category for a selected month.",
		Example: `  actual-cli budgets categories --month 2026-03
  actual-cli budgets categories --month 2026-03 --json
  actual-cli budgets categories --month 2026-03 --agent-json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateMonth(month, "month"); err != nil {
				return err
			}

			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			var res bridge.BudgetCategoriesResponse
			if err := runBridge(cmd.Context(), "budgets-categories", bridge.Request{Config: cfg, Args: bridge.BudgetCategoriesArgs{Month: month}}, &res); err != nil {
				return err
			}

			if useAgentJSON(cmd) {
				return printJSON(successEnvelope(cmd, map[string]any{"month": res.Month, "categories": res.Categories}))
			}
			if asJSON {
				return printJSON(res)
			}

			rows := make([][]string, 0, len(res.Categories))
			for _, raw := range res.Categories {
				var c map[string]any
				if err := json.Unmarshal(raw, &c); err != nil {
					return fmt.Errorf("invalid budget category payload: %w", err)
				}
				rows = append(rows, []string{
					fmt.Sprint(c["category_name"]),
					fmt.Sprint(c["category_group_name"]),
					formatCurrencyCentsBE(firstNumeric(c, "budgeted", "planned")),
					formatCurrencyCentsBE(firstNumeric(c, "spent", "actual")),
					formatCurrencyCentsBE(firstNumeric(c, "remaining", "variance")),
					fmt.Sprint(c["carryover"]),
				})
			}

			printTable([]string{"Category", "Group", "Budgeted", "Spent", "Remaining", "Carryover"}, rows)
			return nil
		},
	}

	cmd.Flags().StringVar(&month, "month", "", "Budget month (YYYY-MM)")
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

func formatCurrencyCentsBE(amount float64) string {
	cents := int64(math.Round(sanitizeFinite(amount)))
	negative := cents < 0
	if negative {
		cents = -cents
	}

	units := cents / 100
	fraction := cents % 100

	formattedUnits := formatThousandsDot(units)
	formatted := fmt.Sprintf("%s,%02d", formattedUnits, fraction)
	if negative {
		return "-" + formatted
	}
	return formatted
}

func formatThousandsDot(n int64) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}

	var b strings.Builder
	first := len(s) % 3
	if first == 0 {
		first = 3
	}
	b.WriteString(s[:first])
	for i := first; i < len(s); i += 3 {
		b.WriteString(".")
		b.WriteString(s[i : i+3])
	}
	return b.String()
}

func sanitizeFinite(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return v
}
