"""Compare two data folders, to check that a change of the tool leaves its output equivalent.

Equivalent means: same tables, columns, rows (in any order) and indexes in data.db; same pixels in every PNG
(encodings may differ between Pillow versions); same bytes in every other file; same manifest except the build
date. Full-text search indexes are compared by their vocabulary (term, documents, occurrences), not by their
shadow tables, whose storage layout depends on insertion history.
"""

import json
import sqlite3
from pathlib import Path

from PIL import Image

SKIPPED_TABLES = ("sqlite_stat",)  # ANALYZE statistics depend on the run


def _tables(db: sqlite3.Connection, schema: str) -> dict[str, str]:
    rows = db.execute(f"SELECT name, sql FROM {schema}.sqlite_master WHERE type = 'table'").fetchall()  # noqa: S608
    virtual = [name for name, sql in rows if sql and sql.upper().startswith("CREATE VIRTUAL TABLE")]
    return {
        name: sql
        for name, sql in rows
        if not name.startswith(SKIPPED_TABLES) and not any(name.startswith(f"{v}_") for v in virtual)
    }


def _indexes(db: sqlite3.Connection, schema: str) -> dict[str, str]:
    rows = db.execute(f"SELECT name, sql FROM {schema}.sqlite_master WHERE type = 'index'").fetchall()  # noqa: S608
    return {name: " ".join(sql.split()) for name, sql in rows if sql}


def _columns(db: sqlite3.Connection, schema: str, table: str) -> list[tuple[str, str]]:
    return [(r[1], r[2]) for r in db.execute(f'PRAGMA {schema}.table_info("{table}")')]


def _rows_differ(db: sqlite3.Connection, table: str, columns: list[tuple[str, str]]) -> bool:
    # Compare multisets inside SQLite: group identical rows with their count, then diff both ways.
    names = ", ".join(f'"{name}"' for name, _ in columns)
    side = 'SELECT {names}, COUNT(*) FROM {schema}."{table}" GROUP BY {names}'
    a = side.format(names=names, schema="main", table=table)
    b = side.format(names=names, schema="other", table=table)
    query = f"SELECT EXISTS ({a} EXCEPT {b}) OR EXISTS ({b} EXCEPT {a})"
    return bool(db.execute(query).fetchone()[0])


def _search_index_differs(db: sqlite3.Connection, table: str) -> bool:
    # fts5vocab tables live in the temp schema, so both read-only databases can be inspected.
    for schema in ("main", "other"):
        db.execute(f'CREATE VIRTUAL TABLE temp."{schema}_{table}" USING fts5vocab({schema}, "{table}", row)')
    names = (f'temp."{schema}_{table}"' for schema in ("main", "other"))
    a, b = (f"SELECT term, doc, cnt FROM {name}" for name in names)  # noqa: S608 - names from sqlite_master
    return bool(db.execute(f"SELECT EXISTS ({a} EXCEPT {b}) OR EXISTS ({b} EXCEPT {a})").fetchone()[0])


def _read_only(path: Path) -> str:
    # A file: URI with mode=ro; as_uri() escapes characters such as % and # that SQLite would interpret.
    return f"{path.resolve().as_uri()}?mode=ro"


def _compare_database(a: Path, b: Path) -> list[str]:
    db = sqlite3.connect(_read_only(a), uri=True)
    db.execute("ATTACH DATABASE ? AS other", (_read_only(b),))
    try:
        differences = []
        tables_a, tables_b = _tables(db, "main"), _tables(db, "other")
        differences += [f"data.db: table {t} only in {a.parent}" for t in sorted(tables_a.keys() - tables_b)]
        differences += [f"data.db: table {t} only in {b.parent}" for t in sorted(tables_b.keys() - tables_a)]
        for table in sorted(tables_a.keys() & tables_b):
            columns = _columns(db, "main", table)
            if columns != _columns(db, "other", table):
                differences.append(f"data.db: {table}: columns differ")
            elif tables_a[table].upper().startswith("CREATE VIRTUAL TABLE"):
                if " USING FTS5" in tables_a[table].upper() and _search_index_differs(db, table):
                    differences.append(f"data.db: {table}: search index differs")
            elif _rows_differ(db, table, columns):
                differences.append(f"data.db: {table}: rows differ")
        indexes_a, indexes_b = _indexes(db, "main"), _indexes(db, "other")
        differences += [
            f"data.db: index {name} differs or is missing"
            for name in sorted(indexes_a.keys() | indexes_b.keys())
            if indexes_a.get(name) != indexes_b.get(name)
        ]
        return differences
    finally:
        db.close()


def _same_pixels(a: Path, b: Path) -> bool:
    with Image.open(a) as image_a, Image.open(b) as image_b:
        return (image_a.mode, image_a.size, image_a.tobytes()) == (
            image_b.mode,
            image_b.size,
            image_b.tobytes(),
        )


def _manifest(path: Path) -> object:
    manifest = json.loads(path.read_text(encoding="utf-8"))
    build = manifest.get("build")
    if isinstance(build, dict):
        build.pop("date", None)  # every build has its own date
    return manifest


def _files(root: Path) -> set[str]:
    return {p.relative_to(root).as_posix() for p in root.rglob("*") if p.is_file()}


def compare(a: Path, b: Path) -> list[str]:
    """Differences between data folders ``a`` and ``b``, as readable messages; empty when equivalent."""
    differences = _compare_database(a / "data.db", b / "data.db")
    if _manifest(a / "manifest.json") != _manifest(b / "manifest.json"):
        differences.append("manifest.json differs (build date ignored)")
    files_a, files_b = _files(a) - {"data.db", "manifest.json"}, _files(b) - {"data.db", "manifest.json"}
    differences += [f"{f}: only in {a}" for f in sorted(files_a - files_b)]
    differences += [f"{f}: only in {b}" for f in sorted(files_b - files_a)]
    for name in sorted(files_a & files_b):
        file_a, file_b = a / name, b / name
        if name.lower().endswith(".png"):
            if not _same_pixels(file_a, file_b):
                differences.append(f"{name}: pixels differ")
        elif file_a.read_bytes() != file_b.read_bytes():
            differences.append(f"{name}: bytes differ")
    return differences
