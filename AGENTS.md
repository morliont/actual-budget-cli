# AGENTS.md — Operator Guide for `actual-budget-cli`

This repo exposes a small, stable CLI surface intended for both humans and subagents.

## Scope

- Use this guide for orchestration and automation flows.
- Runtime behavior is defined by CLI code and `docs/agent-contract.md`.

## Fast Path (agent-safe defaults)

```bash
# Stable machine output + no prompts
actual-cli --agent-json --non-interactive <command>
```

Recommended command order for first-time setup:

1. `auth login` (one-time setup / refresh credentials)
2. `auth-check` intent via `actual-cli auth check`
3. `accounts-list`
4. `transactions-list`
5. `budgets-summary`
6. `reports-monthly-variance` for deterministic monthly analysis

## Required environment / preconditions

- `node` available in `PATH` (bridge runtime)
- Config readable at `~/.config/actual-cli/config.json` for non-auth commands
- For non-interactive auth:
  - provide `--password-stdin`, or
  - set `ACTUAL_CLI_PASSWORD`

Optional env:

- `ACTUAL_CLI_BRIDGE_TIMEOUT` (e.g., `45s`, `2m`, `60`)

## Error handling contract

When `--agent-json` is enabled, parse only the envelope:

- `ok=true` → use `data`
- `ok=false` → branch by `error.code` and `error.retryable`

Canonical codes: `AUTH_FAILED`, `NETWORK_UNREACHABLE`, `TIMEOUT`, `INVALID_INPUT`, `INTERNAL_ERROR`.

## Capability docs

- `docs/capability-map.md`
- `docs/agent-contract.md`
- `docs/workflows/finance-monthly-analysis.md`
- `docs/orchestration.md`
- `skills/*.md` (legacy compatibility pointers)
