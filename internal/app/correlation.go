package app

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const correlationIDFlag = "correlation-id"

func correlationIDForCmd(cmd *cobra.Command) string {
	v, err := cmd.Flags().GetString(correlationIDFlag)
	if err == nil && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	v, err = cmd.InheritedFlags().GetString(correlationIDFlag)
	if err == nil && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return strings.TrimSpace(os.Getenv("ACTUAL_CLI_CORRELATION_ID"))
}
