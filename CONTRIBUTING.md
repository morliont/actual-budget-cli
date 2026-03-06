# Contributing

Thanks for contributing to `actual-budget-cli`.

## Development setup

- Go 1.22+
- Node.js 20+

```bash
make setup
```

## Local quality checks

Before opening a PR, run:

```bash
make lint
make test
make build
make fmt-check
```

## Pull request guidelines

1. Keep changes focused and small.
2. Include tests when behavior changes.
3. Update README/docs for user-facing changes.
4. Use clear commit messages.

## Release process (maintainers)

1. Ensure `main` is green.
2. Create and push a semver tag (for example `v0.2.0`).
3. GitHub Actions release workflow builds cross-platform artifacts and publishes checksums + release notes.
