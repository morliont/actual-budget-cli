package app

import (
	"encoding/json"
	"fmt"

	"github.com/morliont/actual-budget-cli/internal/bridge"
	"github.com/spf13/cobra"
)

func newCategoriesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "categories",
		Short: "Work with categories",
		Long:  "List categories with group metadata.",
	}
	cmd.AddCommand(newCategoriesListCmd())
	return cmd
}

func newCategoriesListCmd() *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List categories",
		Long:  "List category metadata including group linkage and visibility flags.",
		Example: `  actual-cli categories list
  actual-cli categories list --json
  actual-cli categories list --agent-json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			var res bridge.CategoriesListResponse
			if err := runBridge(cmd.Context(), "categories-list", bridge.Request{Config: cfg}, &res); err != nil {
				return err
			}

			if useAgentJSON(cmd) {
				return printJSON(successEnvelope(cmd, map[string]any{"categories": res.Categories}))
			}
			if asJSON {
				return printJSON(res.Categories)
			}

			rows := make([][]string, 0, len(res.Categories))
			for _, raw := range res.Categories {
				var c bridge.CategoryRow
				if err := json.Unmarshal(raw, &c); err != nil {
					return fmt.Errorf("invalid category payload: %w", err)
				}

				archived := "n/a"
				if c.Archived != nil {
					if *c.Archived {
						archived = "yes"
					} else {
						archived = "no"
					}
				}
				hidden := "no"
				if c.Hidden {
					hidden = "yes"
				}

				rows = append(rows, []string{c.ID, c.Name, c.GroupID, c.GroupName, hidden, archived})
			}

			printTable([]string{"ID", "Category", "Group ID", "Group", "Hidden", "Archived"}, rows)
			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "Output JSON")
	return cmd
}
