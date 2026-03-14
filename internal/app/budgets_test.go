package app

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/morliont/actual-budget-cli/internal/bridge"
	"github.com/morliont/actual-budget-cli/internal/config"
	"github.com/morliont/actual-budget-cli/internal/output"
)

func TestBudgetSummaryAgentPayload_CoreAndExtra(t *testing.T) {
	payload, err := budgetSummaryAgentPayload(bridge.BudgetSummaryResponse{
		Month:  "2026-03",
		Budget: json.RawMessage(`{"income":1000,"budgeted":700,"spent":500,"providerField":"x"}`),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload["month"] != "2026-03" || payload["income"] != float64(1000) || payload["budgeted"] != float64(700) || payload["spent"] != float64(500) {
		t.Fatalf("unexpected core payload: %#v", payload)
	}
	extra, ok := payload["extra"].(map[string]any)
	if !ok {
		t.Fatalf("expected extra map, got %T", payload["extra"])
	}
	if extra["providerField"] != "x" {
		t.Fatalf("expected provider field in extra, got %#v", extra)
	}
}

func TestBudgetsSummary_AgentJSONContract(t *testing.T) {
	withAppDeps(t)
	loadConfig = func() (*config.Config, error) { return &config.Config{}, nil }
	runBridge = func(_ context.Context, op string, _ bridge.Request, out any) error {
		if op != "budgets-summary" {
			t.Fatalf("unexpected op: %s", op)
		}
		resp := out.(*bridge.BudgetSummaryResponse)
		resp.Month = "2026-03"
		resp.Budget = json.RawMessage(`{"income":1200,"budgeted":800,"spent":650,"upstream":"keep"}`)
		return nil
	}
	var got any
	printJSON = func(v any) error {
		got = v
		return nil
	}

	cmd := newBudgetsSummaryCmd()
	cmd.Flags().Bool(agentJSONFlag, false, "")
	cmd.SetArgs([]string{"--agent-json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	env, ok := got.(output.Envelope)
	if !ok || !env.OK {
		t.Fatalf("expected success envelope, got %#v", got)
	}
	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected map data, got %T", env.Data)
	}
	budget, ok := data["budget"].(map[string]any)
	if !ok {
		t.Fatalf("expected budget object, got %T", data["budget"])
	}
	if budget["month"] != "2026-03" || budget["income"] != float64(1200) || budget["budgeted"] != float64(800) || budget["spent"] != float64(650) {
		t.Fatalf("unexpected core budget schema: %#v", budget)
	}
	extra, ok := budget["extra"].(map[string]any)
	if !ok {
		t.Fatalf("expected extra map, got %T", budget["extra"])
	}
	if extra["upstream"] != "keep" {
		t.Fatalf("expected upstream field under extra: %#v", extra)
	}
}
