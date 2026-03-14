package app

import (
	"context"
	"strings"
	"testing"

	"github.com/morliont/actual-budget-cli/internal/bridge"
	"github.com/morliont/actual-budget-cli/internal/config"
	"github.com/morliont/actual-budget-cli/internal/output"
)

func TestAuthLogin_NonInteractiveRequiresPassword(t *testing.T) {
	withAppDeps(t)
	getenv = func(string) string { return "" }
	called := false
	runBridge = func(context.Context, string, bridge.Request, any) error {
		called = true
		return nil
	}

	cmd := newAuthLoginCmd()
	cmd.Flags().Bool(nonInteractiveFlag, false, "")
	cmd.SetArgs([]string{"--non-interactive", "--server", "http://localhost:5006", "--budget", "sync-id"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "server password is required in non-interactive mode") {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("bridge should not be called when password is missing")
	}
}

func TestAuthLogin_NonInteractiveUsesEnvPassword(t *testing.T) {
	withAppDeps(t)
	getenv = func(key string) string {
		if key == serverPasswordEnv {
			return "env-secret"
		}
		return ""
	}
	saveConfig = func(*config.Config) error { return nil }

	called := false
	runBridge = func(_ context.Context, op string, req bridge.Request, out any) error {
		called = true
		if op != "auth-check" {
			t.Fatalf("unexpected op: %s", op)
		}
		cfg, ok := req.Config.(*config.Config)
		if !ok {
			t.Fatalf("unexpected config type: %T", req.Config)
		}
		if cfg.Password != "env-secret" {
			t.Fatalf("unexpected password in config: %+v", cfg)
		}
		return nil
	}

	cmd := newAuthLoginCmd()
	cmd.Flags().Bool(nonInteractiveFlag, false, "")
	cmd.SetArgs([]string{"--non-interactive", "--server", "http://localhost:5006", "--budget", "sync-id"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected bridge to be called")
	}
}

func TestAuthCheck_NoSideEffects(t *testing.T) {
	withAppDeps(t)
	loadConfig = func() (*config.Config, error) {
		return &config.Config{ServerURL: "http://localhost:5006", BudgetID: "sync-id", Password: "secret"}, nil
	}
	saved := false
	saveConfig = func(*config.Config) error {
		saved = true
		return nil
	}
	called := false
	runBridge = func(_ context.Context, op string, req bridge.Request, out any) error {
		called = true
		if op != "auth-check" {
			t.Fatalf("unexpected op: %s", op)
		}
		if _, ok := req.Config.(*config.Config); !ok {
			t.Fatalf("unexpected config type: %T", req.Config)
		}
		resp := out.(*bridge.AuthCheckResponse)
		resp.OK = true
		return nil
	}

	cmd := newAuthCheckCmd()
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected bridge to be called")
	}
	if saved {
		t.Fatal("auth check must not persist config")
	}
}

func TestAuthCheck_AgentJSONDeterministic(t *testing.T) {
	withAppDeps(t)
	loadConfig = func() (*config.Config, error) { return &config.Config{}, nil }
	runBridge = func(_ context.Context, _ string, _ bridge.Request, out any) error {
		out.(*bridge.AuthCheckResponse).OK = true
		return nil
	}
	var got any
	printJSON = func(v any) error {
		got = v
		return nil
	}

	cmd := newAuthCheckCmd()
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
	if data["authenticated"] != true || data["message"] != "Credentials are valid" {
		t.Fatalf("unexpected auth-check payload: %#v", data)
	}
}

func TestAuthLogin_PasswordStdinOverridesEnv(t *testing.T) {
	withAppDeps(t)
	getenv = func(key string) string {
		if key == serverPasswordEnv {
			return "env-secret"
		}
		return ""
	}
	saveConfig = func(*config.Config) error { return nil }

	runBridge = func(_ context.Context, _ string, req bridge.Request, _ any) error {
		cfg, ok := req.Config.(*config.Config)
		if !ok {
			t.Fatalf("unexpected config type: %T", req.Config)
		}
		if cfg.Password != "stdin-secret" {
			t.Fatalf("expected stdin-secret, got %+v", cfg)
		}
		return nil
	}

	cmd := newAuthLoginCmd()
	cmd.Flags().Bool(nonInteractiveFlag, false, "")
	cmd.SetIn(strings.NewReader("stdin-secret\n"))
	cmd.SetArgs([]string{"--non-interactive", "--server", "http://localhost:5006", "--budget", "sync-id", "--password-stdin"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
