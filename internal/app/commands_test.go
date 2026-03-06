package app

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/morliont/actual-budget-cli/internal/bridge"
	"github.com/morliont/actual-budget-cli/internal/config"
)

func withAppDeps(t *testing.T) {
	t.Helper()
	origLoadConfig := loadConfig
	origRunBridge := runBridge
	origPrintJSON := printJSON
	origPrintTable := printTable
	t.Cleanup(func() {
		loadConfig = origLoadConfig
		runBridge = origRunBridge
		printJSON = origPrintJSON
		printTable = origPrintTable
	})
}

func TestTransactionsList_ValidationBeforeBridge(t *testing.T) {
	withAppDeps(t)
	loadConfig = func() (*config.Config, error) { return &config.Config{}, nil }
	called := false
	runBridge = func(context.Context, string, bridge.Request, any) error {
		called = true
		return nil
	}

	cmd := newTransactionsListCmd()
	cmd.SetArgs([]string{"--from", "03-01-2026"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid --from value") {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("bridge should not be called when input is invalid")
	}
}

func TestTransactionsList_LimitValidationBeforeBridge(t *testing.T) {
	withAppDeps(t)
	loadConfig = func() (*config.Config, error) { return &config.Config{}, nil }
	called := false
	runBridge = func(context.Context, string, bridge.Request, any) error {
		called = true
		return nil
	}

	cmd := newTransactionsListCmd()
	cmd.SetArgs([]string{"--limit", "0"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid --limit value 0") {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("bridge should not be called when limit is invalid")
	}
}

func TestTransactionsList_DefaultArgsPassedToBridge(t *testing.T) {
	withAppDeps(t)
	loadConfig = func() (*config.Config, error) { return &config.Config{}, nil }
	printTable = func(headers []string, rows [][]string) {}

	var gotArgs bridge.TransactionsListArgs
	runBridge = func(_ context.Context, op string, req bridge.Request, out any) error {
		if op != "transactions-list" {
			return errors.New("unexpected op")
		}
		args, ok := req.Args.(bridge.TransactionsListArgs)
		if !ok {
			return errors.New("args are not typed transactions args")
		}
		gotArgs = args
		resp := out.(*bridge.TransactionsListResponse)
		resp.Transactions = []json.RawMessage{json.RawMessage(`{"date":"2026-03-01","account":"Checking","payee_name":"Coffee","amount":-350,"notes":"Latte"}`)}
		return nil
	}

	cmd := newTransactionsListCmd()
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := bridge.TransactionsListArgs{From: "1900-01-01", To: "2999-12-31", Limit: 100}
	if !reflect.DeepEqual(gotArgs, want) {
		t.Fatalf("unexpected bridge args: got %+v want %+v", gotArgs, want)
	}
}

func TestAccountsList_TableDecodeFailure(t *testing.T) {
	withAppDeps(t)
	loadConfig = func() (*config.Config, error) { return &config.Config{}, nil }
	runBridge = func(_ context.Context, _ string, _ bridge.Request, out any) error {
		resp := out.(*bridge.AccountsListResponse)
		resp.Accounts = []json.RawMessage{json.RawMessage(`{"id":`)}
		return nil
	}

	cmd := newAccountsListCmd()
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected decode error")
	}
	if !strings.Contains(err.Error(), "invalid account payload") {
		t.Fatalf("unexpected error: %v", err)
	}
}
