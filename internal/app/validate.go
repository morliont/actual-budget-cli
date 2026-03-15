package app

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func validateServerURL(raw string) error {
	clean := strings.TrimSpace(raw)
	if clean == "" {
		return fmt.Errorf("server URL is required (example: --server http://localhost:5006)")
	}

	u, err := url.ParseRequestURI(clean)
	if err != nil {
		return fmt.Errorf("invalid server URL %q: use full URL with scheme and host (example: http://localhost:5006)", clean)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("invalid server URL %q: scheme must be http or https", clean)
	}
	if u.Host == "" {
		return fmt.Errorf("invalid server URL %q: host is required (example: http://localhost:5006)", clean)
	}
	return nil
}

func validateDate(raw, name string) error {
	clean := strings.TrimSpace(raw)
	if _, err := time.Parse("2006-01-02", clean); err != nil {
		return fmt.Errorf("invalid --%s value %q: expected YYYY-MM-DD (example: --%s 2026-01-31)", name, raw, name)
	}
	return nil
}

func validateMonth(raw, name string) error {
	clean := strings.TrimSpace(raw)
	if clean == "" {
		return fmt.Errorf("--%s is required (expected YYYY-MM)", name)
	}
	if _, err := time.Parse("2006-01", clean); err != nil {
		return fmt.Errorf("invalid --%s value %q: expected YYYY-MM (example: --%s 2026-03)", name, raw, name)
	}
	return nil
}

func validateLimit(limit int) error {
	if limit <= 0 {
		return fmt.Errorf("invalid --limit value %s: must be greater than 0", strconv.Itoa(limit))
	}
	return nil
}

func validateDateRange(from, to string) error {
	fromDate, err := time.Parse("2006-01-02", from)
	if err != nil {
		return err
	}
	toDate, err := time.Parse("2006-01-02", to)
	if err != nil {
		return err
	}
	if fromDate.After(toDate) {
		return fmt.Errorf("invalid date range: --from (%s) cannot be after --to (%s)", from, to)
	}
	return nil
}
