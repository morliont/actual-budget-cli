# actual-budget-cli

CLI for [Actual Budget](https://actualbudget.org/).

> Actual has no public REST API. This CLI uses a Node bridge powered by `@actual-app/api` with a Go command UX.

## Quick install

### Homebrew (macOS/Linux)

```bash
brew install morliont/tap/actual-cli
```

### One-line installer (Linux/macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/morliont/actual-budget-cli/main/scripts/install.sh | sh
```

### Go install

```bash
go install github.com/morliont/actual-budget-cli/cmd/actual-cli@latest
```

## Quick start

```bash
actual-cli init
actual-cli accounts list
actual-cli transactions list --limit 20
actual-cli budgets summary
```

## Commands

- `actual-cli init` — first-run setup wizard (TTY-aware)
- `actual-cli auth login` — explicit login flow
- `actual-cli accounts list`
- `actual-cli transactions list`
- `actual-cli budgets summary`

## Docs

- [Install guide](docs/install.md)
- [Release verification](docs/release-verification.md)
- [Security](docs/security.md)
- [Troubleshooting](docs/troubleshooting.md)
- [Advanced usage](docs/advanced.md)
- [Release process](RELEASE.md)

## Orchestration

For ticket execution flow (Rudy orchestrates, Francois codes, Linear tracks), see:

- [Workflow definition](WORKFLOW.md)
- [Orchestration guide + operator runbook](docs/orchestration.md)
