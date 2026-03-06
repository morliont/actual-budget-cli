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

## Installation

### Option A: Install from source (Go)

```bash
go install github.com/morliont/actual-budget-cli/cmd/actual-cli@latest
```

### Option B: Build locally

```bash
git clone https://github.com/morliont/actual-budget-cli.git
cd actual-budget-cli
make setup
make build
./bin/actual-cli --version
```

### Option C: Prebuilt binaries

For tagged releases, download the archive/binary for your platform from GitHub Releases, then verify with `checksums.txt`.

## Quick start

```bash
./bin/actual-cli auth login --server http://localhost:5006 --budget <SYNC_ID>
```

You will be prompted for your server password (not echoed).

## Usage examples

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

## Command structure

- `auth` — authentication flows (`login`)
- `accounts` — account queries (`list`)
- `transactions` — transaction queries (`list`)
- `budgets` — budget summaries (`summary`)

For command-specific help:

```bash
./bin/actual-cli --help
./bin/actual-cli <command> --help
```

## Version information

Builds include version metadata wired via Go ldflags.

```bash
./bin/actual-cli --version
# Example output:
# v0.1.0 (commit a1b2c3d, built 2026-03-06T06:00:00Z)
```

## Validation rules

The CLI validates input before bridge execution:

- `auth login --server` must include both scheme and host
- `transactions list --from/--to` must use `YYYY-MM-DD`
- `transactions list --from` must not be after `--to`
- `transactions list --limit` must be greater than `0`

## Config & security

Config is stored at:

- `~/.config/actual-cli/config.json` (permission `0600`)
- budget cache at `~/.local/share/actual-cli`

Security notes:

- Never commit config files or secrets
- Use environment/secret managers in CI
- For self-signed certs, set `NODE_EXTRA_CA_CERTS` if needed
- Credentials and bridge artifacts are written with locked-down permissions
- Bridge request payloads are sent over stdin (not process args)

Bridge execution timeout:

- Default timeout is `30s`
- Configure via `ACTUAL_CLI_BRIDGE_TIMEOUT` (`45s`, `2m`, or integer seconds)

## Release process

Maintainers can create a tagged release with:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Tagging triggers GitHub Actions to:

1. run lint/tests,
2. build deterministic multi-platform artifacts,
3. generate release notes from git history,
4. publish artifacts + `checksums.txt` to GitHub Releases.

For local dry runs:

```bash
make release-artifacts
make release-notes
```

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for development and PR workflow.

Project policies:

- [Security policy](./SECURITY.md)
- [Code of Conduct](./CODE_OF_CONDUCT.md)

## Development

```bash
make lint
make test
make build
make fmt-check
```

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

### Internal bridge contracts

The app layer uses typed DTOs in `internal/bridge/types.go` for core bridge request/response shapes.
For commands that support `--json`, payloads are kept as `json.RawMessage` slices to preserve machine-readable output compatibility while still using typed row decoders for table output.

## CI

GitHub Actions workflow runs:

- `npm ci`
- `make lint`
- `make test`
- `make build`
- `make fmt-check`
