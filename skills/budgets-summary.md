# Skill: budgets-summary

## Intent
Fetch current-month budget summary totals.

## Command

```bash
actual-cli --agent-json --non-interactive budgets summary
```

## Required input

- Existing valid config at `~/.config/actual-cli/config.json`

## Expected output shape (`--agent-json`)

```json
{
  "ok": true,
  "data": {
    "budget": {
      "month": "YYYY-MM",
      "income": 0,
      "budgeted": 0,
      "spent": 0,
      "extra": {}
    }
  }
}
```

`month/income/budgeted/spent` are stable core fields. Additional provider-specific fields are exposed under `extra`.

## Failure handling

- `AUTH_FAILED`: rerun `auth-check`
- `NETWORK_UNREACHABLE` / `TIMEOUT`: retry with backoff
- `INVALID_INPUT`: inspect global flags/config state
- `INTERNAL_ERROR`: escalate with command + stderr context
