package main

import (
	"fmt"
	"os"

	"github.com/morliont/actual-budget-cli/internal/app"
)

func main() {
	if err := app.NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
