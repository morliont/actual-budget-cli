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

func TestBudgetsCategories_InvalidMonth(t *testing.T) {
	withAppDeps(t)
	loadConfig = func() (*config.Config, error) { return &config.Config{}, nil }
	called := false
	runBridge = func(_ context.Context, _ string, _ bridge.Request, _ any) error { called = true; return nil }

	cmd := newBudgetsCategoriesCmd()
	cmd.SetArgs([]string{"--month", "2026/03"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "expected YYYY-MM") {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("bridge should not be called on invalid month")
	}
}

func TestBudgetsCategories_AgentJSONContract(t *testing.T) {
	withAppDeps(t)
	loadConfig = func() (*config.Config, error) { return &config.Config{}, nil }
	runBridge = func(_ context.Context, op string, req bridge.Request, out any) error {
		if op != "budgets-categories" {
			t.Fatalf("unexpected op: %s", op)
		}
		if req.Args.(bridge.BudgetCategoriesArgs).Month != "2026-03" {
			t.Fatalf("unexpected month: %#v", req.Args)
		}
		resp := out.(*bridge.BudgetCategoriesResponse)
		resp.Month = "2026-03"
		resp.Categories = []json.RawMessage{json.RawMessage(`{"month":"2026-03","category_id":"c1","category_name":"Groceries","category_group_id":"g1","category_group_name":"Living","budgeted":5000,"planned":5000,"spent":4200,"actual":4200,"remaining":800,"variance":800,"carryover":true}`)}
		return nil
	}
	var got any
	printJSON = func(v any) error { got = v; return nil }

	cmd := newBudgetsCategoriesCmd()
	cmd.Flags().Bool(agentJSONFlag, false, "")
	cmd.SetArgs([]string{"--month", "2026-03", "--agent-json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	env, ok := got.(output.Envelope)
	if !ok || !env.OK {
		t.Fatalf("expected success envelope, got %#v", got)
	}
	data := env.Data.(map[string]any)
	if data["month"] != "2026-03" {
		t.Fatalf("unexpected month: %#v", data)
	}
}
