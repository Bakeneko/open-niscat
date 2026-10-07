"""Build the Open Niscat data folder from a NISCAT 01/2015 installer.

Usage: python tools/data/build_data.py --installer <NISCAT installer folder> [--out data]

Steps: export the Access 97 tables (Windows only: 32-bit PowerShell and its Jet driver), build data.db,
convert the drawings, write manifest.json. Everything is written into <out>.tmp, which replaces <out> only
when every step succeeded: a failed build never leaves a half-written data folder.
"""

import argparse
import contextlib
import os
import shutil
import sqlite3
import subprocess
import sys
import tempfile
import time
from pathlib import Path

if sys.version_info < (3, 12):  # noqa: UP036 - give older interpreters a message, not a SyntaxError later
    print("error: Python 3.12 or newer is required", file=sys.stderr)
    sys.exit(1)

try:
    from niscat_data.database import build_database
    from niscat_data.drawings import convert_drawings
    from niscat_data.manifest import EDITION, write_manifest
except ModuleNotFoundError as missing:  # Pillow is the only third-party dependency
    print(
        f"error: {missing.name} is missing: install Pillow"
        " (python -m pip install pillow, or make tools-install)",
        file=sys.stderr,
    )
    sys.exit(1)

# Jet 4.0, the only driver able to read the Access 97 files, exists for 32-bit programs only.
POWERSHELL_32 = Path(r"C:\Windows\SysWOW64\WindowsPowerShell\v1.0\powershell.exe")
EXPORT_SCRIPT = Path(__file__).with_name("export_mdb.ps1")


class BuildError(Exception):
    """A problem the user can fix; reported as one message."""


def _jobs(value: str) -> int:
    # Windows' process pools accept at most 61 workers.
    jobs = int(value)
    if jobs < 1:
        msg = "must be at least 1"
        raise argparse.ArgumentTypeError(msg)
    return min(jobs, 61)


def _parse(argv: list[str] | None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Build the Open Niscat data folder from a NISCAT installer.")
    parser.add_argument("--installer", type=Path, required=True, help="NISCAT 01/2015 installer folder")
    parser.add_argument(
        "--out", type=Path, default=Path("data"), help="data folder to create (default: data)"
    )
    parser.add_argument("--staging", type=Path, help="folder for the exported tables (default: temporary)")
    parser.add_argument(
        "--skip-export", action="store_true", help="reuse the tables already exported in --staging"
    )
    parser.add_argument("--force", action="store_true", help="replace an existing --out folder")
    parser.add_argument(
        "--jobs",
        type=_jobs,
        default=min(os.cpu_count() or 1, 61),
        help="parallel drawing conversions (default: number of CPUs, 61 at most)",
    )
    parser.add_argument("--edition", default=EDITION, help=f"NISCAT edition as YYYY-MM (default: {EDITION})")
    args = parser.parse_args(argv)
    if args.skip_export and args.staging is None:
        parser.error("--skip-export needs --staging (the folder holding a previous export)")
    # Absolute paths: "--out ." and "data/" name a real folder whose siblings (.tmp, .old) can be made.
    args.out = args.out.resolve()
    if args.staging is not None:
        args.staging = args.staging.resolve()
    return args


def _check(args: argparse.Namespace) -> None:
    spa2 = args.installer / "spa2"
    if not (spa2 / "niscat.mdb").is_file() or not (spa2 / "series").is_dir():
        msg = f"{args.installer} is not a NISCAT installer (spa2/niscat.mdb and spa2/series/ expected)"
        raise BuildError(msg)
    if args.out.exists() and not args.force:
        msg = f"{args.out} already exists: add --force to replace it"
        raise BuildError(msg)
    if args.staging is not None and args.staging.is_relative_to(args.out):
        msg = f"the staging folder {args.staging} is inside {args.out}, which the build replaces"
        raise BuildError(msg)
    if not args.skip_export and (sys.platform != "win32" or not POWERSHELL_32.is_file()):
        msg = (
            "exporting the NISCAT tables needs Windows and its 32-bit PowerShell; on another system, "
            "export on Windows first, then run with --staging <folder> --skip-export"
        )
        raise BuildError(msg)


def _export(installer: Path, staging: Path) -> None:
    print(f"Exporting the NISCAT tables to {staging}...")
    command = [str(POWERSHELL_32), "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", str(EXPORT_SCRIPT)]
    result = subprocess.run(  # noqa: S603 - fixed program and script, paths given by the user
        [*command, "-SrcDir", str(installer), "-OutDir", str(staging)], check=False
    )
    if result.returncode != 0:
        msg = (
            "the export of the NISCAT tables failed: is the Microsoft Jet 4.0 driver installed,"
            " and are PowerShell scripts allowed (a group policy can forbid them)?"
        )
        raise BuildError(msg)


def _build(args: argparse.Namespace, staging: Path, work: Path) -> None:
    if not args.skip_export:
        _export(args.installer, staging)
    print("Building data.db...")
    build_database(staging, args.installer, work / "data.db")
    print("Converting the drawings...")
    drawings, documents = convert_drawings(args.installer, work, args.jobs)
    manifest = write_manifest(work, args.edition)
    with contextlib.closing(sqlite3.connect(work / "data.db")) as db:
        parts, vins = (db.execute(f"SELECT COUNT(*) FROM {t}").fetchone()[0] for t in ("part", "vin"))  # noqa: S608
    print(
        f"{parts} part lines, {vins} VINs, {drawings} drawings, {documents} documents ({manifest['version']})"
    )


def _replace(work: Path, out: Path) -> None:
    # Keep the previous folder until the new one is in place, so a failed rename loses nothing.
    old = out.with_name(out.name + ".old")
    if out.exists():
        shutil.rmtree(old, ignore_errors=True)
        try:
            out.rename(old)
        except PermissionError as error:
            msg = (
                f"{out} is in use: close any program using it (for example the Open Niscat server)"
                " and run again"
            )
            raise BuildError(msg) from error
    try:
        work.rename(out)
    except OSError as error:
        if old.exists():
            try:
                old.rename(out)
            except OSError:
                msg = f"could not put the new data in place; the previous data is in {old}"
                raise BuildError(msg) from error
        raise
    shutil.rmtree(old, ignore_errors=True)


def main(argv: list[str] | None = None) -> int:
    """Run the build; return the exit code (0 success, 1 failure, 2 usage error)."""
    args = _parse(argv)
    start = time.monotonic()
    work = args.out.with_name(args.out.name + ".tmp")
    temporary = None
    try:
        _check(args)
        shutil.rmtree(work, ignore_errors=True)  # a previous interrupted build
        work.mkdir(parents=True)
        if args.staging is None:
            temporary = tempfile.TemporaryDirectory(prefix="niscat-staging-", ignore_cleanup_errors=True)
            staging = Path(temporary.name)
        else:
            staging = args.staging
        _build(args, staging, work)
        _replace(work, args.out)
    except Exception as error:  # noqa: BLE001 - the command reports any failure as one message
        print(f"error: {error}", file=sys.stderr)
        return 1
    finally:
        shutil.rmtree(work, ignore_errors=True)
        if temporary is not None:
            temporary.cleanup()
    print(f"Data folder ready: {args.out} ({time.monotonic() - start:.0f} s)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
