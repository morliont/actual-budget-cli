package output

import "testing"

func TestMapErrorCoreCodes(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantCode  string
		retryable bool
	}{
		{name: "auth", err: testErr("auth failed: unauthorized"), wantCode: "AUTH_FAILED", retryable: false},
		{name: "network", err: testErr("network error while contacting Actual server"), wantCode: "NETWORK_UNREACHABLE", retryable: true},
		{name: "timeout", err: testErr("request timed out after 30s"), wantCode: "TIMEOUT", retryable: true},
		{name: "invalid", err: testErr("invalid --from value \"03-01-2026\""), wantCode: "INVALID_INPUT", retryable: false},
		{name: "internal", err: testErr("boom"), wantCode: "INTERNAL_ERROR", retryable: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapError(tt.err)
			if got.Code != tt.wantCode {
				t.Fatalf("unexpected code: got %s want %s", got.Code, tt.wantCode)
			}
			if got.Retryable != tt.retryable {
				t.Fatalf("unexpected retryable: got %v want %v", got.Retryable, tt.retryable)
			}
		})
	}
}

func TestFailureEnvelopeShape(t *testing.T) {
	env := Failure(testErr("request timed out"))
	if env.OK {
		t.Fatal("expected ok=false")
	}
	if env.Error == nil {
		t.Fatal("expected error payload")
	}
	if env.Error.Code == "" || env.Error.Message == "" {
		t.Fatalf("expected non-empty error fields, got %+v", env.Error)
	}
}

func TestEnvelopeCorrelationID(t *testing.T) {
	s := SuccessWithCorrelationID(map[string]any{"x": 1}, "corr-1")
	if s.Meta == nil || s.Meta.CorrelationID != "corr-1" {
		t.Fatalf("expected correlation id on success envelope, got %+v", s.Meta)
	}

	f := FailureWithCorrelationID(testErr("boom"), "corr-2")
	if f.Meta == nil || f.Meta.CorrelationID != "corr-2" {
		t.Fatalf("expected correlation id on failure envelope, got %+v", f.Meta)
	}
}

type testErr string

func (e testErr) Error() string { return string(e) }
