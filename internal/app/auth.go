package app

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/morliont/actual-budget-cli/internal/bridge"
	"github.com/morliont/actual-budget-cli/internal/config"
	"github.com/spf13/cobra"
)

const serverPasswordEnv = "ACTUAL_CLI_PASSWORD"

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with an Actual server",
		Long:  "Authenticate and store local credentials used by other commands.",
	}
	cmd.AddCommand(newAuthLoginCmd(), newAuthCheckCmd())
	return cmd
}

func newAuthLoginCmd() *cobra.Command {
	var serverURL, budgetID, budgetPassword string
	var passwordFromStdin bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in and save local config",
		Long:  "Prompt for credentials (when missing) and store them in local config.",
		Example: `  actual-cli auth login --server http://localhost:5006 --budget <SYNC_ID>
  actual-cli auth login
  printf '%s\n' "$ACTUAL_PASSWORD" | actual-cli --non-interactive auth login --server http://localhost:5006 --budget <SYNC_ID> --password-stdin`,
		RunE: func(cmd *cobra.Command, args []string) error {
			nonInteractive := useNonInteractive(cmd)
			reader := bufio.NewReader(cmd.InOrStdin())

			if serverURL == "" {
				if nonInteractive {
					return fmt.Errorf("--server is required in non-interactive mode")
				}
				fmt.Print("Actual server URL: ")
				txt, _ := reader.ReadString('\n')
				serverURL = strings.TrimSpace(txt)
			}
			if budgetID == "" {
				if nonInteractive {
					return fmt.Errorf("--budget is required in non-interactive mode")
				}
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

			password, err := resolveServerPassword(cmd, passwordFromStdin, nonInteractive)
			if err != nil {
				return err
			}

			cfg := &config.Config{ServerURL: strings.TrimSpace(serverURL), Password: password, BudgetID: strings.TrimSpace(budgetID), BudgetPassword: budgetPassword}
			var check bridge.AuthCheckResponse
			if err := runBridge(cmd.Context(), "auth-check", bridge.Request{Config: cfg}, &check); err != nil {
				return fmt.Errorf("auth failed: %w", err)
			}
			if err := saveConfig(cfg); err != nil {
				return err
			}
			if useAgentJSON(cmd) {
				return printJSON(successEnvelope(cmd, map[string]any{"message": "Saved config to ~/.config/actual-cli/config.json"}))
			}
			fmt.Println("Saved config to ~/.config/actual-cli/config.json")
			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server", "", "Actual server URL (http:// or https://)")
	cmd.Flags().StringVar(&budgetID, "budget", "", "Budget sync ID")
	cmd.Flags().StringVar(&budgetPassword, "budget-password", "", "Budget encryption password (optional)")
	cmd.Flags().BoolVar(&passwordFromStdin, "password-stdin", false, "Read server password from stdin (preferred for automation)")
	markMutating(cmd)
	return cmd
}

func newAuthCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "check",
		Aliases: []string{"status"},
		Short:   "Validate saved credentials/session",
		Long:    "Validate configured Actual credentials/session without changing local config.",
		Example: `  actual-cli auth check
  actual-cli --agent-json --non-interactive auth check`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			var check bridge.AuthCheckResponse
			if err := runBridge(cmd.Context(), "auth-check", bridge.Request{Config: cfg}, &check); err != nil {
				return fmt.Errorf("auth failed: %w", err)
			}
			if useAgentJSON(cmd) {
				return printJSON(successEnvelope(cmd, map[string]any{"authenticated": true, "message": "Credentials are valid"}))
			}
			fmt.Println("Credentials are valid")
			return nil
		},
	}
	return cmd
}

func resolveServerPassword(cmd *cobra.Command, passwordFromStdin, nonInteractive bool) (string, error) {
	if passwordFromStdin {
		buf, err := io.ReadAll(cmd.InOrStdin())
		if err != nil {
			return "", fmt.Errorf("failed to read password from stdin: %w", err)
		}
		password := strings.TrimRight(string(buf), "\r\n")
		if strings.TrimSpace(password) == "" {
			return "", fmt.Errorf("--password-stdin provided but stdin was empty")
		}
		return password, nil
	}

	if envPassword := getenv(serverPasswordEnv); strings.TrimSpace(envPassword) != "" {
		return envPassword, nil
	}

	if nonInteractive {
		return "", fmt.Errorf("server password is required in non-interactive mode (use --password-stdin or %s)", serverPasswordEnv)
	}

	fmt.Print("Server password: ")
	pw, err := readPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return "", err
	}
	password := strings.TrimSpace(string(pw))
	if password == "" {
		return "", fmt.Errorf("server password is required")
	}
	return password, nil
}
