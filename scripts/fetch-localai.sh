#!/usr/bin/env bash
# Fetch pinned llama-server binaries + GGUF into third_party/ (packaging only; not used by the running app).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LLAMA_TAG="${LOREMETRY_LLAMA_TAG:-b10447}"
MODEL_URL="${LOREMETRY_MODEL_URL:-https://huggingface.co/Qwen/Qwen2.5-3B-Instruct-GGUF/resolve/main/qwen2.5-3b-instruct-q4_k_m.gguf}"
MODEL_NAME="Qwen2.5-3B-Instruct-Q4_K_M.gguf"
HOST_ONLY=0
SKIP_MODEL=0
for arg in "$@"; do
  case "$arg" in
    --host) HOST_ONLY=1 ;;
    --skip-model) SKIP_MODEL=1 ;;
  esac
done

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

fetch_archive() {
  local url="$1" dest="$2"
  echo "Downloading $url"
  curl -fL --progress-bar -o "$dest" "$url"
}

extract_server() {
  local archive="$1" outdir="$2" binname="$3"
  mkdir -p "$outdir"
  local extract="$tmpdir/ex"
  rm -rf "$extract"
  mkdir -p "$extract"
  case "$archive" in
    *.tar.gz|*.tgz) tar -xzf "$archive" -C "$extract" ;;
    *.zip) unzip -q "$archive" -d "$extract" ;;
    *) echo "unknown archive $archive"; return 1 ;;
  esac
  local found
  found="$(find "$extract" -type f \( -name 'llama-server' -o -name 'llama-server.exe' \) | head -n 1)"
  if [[ -z "$found" ]]; then
    echo "llama-server not found in $archive" >&2
    return 1
  fi
  local libdir
  libdir="$(dirname "$found")"
  # Preserve versioned dylib/dll symlinks required by the loader.
  rm -rf "$outdir"
  mkdir -p "$outdir"
  cp -a "$found" "$outdir/$binname"
  chmod +x "$outdir/$binname" 2>/dev/null || true
  # Copy every library next to the binary (files + symlinks).
  shopt -s nullglob
  for f in "$libdir"/*.dylib "$libdir"/*.so "$libdir"/*.so.* "$libdir"/*.dll; do
    [[ -e "$f" ]] || continue
    cp -a "$f" "$outdir/"
  done
  shopt -u nullglob
  echo "Installed $outdir/$binname"
}

BASE="https://github.com/ggml-org/llama.cpp/releases/download/${LLAMA_TAG}"

install_platform() {
  local plat="$1" asset="$2" binname="$3"
  local archfile="$tmpdir/$asset"
  fetch_archive "$BASE/$asset" "$archfile"
  extract_server "$archfile" "$ROOT/third_party/llama-server/$plat" "$binname"
}

host_key() {
  local sys mach
  sys="$(uname -s)"
  mach="$(uname -m)"
  if [[ "$sys" == Darwin && "$mach" == arm64 ]]; then echo darwin-arm64
  elif [[ "$sys" == Darwin && "$mach" == x86_64 ]]; then echo darwin-amd64
  else echo ""
  fi
}

if [[ "$HOST_ONLY" -eq 1 ]]; then
  HK="$(host_key)"
  case "$HK" in
    darwin-arm64) install_platform darwin-arm64 "llama-${LLAMA_TAG}-bin-macos-arm64.tar.gz" llama-server ;;
    darwin-amd64) install_platform darwin-amd64 "llama-${LLAMA_TAG}-bin-macos-x64.tar.gz" llama-server ;;
    *)
      echo "Host-only fetch supports Darwin arm64/x86_64. Got: $(uname -s) $(uname -m)" >&2
      echo "For Windows, run full fetch or copy llama-server.exe into third_party/llama-server/windows-amd64/" >&2
      exit 1
      ;;
  esac
else
  install_platform darwin-arm64 "llama-${LLAMA_TAG}-bin-macos-arm64.tar.gz" llama-server
  install_platform darwin-amd64 "llama-${LLAMA_TAG}-bin-macos-x64.tar.gz" llama-server
  install_platform windows-amd64 "llama-${LLAMA_TAG}-bin-win-cpu-x64.zip" llama-server.exe
fi

mkdir -p "$ROOT/third_party/models"
MODEL_PATH="$ROOT/third_party/models/$MODEL_NAME"
if [[ "$SKIP_MODEL" -eq 1 ]]; then
  echo "Skipping model download (--skip-model)"
elif [[ -f "$MODEL_PATH" ]]; then
  echo "Model already present: $MODEL_PATH"
else
  echo "Downloading model (~2GB)…"
  fetch_archive "$MODEL_URL" "$MODEL_PATH"
fi

echo "Done. llama.cpp ${LLAMA_TAG}"
