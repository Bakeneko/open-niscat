from pathlib import Path

from PIL import Image

from niscat_data.drawings import convert_drawings


def g4_tiff(path: Path, size: tuple[int, int]) -> Image.Image:
    """A bilevel CCITT Group 4 TIFF like NISCAT's drawings, with a few black pixels."""
    path.parent.mkdir(parents=True, exist_ok=True)
    image = Image.new("1", size, 1)
    for xy in ((0, 0), (3, 2), (size[0] - 1, size[1] - 1)):
        image.putpixel(xy, 0)
    image.save(path, format="TIFF", compression="group4")
    return image


def test_convert_drawings(installer: Path, tmp_path: Path) -> None:
    series = installer / "spa2" / "series"
    plate = g4_tiff(series / "img" / "aa" / "aa230.etd", (40, 20))  # .etd: a TIFF under another extension
    group = g4_tiff(series / "gindex" / "aa" / "B.TIF", (12, 30))  # some sources are upper-case
    out = tmp_path / "out"

    assert convert_drawings(installer, out, jobs=2) == (2, 1)

    for png, source in ((out / "img" / "AA" / "AA230.png", plate), (out / "gindex" / "AA" / "B.png", group)):
        with Image.open(png) as converted:
            assert converted.format == "PNG"
            assert converted.size == source.size  # hotspot coordinates are in source pixels
            assert converted.mode == "1"  # bilevel kept
            assert converted.tobytes() == source.tobytes()
    assert (out / "cinfo" / "AA" / "g0101.pdf").read_bytes() == b"%PDF-1.4 test"


def test_convert_drawings_without_documents(installer: Path, tmp_path: Path) -> None:
    for pdf in (installer / "spa2" / "series" / "cinfo").rglob("*.pdf"):
        pdf.unlink()
    assert convert_drawings(installer, tmp_path / "out", jobs=1) == (0, 0)
