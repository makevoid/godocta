#!/usr/bin/env python3
"""Run the same checks locally and in CI, using only Python's standard library."""

import argparse
import os
from pathlib import Path
import re
import shlex
import shutil
import subprocess
import sys
import tempfile

ROOT = Path(__file__).resolve().parent.parent
GODOT_ERROR = re.compile(r"SCRIPT ERROR:|Parse Error:|^ERROR:", re.MULTILINE)
ANSI_ESCAPE = re.compile(r"\x1b\[[0-9;]*m")


def run(command, *, cwd=ROOT, marker=None, godot=False, timeout=600):
    print("\n> " + shlex.join([str(arg) for arg in command]), flush=True)
    result = subprocess.run(
        command, cwd=cwd, stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
        text=True, encoding="utf-8", errors="replace", timeout=timeout,
    )
    output = ANSI_ESCAPE.sub("", result.stdout)
    print(output, end="" if output.endswith("\n") else "\n", flush=True)
    if result.returncode:
        raise RuntimeError(f"Command failed with exit code {result.returncode}")
    if godot and GODOT_ERROR.search(output):
        raise RuntimeError("Godot reported an error even though it exited successfully")
    if marker and not re.search(marker, output, re.MULTILINE):
        raise RuntimeError("Expected test success summary was not printed")
    return output


def check_godot(executable):
    executable = executable or shutil.which("godot")
    if not executable:
        app = Path(os.environ.get("GODOT_APP", "/Applications/Godot.app"))
        binary = app / "Contents/MacOS/Godot"
        if binary.is_file():
            executable = str(binary)
    if not executable:
        raise RuntimeError("Godot not found; put godot on PATH or pass --godot /path/to/Godot")
    run([executable, "--version"])
    with tempfile.TemporaryDirectory(prefix="octa-check-") as log_dir:
        base = [executable, "--headless", "--path", str(ROOT),
                "--log-file", str(Path(log_dir) / "godot.log")]
        run(base + ["--editor", "--import", "--quit"], godot=True)
        run(base + ["--script", "tests/test_core.gd"], godot=True,
            marker=r"^Core checks: [1-9][0-9]* passed, 0 failed\.$")
        run(base + ["--script", "tests/test_scene.gd"], godot=True,
            marker=r"^Scene integration: PASS$")


def check_go():
    module = ROOT / "gen"
    run(["go", "version"], cwd=module)
    unformatted = run(["gofmt", "-l", "."], cwd=module)
    if unformatted.strip():
        raise RuntimeError("Run gofmt on the files listed above")
    run(["go", "vet", "-composites=false", "./..."], cwd=module)
    run(["go", "test", "./..."], cwd=module)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--only", choices=("godot", "go"), help="Run one toolchain's checks")
    parser.add_argument("--godot", default=os.environ.get("GODOT"), help="Godot executable path")
    args = parser.parse_args()
    try:
        if args.only != "go":
            check_godot(args.godot)
        if args.only != "godot":
            check_go()
    except (OSError, RuntimeError, subprocess.TimeoutExpired) as error:
        print(f"Check failed: {error}", file=sys.stderr)
        return 1
    print("\nAll requested checks passed.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
