package app

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

func validateServerURL(raw string) error {
	u, err := url.ParseRequestURI(strings.TrimSpace(raw))
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("invalid server URL: scheme must be http or https")
	}
	if u.Host == "" {
		return fmt.Errorf("invalid server URL: host is required")
	}
	return nil
}

func validateDate(raw, name string) error {
	if _, err := time.Parse("2006-01-02", raw); err != nil {
		return fmt.Errorf("invalid --%s date %q: expected YYYY-MM-DD", name, raw)
	}
	return nil
}
