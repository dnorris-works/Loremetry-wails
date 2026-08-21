# Bundled local AI (llama.cpp)

Shipped inside the Loremetry install. The **running app never downloads** models or binaries.

## Layout

```text
third_party/
  llama-server/
    darwin-arm64/llama-server   (+ any .dylib from the release)
    darwin-amd64/llama-server
    windows-amd64/llama-server.exe  (+ any .dll from the release)
  models/
    Qwen2.5-3B-Instruct-Q4_K_M.gguf
```

## Pins

| Item | Value |
| --- | --- |
| llama.cpp release | `b10447` |
| Model | Qwen2.5-3B-Instruct Q4_K_M GGUF |

## Populate (release / local package)

From the repo root (needs network once, for packaging only):

```bash
./scripts/fetch-localai.sh              # all platforms + model (~2GB)
./scripts/fetch-localai.sh --host       # this Mac’s sidecar + model
./scripts/fetch-localai.sh --host --skip-model   # sidecar only (dev)
```

Then build and copy into the app:

```bash
wails build
./scripts/package-localai.sh
```

Or one shot:

```bash
./scripts/build-with-localai.sh
```

`package-localai.sh` copies the matching platform sidecar + GGUF into:

- macOS: `build/bin/loremetry.app/Contents/Resources/localai/`
- Windows: `build/bin/localai/` next to `loremetry.exe`

## Dev without the GGUF

If `third_party` is empty, local AI will not start and AI analyses fail until the bundle is present.
