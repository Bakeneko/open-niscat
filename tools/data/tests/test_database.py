import shutil
import sqlite3
from pathlib import Path

import pytest

from niscat_data.database import build_database, clean, normalize_vin, part_key, read_hotspots

INDEXES = {
    "vin_vin",
    "vin_codenis",
    "vin_rev",
    "section_plate",
    "section_group",
    "part_plate",
    "part_key",
    "part_pnc",
    "part_pospie",
    "part_ree_key",
    "modelnis_codenis",
    "hotspot_image",
    "infosec_scope",
}


def test_clean_strips_and_maps_blank_to_none() -> None:
    assert clean(" x ") == "x"
    assert clean("   ") is None
    assert clean(0) == 0
    assert clean(None) is None


def test_part_key_and_vin_normalisation() -> None:
    assert part_key("-23319-D9700") == "23319D9700"
    assert part_key("d-4100-17c90") == "D410017C90"
    assert part_key(None) is None
    assert normalize_vin(" vsk bec\t220 ") == "VSKBEC220"
    assert normalize_vin("   ") is None


def test_read_hotspots_skips_zones_without_caption(tmp_path: Path) -> None:
    ini = tmp_path / "x.ini"
    ini.write_text(
        "[Zone1]\nzonecaptionleft=230\nzoneleft=10\nzonetop=20\nzonewidth=30\nzoneheight=40\n"
        "[Zone2]\nzoneleft=1\n"
        "[Zone3]\nzonecaptionleft=2é\n",
        encoding="latin-1",
    )
    assert read_hotspots(ini) == [("230", 10, 20, 30, 40), ("2é", 0, 0, 0, 0)]


def build(staging: Path, installer: Path, tmp_path: Path) -> sqlite3.Connection:
    path = tmp_path / "data.db"
    build_database(staging, installer, path)
    return sqlite3.connect(path)


def test_build_database_tables(staging: Path, installer: Path, tmp_path: Path) -> None:
    db = build(staging, installer, tmp_path)
    count = {
        t: db.execute(f"SELECT COUNT(*) FROM {t}").fetchone()[0]
        for t in ("catalog", "vin", "section", "part")
    }
    assert count == {"catalog": 2, "vin": 2, "section": 4, "part": 8}
    assert db.execute("SELECT langs FROM catalog WHERE etd = 'AA'").fetchone() == ("en,fr",)
    assert db.execute("SELECT date_from, date_to FROM catalog WHERE etd = 'AA'").fetchone() == (
        "04/87",
        "11/94",
    )
    assert db.execute("SELECT label FROM main_group WHERE lang = 'fr'").fetchone() == ("ELECTRICITE MOTEUR",)
    assert db.execute("SELECT COUNT(*) FROM main_group").fetchone() == (4,)  # Italian label is NULL: skipped
    assert db.execute("SELECT DISTINCT plate FROM section ORDER BY 1").fetchall() == [("AA230",), ("AA231A",)]


def test_build_database_vins(staging: Path, installer: Path, tmp_path: Path) -> None:
    db = build(staging, installer, tmp_path)
    rows = db.execute("SELECT vin, vin_raw, vin_rev FROM vin ORDER BY vin").fetchall()
    assert rows[1] == ("VSKBEC220U0990494", "VSK BEC220U0990494", "4940990U022CEBKSV")


def test_build_database_carries_items_forward(staging: Path, installer: Path, tmp_path: Path) -> None:
    db = build(staging, installer, tmp_path)
    rows = db.execute(
        "SELECT pospie, item, item_eff, sec, s, part_key, ree_key, pnc FROM part"
        " WHERE lang = 'en' ORDER BY rowid"
    ).fetchall()
    assert rows == [
        (1, "01", "01", "01", "*", "11715D5500", None, None),
        (2, None, "01", "02", None, "11716D5500", None, None),  # no item of its own: carried from line 1
        (3, "02", "02", "01", "#", "23319D9700", None, "23319"),
        (4, None, None, "01", None, "23100D9701", "23100D9700", None),  # new plate: nothing to carry
    ]


def test_build_database_hotspots_and_documents(staging: Path, installer: Path, tmp_path: Path) -> None:
    db = build(staging, installer, tmp_path)
    assert db.execute(
        "SELECT etd, kind, image, caption, x, y, w, h FROM hotspot ORDER BY kind"
    ).fetchall() == [
        ("AA", "group", "B", "230", 1, 2, 3, 4),
        ("AA", "plate", "AA230", "01", 10, 20, 30, 40),
        ("AA", "plate", "AA230", "02", 50, 0, 0, 0),
    ]
    assert db.execute("SELECT etd, file FROM cinfo").fetchall() == [("AA", "g0101.pdf")]


def test_build_database_indexes_search_and_statistics(staging: Path, installer: Path, tmp_path: Path) -> None:
    db = build(staging, installer, tmp_path)
    names = {r[0] for r in db.execute("SELECT name FROM sqlite_master WHERE type = 'index'") if r[0]}
    assert names >= INDEXES
    hits = db.execute(
        "SELECT part_no FROM part WHERE rowid IN (SELECT rowid FROM part_fts WHERE part_fts MATCH 'palier')"
    ).fetchall()
    assert {h[0] for h in hits} == {"-23319-D9700"}
    assert db.execute("SELECT COUNT(*) FROM sqlite_stat1").fetchone()[0] > 0
    infosec = db.execute("SELECT variant, c01, c10 FROM infosec ORDER BY variant").fetchall()
    assert infosec == [("", "-", "-"), ("F", "1", "-")]


@pytest.mark.parametrize(
    ("missing", "message"),
    [
        ("staging/spa2__niscat/SERIES.jsonl", "SERIES.jsonl"),
        ("staging/spa2__vn/v1t.jsonl", "v1t.jsonl"),
        ("staging/spa2__series__AA", "spa2__series__"),
        ("installer/spa2", "series"),
    ],
)
def test_build_database_refuses_incomplete_inputs(
    staging: Path, installer: Path, tmp_path: Path, missing: str, message: str
) -> None:
    shutil.rmtree(staging / "spa2__series__AB")  # leave AA as the only series
    target = tmp_path / missing
    if target.is_dir():
        shutil.rmtree(target)
    else:
        target.unlink()
    with pytest.raises(FileNotFoundError, match=message):
        build_database(staging, installer, tmp_path / "data.db")


def test_build_database_finds_upper_case_zone_files(staging: Path, installer: Path, tmp_path: Path) -> None:
    group = installer / "spa2" / "series" / "gindex" / "aa"
    (group / "b.ini").rename(group / "B.INI")
    db = build(staging, installer, tmp_path)
    assert db.execute("SELECT COUNT(*) FROM hotspot WHERE kind = 'group'").fetchone() == (1,)
