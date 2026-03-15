# Capability Map (Intent → Command)

Use this map for tool-agnostic orchestration (Codex/subagents/scripts).

## Routing Table

| Intent | Command(s) | Required flags/env | Expected output (`--agent-json`) | Failure handling |
|---|---|---|---|---|
| `auth-check` | `actual-cli --agent-json --non-interactive auth check` | Valid saved config (`~/.config/actual-cli/config.json`) | `data.authenticated` boolean + `data.message` | `AUTH_FAILED`: verify saved creds/ids (refresh via `auth login`); `NETWORK_UNREACHABLE`/`TIMEOUT`: retry with backoff |
| `accounts-list` | `actual-cli --agent-json --non-interactive accounts list` | Valid saved config (`~/.config/actual-cli/config.json`) | `data.accounts[]` rows (`id,name,type,offbudget,closed`) | On `AUTH_FAILED` rerun auth-check; transient network/timeouts retry |
| `categories-list` | `actual-cli --agent-json --non-interactive categories list` | Valid config | `data.categories[]` rows (`id,name,group_id,group_name,hidden,archived`) | `AUTH_FAILED`: rerun auth-check; transient network/timeouts retry |
| `transactions-list` | `actual-cli --agent-json --non-interactive transactions list [--account <id>] [--from YYYY-MM-DD] [--to YYYY-MM-DD] [--limit N] [--include-category-names]` | Valid config. Optional flags must validate (`from/to` format, `from<=to`, `limit>0`) | `data.transactions[]` rows (base fields preserved; optional `category_name,category_group_name`) | `INVALID_INPUT`: correct filters; `AUTH_FAILED`: rerun auth-check; `TIMEOUT`: retry/raise `ACTUAL_CLI_BRIDGE_TIMEOUT` |
| `budgets-summary` | `actual-cli --agent-json --non-interactive budgets summary` | Valid config | `data.budget` stable core (`month,income,budgeted,spent`) + `extra` object for provider-specific fields | `AUTH_FAILED`: rerun auth-check; transient network/timeouts retry |
| `budgets-categories` | `actual-cli --agent-json --non-interactive budgets categories --month YYYY-MM` | Valid config + valid month (`YYYY-MM`) | `data.month` + `data.categories[]` rows (`budgeted/planned,spent/actual,remaining/variance,carryover`) | `INVALID_INPUT`: fix month; auth/network handling same as above |
| `doctor` | `actual-cli --agent-json doctor` | Node runtime in PATH; config recommended for full readiness | `data.ready` + `data.checks[]` + `data.summary` | If `data.ready=false`, inspect failed checks and remediate locally before orchestration |
| `reports-monthly-variance` | `actual-cli --agent-json --non-interactive reports monthly-variance --from YYYY-MM --to YYYY-MM [--strict]` | Valid config + valid month range (`from<=to`) | `data.from`, `data.to`, `data.months[]` with explicit `raw.net_spent/outflow_spent/inflow_offsets/net_variance/planning_variance` (plus legacy aliases `spent`,`variance`), `normalized`, reconciliation `checks[]`, and `quality` metadata | `INVALID_INPUT`: fix range; with `--strict`, any reconciliation mismatch fails command (non-retryable) |

## Global execution notes

- Prefer `--agent-json` for deterministic parsing.
- Prefer `--non-interactive` for unattended runs.
- For automation safety, enable read-only mode with `ACTUAL_CLI_READ_ONLY=true` (or `--read-only`).
- If a mutating command is called while read-only is active, the command fails with `READ_ONLY_BLOCKED`.
- Override env default explicitly with `--read-only=false` when mutation is intentional.
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
- `./workflows/finance-monthly-analysis.md`
- `../skills/*.md` (legacy compatibility pointers)
