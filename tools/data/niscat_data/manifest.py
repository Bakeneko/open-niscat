"""Write manifest.json, which describes a data folder to the Open Niscat program."""

import json
import re
from datetime import UTC, date, datetime
from pathlib import Path

# Schema of data.db that the Go program supports (catalog.SchemaVersion, checked by a test).
SCHEMA = 1
# NISCAT edition this tool reads, as YYYY-MM. The source tables do not store it: NISCAT shows "Ed. 01/2015".
EDITION = "2015-01"
# Bump when the tool's output changes for the same edition; it is part of the data version.
REVISION = 2


def write_manifest(out: Path, edition: str = EDITION, today: date | None = None) -> dict[str, object]:
    """Write ``out/manifest.json`` and return its content.

    ``version`` identifies a delivery of the data and orders them: the edition as YYYY.MM, then the revision
    of this tool ("2015.01-2").
    """
    match = re.fullmatch(r"(\d{4})-(0[1-9]|1[0-2])", edition)
    if not match:
        msg = f"edition must be YYYY-MM, got {edition!r}"
        raise ValueError(msg)
    manifest: dict[str, object] = {
        "schema": SCHEMA,
        "version": f"{match[1]}.{match[2]}-{REVISION}",
        "source": {"name": "NISCAT", "publisher": "Nissan Motor Ibérica", "edition": edition},
        "build": {
            "tool": "niscat-data",
            "revision": REVISION,
            "date": (today or datetime.now(UTC).astimezone().date()).isoformat(),
        },
    }
    text = json.dumps(manifest, indent=2, ensure_ascii=False) + "\n"
    (out / "manifest.json").write_text(text, encoding="utf-8")
    return manifest
