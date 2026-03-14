# Agent Contract (Phase 1 + Phase 2)

This document defines the stable machine-readable contract for `actual-cli` when used by agents.

## Enable agent contract mode

Use global flags as needed:

```bash
actual-cli --agent-json <command>
actual-cli --non-interactive <command>
```

- `--agent-json`: stable machine-readable JSON envelope output.
- `--non-interactive`: disables prompts and fails fast if required input is missing.
- `--correlation-id`: optional trace identifier (or `ACTUAL_CLI_CORRELATION_ID`) echoed in envelope `meta.correlationId`.

Without these flags, existing human-oriented interactive behavior stays unchanged.

---

## Output envelope schema

All command responses in agent mode use this envelope:

```json
{
  "ok": true,
  "data": {},
  "meta": {
    "correlationId": "optional-trace-id"
  }
}
```

or on error:

```json
{
  "ok": false,
  "error": {
    "code": "INVALID_INPUT",
    "message": "invalid --from value \"03-01-2026\": expected YYYY-MM-DD (example: --from 2026-01-31)",
    "retryable": false
  },
  "meta": {
    "correlationId": "optional-trace-id"
  }
}
```

### Fields

- `ok` (boolean): success/failure indicator.
- `data` (object, optional): successful payload.
- `error` (object, optional): present only when `ok=false`.
  - `code` (string): canonical error code.
  - `message` (string): user-facing error detail.
  - `retryable` (boolean): whether immediate retry is likely useful.
- `meta` (object, optional): envelope metadata.
  - `correlationId` (string, optional): caller-provided trace ID.

---

## Command I/O shape in agent mode

- `auth check` → `data.authenticated` (bool), `data.message` (string)
- `auth login` → `data.message`
- `accounts list` → `data.accounts` (array)
- `transactions list` → `data.transactions` (array)
- `budgets summary` → `data.budget` (stable core schema + extensible extras)
- `doctor` → `data.ready` (bool), `data.checks[]`, `data.summary`

### `budgets summary` schema (`--agent-json`)

`data.budget` has a stable core contract:

```json
{
  "month": "YYYY-MM",
  "income": 0,
  "budgeted": 0,
  "spent": 0,
  "extra": {}
}
```

- Core fields (`month`, `income`, `budgeted`, `spent`) are stable.
- Any additional/provider-specific fields are placed under `extra` for forward compatibility.

### Auth password sources (`auth login`)

For automation, server password is resolved with explicit precedence:

1. `--password-stdin` (reads password from stdin)
2. `ACTUAL_CLI_PASSWORD` environment variable
3. Interactive hidden prompt (only when `--non-interactive` is not set)

In `--non-interactive` mode, if no password source is provided, command fails deterministically with `INVALID_INPUT` in agent-json mode.

Input flags/arguments otherwise remain backward-compatible with current CLI behavior.

---

## Error code taxonomy

Canonical codes introduced in Phase 1:

- `AUTH_FAILED`
  - Authentication/authorization failures.
  - Typical cause: wrong credentials, unauthorized access.
- `NETWORK_UNREACHABLE`
  - Network/connectivity failures.
  - Typical cause: server down, DNS/connect issues.
- `TIMEOUT`
  - Request exceeded timeout.
  - Typical cause: server too slow or connectivity degradation.
- `INVALID_INPUT`
  - Client-side validation errors.
  - Typical cause: invalid flags, missing required values, bad date formats/ranges.
- `INTERNAL_ERROR`
  - Unclassified/internal failures.

---

## Retry guidance

Use `error.retryable` as the primary signal.

Recommended behavior:

- `retryable=true` (`NETWORK_UNREACHABLE`, `TIMEOUT`):
  - Retry with backoff.
  - For repeated `TIMEOUT`, consider increasing `ACTUAL_CLI_BRIDGE_TIMEOUT`.
- `retryable=false` (`AUTH_FAILED`, `INVALID_INPUT`):
  - Do not blind-retry.
  - Fix credentials or inputs first.
- `INTERNAL_ERROR`:
  - Inspect `error.message` and logs.
  - Optional limited retry if environment suggests transient failure.
