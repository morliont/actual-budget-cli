package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/morliont/actual-budget-cli/internal/bridge"
	"github.com/morliont/actual-budget-cli/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newInitCmd() *cobra.Command {
	var serverURL, budgetID, password, budgetPassword string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Interactive first-run setup wizard",
		Long:  "Configure server URL, budget sync ID, password, test connection, and save local config.",
		Example: `  actual-cli init
  actual-cli init --server http://localhost:5006 --budget <SYNC_ID> --password "$ACTUAL_PASSWORD"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			isTTY := term.IsTerminal(int(os.Stdin.Fd()))
			reader := bufio.NewReader(os.Stdin)

			if serverURL == "" {
				if !isTTY {
					return fmt.Errorf("--server is required in non-interactive mode")
				}
				fmt.Print("Actual server URL: ")
				txt, _ := reader.ReadString('\n')
				serverURL = strings.TrimSpace(txt)
			}
			if budgetID == "" {
				if !isTTY {
					return fmt.Errorf("--budget is required in non-interactive mode")
				}
				fmt.Print("Budget Sync ID: ")
				txt, _ := reader.ReadString('\n')
				budgetID = strings.TrimSpace(txt)
			}
			if password == "" {
				if !isTTY {
					return fmt.Errorf("--password is required in non-interactive mode")
				}
				fmt.Print("Server password: ")
				pw, err := term.ReadPassword(int(os.Stdin.Fd()))
				fmt.Println()
				if err != nil {
					return err
				}
				password = string(pw)
			}

			if err := validateServerURL(serverURL); err != nil {
				return err
			}
			if strings.TrimSpace(budgetID) == "" {
				return fmt.Errorf("budget sync ID is required")
			}
			if strings.TrimSpace(password) == "" {
				return fmt.Errorf("password is required")
			}

			cfg := &config.Config{
				ServerURL:      strings.TrimSpace(serverURL),
				Password:       password,
				BudgetID:       strings.TrimSpace(budgetID),
				BudgetPassword: budgetPassword,
			}

			fmt.Print("Testing connection... ")
			if err := runBridge(cmd.Context(), "auth-check", bridge.Request{Config: cfg}, &bridge.AuthCheckResponse{}); err != nil {
				fmt.Println("failed")
				return fmt.Errorf("connection test failed: %w", err)
			}
			fmt.Println("ok")

			if err := config.Save(cfg); err != nil {
				return err
			}
			fmt.Println("Saved config to ~/.config/actual-cli/config.json")
			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server", "", "Actual server URL (http:// or https://)")
	cmd.Flags().StringVar(&budgetID, "budget", "", "Budget sync ID")
	cmd.Flags().StringVar(&password, "password", "", "Actual server password")
	cmd.Flags().StringVar(&budgetPassword, "budget-password", "", "Budget encryption password (optional)")

	return cmd
}
