package bridge

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTimeoutFromEnv(t *testing.T) {
	t.Setenv(timeoutEnvVar, "")
	d, err := timeoutFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != defaultTimeout {
		t.Fatalf("expected default timeout %s, got %s", defaultTimeout, d)
	}

	t.Setenv(timeoutEnvVar, "45s")
	d, err = timeoutFromEnv()
	if err != nil {
		t.Fatalf("unexpected error for duration: %v", err)
	}
	if d != 45*time.Second {
		t.Fatalf("expected 45s, got %s", d)
	}

	t.Setenv(timeoutEnvVar, "12")
	d, err = timeoutFromEnv()
	if err != nil {
		t.Fatalf("unexpected error for seconds: %v", err)
	}
	if d != 12*time.Second {
		t.Fatalf("expected 12s, got %s", d)
	}

	t.Setenv(timeoutEnvVar, "0")
	if _, err := timeoutFromEnv(); err == nil {
		t.Fatal("expected error for zero timeout")
	}
}

func TestMaterializeBridgeScript(t *testing.T) {
	path, cleanup, err := materializeBridgeScript()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer cleanup()

	st, err := os.Stat(path)
	if err != nil {
		t.Fatalf("script path should exist: %v", err)
	}
	if st.Size() == 0 {
		t.Fatal("embedded script should not be empty")
	}
	if st.Mode().Perm() != 0o600 {
		t.Fatalf("expected 0600 script perms, got %o", st.Mode().Perm())
	}
}

func TestRunUsesStdinAndDoesNotLeakSecretsInArgv(t *testing.T) {
	t.Setenv(timeoutEnvVar, "5s")

	d := t.TempDir()
	argsPath := filepath.Join(d, "args.txt")
	stdinPath := filepath.Join(d, "stdin.txt")

	fakeNode := filepath.Join(d, "node")
	script := "#!/usr/bin/env sh\n" +
		"printf '%s\\n' \"$@\" > \"" + argsPath + "\"\n" +
		"cat > \"" + stdinPath + "\"\n" +
		"printf '{\"ok\":true}'\n"
	if err := os.WriteFile(fakeNode, []byte(script), 0o700); err != nil {
		t.Fatalf("write fake node: %v", err)
	}

	t.Setenv("PATH", d+string(os.PathListSeparator)+os.Getenv("PATH"))

	secret := "super-secret-password"
	var out map[string]any
	err := Run(context.Background(), "auth-check", Request{Config: map[string]any{"password": secret}}, &out)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	argsRaw, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatalf("read argv capture: %v", err)
	}
	args := strings.TrimSpace(string(argsRaw))
	if strings.Contains(args, secret) {
		t.Fatalf("secret leaked into argv: %q", args)
	}
	if !strings.Contains(args, "auth-check") {
		t.Fatalf("expected operation in argv, got %q", args)
	}

	stdinRaw, err := os.ReadFile(stdinPath)
	if err != nil {
		t.Fatalf("read stdin capture: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(stdinRaw, &payload); err != nil {
		t.Fatalf("stdin should carry JSON payload: %v", err)
	}
	cfg, _ := payload["config"].(map[string]any)
	if cfg["password"] != secret {
		t.Fatalf("expected secret in stdin payload")
	}
}

func TestRunReturnsBridgeStderrWithoutSecretEcho(t *testing.T) {
	t.Setenv(timeoutEnvVar, "5s")

	d := t.TempDir()
	fakeNode := filepath.Join(d, "node")
	script := "#!/usr/bin/env sh\n" +
		"cat >/dev/null\n" +
		"echo 'bridge exploded' 1>&2\n" +
		"exit 1\n"
	if err := os.WriteFile(fakeNode, []byte(script), 0o700); err != nil {
		t.Fatalf("write fake node: %v", err)
	}
	secret := "very-secret"
	t.Setenv("PATH", d+string(os.PathListSeparator)+os.Getenv("PATH"))

	var out map[string]any
	err := Run(context.Background(), "accounts-list", Request{Config: map[string]any{"password": secret}}, &out)
	if err == nil {
		t.Fatal("expected run error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "bridge error: bridge exploded") {
		t.Fatalf("unexpected error: %q", msg)
	}
	if strings.Contains(msg, secret) {
		t.Fatalf("secret leaked in error message: %q", msg)
	}
}

func TestBridgeUserMessageNetworkErrors(t *testing.T) {
	msg := bridgeUserMessage("request failed: ECONNREFUSED 127.0.0.1:5006")
	if !strings.Contains(msg, "network error") {
		t.Fatalf("expected network guidance, got %q", msg)
	}

	msg = bridgeUserMessage("connect ETIMEDOUT")
	if !strings.Contains(msg, timeoutEnvVar) {
		t.Fatalf("expected timeout env guidance, got %q", msg)
	}
}

func TestRunTimeoutMessage(t *testing.T) {
	t.Setenv(timeoutEnvVar, "100ms")

	d := t.TempDir()
	fakeNode := filepath.Join(d, "node")
	script := "#!/usr/bin/env sh\n" +
		"cat >/dev/null\n" +
		"sleep 1\n"
	if err := os.WriteFile(fakeNode, []byte(script), 0o700); err != nil {
		t.Fatalf("write fake node: %v", err)
	}
	t.Setenv("PATH", d+string(os.PathListSeparator)+os.Getenv("PATH"))

	var out map[string]any
	err := Run(context.Background(), "accounts-list", Request{Config: map[string]any{}}, &out)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "request timed out") {
		t.Fatalf("unexpected timeout error: %q", err.Error())
	}
	if !strings.Contains(err.Error(), timeoutEnvVar) {
		t.Fatalf("expected timeout env var mention, got %q", err.Error())
	}
}
