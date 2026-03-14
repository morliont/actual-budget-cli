# Skill: transactions-list

## Intent
Fetch transactions with optional account/date/limit filters.

## Command

```bash
actual-cli --agent-json --non-interactive transactions list \
  [--account "$ACCOUNT_ID"] \
  [--from YYYY-MM-DD] \
  [--to YYYY-MM-DD] \
  [--limit N]
```

## Required input

- Existing valid config at `~/.config/actual-cli/config.json`
- If provided:
  - `--from`, `--to` in `YYYY-MM-DD`
  - `--from <= --to`
  - `--limit > 0`

Defaults if omitted:

- `from=1900-01-01`
- `to=2999-12-31`
- `limit=100`

## Expected output shape (`--agent-json`)

```json
{
  "ok": true,
  "data": {
    "transactions": [
      {
        "date": "YYYY-MM-DD",
        "account": "...",
        "payee_name": "...",
        "amount": 1234,
        "notes": "..."
      }
    ]
  }
}
```

## Failure handling

- `INVALID_INPUT`: correct date/range/limit and rerun
- `AUTH_FAILED`: rerun `auth-check`
- `NETWORK_UNREACHABLE` / `TIMEOUT`: retry with backoff; optionally raise `ACTUAL_CLI_BRIDGE_TIMEOUT`
- `INTERNAL_ERROR`: return error details to orchestrator
