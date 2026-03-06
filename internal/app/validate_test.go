package app

import (
	"strings"
	"testing"
)

func TestValidateServerURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{name: "valid http", url: "http://localhost:5006"},
		{name: "valid https", url: "https://actual.example.com"},
		{name: "missing scheme", url: "actual.example.com", wantErr: true},
		{name: "empty", url: "   ", wantErr: true},
		{name: "unsupported scheme", url: "ftp://example.com", wantErr: true},
		{name: "missing host", url: "https:///", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateServerURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateServerURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateServerURLMessageIsActionable(t *testing.T) {
	err := validateServerURL("actual.example.com")
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "scheme and host") {
		t.Fatalf("expected actionable server URL message, got %q", err.Error())
	}
}

func TestValidateDate(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "valid", value: "2026-03-01"},
		{name: "invalid format", value: "01-03-2026", wantErr: true},
		{name: "invalid day", value: "2026-02-30", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDate(tt.value, "from")
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateDate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateLimit(t *testing.T) {
	if err := validateLimit(1); err != nil {
		t.Fatalf("expected valid limit: %v", err)
	}
	if err := validateLimit(0); err == nil {
		t.Fatal("expected validation error for zero limit")
	}
}

func TestValidateDateRange(t *testing.T) {
	if err := validateDateRange("2026-01-01", "2026-01-31"); err != nil {
		t.Fatalf("expected valid date range: %v", err)
	}
	err := validateDateRange("2026-02-01", "2026-01-31")
	if err == nil {
		t.Fatal("expected error when --from is after --to")
	}
	if !strings.Contains(err.Error(), "cannot be after") {
		t.Fatalf("unexpected date range error: %q", err.Error())
	}
}
