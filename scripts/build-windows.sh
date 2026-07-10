#!/usr/bin/env bash
# Build a portable Windows x64 EXE with embedded WebView2 bootstrapper.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

WAILS_BIN="${WAILS_BIN:-wails}"

echo "==> StegSolve Go Windows amd64 build"
echo "    wails: $($WAILS_BIN version 2>/dev/null | head -1 || true)"

$WAILS_BIN build \
  -clean \
  -trimpath \
  -platform windows/amd64 \
  -webview2 embed

OUT="build/bin/stegsolve-go.exe"
if [[ -f "$OUT" ]]; then
  ls -lh "$OUT"
  echo "OK: $OUT"
else
  # Wails may place under platform-specific path
  find build/bin -type f -name '*.exe' -print -exec ls -lh {} \;
fi
