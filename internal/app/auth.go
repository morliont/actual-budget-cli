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

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "auth", Short: "Authentication commands"}
	cmd.AddCommand(newAuthLoginCmd())
	return cmd
}

func newAuthLoginCmd() *cobra.Command {
	var serverURL, budgetID, budgetPassword string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login and store local config",
		RunE: func(cmd *cobra.Command, args []string) error {
			reader := bufio.NewReader(os.Stdin)

			if serverURL == "" {
				fmt.Print("Actual server URL: ")
				txt, _ := reader.ReadString('\n')
				serverURL = strings.TrimSpace(txt)
			}
			if budgetID == "" {
				fmt.Print("Budget Sync ID: ")
				txt, _ := reader.ReadString('\n')
				budgetID = strings.TrimSpace(txt)
			}
			if err := validateServerURL(serverURL); err != nil {
				return err
			}
			if strings.TrimSpace(budgetID) == "" {
				return fmt.Errorf("budget sync ID is required")
			}

			fmt.Print("Server password: ")
			pw, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println()
			if err != nil {
				return err
			}

			cfg := &config.Config{ServerURL: serverURL, Password: string(pw), BudgetID: budgetID, BudgetPassword: budgetPassword}
			var check map[string]any
			if err := bridge.Run(cmd.Context(), "auth-check", bridge.Request{Config: cfg}, &check); err != nil {
				return fmt.Errorf("auth failed: %w", err)
			}
			if err := config.Save(cfg); err != nil {
				return err
			}
			fmt.Println("Saved config to ~/.config/actual-cli/config.json")
			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server", "", "Actual server URL")
	cmd.Flags().StringVar(&budgetID, "budget", "", "Budget sync ID")
	cmd.Flags().StringVar(&budgetPassword, "budget-password", "", "Budget encryption password (optional)")
	return cmd
}
