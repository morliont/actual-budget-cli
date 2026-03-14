package app

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/morliont/actual-budget-cli/internal/config"
	"github.com/morliont/actual-budget-cli/internal/output"
)

func TestDoctor_AgentJSONReady(t *testing.T) {
	withAppDeps(t)
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg := &config.Config{ServerURL: "http://localhost:5006", BudgetID: "budget-1", DataDir: filepath.Join(home, "data")}
	loadConfig = func() (*config.Config, error) { return cfg, nil }
	lookPath = func(file string) (string, error) { return "/usr/bin/node", nil }

	if err := os.MkdirAll(filepath.Join(home, ".config", "actual-cli"), 0o700); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(home, ".config", "actual-cli", "config.json"), []byte(`{"serverUrl":"http://localhost:5006","budgetId":"budget-1"}`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	var got any
	printJSON = func(v any) error {
		got = v
		return nil
	}

	cmd := newDoctorCmd()
	cmd.Flags().Bool(agentJSONFlag, false, "")
	cmd.Flags().String(correlationIDFlag, "", "")
	cmd.SetArgs([]string{"--agent-json", "--correlation-id", "trace-1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	env, ok := got.(output.Envelope)
	if !ok {
		t.Fatalf("expected output.Envelope, got %T", got)
	}
	if !env.OK {
		t.Fatalf("expected success envelope, got %+v", env)
	}
	if env.Meta == nil || env.Meta.CorrelationID != "trace-1" {
		t.Fatalf("expected correlation id meta, got %+v", env.Meta)
	}
}

func TestDoctor_HumanModeFailsWhenNotReady(t *testing.T) {
	withAppDeps(t)
	loadConfig = func() (*config.Config, error) { return nil, errors.New("config not found") }
	lookPath = func(file string) (string, error) { return "", errors.New("not found") }

	cmd := newDoctorCmd()
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error when checks fail")
	}
}
