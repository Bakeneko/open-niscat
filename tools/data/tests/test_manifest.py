import json
import re
from datetime import date
from pathlib import Path

import pytest

from niscat_data.manifest import SCHEMA, write_manifest

REPO = Path(__file__).resolve().parents[3]


def test_write_manifest(tmp_path: Path) -> None:
    manifest = write_manifest(tmp_path, today=date(2026, 10, 7))
    text = (tmp_path / "manifest.json").read_text(encoding="utf-8")
    assert (
        json.loads(text)
        == manifest
        == {
            "schema": 1,
            "version": "2015.01-2",
            "source": {"name": "NISCAT", "publisher": "Nissan Motor Ibérica", "edition": "2015-01"},
            "build": {"tool": "niscat-data", "revision": 2, "date": "2026-10-07"},
        }
    )
    assert "Ibérica" in text  # written as UTF-8, not escaped
    assert text.endswith("}\n")
    assert '\n  "schema": 1,' in text  # 2-space indent


@pytest.mark.parametrize("edition", ["01/2015", "2015-13", "2015-1", ""])
def test_malformed_edition(tmp_path: Path, edition: str) -> None:
    with pytest.raises(ValueError, match="edition"):
        write_manifest(tmp_path, edition=edition)


def test_schema_matches_the_go_program() -> None:
    source = (REPO / "internal" / "catalog" / "store.go").read_text(encoding="utf-8")
    match = re.search(r"const SchemaVersion = (\d+)", source)
    assert match is not None
    assert int(match[1]) == SCHEMA
