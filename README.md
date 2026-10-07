# Open Niscat

Browse the NISCAT parts catalog (Nissan Motor Ibérica light commercial vehicles and trucks, edition 01/2015) in a web browser: identify a vehicle by VIN, navigate the exploded views, search parts, follow supersessions, and build a parts list to copy, export or share.

The catalog data is **not** included: it is licensed by Nissan. You need your own `data/` folder (`data.db`, `manifest.json`, `img/`, `gindex/`, `cinfo/`).

*Version française : [README.fr.md](README.fr.md).*

## Run

1. Put the binary and the `data/` folder side by side:
   ```
   open-niscat-windows-amd64.exe   (or open-niscat-linux-amd64)
   data/
   ```
2. Start the binary (double-click on Windows). Your browser opens on http://127.0.0.1:8080/.

Options (also settable in `open-niscat.toml` next to the binary — see `open-niscat.example.toml`):

| Flag | Default | Meaning |
|---|---|---|
| `--data` | `./data` next to the binary | data directory |
| `--addr` | `127.0.0.1:8080` | listen address; `0.0.0.0:8080` to allow other machines of the local network |
| `--open-browser` | `true` | open the default browser on startup |
| `--default-lang` | `en` | language offered on `/` (`en` or `fr`) |
| `--config` | `open-niscat.toml` next to the binary | configuration file |

Priority: flags, then the file, then defaults. An unknown key or invalid value stops the program with a message.

## Build

Requirements: Go 1.26+, Node 24+, GNU Make, golangci-lint 2.x.

```
make web-install   # once, and after a change of web/package-lock.json
make lint test     # Go + frontend
make build         # frontend, then binaries for Linux and Windows in dist/
```

Development: `make run` (API on :8080 with `./data`) and `npm --prefix web run dev` (Vite with hot reload, proxies `/api` and `/files`).

## License and data

The source code is yours to use. The catalog data belongs to Nissan and must never be committed or published.
