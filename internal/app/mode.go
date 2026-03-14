package app

import "github.com/spf13/cobra"

const (
	agentJSONFlag      = "agent-json"
	nonInteractiveFlag = "non-interactive"
)

func useAgentJSON(cmd *cobra.Command) bool {
	v, err := cmd.Flags().GetBool(agentJSONFlag)
	if err == nil && v {
		return true
	}
	v, err = cmd.InheritedFlags().GetBool(agentJSONFlag)
	return err == nil && v
}

func useNonInteractive(cmd *cobra.Command) bool {
	v, err := cmd.Flags().GetBool(nonInteractiveFlag)
	if err == nil && v {
		return true
	}
	v, err = cmd.InheritedFlags().GetBool(nonInteractiveFlag)
	return err == nil && v
}
