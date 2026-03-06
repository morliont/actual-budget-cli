package bridge

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

//go:embed actual-bridge.mjs
var bridgeScript embed.FS

const (
	defaultTimeout = 30 * time.Second
	timeoutEnvVar  = "ACTUAL_CLI_BRIDGE_TIMEOUT"
)

type Request struct {
	Config any            `json:"config"`
	Args   map[string]any `json:"args,omitempty"`
}

func Run(parent context.Context, op string, req Request, out any) error {
	timeout, err := timeoutFromEnv()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	payload, err := json.Marshal(req)
	if err != nil {
		return err
	}

	scriptPath, cleanup, err := materializeBridgeScript()
	if err != nil {
		return err
	}
	defer cleanup()

	cmd := exec.CommandContext(ctx, "node", scriptPath, op)
	cmd.Stdin = bytes.NewReader(payload)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("request timed out after %s; increase %s (example: export %s=60s)", timeout, timeoutEnvVar, timeoutEnvVar)
		}

		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg != "" {
			return fmt.Errorf("%s", bridgeUserMessage(stderrMsg))
		}

		if errors.Is(err, exec.ErrNotFound) {
			return fmt.Errorf("bridge runtime unavailable: Node.js is not installed or not in PATH")
		}
		return fmt.Errorf("failed to execute bridge process: %w", err)
	}

	if err := json.Unmarshal(stdout.Bytes(), out); err != nil {
		return fmt.Errorf("invalid bridge response: %w", err)
	}
	return nil
}

func timeoutFromEnv() (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(timeoutEnvVar))
	if raw == "" {
		return defaultTimeout, nil
	}
	v, err := time.ParseDuration(raw)
	if err == nil {
		if v <= 0 {
			return 0, fmt.Errorf("invalid %s=%q: value must be greater than 0", timeoutEnvVar, raw)
		}
		return v, nil
	}
	seconds, secErr := strconv.Atoi(raw)
	if secErr != nil || seconds <= 0 {
		return 0, fmt.Errorf("invalid %s=%q; use Go duration (e.g. 45s, 2m) or positive integer seconds", timeoutEnvVar, raw)
	}
	return time.Duration(seconds) * time.Second, nil
}

func bridgeUserMessage(stderrMsg string) string {
	lower := strings.ToLower(stderrMsg)
	if strings.Contains(lower, "econnrefused") || strings.Contains(lower, "enotfound") || strings.Contains(lower, "fetch failed") || strings.Contains(lower, "network") {
		return "network error while contacting Actual server; verify --server URL, server availability, and connectivity"
	}
	if strings.Contains(lower, "etimedout") || strings.Contains(lower, "timeout") {
		return fmt.Sprintf("request timed out contacting Actual server; retry or increase %s", timeoutEnvVar)
	}
	return fmt.Sprintf("bridge error: %s", stderrMsg)
}

func materializeBridgeScript() (string, func(), error) {
	content, err := bridgeScript.ReadFile("actual-bridge.mjs")
	if err != nil {
		return "", nil, fmt.Errorf("read embedded bridge script: %w", err)
	}

	tmp, err := os.CreateTemp("", "actual-bridge-*.mjs")
	if err != nil {
		return "", nil, fmt.Errorf("create bridge script temp file: %w", err)
	}
	path := tmp.Name()

	cleanup := func() {
		_ = os.Remove(path)
	}

	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		cleanup()
		return "", nil, fmt.Errorf("secure bridge script temp file: %w", err)
	}
	if _, err := tmp.Write(content); err != nil {
		_ = tmp.Close()
		cleanup()
		return "", nil, fmt.Errorf("write bridge script temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("close bridge script temp file: %w", err)
	}

	return path, cleanup, nil
}
