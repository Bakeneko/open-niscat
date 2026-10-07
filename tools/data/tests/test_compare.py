import json
import shutil
import sqlite3
import subprocess
import sys
from pathlib import Path

import pytest
from PIL import Image

from compare_data import main
from niscat_data.compare import compare


def make_data(root: Path) -> Path:
    """A small data folder: a database with two tables and an index, a drawing, a document, a manifest."""
    root.mkdir(parents=True)
    db = sqlite3.connect(root / "data.db")
    db.executescript(
        "CREATE TABLE part (etd TEXT, pospie INTEGER, des TEXT);"
        "CREATE TABLE vin (vin TEXT);"
        "CREATE INDEX part_etd ON part(etd);"
        "INSERT INTO part VALUES ('AA', 1, 'PALIER'), ('AA', 2, NULL);"
        "INSERT INTO vin VALUES ('V1');"
        "CREATE VIRTUAL TABLE part_fts USING fts5(des, content='part', content_rowid='rowid');"
        "INSERT INTO part_fts(part_fts) VALUES ('rebuild');"
        "ANALYZE;"
    )
    db.commit()
    db.close()
    (root / "img" / "AA").mkdir(parents=True)
    image = Image.new("1", (8, 4), 1)
    image.putpixel((1, 1), 0)
    image.save(root / "img" / "AA" / "AA230.png", optimize=False)
    (root / "cinfo" / "AA").mkdir(parents=True)
    (root / "cinfo" / "AA" / "g0101.pdf").write_bytes(b"%PDF-1.4 test")
    manifest = {"schema": 1, "version": "2015.01-2", "build": {"revision": 2, "date": "2026-10-07"}}
    (root / "manifest.json").write_text(json.dumps(manifest), encoding="utf-8")
    return root


@pytest.fixture
def pair(tmp_path: Path) -> tuple[Path, Path]:
    a = make_data(tmp_path / "a")
    b = tmp_path / "b"
    shutil.copytree(a, b)
    return a, b


def sql(folder: Path, statement: str) -> None:
    db = sqlite3.connect(folder / "data.db")
    db.executescript(statement)
    db.commit()
    db.close()


def test_identical_folders(pair: tuple[Path, Path]) -> None:
    assert compare(*pair) == []


def test_row_order_does_not_matter(pair: tuple[Path, Path]) -> None:
    a, b = pair
    sql(b, "DELETE FROM part; INSERT INTO part VALUES ('AA', 2, NULL), ('AA', 1, 'PALIER');")
    assert compare(a, b) == []


@pytest.mark.parametrize(
    ("change", "message"),
    [
        ("UPDATE part SET des = 'POULIE' WHERE pospie = 1", "part: rows differ"),
        ("ALTER TABLE vin ADD COLUMN vin_rev TEXT", "vin: columns differ"),
        ("DROP INDEX part_etd", "index part_etd"),
        ("CREATE TABLE extra (x)", "table extra"),
    ],
)
def test_database_differences(pair: tuple[Path, Path], change: str, message: str) -> None:
    a, b = pair
    sql(b, change)
    differences = compare(a, b)
    assert any(message in d for d in differences), differences


def test_drawing_pixels_not_bytes(pair: tuple[Path, Path]) -> None:
    a, b = pair
    png = b / "img" / "AA" / "AA230.png"
    with Image.open(png) as image:
        image.copy().save(png, optimize=True, compress_level=9)  # same pixels, other encoding
    assert compare(a, b) == []
    with Image.open(png) as image:
        changed = image.copy()
    changed.putpixel((0, 0), 0)
    changed.save(png)
    assert any("img/AA/AA230.png: pixels differ" in d for d in compare(a, b))


def test_documents_files_and_manifest(pair: tuple[Path, Path]) -> None:
    a, b = pair
    (b / "cinfo" / "AA" / "g0101.pdf").write_bytes(b"%PDF-1.4 other")
    (b / "img" / "AA" / "AA231.png").write_bytes(b"")
    manifest = json.loads((b / "manifest.json").read_text(encoding="utf-8"))
    manifest["build"]["date"] = "2030-01-01"  # ignored: every build has its own date
    (b / "manifest.json").write_text(json.dumps(manifest), encoding="utf-8")
    differences = compare(a, b)
    assert any("cinfo/AA/g0101.pdf: bytes differ" in d for d in differences)
    assert any("img/AA/AA231.png: only in" in d for d in differences)
    assert not any("manifest" in d for d in differences)
    manifest["version"] = "2015.01-3"
    (b / "manifest.json").write_text(json.dumps(manifest), encoding="utf-8")
    assert any("manifest.json" in d for d in compare(a, b))


def test_command_line(pair: tuple[Path, Path], capsys: pytest.CaptureFixture[str]) -> None:
    a, b = pair
    assert main([str(a), str(b)]) == 0
    assert "equivalent" in capsys.readouterr().out
    sql(b, "DELETE FROM vin")
    assert main([str(a), str(b)]) == 1
    assert "vin: rows differ" in capsys.readouterr().out


def test_search_index(pair: tuple[Path, Path]) -> None:
    a, b = pair
    sql(b, "INSERT INTO part_fts(part_fts) VALUES ('delete-all')")  # a build that forgot the index
    assert any("part_fts: search index differs" in d for d in compare(a, b))


def test_null_and_empty_text_differ(pair: tuple[Path, Path]) -> None:
    a, b = pair
    sql(b, "UPDATE part SET des = '' WHERE des IS NULL")
    assert any("part: rows differ" in d for d in compare(a, b))


def test_paths_needing_escaping(tmp_path: Path) -> None:
    a = make_data(tmp_path / "c%23d é")
    b = tmp_path / "other #1"
    shutil.copytree(a, b)
    assert compare(a, b) == []


def test_missing_pillow_is_one_message(tmp_path: Path) -> None:
    script = Path(__file__).parents[1] / "compare_data.py"
    code = (
        "import runpy, sys; sys.modules['PIL'] = None; "
        f"sys.argv = ['compare_data.py', {str(tmp_path)!r}, {str(tmp_path)!r}]; "
        f"runpy.run_path({str(script)!r}, run_name='__main__')"
    )
    result = subprocess.run([sys.executable, "-c", code], capture_output=True, text=True, check=False)
    assert result.returncode == 1
    assert "Pillow" in result.stderr
    assert "Traceback" not in result.stderr
