package app

import (
	"github.com/morliont/actual-budget-cli/internal/bridge"
	"github.com/morliont/actual-budget-cli/internal/config"
	"github.com/morliont/actual-budget-cli/internal/output"
)

var (
	loadConfig = config.Load
	runBridge  = bridge.Run
	printJSON  = output.PrintJSON
	printTable = output.PrintTable
)
