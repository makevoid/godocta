#!/usr/bin/env python3
"""Container entrypoint: validate an isolated checkout, export, and package it."""

from pathlib import Path
import hashlib
import re
import shutil
import subprocess
import tarfile
import zipfile

WORK = Path("/work")
OUT = Path("/out")
ERROR = re.compile(r"SCRIPT ERROR:|Parse Error:|^ERROR:", re.MULTILINE)


def run(command, log_name, marker=None):
    print("Running: " + " ".join(map(str, command)), flush=True)
    result = subprocess.run(command, cwd=WORK, text=True, stdout=subprocess.PIPE,
                            stderr=subprocess.STDOUT, timeout=600)
    (OUT / log_name).write_text(result.stdout)
    if result.returncode or ERROR.search(result.stdout) or (marker and marker not in result.stdout):
        print(result.stdout, flush=True)
        raise RuntimeError(f"Failed: {log_name}")
    print(f"Passed: {log_name}", flush=True)


def main():
    shutil.copytree("/source", WORK)
    version = re.search(r'^config/version="([0-9]+\.[0-9]+\.[0-9]+)"$',
                        (WORK / "project.godot").read_text(), re.MULTILINE).group(1)
    godot = ["godot", "--headless", "--path", str(WORK)]
    run(["python3", "tools/check.py", "--only", "godot"], "checks.log",
        "All requested checks passed.")
    # Inspect the real export file list and bytes before creating platform packages.
    run(godot + ["--export-pack", "Linux x86_64", "/tmp/contents.zip"], "contents.log")
    with zipfile.ZipFile("/tmp/contents.zip") as pack:
        expected = {str(p.relative_to(WORK)) for p in (WORK / "structures").glob("*.octa")}
        actual = {p for p in pack.namelist() if p.endswith(".octa")}
        if actual != expected:
            raise RuntimeError(f"Exported builds differ: {actual ^ expected}")
        for name in expected:
            if pack.read(name) != (WORK / name).read_bytes():
                raise RuntimeError(f"Export changed {name}")
        excluded = ("experiments/", "captures/", "tests/", "tools/", "docs/", "gen/")
        if any(name.startswith(excluded) for name in pack.namelist()):
            raise RuntimeError("Development or experiment files leaked into the export")
    print(f"Verified {len(expected) - 1} builds plus the neighbors fixture in the export.", flush=True)

    linux = OUT / "Octa-linux-x86_64"
    linux.mkdir(exist_ok=True)
    run(godot + ["--export-release", "Linux x86_64", str(linux / "Octa.x86_64")], "linux-export.log")
    mac = OUT / "Octa-macos.zip"
    run(godot + ["--export-release", "macOS", str(mac)], "macos-export.log")

    notices = {
        "LICENSE.txt": (WORK / "LICENSE").read_bytes(),
        "CREDITS.md": (WORK / "docs/credits.md").read_bytes(),
        "CONTROLS.md": (WORK / "docs/controls.md").read_bytes(),
    }
    for path in Path("/opt/godot-notices").glob("*.txt"):
        notices[path.name] = path.read_bytes()
    common = f"""Octa Structure Studio {version}

100 procedural builds, painting, assembly animation and X-ray inspection.
No Godot installation, Go installation, account or internet connection required.

Source and updates: https://github.com/makevoid/godocta
Controls: CONTROLS.md
Licenses: LICENSE.txt, CREDITS.md, GODOT_LICENSE.txt, GODOT_COPYRIGHT.txt
"""
    linux_readme = common + """
Linux x86-64: extract the entire archive, then run ./Octa.x86_64.
Keep Octa.pck beside the executable. OpenGL 3.3-capable graphics are required.
If a file manager drops executable permissions, run chmod +x Octa.x86_64.
"""
    mac_readme = common + """
macOS (Apple silicon and Intel): unzip, then move the .app to Applications.
This build is ad-hoc signed and is NOT notarized by Apple.
If macOS blocks it, use System Settings > Privacy & Security > Open Anyway
after attempting to open the app. Only do this for a download you trust.
"""
    for name, data in notices.items():
        (linux / name).write_bytes(data)
    (linux / "README.txt").write_text(linux_readme)
    with zipfile.ZipFile(mac, "a", compression=zipfile.ZIP_DEFLATED) as pack:
        for name, data in notices.items():
            pack.writestr(name, data)
        pack.writestr("README.txt", mac_readme)
    with tarfile.open(OUT / "Octa-linux-x86_64.tar.gz", "w:gz") as pack:
        pack.add(linux, arcname=linux.name)
    artifacts = [mac, OUT / "Octa-linux-x86_64.tar.gz"]
    (OUT / "SHA256SUMS.txt").write_text("".join(
        f"{hashlib.sha256(path.read_bytes()).hexdigest()}  {path.name}\n" for path in artifacts))
    shutil.copy2(WORK / "tools/release/smoke.gd", OUT / "smoke.gd")
    print(f"Packaged Octa {version}: " + ", ".join(path.name for path in artifacts), flush=True)


if __name__ == "__main__":
    main()
