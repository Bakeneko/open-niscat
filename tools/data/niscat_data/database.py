"""Build data.db from the NISCAT tables exported by export_mdb.ps1.

The staging folder holds one sub-folder per source .mdb (``spa2__niscat``, ``spa2__vn``,
``spa2__series__AA``...) with one JSON Lines file per table. This module normalises them into the SQLite
database Open Niscat reads; its schema is the one the Go program expects (``catalog.SchemaVersion``).
"""

import configparser
import json
import re
import sqlite3
from collections.abc import Iterator
from pathlib import Path
from typing import Any

# Source tables and columns are suffixed or keyed by a one-letter language code. Italian is part of the format
# but the 01/2015 edition has no Italian content.
LANG_CODES = {"L": "es", "Q": "en", "D": "de", "F": "fr", "I": "it"}

SCHEMA = """
CREATE TABLE catalog (            -- niscat.mdb/SERIES: one row per catalogue entry shown in the UI
  etd TEXT, grupo TEXT, model TEXT, cmodel TEXT, drive TEXT, date_from TEXT, date_to TEXT,
  serie TEXT, data TEXT, orden TEXT, langs TEXT);
CREATE TABLE vin (                -- vn.mdb/v1t
  vin TEXT, vin_raw TEXT, codenis TEXT, etd TEXT, tipo TEXT, prodata TEXT);
CREATE TABLE grupo (etd TEXT, grupo TEXT, esp TEXT);
CREATE TABLE main_group (         -- INDIGRAL: A=engine ... K=misc
  etd TEXT, cetd TEXT, lang TEXT, label TEXT);
CREATE TABLE section (            -- <lang>_sec: one section = one plate = one drawing
  etd TEXT, lang TEXT, plate TEXT, codigrup TEXT, numsec TEXT, nomsec TEXT, notas TEXT,
  desde TEXT, hasta TEXT);
CREATE TABLE part (               -- <lang>_PAR: parts list lines
  etd TEXT, lang TEXT, pospie INTEGER, plate TEXT, item TEXT, item_eff TEXT, sec TEXT,
  nc TEXT, s TEXT, ind1 TEXT, ind2 TEXT, ind3 TEXT, ind4 TEXT, ind5 TEXT,
  part_no TEXT, part_key TEXT, des TEXT, esp TEXT, qty TEXT, cap TEXT, pmc TEXT, app TEXT,
  dataplic TEXT, ica TEXT, ree TEXT, ree_key TEXT, kdf TEXT, pnc TEXT);
CREATE TABLE attr_table (         -- csel (kind T) / cinf (kind I): vehicle attribute tables
  etd TEXT, grupo TEXT, kind TEXT, tabla TEXT, lang TEXT, label TEXT);
CREATE TABLE attr_value (         -- G01T01, G01I02...: attribute values
  etd TEXT, grupo TEXT, tabla TEXT, cetd TEXT, code TEXT, lang TEXT, label TEXT, orden TEXT);
CREATE TABLE modelnis (           -- model code -> attribute values (c01..c10)
  etd TEXT, codenis TEXT, grupo TEXT,
  c01 TEXT, c02 TEXT, c03 TEXT, c04 TEXT, c05 TEXT, c06 TEXT, c07 TEXT, c08 TEXT, c09 TEXT, c10 TEXT);
CREATE TABLE infosec (            -- INFOSEC / INFOSECF: section applicability by attributes
  etd TEXT, variant TEXT, secc TEXT, grupo TEXT, codigrup TEXT, desdat TEXT, finsdat TEXT,
  c01 TEXT, c02 TEXT, c03 TEXT, c04 TEXT, c05 TEXT, c06 TEXT, c07 TEXT, c08 TEXT, c09 TEXT, c10 TEXT);
CREATE TABLE hotspot (            -- KView zones (*.ini next to drawings); caption = item no. or section no.
  etd TEXT, kind TEXT, image TEXT, caption TEXT, x INTEGER, y INTEGER, w INTEGER, h INTEGER);
CREATE TABLE cinfo (etd TEXT, file TEXT);
"""

INDEXES = """
CREATE INDEX vin_vin ON vin(vin);
CREATE INDEX vin_codenis ON vin(codenis);
CREATE INDEX section_plate ON section(etd, plate, lang);
CREATE INDEX part_plate ON part(etd, plate, lang);
CREATE INDEX part_key ON part(part_key);
CREATE INDEX part_pnc ON part(pnc);
CREATE INDEX modelnis_codenis ON modelnis(codenis);
CREATE INDEX hotspot_image ON hotspot(etd, image);
CREATE VIRTUAL TABLE part_fts USING fts5(des, part_no, part_key, pnc, content='part', content_rowid='rowid');
INSERT INTO part_fts(rowid, des, part_no, part_key, pnc) SELECT rowid, des, part_no, part_key, pnc FROM part;
"""

# Lookups the program needs, added after loading (reversed VINs serve the "end of VIN" search).
SERVING_INDEXES = """
CREATE INDEX vin_rev ON vin(vin_rev);
CREATE INDEX part_ree_key ON part(ree_key);
CREATE INDEX part_pospie ON part(etd, pospie, lang);
CREATE INDEX infosec_scope ON infosec(etd, grupo, variant);
CREATE INDEX section_group ON section(etd, lang, codigrup);
"""

Row = dict[str, Any]


def clean(value: object) -> object:
    """Strip strings and map blank ones to None: source columns are fixed-width and space-padded."""
    if isinstance(value, str):
        return value.strip() or None
    return value


def text(row: Row, column: str) -> str | None:
    """A cleaned text column (None when blank or absent)."""
    value = clean(row.get(column))
    return None if value is None else str(value)


def part_key(number: str | None) -> str | None:
    """Search key of a part number: uppercase letters and digits only ("-23319-D9700" -> "23319D9700")."""
    return re.sub(r"[^0-9A-Z]", "", number.upper()) if number else None


def normalize_vin(raw: str | None) -> str | None:
    """VIN without any whitespace, uppercased; None when nothing is left."""
    return re.sub(r"\s+", "", raw or "").upper() or None


def read_jsonl(path: Path) -> Iterator[Row]:
    """Rows of one exported table; a missing table reads as empty."""
    if not path.exists():
        return
    with path.open(encoding="utf-8") as f:
        for line in f:
            yield json.loads(line)


def read_hotspots(ini: Path) -> list[tuple[str | None, int, int, int, int]]:
    """Clickable zones of a drawing, from a KView zone file: (caption, x, y, w, h) in source pixels.

    Sections without ``zonecaptionleft`` are not zones (KView settings) and are skipped.
    """
    parser = configparser.RawConfigParser(strict=False)
    parser.read(ini, encoding="latin-1")
    zones = []
    for name in parser.sections():
        zone = parser[name]
        if "zonecaptionleft" not in zone:
            continue
        zones.append(
            (
                text(dict(zone), "zonecaptionleft"),
                int(zone.get("zoneleft", "0")),
                int(zone.get("zonetop", "0")),
                int(zone.get("zonewidth", "0")),
                int(zone.get("zoneheight", "0")),
            )
        )
    return zones


def _insert(db: sqlite3.Connection, table: str, values: tuple[object, ...]) -> None:
    db.execute(f"INSERT INTO {table} VALUES ({','.join('?' * len(values))})", values)  # noqa: S608 - fixed names


def _load_attributes(db: sqlite3.Connection, etd: str, folder: Path) -> None:
    _load_attribute_tables(db, etd, folder)
    _load_attribute_values(db, etd, folder)
    _load_model_codes(db, etd, folder)
    _load_applicability(db, etd, folder)


def _load_attribute_tables(db: sqlite3.Connection, etd: str, folder: Path) -> None:
    for kind, table in (("T", "csel"), ("I", "cinf")):
        for r in read_jsonl(folder / f"{table}.jsonl"):
            for letter, lang in LANG_CODES.items():
                _insert(
                    db, "attr_table", (etd, text(r, "grupo"), kind, text(r, "tabla"), lang, text(r, letter))
                )


def _load_attribute_values(db: sqlite3.Connection, etd: str, folder: Path) -> None:
    # Attribute values live in one table per attribute: G01T01 = period G01, csel table T01.
    for path in sorted(folder.glob("*.jsonl")):
        match = re.fullmatch(r"(G\d\d)([TI]\d\d)", path.stem)
        if not match:
            continue
        for r in read_jsonl(path):
            for letter, lang in LANG_CODES.items():
                values = (etd, match[1], match[2], text(r, "CETD"), text(r, "CODE"), lang, text(r, letter))
                _insert(db, "attr_value", (*values, clean(r.get("Orden"))))


def _load_model_codes(db: sqlite3.Connection, etd: str, folder: Path) -> None:
    for r in read_jsonl(folder / "MODELNIS.jsonl"):
        lower = {k.lower(): v for k, v in r.items()}
        columns = tuple(clean(lower.get(f"c{i:02d}")) for i in range(1, 11))
        _insert(db, "modelnis", (etd, text(lower, "codenis"), text(lower, "grupo"), *columns))


def _load_applicability(db: sqlite3.Connection, etd: str, folder: Path) -> None:
    # INFOSECF (variant F) is the table NISCAT uses for applicability; INFOSEC is kept for reference.
    # The tenth attribute column is named C010, not C10.
    for variant, table in (("", "INFOSEC"), ("F", "INFOSECF")):
        for r in read_jsonl(folder / f"{table}.jsonl"):
            columns = tuple(clean(r.get(f"C{i:02d}" if i < 10 else "C010")) for i in range(1, 11))
            head = (etd, variant, text(r, "SECC162"), text(r, "grupo"), text(r, "CODIGRUP"))
            _insert(db, "infosec", (*head, text(r, "DESDAT"), text(r, "FINSDAT"), *columns))


def _load_parts(db: sqlite3.Connection, etd: str, lang: str, rows: list[Row]) -> None:
    # NISCAT prints an item number (ITE164) only on the first line of an item: the following lines of the same
    # plate inherit it (item_eff), in POSPIE order. A new plate starts without an item.
    last_plate, last_item = None, None
    for r in sorted(rows, key=lambda row: int(row["POSPIE"])):
        plate = (text(r, "CODILAM") or "").upper()
        item = text(r, "ITE164")
        if plate != last_plate:
            last_plate, last_item = plate, None
        if item:
            last_item = item
        number, replaced = text(r, "PIE164"), text(r, "REE164")
        levels = tuple(text(r, f"IND{i}164") for i in range(1, 6))
        _insert(
            db,
            "part",
            (
                etd, lang, r["POSPIE"], plate, item, last_item, text(r, "SEC164"),
                text(r, "NC164"), text(r, "S164"), *levels,
                number, part_key(number), text(r, "DES164"), text(r, "ESP164"), text(r, "CAN164"),
                text(r, "CAP164"), text(r, "PMC164"), text(r, "APP164"), text(r, "DATAPLIC"),
                text(r, "ICA164"), replaced, part_key(replaced), text(r, "KDF164"), text(r, "OBS164"),
            ),
        )  # fmt: skip


def _load_series(db: sqlite3.Connection, etd: str, folder: Path) -> None:
    for r in read_jsonl(folder / "grupos.jsonl"):
        _insert(db, "grupo", (etd, text(r, "GRUPO"), text(r, "ESP")))
    for r in read_jsonl(folder / "INDIGRAL.jsonl"):
        for letter, lang in LANG_CODES.items():
            if text(r, letter):
                _insert(db, "main_group", (etd, text(r, "CETD"), lang, text(r, letter)))
    _load_attributes(db, etd, folder)
    for letter, lang in LANG_CODES.items():
        if not (folder / f"{letter}_sec.jsonl").exists():
            continue  # a series may lack a language (some exist in English only)
        for r in read_jsonl(folder / f"{letter}_sec.jsonl"):
            number = text(r, "NUMSEC")
            plate = (
                f"{etd}{number}".upper()
            )  # the plate key, also the drawing name (AA230 -> img/AA/AA230.png)
            values = (etd, lang, plate, text(r, "CODIGRUP"), number, text(r, "NOMSEC"), text(r, "NOTAS"))
            _insert(db, "section", (*values, text(r, "DESDE"), text(r, "HASTA")))
        _load_parts(db, etd, lang, list(read_jsonl(folder / f"{letter}_PAR.jsonl")))


def _load_hotspots(db: sqlite3.Connection, series: Path) -> None:
    for kind, folder in (("plate", "img"), ("group", "gindex")):
        for ini in sorted(series.glob(f"{folder}/*/*.ini", case_sensitive=False)):
            if ini.name.lower().startswith("f_"):  # f_*.ini hold KView font settings, not zones
                continue
            etd, image = ini.parent.name.upper(), ini.stem.upper()
            for zone in read_hotspots(ini):
                _insert(db, "hotspot", (etd, kind, image, *zone))


def _load_catalogs(db: sqlite3.Connection, staging: Path) -> None:
    for r in read_jsonl(staging / "spa2__niscat" / "SERIES.jsonl"):
        langs = ",".join(lang for letter, lang in LANG_CODES.items() if r.get(letter))
        head = (text(r, "ETD"), text(r, "GRUPO"), text(r, "MODEL"), text(r, "CMODEL"), text(r, "TIPUS"))
        tail = (text(r, "FROM"), text(r, "UP_TO"), text(r, "SERIE"), text(r, "DATA"), text(r, "ORDEN"), langs)
        _insert(db, "catalog", (*head, *tail))
    for r in read_jsonl(staging / "spa2__vn" / "v1t.jsonl"):
        raw = r.get("VIN")
        values = (normalize_vin(raw), raw, text(r, "CODENIS"), text(r, "MIMODEL"), text(r, "TIPO"))
        _insert(db, "vin", (*values, text(r, "PRODATA")))


def _optimise(db: sqlite3.Connection) -> None:
    db.create_function("rev", 1, lambda s: s[::-1] if s else None, deterministic=True)
    db.execute("ALTER TABLE vin ADD COLUMN vin_rev TEXT")
    db.execute("UPDATE vin SET vin_rev = rev(vin)")
    db.executescript(SERVING_INDEXES)
    db.commit()
    db.execute("ANALYZE")
    db.commit()
    db.execute("VACUUM")


def _require(staging: Path, installer: Path) -> None:
    # Optional tables read as empty, so a wrong folder would otherwise build an empty database silently.
    required = [staging / "spa2__niscat" / "SERIES.jsonl", staging / "spa2__vn" / "v1t.jsonl"]
    required.append(installer / "spa2" / "series")
    for path in required:
        if not path.exists():
            msg = f"{path} is missing: is it a NISCAT export (staging) and installer?"
            raise FileNotFoundError(msg)
    if not any(staging.glob("spa2__series__*")):
        msg = f"no spa2__series__* folder in {staging}: the series tables were not exported"
        raise FileNotFoundError(msg)


def build_database(staging: Path, installer: Path, path: Path) -> None:
    """Create ``path`` (overwritten) from the staging export and the installer's zone files and documents.

    Raises:
        FileNotFoundError: the staging export or the installer lacks a required part.
    """
    _require(staging, installer)
    path.unlink(missing_ok=True)
    db = sqlite3.connect(path)
    try:
        db.executescript(SCHEMA)
        _load_catalogs(db, staging)
        for folder in sorted(staging.glob("spa2__series__*")):
            _load_series(db, folder.name.rsplit("__", 1)[1], folder)
        series = installer / "spa2" / "series"
        _load_hotspots(db, series)
        for document in sorted(series.glob("cinfo/*/*")):
            _insert(db, "cinfo", (document.parent.name.upper(), document.name))
        db.executescript(INDEXES)
        db.commit()
        _optimise(db)
    finally:
        db.close()
