#!/usr/bin/env bash
# Copy third_party local AI into the Wails build output for the current (or specified) platform.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BIN="$ROOT/build/bin"
MODEL_NAME="Qwen2.5-3B-Instruct-Q4_K_M.gguf"
MODEL_SRC="$ROOT/third_party/models/$MODEL_NAME"

if [[ ! -f "$MODEL_SRC" ]]; then
  echo "Missing $MODEL_SRC — run ./scripts/fetch-localai.sh first" >&2
  exit 1
fi

detect_plat() {
  case "$(uname -s)-$(uname -m)" in
    Darwin-arm64|Darwin-aarch64) echo darwin-arm64 ;;
    Darwin-x86_64) echo darwin-amd64 ;;
    MINGW*|MSYS*|CYGWIN*|Windows*) echo windows-amd64 ;;
    *) echo "${LOREMETRY_PACKAGE_PLAT:-}" ;;
  esac
}

PLAT="${1:-$(detect_plat)}"
if [[ -z "$PLAT" ]]; then
  echo "Usage: $0 darwin-arm64|darwin-amd64|windows-amd64" >&2
  exit 1
fi

SRC="$ROOT/third_party/llama-server/$PLAT"
if [[ ! -d "$SRC" ]]; then
  echo "Missing $SRC — run ./scripts/fetch-localai.sh" >&2
  exit 1
fi

case "$PLAT" in
  darwin-arm64|darwin-amd64)
    APP="$(find "$BIN" -maxdepth 2 -name '*.app' -print -quit 2>/dev/null || true)"
    if [[ -z "$APP" ]]; then
      echo "No .app in $BIN — run wails build first" >&2
      exit 1
    fi
    DEST="$APP/Contents/Resources/localai"
    mkdir -p "$DEST/llama-server/$PLAT" "$DEST/models"
    cp -a "$SRC/." "$DEST/llama-server/$PLAT/"
    cp "$MODEL_SRC" "$DEST/models/$MODEL_NAME"
    echo "Packaged local AI into $DEST"
    ;;
  windows-amd64)
    DEST="$BIN/localai"
    mkdir -p "$DEST/llama-server/$PLAT" "$DEST/models"
    cp -a "$SRC/." "$DEST/llama-server/$PLAT/"
    cp "$MODEL_SRC" "$DEST/models/$MODEL_NAME"
    echo "Packaged local AI into $DEST (place beside loremetry.exe)"
    ;;
  *)
    echo "Unknown platform $PLAT" >&2
    exit 1
    ;;
esac
