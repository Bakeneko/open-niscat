# Open Niscat

Browse the NISCAT parts catalog (Nissan Motor Ibérica light commercial vehicles and trucks, edition 01/2015) in a web browser: identify a vehicle by VIN, navigate the exploded views, search parts, follow supersessions, and build a parts list to copy, export or share.

A single binary serves the catalog and its web interface: copy it next to a `data/` folder on a workshop PC and open a browser. Phones and tablets of the local network can use it too.

The catalog data is **not** included: it is licensed by Nissan. You need your own `data/` folder (see [Data](#data)).

*French version: [README.fr.md](README.fr.md).*

## Features

- **Vehicle identification**: by full VIN or by its last 6+ characters, or by catalog and model code (filters on engine, body, transmission…). The vehicle sheet lists its attributes, documents and the general index.
- **Exploded views**: group indexes and parts plates with zoom (wheel, pinch), pan, clickable callouts and quick tooltips; the callout numbers are written over the drawings as NISCAT does.
- **Applicability**: sections and parts that do not apply to the vehicle are hidden or greyed; parts outside the vehicle's production date are flagged, never hidden.
- **Search**: VINs, sections and parts (reference or text), limited to the current vehicle or across all vehicles.
- **Supersessions**: alternative and latest known part numbers, and every plate where a part is used.
- **Parts list**: kept in the browser, quantities, copy for a spreadsheet, CSV export, print / PDF, and a share link that opens the same list elsewhere.
- **Shareable URLs**: every page (vehicle, plate, selected callout, search, list) has its own link.
- English, French, Spanish and German interface (the languages of the NISCAT data); designed for desktop and mobile; plates print as drawing then parts table.

## Data

The `data/` folder holds the catalog converted from a NISCAT installation (edition 01/2015):

```
data/
├── manifest.json   data schema, data version, source edition, build information
├── data.db         SQLite database (catalogs, VINs, model codes, sections, parts, hotspots)
├── img/<series>/   parts plates (PNG)
├── gindex/<series>/  group index drawings (PNG)
└── cinfo/<series>/   catalog documents (PDF)
```

The program opens `data.db` read-only and refuses a folder whose `manifest.json` schema it does not support.

### Build the data folder

If you own the NISCAT 01/2015 installer, build the folder yourself (Windows, Python 3.12+ with Pillow):

```
python tools/data/build_data.py --installer <installer folder> --out data
```

See [tools/data/README.md](tools/data/README.md) for the requirements, options and other systems.

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
| `--default-lang` | `en` | language offered on `/` (`en`, `fr`, `es` or `de`) |
| `--config` | `open-niscat.toml` next to the binary | configuration file |

Priority: flags, then the file, then defaults. An unknown key or invalid value stops the program with a message.

## Build

Requirements: Go 1.26+, Node 24+, GNU Make, golangci-lint 2.x.

```
make web-install   # once, and after a change of web/package-lock.json
make tools-install # once: Python dependencies of tools/data (Pillow, ruff, mypy, pytest)
make lint test     # Go, frontend and data tooling
make build         # frontend, then binaries for Linux and Windows in dist/
```

Development: `make run` (API on :8080 with `./data`) and `npm --prefix web run dev` (Vite with hot reload on :5173, proxies `/api` and `/files`).

Tests use a small synthetic catalog; the tests against the real catalog run only when `data/` is present.

## License and data

The source code is released under the [MIT License](LICENSE). The catalog data belongs to Nissan: it is not distributed with this project and must never be committed or published.
