# actual-budget-cli

Open-source-ready Go CLI for Actual Budget.

> Note: Actual has no public REST API. This CLI uses a Node bridge powered by `@actual-app/api` while keeping CLI UX and architecture in Go.

## Features (MVP)

- `actual-cli auth login`
- `actual-cli accounts list [--json]`
- `actual-cli transactions list [--account <id>] [--from YYYY-MM-DD] [--to YYYY-MM-DD] [--limit N] [--json]`
- `actual-cli budgets summary [--json]`
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

The CLI validates server URLs (`http/https`), transaction dates (`YYYY-MM-DD`), and `--limit` (>0) before calling the bridge.

## Command Structure

Top-level command tree:

- `auth` — authentication flows (`login`)
- `accounts` — account queries (`list`)
- `transactions` — transaction queries (`list`)
- `budgets` — budget summaries (`summary`)

For command-specific help:

```bash
./bin/actual-cli --help
./bin/actual-cli <command> --help
```

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
```

## Development

```bash
make lint
make test
make build
```

## CI

GitHub Actions workflow runs:

- `npm ci`
- `go vet ./...`
- `go test ./...`
- `go build ./cmd/actual-cli`
