#!/usr/bin/env bash
# Build portable packages for the current host (or cross-target via PLATFORM).
# Examples:
#   ./scripts/build.sh
#   PLATFORM=windows/amd64 ./scripts/build.sh
#   PLATFORM=darwin/arm64 ./scripts/build.sh
#   PLATFORM=linux/amd64 ./scripts/build.sh
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

WAILS_BIN="${WAILS_BIN:-wails}"
PLATFORM="${PLATFORM:-}"

detect_platform() {
  local os arch
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64) arch=amd64 ;;
    aarch64|arm64) arch=arm64 ;;
    *) echo "unsupported arch: $arch" >&2; exit 1 ;;
  esac
  case "$os" in
    darwin) echo "darwin/${arch}" ;;
    linux) echo "linux/${arch}" ;;
    mingw*|msys*|cygwin*) echo "windows/${arch}" ;;
    *) echo "unsupported os: $os" >&2; exit 1 ;;
  esac
}

if [[ -z "$PLATFORM" ]]; then
  PLATFORM="$(detect_platform)"
fi

EXTRA=()
case "$PLATFORM" in
  windows/*)
    EXTRA+=(-webview2 embed)
    ;;
esac

echo "==> StegSolve Go build"
echo "    platform: $PLATFORM"
echo "    wails:    $($WAILS_BIN version 2>/dev/null | head -1 || true)"

$WAILS_BIN build \
  -clean \
  -trimpath \
  -platform "$PLATFORM" \
  "${EXTRA[@]}"

echo "==> artifacts under build/bin/"
find build/bin -maxdepth 3 \( -type f -o -type d -name '*.app' \) 2>/dev/null | head -40
