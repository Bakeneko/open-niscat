# Open Niscat

Browse the NISCAT parts catalog (Nissan Motor Ibérica light commercial vehicles and trucks, edition 01/2015) in a web browser: identify a vehicle by VIN, navigate the exploded views, search parts, follow supersessions, and build a parts list to copy, export or share.

A single binary serves the catalog and its web interface: copy it next to a `data/` folder on a workshop PC and open a browser, or run it with Docker on a server. Phones and tablets of the local network can use it too.

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

Download the binary for your system from the [releases page](https://github.com/Bakeneko/open-niscat/releases): `windows-amd64`, `linux-amd64`, `linux-arm64` or `darwin-arm64` (Apple Silicon Mac). `SHA256SUMS` lets you check the download (`sha256sum -c SHA256SUMS --ignore-missing`). You can also build it yourself (see [Build](#build)).

1. Put the binary and the `data/` folder side by side:
   ```
   open-niscat-windows-amd64.exe   (or open-niscat-linux-amd64, …)
   data/
   ```
2. Start the binary (double-click on Windows). Your browser opens on http://127.0.0.1:8080/.
   - Linux and macOS: make it executable first (`chmod +x open-niscat-*`).
   - macOS blocks the unsigned binary on first launch: right click → Open, or `xattr -d com.apple.quarantine open-niscat-darwin-arm64`.

Options (also settable in `open-niscat.toml` next to the binary — see `open-niscat.example.toml` — or as environment variables):

| Flag | Environment variable | Default | Meaning |
|---|---|---|---|
| `--data` | `OPEN_NISCAT_DATA` | `./data` next to the binary | data directory |
| `--addr` | `OPEN_NISCAT_ADDR` | `127.0.0.1:8080` | listen address; `0.0.0.0:8080` to allow other machines of the local network |
| `--open-browser` | `OPEN_NISCAT_OPEN_BROWSER` | `true` | open the default browser on startup |
| `--default-lang` | `OPEN_NISCAT_DEFAULT_LANG` | `en` | language offered on `/` (`en`, `fr`, `es` or `de`) |
| `--config` | | `open-niscat.toml` next to the binary | configuration file |

Priority: flags, then environment variables, then the file, then defaults. A relative path given by a flag or a variable is relative to the working directory, one from the file to the file's folder. An unknown key or invalid value stops the program with a message.

### Docker

The image `ghcr.io/bakeneko/open-niscat` (amd64 and arm64) contains the program only: mount your `data/` folder read-only at `/data`. Copy [compose.example.yaml](compose.example.yaml) as `compose.yaml` next to the `data/` folder, then:

```
docker compose up -d
```

or without compose:

```
docker run -d --name open-niscat --restart unless-stopped -p 8080:8080 -v ./data:/data:ro ghcr.io/bakeneko/open-niscat
```

- Settings are environment variables (`-e OPEN_NISCAT_DEFAULT_LANG=fr`, or `environment:` in `compose.yaml`); the image already sets `OPEN_NISCAT_DATA=/data`, `OPEN_NISCAT_ADDR=0.0.0.0:8080` and `OPEN_NISCAT_OPEN_BROWSER=false`. Keep the container port at 8080 and change the published port instead (`-p 80:8080`).
- On Linux, the data files must be readable by the container user (uid 10001), e.g. `chmod -R a+rX data`.
- `GET /health` answers 200 while the catalog is readable; the image health check uses it (`docker ps` shows `healthy`).

## Build

Requirements: Go 1.26+, Node 24+, GNU Make, golangci-lint 2.x.

```
make web-install   # once, and after a change of web/package-lock.json
make tools-install # once: Python dependencies of tools/data (Pillow, ruff, mypy, pytest)
make lint test     # Go, frontend and data tooling
make vuln          # known vulnerabilities in the Go dependencies (govulncheck)
make build         # frontend, then the four release binaries in dist/
```

Development: `make run` (API on :8080 with `./data`) and `npm --prefix web run dev` (Vite with hot reload on :5173, proxies `/api` and `/files`).

Tests use a small synthetic catalog; the tests against the real catalog run only when `data/` is present.

`make docker` builds the image for your machine (`open-niscat:<version>`).

### Release

Tag a version and push the tag: `git tag v1.2.0 && git push origin v1.2.0`. The release workflow runs the checks, attaches the four binaries and `SHA256SUMS` to a GitHub release and publishes the image as `1.2.0`, `1.2` and `latest`. A suffix (`v1.2.0-rc.1`, `v1.2.0-beta.1`) makes a pre-release, published under its own tag only.

A release can also be published from the GitHub web interface with a new tag: the workflow then attaches the files to it and keeps its notes. The image tags still follow the tag name, not the pre-release checkbox.

## License and data

The source code is released under the [MIT License](LICENSE). The catalog data belongs to Nissan: it is not distributed with this project and must never be committed or published.
