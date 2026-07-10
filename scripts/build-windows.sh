#!/usr/bin/env bash
# Convenience wrapper: Windows amd64 portable EXE with embedded WebView2 bootstrapper.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
export PLATFORM=windows/amd64
exec "$ROOT/scripts/build.sh"
