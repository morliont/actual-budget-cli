package app

import (
	"context"
	"strings"
	"testing"

	"github.com/morliont/actual-budget-cli/internal/bridge"
	"github.com/morliont/actual-budget-cli/internal/config"
)

func TestReadOnlyMode_BlocksMutatingCommand(t *testing.T) {
	withAppDeps(t)
	getenv = func(key string) string {
		if key == readOnlyEnv {
			return "true"
		}
		if key == serverPasswordEnv {
			return "pw"
		}
		return ""
	}

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"auth", "login", "--server", "http://localhost:5006", "--budget", "sync-id"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected read-only error")
	}
	if !strings.Contains(err.Error(), "read-only mode blocked mutating command") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReadOnlyMode_AllowsReadCommand(t *testing.T) {
	withAppDeps(t)
	getenv = func(key string) string {
		if key == readOnlyEnv {
			return "true"
		}
		return ""
	}
	loadConfig = func() (*config.Config, error) { return &config.Config{}, nil }
	called := false
	runBridge = func(_ context.Context, op string, req bridge.Request, out any) error {
		called = true
		if op != "auth-check" {
			t.Fatalf("unexpected op: %s", op)
		}
		out.(*bridge.AuthCheckResponse).OK = true
		return nil
	}

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"auth", "check"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected read command to execute")
	}
}
