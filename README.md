# actual-budget-cli

Open-source-ready Go CLI for Actual Budget.

> Note: Actual has no public REST API. This CLI uses a Node bridge powered by `@actual-app/api` while keeping CLI UX and architecture in Go.

## Features (MVP)

- `actual-cli auth login`
- `actual-cli accounts list [--json]`
- `actual-cli transactions list [--account <id>] [--from YYYY-MM-DD] [--to YYYY-MM-DD] [--limit N] [--json]`
- `actual-cli budgets summary [--json]`
- `actual-cli doctor` (readiness checks; supports `--agent-json`)
- `actual-cli --version` (includes build metadata)

## Requirements

- Go 1.22+
- Node.js 20+
- Access to an Actual server

## Setup

```bash
make setup
make build
./bin/actual-cli auth login --server http://localhost:5006 --budget <SYNC_ID>
```

You will be prompted for your server password (not echoed).

The CLI validates server URLs (`http/https` + host), transaction dates (`YYYY-MM-DD`), date ranges (`--from` must be on/before `--to`), and `--limit` (>0) before calling the bridge.

## Command Structure

Top-level command tree:

- `auth` — authentication flows (`login`)
- `accounts` — account queries (`list`)
- `transactions` — transaction queries (`list`)
- `budgets` — budget summaries (`summary`)
- `doctor` — environment readiness checks

For command-specific help:

```bash
./bin/actual-cli --help
./bin/actual-cli <command> --help
```

## Agent / Subagent Docs

- [CLAUDE.md](./CLAUDE.md) — canonical agent entrypoint and operating rules
- [AGENTS.md](./AGENTS.md) — operator-focused quickstart for orchestration flows
- [docs/capability-map.md](./docs/capability-map.md) — intent routing map (intent → command → output/errors)
- [docs/agent-contract.md](./docs/agent-contract.md) — JSON envelope contract (`--agent-json`)
- [docs/agent-contract-changelog.md](./docs/agent-contract-changelog.md) — versioned planner-facing contract changes
- Canonical skill layout (`.claude/skills/<skill>/SKILL.md`):
  - [auth-check](./.claude/skills/auth-check/SKILL.md)
  - [accounts-list](./.claude/skills/accounts-list/SKILL.md)
  - [transactions-list](./.claude/skills/transactions-list/SKILL.md)
  - [budgets-summary](./.claude/skills/budgets-summary/SKILL.md)
- Legacy compatibility pointers remain under [`./skills/`](./skills/)

## Version Information

Builds include version metadata wired via Go ldflags.

```bash
./bin/actual-cli --version
# Example output:
# v0.1.0 (commit a1b2c3d, built 2026-03-06T06:00:00Z)
```

## Config & Security

Config is stored at:

- `~/.config/actual-cli/config.json` (permission `0600`)
- budget cache at `~/.local/share/actual-cli`

Security notes:

- Never commit config files or secrets.
- Use environment/secret managers in CI.
- For self-signed certs, set `NODE_EXTRA_CA_CERTS` if needed.
- Credentials are written with locked-down permissions (`~/.config/actual-cli` = `0700`, `config.json` = `0600`, data dir = `0700`).
- Bridge request payloads are sent over stdin (not process args) to avoid leaking secrets via argv/process listings.
- The bridge script is embedded in the Go binary and materialized as a temporary file with `0600` permissions for execution (no cwd-relative script lookup).
- Bridge errors only surface sanitized stderr text and do not echo request payloads.

Bridge execution timeout:

- Default timeout is `30s`.
- Configure via `ACTUAL_CLI_BRIDGE_TIMEOUT` (supports Go duration values like `45s`, `2m`, or plain positive integer seconds).

Correlation ID support:

- Set `--correlation-id <id>` (or `ACTUAL_CLI_CORRELATION_ID`) to tag agent-json envelopes and error output with a trace identifier.

## Usage Examples

```bash
# login (non-interactive flags for server/budget)
./bin/actual-cli auth login --server http://localhost:5006 --budget <SYNC_ID>

# list accounts
./bin/actual-cli accounts list

# list transactions for one account and date range
./bin/actual-cli transactions list --account <ACCOUNT_ID> --from 2026-01-01 --to 2026-01-31 --limit 50

# machine-readable output
./bin/actual-cli transactions list --json

# budget summary (current month)
./bin/actual-cli budgets summary

# readiness checks for automation
./bin/actual-cli --agent-json --correlation-id run-42 doctor
```

## Validation Rules

- `auth login --server` must include both scheme and host (for example `http://localhost:5006` or `https://actual.example.com`).
- `transactions list --from/--to` must use `YYYY-MM-DD`.
- `transactions list --from` must not be after `--to`.
- `transactions list --limit` must be greater than `0`.

Validation happens before bridge execution so you get fast, actionable feedback locally.

## Troubleshooting

- **`request timed out after ...`**
  - Increase bridge timeout: `export ACTUAL_CLI_BRIDGE_TIMEOUT=60s`
  - Retry and verify server responsiveness.
- **`network error while contacting Actual server`**
  - Confirm `--server` URL and port.
  - Ensure the Actual server is reachable from your machine.
  - Check VPN/firewall/proxy settings.
- **`bridge runtime unavailable`**
  - Install Node.js 20+ and ensure `node` is available in `PATH`.
- **TLS / self-signed cert issues**
  - Set `NODE_EXTRA_CA_CERTS` to your CA bundle when needed.

## Development

```bash
make lint
make test
make build
```

### Internal bridge contracts

The app layer uses typed DTOs in `internal/bridge/types.go` for core bridge request/response shapes.
For commands that support `--json`, payloads are kept as `json.RawMessage` slices to preserve machine-readable output compatibility while still using typed row decoders for table output.

## CI

GitHub Actions workflow runs:

- `npm ci`
- `go vet ./...`
- `go test ./...`
- `go build ./cmd/actual-cli`
