"""Convert the NISCAT drawings to PNG and copy the catalog documents.

The installer stores bilevel CCITT Group 4 TIFF drawings: parts plates as ``img/<series>/<plate>.etd``
and group indexes as ``gindex/<series>/<group>.tif``. Browsers cannot show them, so they become PNG files of
the same size: the hotspot coordinates read from the KView zone files are in source pixels. Folder and file
names are upper-cased, as the database keys are (series AA, plate AA230).
"""

import shutil
from concurrent.futures import ProcessPoolExecutor
from pathlib import Path

from PIL import Image

# (installer sub-folder, drawing file pattern) -> converted into <out>/<sub-folder>/<SERIES>/<NAME>.png
DRAWINGS = (("img", "*.etd"), ("gindex", "*.tif"))


def _convert(job: tuple[Path, Path]) -> None:
    source, target = job
    with Image.open(source) as image:
        image.save(target, "PNG")


def convert_drawings(installer: Path, out: Path, jobs: int) -> tuple[int, int]:
    """Write the PNG drawings and the documents under ``out``; returns (drawings, documents) counts."""
    series = installer / "spa2" / "series"
    work: list[tuple[Path, Path]] = []
    for folder, pattern in DRAWINGS:
        # Some sources are upper-case (B.TIF): match patterns regardless of case on every system.
        for source in sorted(series.glob(f"{folder}/*/{pattern}", case_sensitive=False)):
            target = out / folder / source.parent.name.upper() / f"{source.stem.upper()}.png"
            target.parent.mkdir(parents=True, exist_ok=True)
            work.append((source, target))
    with ProcessPoolExecutor(max_workers=max(1, jobs)) as pool:
        list(pool.map(_convert, work, chunksize=32))

    # Documents keep their own names (G0103.PDF stays upper-case); the database lists them as found.
    documents = 0
    for source in sorted(p for p in series.glob("cinfo/*/*") if p.is_file()):
        target = out / "cinfo" / source.parent.name.upper() / source.name
        target.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(source, target)
        documents += 1
    return len(work), documents
