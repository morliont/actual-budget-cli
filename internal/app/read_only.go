package app

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

const (
	readOnlyFlag      = "read-only"
	readOnlyEnv       = "ACTUAL_CLI_READ_ONLY"
	mutationAnnoKey   = "actual-cli/mutation"
	mutationAnnoValue = "true"
)

func markMutating(cmd *cobra.Command) {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[mutationAnnoKey] = mutationAnnoValue
}

func isMutatingCommand(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cmd.Annotations[mutationAnnoKey]), mutationAnnoValue)
}

func isReadOnlyMode(cmd *cobra.Command) bool {
	v, err := cmd.Flags().GetBool(readOnlyFlag)
	if err == nil {
		return v
	}
	v, err = cmd.InheritedFlags().GetBool(readOnlyFlag)
	if err == nil {
		return v
	}
	return readOnlyDefaultFromEnv()
}

func readOnlyDefaultFromEnv() bool {
	raw := strings.TrimSpace(getenv(readOnlyEnv))
	if raw == "" {
		return false
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return false
	}
	return v
}

func enforceReadOnly(cmd *cobra.Command) error {
	if !isReadOnlyMode(cmd) || !isMutatingCommand(cmd) {
		return nil
	}
	return fmt.Errorf("read-only mode blocked mutating command: %s (set --read-only=false or %s=false to allow)", cmd.CommandPath(), readOnlyEnv)
}
