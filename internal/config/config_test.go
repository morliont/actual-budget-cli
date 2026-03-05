package config

import "testing"

func TestConfigStruct(t *testing.T) {
	c := &Config{ServerURL: "http://localhost:5006", BudgetID: "id", Password: "pw"}
	if c.ServerURL == "" || c.BudgetID == "" || c.Password == "" {
		t.Fatal("expected required fields")
	}
}
