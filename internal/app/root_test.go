package app

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommandMetadata(t *testing.T) {
	cmd := NewRootCmd()
	if cmd.Use != "actual-cli" {
		t.Fatalf("unexpected root use: %s", cmd.Use)
	}
	if cmd.Short == "" || cmd.Long == "" {
		t.Fatalf("expected short and long descriptions to be set")
	}
	if cmd.Version == "" {
		t.Fatalf("expected version metadata to be set")
	}

	for _, name := range []string{"init", "accounts", "auth", "budgets", "transactions"} {
		if _, _, err := cmd.Find([]string{name}); err != nil {
			t.Fatalf("expected %q subcommand: %v", name, err)
		}
	}
}

func TestRootHelpIncludesMajorCommands(t *testing.T) {
	cmd := NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute --help: %v", err)
	}

	help := buf.String()
	for _, section := range []string{"init", "auth", "accounts", "transactions", "budgets"} {
		if !strings.Contains(help, section) {
			t.Fatalf("expected help output to include %q", section)
		}
	}
}

func TestRootVersionFlag(t *testing.T) {
	cmd := NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute --version: %v", err)
	}

	out := strings.TrimSpace(buf.String())
	if out == "" {
		t.Fatal("expected version output")
	}
	if !strings.Contains(out, "commit") {
		t.Fatalf("expected version output to include build metadata, got %q", out)
	}
}
