package app

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/morliont/actual-budget-cli/internal/bridge"
	"github.com/morliont/actual-budget-cli/internal/config"
	"github.com/morliont/actual-budget-cli/internal/output"
)

func TestCategoriesList_AgentJSONEnvelope(t *testing.T) {
	withAppDeps(t)
	loadConfig = func() (*config.Config, error) { return &config.Config{}, nil }
	runBridge = func(_ context.Context, op string, _ bridge.Request, out any) error {
		if op != "categories-list" {
			t.Fatalf("unexpected op: %s", op)
		}
		resp := out.(*bridge.CategoriesListResponse)
		resp.Categories = []json.RawMessage{json.RawMessage(`{"id":"c1","name":"Groceries","group_id":"g1","group_name":"Living","hidden":false,"archived":null}`)}
		return nil
	}

	var got any
	printJSON = func(v any) error { got = v; return nil }

	cmd := newCategoriesListCmd()
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
	cats, ok := data["categories"].([]json.RawMessage)
	if !ok || len(cats) != 1 {
		t.Fatalf("expected one category, got %#v", data["categories"])
	}
}
