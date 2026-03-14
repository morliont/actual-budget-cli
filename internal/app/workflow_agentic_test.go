package app

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/morliont/actual-budget-cli/internal/bridge"
	"github.com/morliont/actual-budget-cli/internal/config"
	"github.com/morliont/actual-budget-cli/internal/output"
)

type authCheckFlowFixture struct {
	Name            string `json:"name"`
	BridgeError     string `json:"bridgeError"`
	WantErrContains string `json:"wantErrContains"`
	WantRetryable   bool   `json:"wantRetryable"`
}

type authRetryFlowFixture struct {
	Name         string   `json:"name"`
	Args         []string `json:"args"`
	BridgeErrors []string `json:"bridgeErrors"`
	MaxAttempts  int      `json:"maxAttempts"`
	WantAttempts int      `json:"wantAttempts"`
	WantSuccess  bool     `json:"wantSuccess"`
}

type invalidInputFlowFixture struct {
	Name            string   `json:"name"`
	Args            []string `json:"args"`
	WantErrContains string   `json:"wantErrContains"`
	WantBridgeCalls int      `json:"wantBridgeCalls"`
}

func TestAgenticWorkflow_AuthCheckSuccessAndFail(t *testing.T) {
	withAppDeps(t)
	loadConfig = func() (*config.Config, error) { return &config.Config{}, nil }

	var fixtures []authCheckFlowFixture
	readFixture(t, "tests/fixtures/auth_check_flow.json", &fixtures)

	for _, tc := range fixtures {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			withAppDeps(t)
			loadConfig = func() (*config.Config, error) { return &config.Config{}, nil }
			runBridge = func(_ context.Context, _ string, _ bridge.Request, out any) error {
				if tc.BridgeError != "" {
					return errors.New(tc.BridgeError)
				}
				out.(*bridge.AuthCheckResponse).OK = true
				return nil
			}

			cmd := NewRootCmd()
			cmd.SetArgs([]string{"auth", "check"})
			err := cmd.Execute()
			if tc.WantErrContains == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tc.WantErrContains) {
				t.Fatalf("unexpected error: %v", err)
			}
			mapped := output.MapError(err)
			if mapped.Retryable != tc.WantRetryable {
				t.Fatalf("unexpected retryability: got %v want %v", mapped.Retryable, tc.WantRetryable)
			}
		})
	}
}

func TestAgenticWorkflow_TransientNetworkFailureThenRetrySuccess(t *testing.T) {
	withAppDeps(t)
	loadConfig = func() (*config.Config, error) { return &config.Config{}, nil }

	var tc authRetryFlowFixture
	readFixture(t, "tests/fixtures/auth_retry_flow.json", &tc)

	attempt := 0
	runBridge = func(_ context.Context, _ string, _ bridge.Request, out any) error {
		if attempt >= len(tc.BridgeErrors) {
			return errors.New("unexpected extra attempt")
		}
		errMsg := tc.BridgeErrors[attempt]
		attempt++
		if errMsg != "" {
			return errors.New(errMsg)
		}
		resp, ok := out.(*bridge.AccountsListResponse)
		if !ok {
			return errors.New("unexpected response type")
		}
		resp.Accounts = nil
		return nil
	}

	err := runCommandWithRetry(tc.Args, tc.MaxAttempts)
	if tc.WantSuccess && err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if attempt != tc.WantAttempts {
		t.Fatalf("unexpected attempts: got %d want %d", attempt, tc.WantAttempts)
	}
}

func TestAgenticWorkflow_InvalidInputShortCircuitsBridge(t *testing.T) {
	withAppDeps(t)
	loadConfig = func() (*config.Config, error) { return &config.Config{}, nil }

	var tc invalidInputFlowFixture
	readFixture(t, "tests/fixtures/invalid_input_flow.json", &tc)

	bridgeCalls := 0
	runBridge = func(_ context.Context, _ string, _ bridge.Request, _ any) error {
		bridgeCalls++
		return nil
	}

	cmd := NewRootCmd()
	cmd.SetArgs(tc.Args)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), tc.WantErrContains) {
		t.Fatalf("unexpected error: %v", err)
	}
	if bridgeCalls != tc.WantBridgeCalls {
		t.Fatalf("unexpected bridge calls: got %d want %d", bridgeCalls, tc.WantBridgeCalls)
	}
}

func runCommandWithRetry(args []string, maxAttempts int) error {
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	var lastErr error
	for i := 0; i < maxAttempts; i++ {
		cmd := NewRootCmd()
		cmd.SetArgs(args)
		err := cmd.Execute()
		if err == nil {
			return nil
		}
		lastErr = err
		if !output.MapError(err).Retryable {
			return err
		}
	}
	return lastErr
}

func readFixture(t *testing.T, rel string, out any) {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test file path")
	}
	absPath := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", rel))
	data, err := os.ReadFile(absPath)
	if err != nil {
		t.Fatalf("read fixture %s: %v", rel, err)
	}
	if err := json.Unmarshal(data, out); err != nil {
		t.Fatalf("parse fixture %s: %v", rel, err)
	}
}
