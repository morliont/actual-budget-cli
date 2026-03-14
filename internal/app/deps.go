package app

import (
	"os"

	"github.com/morliont/actual-budget-cli/internal/bridge"
	"github.com/morliont/actual-budget-cli/internal/config"
	"github.com/morliont/actual-budget-cli/internal/output"
	"golang.org/x/term"
)

var (
	loadConfig   = config.Load
	saveConfig   = config.Save
	runBridge    = bridge.Run
	printJSON    = output.PrintJSON
	printTable   = output.PrintTable
	readPassword = term.ReadPassword
	getenv       = os.Getenv
)
