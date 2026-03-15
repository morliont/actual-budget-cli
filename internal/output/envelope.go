package output

import "strings"

type Envelope struct {
	OK    bool           `json:"ok"`
	Data  any            `json:"data,omitempty"`
	Error *EnvelopeError `json:"error,omitempty"`
	Meta  *EnvelopeMeta  `json:"meta,omitempty"`
}

type EnvelopeMeta struct {
	CorrelationID string `json:"correlationId,omitempty"`
}

type EnvelopeError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
}

func Success(data any) Envelope {
	return Envelope{OK: true, Data: data}
}

func SuccessWithCorrelationID(data any, correlationID string) Envelope {
	env := Success(data)
	if strings.TrimSpace(correlationID) != "" {
		env.Meta = &EnvelopeMeta{CorrelationID: strings.TrimSpace(correlationID)}
	}
	return env
}

func Failure(err error) Envelope {
	mapped := MapError(err)
	return Envelope{OK: false, Error: &mapped}
}

func FailureWithCorrelationID(err error, correlationID string) Envelope {
	env := Failure(err)
	if strings.TrimSpace(correlationID) != "" {
		env.Meta = &EnvelopeMeta{CorrelationID: strings.TrimSpace(correlationID)}
	}
	return env
}

func MapError(err error) EnvelopeError {
	if err == nil {
		return EnvelopeError{Code: "INTERNAL_ERROR", Message: "unknown error", Retryable: false}
	}
	msg := err.Error()
	lower := strings.ToLower(msg)

	switch {
	case strings.Contains(lower, "read-only mode blocked mutating command"):
		return EnvelopeError{Code: "READ_ONLY_BLOCKED", Message: msg, Retryable: false}
	case strings.Contains(lower, "auth failed"),
		strings.Contains(lower, "unauthorized"),
		strings.Contains(lower, "forbidden"),
		strings.Contains(lower, "invalid password"),
		strings.Contains(lower, "authentication"):
		return EnvelopeError{Code: "AUTH_FAILED", Message: msg, Retryable: false}
	case strings.Contains(lower, "econnrefused"),
		strings.Contains(lower, "enotfound"),
		strings.Contains(lower, "network error"),
		strings.Contains(lower, "connection refused"),
		strings.Contains(lower, "no such host"),
		strings.Contains(lower, "fetch failed"):
		return EnvelopeError{Code: "NETWORK_UNREACHABLE", Message: msg, Retryable: true}
	case strings.Contains(lower, "timed out"),
		strings.Contains(lower, "timeout"),
		strings.Contains(lower, "deadline exceeded"):
		return EnvelopeError{Code: "TIMEOUT", Message: msg, Retryable: true}
	case strings.Contains(lower, "invalid --"),
		strings.Contains(lower, " is required"),
		strings.Contains(lower, "invalid server url"),
		strings.Contains(lower, "invalid date range"),
		strings.Contains(lower, "expected yyyy-mm-dd"),
		strings.Contains(lower, "expected yyyy-mm"):
		return EnvelopeError{Code: "INVALID_INPUT", Message: msg, Retryable: false}
	default:
		return EnvelopeError{Code: "INTERNAL_ERROR", Message: msg, Retryable: false}
	}
}
