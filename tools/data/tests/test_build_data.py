import json
import subprocess
import sys
from pathlib import Path

import pytest

import build_data


def run(installer: Path, staging: Path, out: Path, *extra: str) -> int:
    args = ["--installer", str(installer), "--staging", str(staging), "--skip-export", "--out", str(out)]
    return build_data.main([*args, "--jobs", "1", *extra])


def test_builds_a_complete_data_folder(staging: Path, installer: Path, tmp_path: Path) -> None:
    out = tmp_path / "data"
    assert run(installer, staging, out) == 0
    assert (out / "data.db").is_file()
    assert json.loads((out / "manifest.json").read_text(encoding="utf-8"))["version"] == "2015.01-2"
    assert (out / "cinfo" / "AA" / "g0101.pdf").is_file()
    assert not (tmp_path / "data.tmp").exists()


def test_keeps_an_existing_folder_without_force(
    staging: Path, installer: Path, tmp_path: Path, capsys: pytest.CaptureFixture[str]
) -> None:
    out = tmp_path / "data"
    out.mkdir()
    (out / "keep.txt").write_text("previous", encoding="utf-8")
    assert run(installer, staging, out) == 1
    assert "--force" in capsys.readouterr().err
    assert (out / "keep.txt").read_text(encoding="utf-8") == "previous"
    assert run(installer, staging, out, "--force") == 0
    assert not (out / "keep.txt").exists()
    assert not (tmp_path / "data.old").exists()


def test_a_failure_leaves_the_previous_folder_untouched(
    staging: Path, installer: Path, tmp_path: Path, monkeypatch: pytest.MonkeyPatch
) -> None:
    out = tmp_path / "data"
    assert run(installer, staging, out) == 0
    before = (out / "manifest.json").read_text(encoding="utf-8")

    def broken(*_: object, **__: object) -> tuple[int, int]:
        msg = "corrupt drawing"
        raise OSError(msg)

    monkeypatch.setattr(build_data, "convert_drawings", broken)
    assert run(installer, staging, out, "--force") == 1
    assert (out / "manifest.json").read_text(encoding="utf-8") == before
    assert not (tmp_path / "data.tmp").exists()


def test_rejects_a_folder_that_is_not_an_installer(
    staging: Path, tmp_path: Path, capsys: pytest.CaptureFixture[str]
) -> None:
    assert run(tmp_path, staging, tmp_path / "data") == 1
    assert "not a NISCAT installer" in capsys.readouterr().err


def test_skip_export_needs_a_staging_folder(installer: Path, tmp_path: Path) -> None:
    with pytest.raises(SystemExit) as exit_info:
        build_data.main(["--installer", str(installer), "--skip-export", "--out", str(tmp_path / "data")])
    assert exit_info.value.code == 2


def test_export_requires_windows(
    installer: Path, tmp_path: Path, monkeypatch: pytest.MonkeyPatch, capsys: pytest.CaptureFixture[str]
) -> None:
    monkeypatch.setattr("sys.platform", "linux")
    assert build_data.main(["--installer", str(installer), "--out", str(tmp_path / "data")]) == 1
    assert "--skip-export" in capsys.readouterr().err


TOOL = Path(build_data.__file__)


def test_missing_pillow_is_one_message(tmp_path: Path) -> None:
    code = (
        "import runpy, sys; sys.modules['PIL'] = None; "
        f"sys.argv = ['build_data.py', '--installer', {str(tmp_path)!r}]; "
        f"runpy.run_path({str(TOOL)!r}, run_name='__main__')"
    )
    result = subprocess.run([sys.executable, "-c", code], capture_output=True, text=True, check=False)
    assert result.returncode == 1
    assert "Pillow" in result.stderr
    assert "Traceback" not in result.stderr


def test_a_failed_export_is_reported(
    installer: Path, tmp_path: Path, monkeypatch: pytest.MonkeyPatch, capsys: pytest.CaptureFixture[str]
) -> None:
    monkeypatch.setattr("sys.platform", "win32")
    monkeypatch.setattr(build_data, "POWERSHELL_32", Path(sys.executable))
    monkeypatch.setattr("subprocess.run", lambda *_, **__: subprocess.CompletedProcess([], 1))
    assert build_data.main(["--installer", str(installer), "--out", str(tmp_path / "data")]) == 1
    assert "Jet" in capsys.readouterr().err
    assert not (tmp_path / "data").exists()


def test_a_locked_folder_keeps_the_previous_data(
    staging: Path,
    installer: Path,
    tmp_path: Path,
    monkeypatch: pytest.MonkeyPatch,
    capsys: pytest.CaptureFixture[str],
) -> None:
    out = tmp_path / "data"
    assert run(installer, staging, out) == 0
    rename = Path.rename

    def locked(self: Path, target: Path) -> Path:
        if self == out:
            msg = "in use"
            raise PermissionError(msg)
        return rename(self, target)

    monkeypatch.setattr(Path, "rename", locked)
    assert run(installer, staging, out, "--force") == 1
    assert "close" in capsys.readouterr().err
    assert (out / "manifest.json").is_file()
    assert not (tmp_path / "data.tmp").exists()


def test_staging_inside_out_is_refused(
    staging: Path, installer: Path, tmp_path: Path, capsys: pytest.CaptureFixture[str]
) -> None:
    out = tmp_path
    assert run(installer, staging, out, "--force") == 1
    assert "staging" in capsys.readouterr().err
    assert staging.is_dir()


@pytest.mark.parametrize("jobs", ["0", "-2", "x"])
def test_jobs_must_be_positive(installer: Path, tmp_path: Path, jobs: str) -> None:
    with pytest.raises(SystemExit) as exit_info:
        build_data.main(["--installer", str(installer), "--out", str(tmp_path / "d"), "--jobs", jobs])
    assert exit_info.value.code == 2
