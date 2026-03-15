package app

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/morliont/actual-budget-cli/internal/bridge"
	"github.com/spf13/cobra"
)

type varianceAmounts struct {
	Budgeted         float64 `json:"budgeted"`
	Spent            float64 `json:"spent"` // legacy alias of net_spent
	NetSpent         float64 `json:"net_spent"`
	OutflowSpent     float64 `json:"outflow_spent"`
	InflowOffsets    float64 `json:"inflow_offsets"`
	Remaining        float64 `json:"remaining"`
	Variance         float64 `json:"variance"` // legacy alias of net_variance
	NetVariance      float64 `json:"net_variance"`
	PlanningVariance float64 `json:"planning_variance"`
}

type normalizedAmounts struct {
	BudgetedAbs          float64 `json:"budgetedAbs"`
	SpentAbs             float64 `json:"spentAbs"`
	RemainingSigned      float64 `json:"remainingSigned"`
	VarianceSigned       float64 `json:"varianceSigned"`
	PlanningVarianceSign float64 `json:"planningVarianceSigned"`
}

type varianceGroup struct {
	GroupID       string            `json:"groupId"`
	GroupName     string            `json:"groupName"`
	CategoryCount int               `json:"categoryCount"`
	Raw           varianceAmounts   `json:"raw"`
	Normalized    normalizedAmounts `json:"normalized"`
}

type varianceCheck struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Message string `json:"message"`
}

type monthQuality struct {
	Confidence       string   `json:"confidence"`
	Warnings         []string `json:"warnings"`
	CheckCount       int      `json:"checkCount"`
	FailedCheckCount int      `json:"failedCheckCount"`
}

type monthVariance struct {
	Month         string            `json:"month"`
	CategoryCount int               `json:"categoryCount"`
	GroupCount    int               `json:"groupCount"`
	Raw           varianceAmounts   `json:"raw"`
	Normalized    normalizedAmounts `json:"normalized"`
	Groups        []varianceGroup   `json:"groups"`
	Checks        []varianceCheck   `json:"checks"`
	Quality       monthQuality      `json:"quality"`
}

type analysisQuality struct {
	Confidence       string   `json:"confidence"`
	Warnings         []string `json:"warnings"`
	StrictMode       bool     `json:"strictMode"`
	MonthCount       int      `json:"monthCount"`
	FailedMonthCount int      `json:"failedMonthCount"`
}

type monthlyVarianceReport struct {
	From              string            `json:"from"`
	To                string            `json:"to"`
	Months            []monthVariance   `json:"months"`
	Summary           varianceAmounts   `json:"summary"`
	SummaryNormalized normalizedAmounts `json:"summaryNormalized"`
	Quality           analysisQuality   `json:"quality"`
}

type budgetCategoryRow struct {
	CategoryID   string `json:"category_id"`
	CategoryName string `json:"category_name"`
	GroupID      string `json:"category_group_id"`
	GroupName    string `json:"category_group_name"`
	Budgeted     any    `json:"budgeted"`
	Planned      any    `json:"planned"`
	Spent        any    `json:"spent"`
	Actual       any    `json:"actual"`
	Remaining    any    `json:"remaining"`
	Variance     any    `json:"variance"`
}

const varianceEpsilon = 0.0001

func newReportsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "reports", Short: "Deterministic analysis reports", Long: "Generate deterministic analysis reports for automation and operators."}
	cmd.AddCommand(newReportsMonthlyVarianceCmd())
	return cmd
}

func newReportsMonthlyVarianceCmd() *cobra.Command {
	var from string
	var to string
	var strict bool
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "monthly-variance --from YYYY-MM --to YYYY-MM",
		Short: "Analyze monthly budget variance by group",
		Long:  "Pull monthly category budgets in range and compute deterministic budgeted/spent/variance summaries.",
		Example: `  actual-cli reports monthly-variance --from 2026-01 --to 2026-03
  actual-cli reports monthly-variance --from 2026-01 --to 2026-03 --strict
  actual-cli --agent-json --non-interactive reports monthly-variance --from 2026-01 --to 2026-03 --strict`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateMonth(from, "from"); err != nil {
				return err
			}
			if err := validateMonth(to, "to"); err != nil {
				return err
			}
			months, err := monthRange(from, to)
			if err != nil {
				return err
			}

			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			report := monthlyVarianceReport{From: from, To: to, Months: make([]monthVariance, 0, len(months))}
			strictIssues := make([]string, 0)
			failedMonths := 0
			warnings := make([]string, 0)

			for _, month := range months {
				monthData, err := loadMonthVariance(cmd, cfg, month)
				if err != nil {
					return err
				}
				report.Months = append(report.Months, monthData)
				report.Summary.Budgeted += monthData.Raw.Budgeted
				report.Summary.NetSpent += monthData.Raw.NetSpent
				report.Summary.OutflowSpent += monthData.Raw.OutflowSpent
				report.Summary.InflowOffsets += monthData.Raw.InflowOffsets
				report.Summary.Remaining += monthData.Raw.Remaining
				report.Summary.NetVariance += monthData.Raw.NetVariance
				report.Summary.PlanningVariance += monthData.Raw.PlanningVariance
				if monthData.Quality.FailedCheckCount > 0 {
					failedMonths++
					strictIssues = append(strictIssues, fmt.Sprintf("%s (%d failed checks)", month, monthData.Quality.FailedCheckCount))
				}
				warnings = append(warnings, monthData.Quality.Warnings...)
			}

			report.Summary = sanitizeVariance(report.Summary)
			report.SummaryNormalized = normalizeAmounts(report.Summary)
			report.Quality = analysisQuality{
				Confidence:       confidenceFromFailures(failedMonths),
				Warnings:         dedupeSorted(warnings),
				StrictMode:       strict,
				MonthCount:       len(report.Months),
				FailedMonthCount: failedMonths,
			}

			if strict && len(strictIssues) > 0 {
				return fmt.Errorf("strict mode failed: reconciliation mismatches detected in %d month(s): %s", len(strictIssues), strings.Join(strictIssues, "; "))
			}

			if useAgentJSON(cmd) {
				return printJSON(successEnvelope(cmd, report))
			}
			if asJSON {
				return printJSON(report)
			}

			rows := make([][]string, 0, len(report.Months))
			for _, m := range report.Months {
				rows = append(rows, []string{m.Month, formatCurrencyCentsBE(m.Raw.Budgeted), formatCurrencyCentsBE(m.Raw.NetSpent), formatCurrencyCentsBE(m.Raw.OutflowSpent), formatCurrencyCentsBE(m.Raw.InflowOffsets), formatCurrencyCentsBE(m.Raw.PlanningVariance), m.Quality.Confidence})
			}
			printTable([]string{"Month", "Budgeted", "Net spent", "Outflow spent", "Inflow offsets", "Planning variance", "Confidence"}, rows)
			return nil
		},
	}

	cmd.Flags().StringVar(&from, "from", "", "Start month (YYYY-MM)")
	cmd.Flags().StringVar(&to, "to", "", "End month (YYYY-MM)")
	cmd.Flags().BoolVar(&strict, "strict", false, "Fail when reconciliation checks detect inconsistencies")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output JSON")
	return cmd
}

func loadMonthVariance(cmd *cobra.Command, cfg any, month string) (monthVariance, error) {
	var res bridge.BudgetCategoriesResponse
	if err := runBridge(cmd.Context(), "budgets-categories", bridge.Request{Config: cfg, Args: bridge.BudgetCategoriesArgs{Month: month}}, &res); err != nil {
		return monthVariance{}, err
	}

	groupTotals := map[string]*varianceGroup{}
	groupOrder := make([]string, 0)
	monthTotals := varianceAmounts{}
	categoryCount := 0
	warnings := make([]string, 0)
	checks := make([]varianceCheck, 0)

	for _, raw := range res.Categories {
		var row budgetCategoryRow
		if err := json.Unmarshal(raw, &row); err != nil {
			return monthVariance{}, fmt.Errorf("invalid budget category payload: %w", err)
		}
		categoryCount++
		budgeted := valueOrFallback(row.Budgeted, row.Planned)
		netSpent := valueOrFallback(row.Spent, row.Actual)
		outflowSpent := 0.0
		inflowOffsets := 0.0
		if netSpent < 0 {
			outflowSpent = -netSpent
		} else if netSpent > 0 {
			inflowOffsets = netSpent
		}
		netVariance := budgeted - netSpent
		planningVariance := budgeted - outflowSpent
		remaining := numberWithFallback(row.Remaining, netVariance)
		variance := numberWithFallback(row.Variance, netVariance)

		groupKey := row.GroupID + "|" + row.GroupName
		g, ok := groupTotals[groupKey]
		if !ok {
			g = &varianceGroup{GroupID: row.GroupID, GroupName: row.GroupName}
			groupTotals[groupKey] = g
			groupOrder = append(groupOrder, groupKey)
		}
		g.CategoryCount++
		g.Raw.Budgeted += budgeted
		g.Raw.NetSpent += netSpent
		g.Raw.OutflowSpent += outflowSpent
		g.Raw.InflowOffsets += inflowOffsets
		g.Raw.Remaining += remaining
		g.Raw.NetVariance += variance
		g.Raw.PlanningVariance += planningVariance

		monthTotals.Budgeted += budgeted
		monthTotals.NetSpent += netSpent
		monthTotals.OutflowSpent += outflowSpent
		monthTotals.InflowOffsets += inflowOffsets
		monthTotals.Remaining += remaining
		monthTotals.NetVariance += variance
		monthTotals.PlanningVariance += planningVariance

		catDelta := sanitizeFinite(netVariance - remaining)
		if math.Abs(catDelta) > varianceEpsilon {
			warnings = append(warnings, fmt.Sprintf("%s/%s remaining mismatch", row.GroupName, row.CategoryName))
		}
	}

	groups := make([]varianceGroup, 0, len(groupOrder))
	for _, key := range groupOrder {
		g := groupTotals[key]
		g.Raw = sanitizeVariance(g.Raw)
		g.Normalized = normalizeAmounts(g.Raw)
		groups = append(groups, *g)
	}
	sort.SliceStable(groups, func(i, j int) bool {
		if groups[i].GroupName != groups[j].GroupName {
			return groups[i].GroupName < groups[j].GroupName
		}
		return groups[i].GroupID < groups[j].GroupID
	})

	monthTotals = sanitizeVariance(monthTotals)
	groupSum := varianceAmounts{}
	for _, g := range groups {
		groupSum.Budgeted += g.Raw.Budgeted
		groupSum.NetSpent += g.Raw.NetSpent
		groupSum.OutflowSpent += g.Raw.OutflowSpent
		groupSum.InflowOffsets += g.Raw.InflowOffsets
		groupSum.Remaining += g.Raw.Remaining
		groupSum.NetVariance += g.Raw.NetVariance
		groupSum.PlanningVariance += g.Raw.PlanningVariance
	}
	groupSum = sanitizeVariance(groupSum)

	checks = append(checks,
		varianceCheck{Name: "identity_net_spent", Passed: almostEqual(monthTotals.NetSpent, monthTotals.InflowOffsets-monthTotals.OutflowSpent), Message: "month net_spent equals inflow_offsets - outflow_spent"},
		varianceCheck{Name: "identity_remaining_net", Passed: almostEqual(monthTotals.Budgeted-monthTotals.NetSpent, monthTotals.Remaining), Message: "month remaining equals budgeted - net_spent"},
		varianceCheck{Name: "identity_variance_net", Passed: almostEqual(monthTotals.Budgeted-monthTotals.NetSpent, monthTotals.NetVariance), Message: "month net_variance equals budgeted - net_spent"},
		varianceCheck{Name: "identity_planning_variance", Passed: almostEqual(monthTotals.Budgeted-monthTotals.OutflowSpent, monthTotals.PlanningVariance), Message: "month planning_variance equals budgeted - outflow_spent"},
		varianceCheck{Name: "group_sum_budgeted", Passed: almostEqual(groupSum.Budgeted, monthTotals.Budgeted), Message: "group budgeted totals reconcile with month"},
		varianceCheck{Name: "group_sum_net_spent", Passed: almostEqual(groupSum.NetSpent, monthTotals.NetSpent), Message: "group net_spent totals reconcile with month"},
		varianceCheck{Name: "group_sum_outflow_spent", Passed: almostEqual(groupSum.OutflowSpent, monthTotals.OutflowSpent), Message: "group outflow_spent totals reconcile with month"},
		varianceCheck{Name: "group_sum_inflow_offsets", Passed: almostEqual(groupSum.InflowOffsets, monthTotals.InflowOffsets), Message: "group inflow_offsets totals reconcile with month"},
		varianceCheck{Name: "group_sum_remaining", Passed: almostEqual(groupSum.Remaining, monthTotals.Remaining), Message: "group remaining totals reconcile with month"},
		varianceCheck{Name: "group_sum_net_variance", Passed: almostEqual(groupSum.NetVariance, monthTotals.NetVariance), Message: "group net_variance totals reconcile with month"},
		varianceCheck{Name: "group_sum_planning_variance", Passed: almostEqual(groupSum.PlanningVariance, monthTotals.PlanningVariance), Message: "group planning_variance totals reconcile with month"},
	)

	failed := 0
	for _, c := range checks {
		if !c.Passed {
			failed++
			warnings = append(warnings, fmt.Sprintf("%s: failed", c.Name))
		}
	}

	return monthVariance{
		Month:         month,
		CategoryCount: categoryCount,
		GroupCount:    len(groups),
		Raw:           monthTotals,
		Normalized:    normalizeAmounts(monthTotals),
		Groups:        groups,
		Checks:        checks,
		Quality:       monthQuality{Confidence: confidenceFromFailures(boolToInt(failed > 0)), Warnings: dedupeSorted(warnings), CheckCount: len(checks), FailedCheckCount: failed},
	}, nil
}

func monthRange(from, to string) ([]string, error) {
	start, err := time.Parse("2006-01", from)
	if err != nil {
		return nil, err
	}
	end, err := time.Parse("2006-01", to)
	if err != nil {
		return nil, err
	}
	if start.After(end) {
		return nil, fmt.Errorf("invalid month range: --from (%s) cannot be after --to (%s)", from, to)
	}
	months := make([]string, 0)
	for m := start; !m.After(end); m = m.AddDate(0, 1, 0) {
		months = append(months, m.Format("2006-01"))
	}
	return months, nil
}

func normalizeAmounts(v varianceAmounts) normalizedAmounts {
	v = sanitizeVariance(v)
	return normalizedAmounts{BudgetedAbs: math.Abs(v.Budgeted), SpentAbs: math.Abs(v.NetSpent), RemainingSigned: v.Remaining, VarianceSigned: v.NetVariance, PlanningVarianceSign: v.PlanningVariance}
}

func sanitizeVariance(v varianceAmounts) varianceAmounts {
	v.Budgeted = sanitizeFinite(v.Budgeted)
	v.NetSpent = sanitizeFinite(v.NetSpent)
	v.OutflowSpent = sanitizeFinite(v.OutflowSpent)
	v.InflowOffsets = sanitizeFinite(v.InflowOffsets)
	v.Remaining = sanitizeFinite(v.Remaining)
	v.NetVariance = sanitizeFinite(v.NetVariance)
	v.PlanningVariance = sanitizeFinite(v.PlanningVariance)
	v.Spent = v.NetSpent
	v.Variance = v.NetVariance
	return v
}

func valueOrFallback(primary any, fallback any) float64 {
	if n, ok := numberFromAny(primary); ok {
		return sanitizeFinite(n)
	}
	if n, ok := numberFromAny(fallback); ok {
		return sanitizeFinite(n)
	}
	return 0
}

func numberWithFallback(primary any, fallback float64) float64 {
	if n, ok := numberFromAny(primary); ok {
		return sanitizeFinite(n)
	}
	return sanitizeFinite(fallback)
}

func almostEqual(a, b float64) bool {
	return math.Abs(sanitizeFinite(a)-sanitizeFinite(b)) <= varianceEpsilon
}

func dedupeSorted(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
	sort.Strings(items)
	out := make([]string, 0, len(items))
	for _, item := range items {
		if len(out) == 0 || out[len(out)-1] != item {
			out = append(out, item)
		}
	}
	return out
}

func confidenceFromFailures(failCount int) string {
	if failCount <= 0 {
		return "high"
	}
	if failCount == 1 {
		return "medium"
	}
	return "low"
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
