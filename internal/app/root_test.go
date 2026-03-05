package app

import "testing"

func TestRootCommand(t *testing.T) {
	cmd := NewRootCmd()
	if cmd.Use != "actual-cli" {
		t.Fatalf("unexpected root use: %s", cmd.Use)
	}
}
