# actual-budget-cli

Open-source-ready Go CLI for Actual Budget.

> Note: Actual has no public REST API. This CLI uses a Node bridge powered by `@actual-app/api` while keeping CLI UX and architecture in Go.

## Features (MVP)

- `actual-cli auth login`
- `actual-cli accounts list [--json]`
- `actual-cli transactions list [--account <id>] [--from YYYY-MM-DD] [--to YYYY-MM-DD] [--limit N] [--json]`
- `actual-cli budgets summary [--json]`

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

## Config & Security

Config is stored at:

- `~/.config/actual-cli/config.json` (permission `0600`)
- budget cache at `~/.local/share/actual-cli`

Security notes:

- Never commit config files or secrets.
- Use environment/secret managers in CI.
- For self-signed certs, set `NODE_EXTRA_CA_CERTS` if needed.

## Usage Examples

```bash
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
