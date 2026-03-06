# Install

## Homebrew

```bash
brew install morliont/tap/actual-cli
```

## Script installer (curl | sh)

```bash
curl -fsSL https://raw.githubusercontent.com/morliont/actual-budget-cli/main/scripts/install.sh | sh
```

The installer:
- detects OS/arch (`linux`/`darwin`, `amd64`/`arm64`)
- downloads latest GitHub release archive
- verifies archive checksum from `checksums.txt`
- installs `actual-cli` into `/usr/local/bin` (override with `INSTALL_DIR`)

Example custom install dir:

```bash
curl -fsSL https://raw.githubusercontent.com/morliont/actual-budget-cli/main/scripts/install.sh | INSTALL_DIR="$HOME/.local/bin" sh
```

## Go install

```bash
go install github.com/morliont/actual-budget-cli/cmd/actual-cli@latest
```

## Build from source

```bash
git clone https://github.com/morliont/actual-budget-cli.git
cd actual-budget-cli
make setup
make build
./bin/actual-cli --version
```
