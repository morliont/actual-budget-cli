package bridge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
)

type Request struct {
	Config any            `json:"config"`
	Args   map[string]any `json:"args,omitempty"`
}

func Run(op string, req Request, out any) error {
	payload, err := json.Marshal(req)
	if err != nil {
		return err
	}
	script := filepath.Join("internal", "bridge", "actual-bridge.mjs")
	cmd := exec.Command("node", script, op, string(payload))
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return fmt.Errorf("bridge error: %s", stderr.String())
		}
		return fmt.Errorf("bridge failed: %w", err)
	}
	if err := json.Unmarshal(stdout.Bytes(), out); err != nil {
		return fmt.Errorf("invalid bridge response: %w", err)
	}
	return nil
}
