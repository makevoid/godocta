#!/usr/bin/env python3
"""Build desktop release archives in Docker. Does not commit, tag, or publish."""

from pathlib import Path
import re
import shutil
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parent.parent
IMAGE = "octa-release:godot-4.7.2"


def main():
    version = re.search(r'^config/version="([0-9]+\.[0-9]+\.[0-9]+)"$',
                        (ROOT / "project.godot").read_text(), re.MULTILINE).group(1)
    subprocess.run(["docker", "build", "--tag", IMAGE, str(ROOT / "tools/release")],
                   check=True)
    (ROOT / "build").mkdir(exist_ok=True)
    output = ROOT / "dist" / version
    output.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix="release-source-", dir=ROOT / "build") as temp:
        source = Path(temp)
        # An explicit source allowlist keeps experiments, credentials and local caches out.
        for directory in ("assets", "scripts", "shaders", "scenes", "structures", "tests"):
            shutil.copytree(ROOT / directory, source / directory,
                            ignore=shutil.ignore_patterns(".DS_Store"))
        for filename in ("project.godot", "export_presets.cfg", "LICENSE",
                         "docs/credits.md", "docs/controls.md", "tools/check.py",
                         "tools/release/smoke.gd"):
            destination = source / filename
            destination.parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(ROOT / filename, destination)
        subprocess.run(["docker", "run", "--rm",
                        "--mount", f"type=bind,source={source},target=/source,readonly",
                        "--mount", f"type=bind,source={output},target=/out",
                        IMAGE], check=True)
    print(f"Release {version} ready in {output}")


if __name__ == "__main__":
    main()
