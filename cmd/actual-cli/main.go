package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/morliont/actual-budget-cli/internal/app"
	"github.com/morliont/actual-budget-cli/internal/output"
)

func main() {
	args := os.Args[1:]
	correlationID := correlationIDRequested(args)
	if err := app.NewRootCmd().Execute(); err != nil {
		if agentJSONRequested(args) {
			_ = output.PrintJSON(output.FailureWithCorrelationID(err, correlationID))
		} else {
			if correlationID != "" {
				fmt.Fprintf(os.Stderr, "Error: %v (correlation-id: %s)\n", err, correlationID)
			} else {
				fmt.Fprintln(os.Stderr, "Error:", err)
			}
		}
		os.Exit(1)
	}
}

func agentJSONRequested(args []string) bool {
	for _, arg := range args {
		if arg == "--agent-json" || strings.HasPrefix(arg, "--agent-json=") {
			return true
		}
	}
	return false
}

func correlationIDRequested(args []string) string {
	for i, arg := range args {
		if arg == "--correlation-id" && i+1 < len(args) {
			return strings.TrimSpace(args[i+1])
		}
		if strings.HasPrefix(arg, "--correlation-id=") {
			return strings.TrimSpace(strings.TrimPrefix(arg, "--correlation-id="))
		}
	}
	return strings.TrimSpace(os.Getenv("ACTUAL_CLI_CORRELATION_ID"))
}
