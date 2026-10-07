# NISCAT data tooling

Builds the `data/` folder Open Niscat reads from a **NISCAT 01/2015 installer** (the `SWD-0123` CD: the folder
that holds `Setup.exe` and `spa2/`). The catalog data belongs to Nissan: this tool converts your own copy, it
does not ship any data.

## Requirements

- **Windows** for the first step: the installer's tables are Access 97 files, readable only through the
  Microsoft Jet 4.0 driver, which exists for 32-bit programs. The tool runs it with the 32-bit PowerShell
  (`C:\Windows\SysWOW64\WindowsPowerShell\v1.0\powershell.exe`), present on Windows 10 and 11.
- **Python 3.12+** with Pillow: `python -m pip install pillow`, or `make tools-install` from the repository
  root (also installs the development checkers). On Windows, if `python` opens the Microsoft Store, Python is
  not on the PATH: use the launcher instead (`py -3.12 tools/data/build_data.py ...`, `make PYTHON=py ...`).
  On Linux and macOS distributions that forbid system-wide `pip install`, use a virtual environment:
  `python3 -m venv tools/data/.venv` then `make PYTHON=tools/data/.venv/bin/python tools-install`.

## Usage

```
python tools/data/build_data.py --installer <installer folder> --out data
```

It exports the tables, builds `data.db`, converts the drawings to PNG, copies the documents and writes
`manifest.json`. Everything is written into `<out>.tmp` first and replaces `<out>` only when every step
succeeded. Expect a few minutes (about 640,000 part lines and 5,000 drawings).

| Option | Meaning |
|---|---|
| `--installer` | NISCAT installer folder (required) |
| `--out` | data folder to create (default `data`) |
| `--force` | replace an existing `--out` folder |
| `--staging` | keep the exported tables in this folder (default: temporary, removed at the end) |
| `--skip-export` | reuse the tables already exported in `--staging` |
| `--jobs` | parallel drawing conversions (default: number of CPUs, 61 at most) |
| `--edition` | NISCAT edition written in the manifest (default `2015-01`) |

**Other systems:** export once on Windows with `--staging <folder>`, copy that folder, then run
`build_data.py --installer <installer> --staging <folder> --skip-export` on Linux or macOS.

## Troubleshooting

- *not a NISCAT installer*: point `--installer` at the folder containing `spa2/niscat.mdb`.
- *the export of the NISCAT tables failed*: the Jet 4.0 driver or the 32-bit PowerShell is missing, or a group
  policy forbids PowerShell scripts (it overrides the tool's `-ExecutionPolicy Bypass`).
- *already exists*: add `--force` to replace the data folder.
- *is in use*: stop the Open Niscat server (or any program reading the folder) and run again.
- *Pillow is missing*: `python -m pip install pillow`.

## Output

```
data/
├── manifest.json   {"schema": 1, "version": "2015.01-2", "source": {...}, "build": {...}}
├── data.db         SQLite: catalogs, VINs, model codes, sections, parts (+ full-text index), hotspots
├── img/<SERIES>/   parts plates (PNG, same pixel size as the source: hotspots are in source pixels)
├── gindex/<SERIES>/  group index drawings (PNG)
└── cinfo/<SERIES>/   catalog documents (PDF)
```

`manifest.json`: `schema` is the `data.db` schema the program supports; `version` identifies the delivery
(source edition `YYYY.MM`, then the tool revision); `source` names the catalog and its edition (`YYYY-MM`);
`build` records the tool, its revision and the date.

## Development

```
make tools-install            # once
make lint-tools test-tools    # ruff (all rules), mypy --strict, pytest
```

Tests build tiny synthetic inputs; they never need the real installer. Before and after changing the tool,
rebuild into another folder and compare with the previous output:

```
python tools/data/build_data.py --installer <installer> --out ../data-new
python tools/data/compare_data.py data ../data-new
```

`compare_data.py` checks tables, rows (in any order), indexes and the full-text index of `data.db`, the pixels
of every PNG, the bytes of every other file and the manifest (except the build date). When a change is meant
to alter the output, bump `REVISION` in `niscat_data/manifest.py`.
