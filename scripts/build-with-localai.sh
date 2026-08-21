#!/usr/bin/env bash
# Build Loremetry and embed the local AI bundle for this host.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
if [[ ! -f third_party/models/Qwen2.5-3B-Instruct-Q4_K_M.gguf ]]; then
  echo "Fetching local AI assets (first time may download ~2GB)…"
  ./scripts/fetch-localai.sh --host
fi
wails build "$@"
./scripts/package-localai.sh
echo "Build complete with local AI packaged."
