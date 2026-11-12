import argparse
import datetime
import pathlib
import zipfile

import docx
from docx.oxml.ns import qn
import openpyxl
import pptx


def normalize_package(path: pathlib.Path) -> None:
    with zipfile.ZipFile(path) as package:
        entries = {name: package.read(name) for name in package.namelist()}
    with zipfile.ZipFile(path, "w", zipfile.ZIP_DEFLATED) as package:
        for name, data in sorted(entries.items()):
            entry = zipfile.ZipInfo(name, (2026, 1, 1, 0, 0, 0))
            entry.compress_type = zipfile.ZIP_DEFLATED
            package.writestr(entry, data)


def generate(directory: pathlib.Path) -> None:
    directory.mkdir(parents=True, exist_ok=True)
    fixed_time = datetime.datetime(2026, 1, 1, tzinfo=datetime.timezone.utc)
    document = docx.Document()
    document.core_properties.author = "RightClick"
    document.core_properties.created = fixed_time
    document.core_properties.modified = fixed_time
    for setting in document.settings.element.iter(qn("w:compatSetting")):
        if setting.get(qn("w:name")) == "compatibilityMode":
            setting.set(qn("w:val"), "15")
    document.add_paragraph()
    document.save(directory / "blank.docx")
    workbook = openpyxl.Workbook()
    workbook.properties.creator = "RightClick"
    workbook.properties.created = fixed_time
    workbook.properties.modified = fixed_time
    workbook.save(directory / "blank.xlsx")
    presentation = pptx.Presentation()
    presentation.core_properties.author = "RightClick"
    presentation.core_properties.created = fixed_time
    presentation.core_properties.modified = fixed_time
    presentation.slides.add_slide(presentation.slide_layouts[6])
    presentation.save(directory / "blank.pptx")
    (directory / "blank.rtf").write_bytes(b"{\\rtf1\\ansi\\deff0{\\fonttbl{\\f0 Helvetica;}}\\f0\\fs24\\par}\n")
    for extension in ("docx", "xlsx", "pptx"):
        normalize_package(directory / f"blank.{extension}")


def verify(directory: pathlib.Path) -> None:
    for extension, load in (("docx", docx.Document), ("xlsx", openpyxl.load_workbook), ("pptx", pptx.Presentation)):
        sources = sorted(directory.glob(f"*.{extension}"))
        for source in sources:
            output = directory / f"{source.stem}-roundtrip.{extension}"
            document = load(source)
            document.save(output)
            load(output)
            with zipfile.ZipFile(output) as package:
                if package.testzip() is not None:
                    raise ValueError(f"corrupt document package: {output}")
            output.unlink()
            print(f"validated {source.name}: load, save, reload")


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--verify", type=pathlib.Path)
    options = parser.parse_args()
    if options.verify is not None:
        verify(options.verify)
    else:
        directory = pathlib.Path(__file__).resolve().parents[1] / "internal" / "catalog" / "documents"
        generate(directory)
        verify(directory)


if __name__ == "__main__":
    main()
