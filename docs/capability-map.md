# Capability Map (Intent → Command)

Use this map for tool-agnostic orchestration (Codex/subagents/scripts).

## Routing Table

| Intent | Command(s) | Required flags/env | Expected output (`--agent-json`) | Failure handling |
|---|---|---|---|---|
| `auth-check` | `actual-cli --agent-json --non-interactive auth check` | Valid saved config (`~/.config/actual-cli/config.json`) | `data.authenticated` boolean + `data.message` | `AUTH_FAILED`: verify saved creds/ids (refresh via `auth login`); `NETWORK_UNREACHABLE`/`TIMEOUT`: retry with backoff |
| `accounts-list` | `actual-cli --agent-json --non-interactive accounts list` | Valid saved config (`~/.config/actual-cli/config.json`) | `data.accounts[]` rows (`id,name,type,offbudget,closed`) | On `AUTH_FAILED` rerun auth-check; transient network/timeouts retry |
| `transactions-list` | `actual-cli --agent-json --non-interactive transactions list [--account <id>] [--from YYYY-MM-DD] [--to YYYY-MM-DD] [--limit N]` | Valid config. Optional flags must validate (`from/to` format, `from<=to`, `limit>0`) | `data.transactions[]` rows (at least `date,account,payee_name,amount,notes`) | `INVALID_INPUT`: correct filters; `AUTH_FAILED`: rerun auth-check; `TIMEOUT`: retry/raise `ACTUAL_CLI_BRIDGE_TIMEOUT` |
| `budgets-summary` | `actual-cli --agent-json --non-interactive budgets summary` | Valid config | `data.budget` stable core (`month,income,budgeted,spent`) + `extra` object for provider-specific fields | `AUTH_FAILED`: rerun auth-check; transient network/timeouts retry |
| `doctor` | `actual-cli --agent-json doctor` | Node runtime in PATH; config recommended for full readiness | `data.ready` + `data.checks[]` + `data.summary` | If `data.ready=false`, inspect failed checks and remediate locally before orchestration |

## Global execution notes

- Prefer `--agent-json` for deterministic parsing.
- Prefer `--non-interactive` for unattended runs.
- Envelope contract:
  - success: `{ "ok": true, "data": ..., "meta": { "correlationId"? } }`
  - failure: `{ "ok": false, "error": { "code", "message", "retryable" }, "meta": { "correlationId"? } }`

## Suggested orchestration flow

1. Ensure/refresh auth (`auth-check` intent).
2. Route read intents directly to one of: accounts, transactions, budgets.
3. If `error.retryable=true`, bounded retry with backoff.
4. If `error.retryable=false`, stop and request corrected input/credentials.

## Related docs

- `../AGENTS.md`
- `./agent-contract.md`
- `../skills/auth-check.md`
- `../skills/accounts-list.md`
- `../skills/transactions-list.md`
- `../skills/budgets-summary.md`
