#!/usr/bin/env sh
set -eu

REPO="morliont/actual-budget-cli"
BINARY="actual-cli"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "error: required command not found: $1" >&2
    exit 1
  }
}

need_cmd curl
need_cmd tar
need_cmd mktemp

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "error: unsupported arch: $ARCH" >&2; exit 1 ;;
esac

case "$OS" in
  linux|darwin) ;;
  *) echo "error: unsupported OS: $OS" >&2; exit 1 ;;
esac

TMP_DIR="$(mktemp -d)"
cleanup() { rm -rf "$TMP_DIR"; }
trap cleanup EXIT INT TERM

TAG="$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n1)"
[ -n "$TAG" ] || { echo "error: unable to determine latest release tag" >&2; exit 1; }

ARCHIVE="${BINARY}_${TAG}_${OS}_${ARCH}.tar.gz"
BASE_URL="https://github.com/$REPO/releases/download/$TAG"

curl -fsSL -o "$TMP_DIR/$ARCHIVE" "$BASE_URL/$ARCHIVE"
curl -fsSL -o "$TMP_DIR/checksums.txt" "$BASE_URL/checksums.txt"

if command -v sha256sum >/dev/null 2>&1; then
  (cd "$TMP_DIR" && sha256sum -c checksums.txt --ignore-missing)
elif command -v shasum >/dev/null 2>&1; then
  EXPECTED="$(grep "  $ARCHIVE$" "$TMP_DIR/checksums.txt" | awk '{print $1}')"
  ACTUAL="$(shasum -a 256 "$TMP_DIR/$ARCHIVE" | awk '{print $1}')"
  [ "$EXPECTED" = "$ACTUAL" ] || { echo "error: checksum verification failed" >&2; exit 1; }
else
  echo "error: need sha256sum or shasum to verify checksum" >&2
  exit 1
fi

tar -xzf "$TMP_DIR/$ARCHIVE" -C "$TMP_DIR"

mkdir -p "$INSTALL_DIR"
if [ -w "$INSTALL_DIR" ]; then
  install "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
else
  sudo install "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
fi

echo "Installed $BINARY $TAG to $INSTALL_DIR/$BINARY"
"$INSTALL_DIR/$BINARY" --version
