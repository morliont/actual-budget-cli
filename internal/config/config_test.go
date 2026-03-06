package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigStruct(t *testing.T) {
	c := &Config{ServerURL: "http://localhost:5006", BudgetID: "id", Password: "pw"}
	if c.ServerURL == "" || c.BudgetID == "" || c.Password == "" {
		t.Fatal("expected required fields")
	}
}

func TestSaveCreatesSecurePermissions(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg := &Config{
		ServerURL: "http://localhost:5006",
		BudgetID:  "budget-id",
		Password:  "super-secret",
	}
	if err := Save(cfg); err != nil {
		t.Fatalf("save should succeed: %v", err)
	}

	configDir := filepath.Join(home, ".config", "actual-cli")
	configFile := filepath.Join(configDir, "config.json")
	dataDir := filepath.Join(home, ".local", "share", "actual-cli")

	stDir, err := os.Stat(configDir)
	if err != nil {
		t.Fatalf("config dir should exist: %v", err)
	}
	if stDir.Mode().Perm() != 0o700 {
		t.Fatalf("expected config dir 0700, got %o", stDir.Mode().Perm())
	}

	stFile, err := os.Stat(configFile)
	if err != nil {
		t.Fatalf("config file should exist: %v", err)
	}
	if stFile.Mode().Perm() != 0o600 {
		t.Fatalf("expected config file 0600, got %o", stFile.Mode().Perm())
	}

	stDataDir, err := os.Stat(dataDir)
	if err != nil {
		t.Fatalf("data dir should exist: %v", err)
	}
	if stDataDir.Mode().Perm() != 0o700 {
		t.Fatalf("expected data dir 0700, got %o", stDataDir.Mode().Perm())
	}
}
