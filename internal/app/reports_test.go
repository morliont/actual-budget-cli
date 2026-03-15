package app

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/morliont/actual-budget-cli/internal/bridge"
	"github.com/morliont/actual-budget-cli/internal/config"
	"github.com/morliont/actual-budget-cli/internal/output"
)

func TestReportsMonthlyVariance_AgentJSONDeterministic(t *testing.T) {
	withAppDeps(t)
	loadConfig = func() (*config.Config, error) { return &config.Config{}, nil }

	runBridge = func(_ context.Context, op string, req bridge.Request, out any) error {
		if op != "budgets-categories" {
			t.Fatalf("unexpected op: %s", op)
		}
		args := req.Args.(bridge.BudgetCategoriesArgs)
		resp := out.(*bridge.BudgetCategoriesResponse)
		resp.Month = args.Month
		switch args.Month {
		case "2026-01":
			resp.Categories = []json.RawMessage{
				json.RawMessage(`{"category_id":"c1","category_name":"Groceries","category_group_id":"g1","category_group_name":"Living","budgeted":5000,"spent":4000,"remaining":1000,"variance":1000}`),
				json.RawMessage(`{"category_id":"c2","category_name":"Fuel","category_group_id":"g1","category_group_name":"Living","budgeted":2000,"spent":2500,"remaining":-500,"variance":-500}`),
			}
		case "2026-02":
			resp.Categories = []json.RawMessage{
				json.RawMessage(`{"category_id":"c3","category_name":"Fun","category_group_id":"g2","category_group_name":"Lifestyle","budgeted":3000,"spent":2800,"remaining":200,"variance":200}`),
			}
		default:
			t.Fatalf("unexpected month: %s", args.Month)
		}
		return nil
	}

	var got any
	printJSON = func(v any) error {
		got = v
		return nil
	}

	cmd := newReportsMonthlyVarianceCmd()
	cmd.Flags().Bool(agentJSONFlag, false, "")
	cmd.Flags().String(correlationIDFlag, "", "")
	cmd.SetArgs([]string{"--from", "2026-01", "--to", "2026-02", "--agent-json", "--correlation-id", "corr-variance"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	env, ok := got.(output.Envelope)
	if !ok || !env.OK || env.Error != nil {
		t.Fatalf("expected success envelope, got %#v", got)
	}
	report, ok := env.Data.(monthlyVarianceReport)
	if !ok {
		t.Fatalf("expected monthlyVarianceReport, got %T", env.Data)
	}
	if report.Quality.MonthCount != 2 || report.Quality.FailedMonthCount != 0 {
		t.Fatalf("unexpected quality: %+v", report.Quality)
	}
	if report.Summary.Budgeted != 10000 || report.Summary.Spent != 9300 || report.Summary.Remaining != 700 {
		t.Fatalf("unexpected summary: %+v", report.Summary)
	}
	if report.Months[0].Groups[0].Raw.Budgeted != 7000 {
		t.Fatalf("unexpected group totals: %+v", report.Months[0].Groups)
	}
}

func TestReportsMonthlyVariance_StrictFailsOnMismatch(t *testing.T) {
	withAppDeps(t)
	loadConfig = func() (*config.Config, error) { return &config.Config{}, nil }

	runBridge = func(_ context.Context, op string, req bridge.Request, out any) error {
		args := req.Args.(bridge.BudgetCategoriesArgs)
		resp := out.(*bridge.BudgetCategoriesResponse)
		resp.Month = args.Month
		resp.Categories = []json.RawMessage{
			json.RawMessage(`{"category_id":"c1","category_name":"Groceries","category_group_id":"g1","category_group_name":"Living","budgeted":5000,"spent":4000,"remaining":2000,"variance":2000}`),
		}
		return nil
	}

	cmd := newReportsMonthlyVarianceCmd()
	cmd.SetArgs([]string{"--from", "2026-01", "--to", "2026-01", "--strict"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected strict failure")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "strict mode failed") {
		t.Fatalf("unexpected strict error: %v", err)
	}
}

func TestMonthRange_Validation(t *testing.T) {
	months, err := monthRange("2026-01", "2026-03")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(months) != 3 || months[0] != "2026-01" || months[2] != "2026-03" {
		t.Fatalf("unexpected months: %#v", months)
	}
	if _, err := monthRange("2026-03", "2026-01"); err == nil {
		t.Fatal("expected invalid range error")
	}
}
