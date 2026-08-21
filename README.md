# README

## About

This is a Desktop app with an API

Configure the project in `wails.json`.
See the [Wails project config reference](https://wails.io/docs/reference/project-config).

## Live Development

To run in live development mode, run `wails dev` in the project directory.
This will run a Vite development server with fast hot reload of frontend
changes. To call Go methods from a browser, use the Wails
[dev server](http://localhost:34115) and open it in your browser.

## Building

```bash
wails build
```

To ship **bundled local AI** (llama-server + GGUF, no runtime download):

```bash
./scripts/fetch-localai.sh    # once; ~2GB model + sidecars
./scripts/build-with-localai.sh
```

See [third_party/README.md](third_party/README.md).
