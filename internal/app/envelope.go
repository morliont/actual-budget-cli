package app

import (
	"github.com/morliont/actual-budget-cli/internal/output"
	"github.com/spf13/cobra"
)

func successEnvelope(cmd *cobra.Command, data any) output.Envelope {
	return output.SuccessWithCorrelationID(data, correlationIDForCmd(cmd))
}
