"""Compare two Open Niscat data folders (development tool).

Usage: python tools/data/compare_data.py <data A> <data B>

Run it after any change to the data tooling: rebuild into another folder and compare with the previous
output. Exit code 0 when equivalent, 1 when they differ.
"""

import argparse
import sys
from pathlib import Path

if sys.version_info < (3, 12):  # noqa: UP036 - give older interpreters a message, not a SyntaxError later
    print("error: Python 3.12 or newer is required", file=sys.stderr)
    sys.exit(1)

try:
    from niscat_data.compare import compare
except ModuleNotFoundError as missing:  # Pillow is the only third-party dependency
    print(f"error: {missing.name} is missing: install Pillow (python -m pip install pillow)", file=sys.stderr)
    sys.exit(1)


def main(argv: list[str] | None = None) -> int:
    """Print the differences between two data folders; return the exit code."""
    parser = argparse.ArgumentParser(description="Compare two Open Niscat data folders.")
    parser.add_argument("a", type=Path, help="reference data folder")
    parser.add_argument("b", type=Path, help="data folder to check")
    args = parser.parse_args(argv)
    differences = compare(args.a, args.b)
    for difference in differences:
        print(difference)
    print("equivalent" if not differences else f"{len(differences)} difference(s)")
    return 1 if differences else 0


if __name__ == "__main__":
    sys.exit(main())
