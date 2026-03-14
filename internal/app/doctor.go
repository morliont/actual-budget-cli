package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type doctorCheck struct {
	Name    string `json:"name"`
	OK      bool   `json:"ok"`
	Details string `json:"details,omitempty"`
}

var lookPath = exec.LookPath

func newDoctorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check environment readiness",
		Long:  "Run non-destructive checks for configuration, runtime dependencies, and local filesystem readiness.",
		RunE: func(cmd *cobra.Command, args []string) error {
			checks := runDoctorChecks()
			passed := 0
			for _, c := range checks {
				if c.OK {
					passed++
				}
			}
			ready := passed == len(checks)

			if useAgentJSON(cmd) {
				return printJSON(successEnvelope(cmd, map[string]any{
					"ready":  ready,
					"checks": checks,
					"summary": map[string]int{
						"passed": passed,
						"total":  len(checks),
					},
				}))
			}

			status := "READY"
			if !ready {
				status = "NOT_READY"
			}
			fmt.Printf("Doctor status: %s (%d/%d checks passed)\n", status, passed, len(checks))
			for _, c := range checks {
				label := "PASS"
				if !c.OK {
					label = "FAIL"
				}
				if strings.TrimSpace(c.Details) == "" {
					fmt.Printf("- [%s] %s\n", label, c.Name)
					continue
				}
				fmt.Printf("- [%s] %s: %s\n", label, c.Name, c.Details)
			}
			if ready {
				return nil
			}
			return fmt.Errorf("doctor found %d failing checks", len(checks)-passed)
		},
	}
	return cmd
}

func runDoctorChecks() []doctorCheck {
	checks := []doctorCheck{}

	cfgPath, cfgPathErr := configPathForDoctor()
	if cfgPathErr != nil {
		checks = append(checks, doctorCheck{Name: "config_path", OK: false, Details: cfgPathErr.Error()})
	} else {
		if _, err := os.Stat(cfgPath); err != nil {
			checks = append(checks, doctorCheck{Name: "config_present", OK: false, Details: fmt.Sprintf("missing config at %s; run 'actual-cli auth login'", cfgPath)})
		} else {
			checks = append(checks, doctorCheck{Name: "config_present", OK: true, Details: cfgPath})
		}
	}

	cfg, cfgErr := loadConfig()
	if cfgErr != nil {
		checks = append(checks, doctorCheck{Name: "config_load", OK: false, Details: cfgErr.Error()})
	} else {
		checks = append(checks, doctorCheck{Name: "config_load", OK: true})
		if err := validateServerURL(cfg.ServerURL); err != nil {
			checks = append(checks, doctorCheck{Name: "server_url", OK: false, Details: err.Error()})
		} else {
			checks = append(checks, doctorCheck{Name: "server_url", OK: true, Details: cfg.ServerURL})
		}
		if strings.TrimSpace(cfg.BudgetID) == "" {
			checks = append(checks, doctorCheck{Name: "budget_id", OK: false, Details: "budget sync ID is empty"})
		} else {
			checks = append(checks, doctorCheck{Name: "budget_id", OK: true, Details: cfg.BudgetID})
		}
		checks = append(checks, dirWritableCheck("data_dir_writable", cfg.DataDir))
	}

	nodePath, err := lookPath("node")
	if err != nil {
		checks = append(checks, doctorCheck{Name: "node_runtime", OK: false, Details: "Node.js executable not found in PATH"})
	} else {
		checks = append(checks, doctorCheck{Name: "node_runtime", OK: true, Details: nodePath})
	}

	return checks
}

func configPathForDoctor() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "actual-cli", "config.json"), nil
}

func dirWritableCheck(name, p string) doctorCheck {
	if strings.TrimSpace(p) == "" {
		return doctorCheck{Name: name, OK: false, Details: "path is empty"}
	}
	if err := os.MkdirAll(p, 0o700); err != nil {
		return doctorCheck{Name: name, OK: false, Details: err.Error()}
	}
	probe, err := os.CreateTemp(p, ".doctor-write-test-*")
	if err != nil {
		return doctorCheck{Name: name, OK: false, Details: err.Error()}
	}
	probePath := probe.Name()
	_ = probe.Close()
	_ = os.Remove(probePath)
	return doctorCheck{Name: name, OK: true, Details: p}
}

