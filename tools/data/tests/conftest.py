"""Shared fixtures: synthetic NISCAT inputs built in temporary folders (never real data).

The shapes mirror what export_mdb.ps1 writes (one JSON object per row, source column names, fixed-width
strings padded with spaces) and what the installer holds (KView zone files next to the drawings).
"""

import json
from collections.abc import Mapping, Sequence
from pathlib import Path

import pytest

LANG_COLUMNS = {"L": "es", "Q": "en", "D": "de", "F": "fr", "I": "it"}


def write_jsonl(path: Path, rows: Sequence[Mapping[str, object]]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text("".join(json.dumps(r) + "\n" for r in rows), encoding="utf-8")


def labels(en: str, fr: str | None = None) -> dict[str, object]:
    """One label per source language column; Spanish/German/Italian copy English like many source rows."""
    return {"L": en, "Q": en, "D": en, "F": fr or en, "I": None}


def part_row(
    pospie: int, plate: str, item: str, sec: str, number: str, des: str, **extra: str
) -> dict[str, object]:
    return {
        "CODILAM": plate,
        "NC164": " ",
        "S164": extra.get("mark", " "),
        "ITE164": item.ljust(2),
        "SEC164": sec,
        "IND1164": "-",
        "IND2164": " ",
        "IND3164": " ",
        "IND4164": " ",
        "IND5164": " ",
        "PIE164": number.ljust(16),
        "DES164": des.ljust(25),
        "ESP164": " " * 25,
        "CAN164": " 1",
        "CAP164": "   ",
        "PMC164": " ",
        "APP164": " " * 18,
        "DATAPLIC": extra.get("dataplic", "0487-1295"),
        "ICA164": "   ",
        "REE164": extra.get("ree", "").ljust(16),
        "KDF164": "*",
        "OBS164": extra.get("pnc", "").rjust(14),
        "POSPIE": pospie,
    }


@pytest.fixture
def staging(tmp_path: Path) -> Path:
    """A staging folder as export_mdb.ps1 would write it, for one series (AA) in English and French."""
    root = tmp_path / "staging"
    write_jsonl(
        root / "spa2__niscat" / "SERIES.jsonl",
        [
            {
                "MODEL": "VANETTE",
                "CMODEL": "C220",
                "TIPUS": "LHD",
                "FROM": "04/87",
                "UP_TO": "11/94",
                "SERIE": "0049",
                "GRUPO": "G01",
                "ETD": "AA",
                "DATA": "0049 TEST",
                "ORDEN": "N001",
                "Q": 1,
                "L": 0,
                "D": 0,
                "F": 1,
                "I": 0,
            },
            {
                "MODEL": "TRADE",
                "CMODEL": "KF30",
                "TIPUS": "RHD",
                "FROM": "06/94",
                "UP_TO": "06/96",
                "SERIE": "0111",
                "GRUPO": "G01",
                "ETD": "AB",
                "DATA": "0111 TEST",
                "ORDEN": "N002",
                "Q": 1,
                "L": 0,
                "D": 0,
                "F": 0,
                "I": 0,
            },
        ],
    )
    write_jsonl(
        root / "spa2__vn" / "v1t.jsonl",
        [
            {
                "VIN": "VSK BEC220U0990494",
                "CODENIS": "BELC220QSKVX",
                "MIMODEL": "AA",
                "TIPO": "4",
                "PRODATA": "198905",
            },
            {
                "VIN": "SJNVC220R00000001",
                "CODENIS": "RMODEL",
                "MIMODEL": "AB",
                "TIPO": "4",
                "PRODATA": "198801",
            },
        ],
    )
    aa = root / "spa2__series__AA"
    write_jsonl(aa / "grupos.jsonl", [{"GRUPO": "G01", "ESP": "0049  from 04/87 up to 11/94"}])
    write_jsonl(aa / "INDIGRAL.jsonl", [{"CETD": "B", **labels("ENGINE ELECTRICAL", "ELECTRICITE MOTEUR")}])
    write_jsonl(aa / "csel.jsonl", [{"grupo": "G01", "tabla": "T01", **labels("ENGINE", "MOTEUR")}])
    write_jsonl(aa / "cinf.jsonl", [{"grupo": "G01", "tabla": "I01", **labels("BODY", "CARROSSERIE")}])
    write_jsonl(aa / "G01T01.jsonl", [{"CETD": "1", "CODE": "LD20", **labels("LD20"), "Orden": 1}])
    write_jsonl(
        aa / "MODELNIS.jsonl",
        [
            {
                "codenis": "BELC220QSKVX",
                "grupo": "G01",
                **{f"c{i:02d}": "1" for i in range(1, 10)},
                "c10": None,
            }
        ],
    )
    sec_row = {"SECC162": "230", "grupo": "G01", "CODIGRUP": "B", "DESDAT": "198704", "FINSDAT": "199512"}
    columns = {f"C{i:02d}": "-" for i in range(1, 10)} | {"C010": "-"}  # the tenth column is named C010
    write_jsonl(aa / "INFOSEC.jsonl", [sec_row | columns])
    write_jsonl(aa / "INFOSECF.jsonl", [sec_row | columns | {"C01": "1"}])
    for code, name in (("Q", "ALTERNATOR"), ("F", "ALTERNATEUR")):
        write_jsonl(
            aa / f"{code}_sec.jsonl",
            [
                {
                    "CODIGRUP": "B",
                    "NOMSEC": name,
                    "NOTAS": "LD20",
                    "DESDE": "04-87",
                    "HASTA": "12-95",
                    "NUMSEC": "230",
                },
                {
                    "CODIGRUP": "B",
                    "NOMSEC": name,
                    "NOTAS": None,
                    "DESDE": "04-87",
                    "HASTA": "12-95",
                    "NUMSEC": "231a",
                },
            ],
        )
        des = {"Q": ("PULLEY", "PALIER"), "F": ("POULIE", "PALIER")}[code]
        # Deliberately out of POSPIE order: the item carry-forward must follow POSPIE, not file order.
        write_jsonl(
            aa / f"{code}_PAR.jsonl",
            [
                part_row(3, "AA230", "02", "01", "-23319-D9700", des[1], mark="#", pnc="23319"),
                part_row(1, "AA230", "01", "01", "-11715-D5500", des[0], mark="*"),
                part_row(2, "AA230", "", "02", "-11716-D5500", des[0]),
                part_row(4, "AA231A", "", "01", "-23100-D9701", des[0], ree="-23100-D9700"),
            ],
        )
    write_jsonl(root / "spa2__series__AB" / "grupos.jsonl", [{"GRUPO": "G01", "ESP": "0111"}])
    for table in ("INDIGRAL", "csel", "cinf", "MODELNIS", "INFOSEC", "INFOSECF"):
        write_jsonl(root / "spa2__series__AB" / f"{table}.jsonl", [])
    return root


@pytest.fixture
def installer(tmp_path: Path) -> Path:
    """An installer tree with KView zone files, a font settings file, drawings folders and a PDF."""
    root = tmp_path / "installer"
    series = root / "spa2" / "series"
    (root / "spa2").mkdir(parents=True)
    (root / "spa2" / "niscat.mdb").write_bytes(b"")
    (root / "spa2" / "vn.mdb").write_bytes(b"")
    plate = series / "img" / "aa"
    plate.mkdir(parents=True)
    (plate / "aa230.ini").write_text(
        "[Zone1]\nzonecaptionleft=01\nzoneleft=10\nzonetop=20\nzonewidth=30\nzoneheight=40\n"
        "[Zone2]\nzoneleft=5\n"
        "[Zone3]\nzonecaptionleft=02\nzoneleft=50\n",
        encoding="latin-1",
    )
    (plate / "f_aa230.ini").write_text("[Zone1]\nzonecaptionleft=FONT\n", encoding="latin-1")
    group = series / "gindex" / "aa"
    group.mkdir(parents=True)
    (group / "b.ini").write_text(
        "[Zone1]\nzonecaptionleft=230\nzoneleft=1\nzonetop=2\nzonewidth=3\nzoneheight=4\n"
    )
    docs = series / "cinfo" / "aa"
    docs.mkdir(parents=True)
    (docs / "g0101.pdf").write_bytes(b"%PDF-1.4 test")
    return root
