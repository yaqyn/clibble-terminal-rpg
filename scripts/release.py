#!/usr/bin/env python3
"""Build portable release archives and SHA-256 checksums without dependencies."""
import hashlib
import os
from pathlib import Path
import re
import subprocess
import sys
import tarfile
import tempfile
import zipfile


def main():
    if len(sys.argv) != 2 or not re.fullmatch(r"v\d+\.\d+\.\d+(?:-[a-zA-Z0-9.-]+)?", sys.argv[1]):
        raise SystemExit("Usage: python3 scripts/release.py vMAJOR.MINOR.PATCH")
    version = sys.argv[1]
    root = Path(__file__).resolve().parent.parent
    status = subprocess.check_output(["git", "status", "--porcelain"], cwd=root, text=True)
    if status.strip():
        raise SystemExit("Commit source changes before building release archives.")
    output = root / "dist" / version
    output.mkdir(parents=True, exist_ok=True)
    checksums = []
    for target_os in ("linux", "darwin", "windows"):
        for arch in ("amd64", "arm64"):
            with tempfile.TemporaryDirectory(prefix="pokedex-release-") as temporary:
                binary = Path(temporary) / ("pokedex.exe" if target_os == "windows" else "pokedex")
                env = dict(os.environ, GOOS=target_os, GOARCH=arch, CGO_ENABLED="0")
                subprocess.run(
                    ["go", "build", "-trimpath", "-ldflags", f"-s -w -X main.version={version}", "-o", str(binary), "."],
                    cwd=root, env=env, check=True,
                )
                suffix = ".zip" if target_os == "windows" else ".tar.gz"
                artifact = output / f"pokedex_{version}_{target_os}_{arch}{suffix}"
                files = [(binary, binary.name), (root / "README.md", "README.md"), (root / "LICENSE", "LICENSE"), (root / "THIRD_PARTY_NOTICES.md", "THIRD_PARTY_NOTICES.md")]
                for folder in ("readme-assets", "licenses"):
                    files.extend((source, str(source.relative_to(root))) for source in sorted((root / folder).rglob("*")) if source.is_file())
                if target_os == "windows":
                    with zipfile.ZipFile(artifact, "w", zipfile.ZIP_DEFLATED) as archive:
                        for source, name in files:
                            archive.write(source, name)
                else:
                    with tarfile.open(artifact, "w:gz") as archive:
                        for source, name in files:
                            archive.add(source, arcname=name)
                checksums.append(f"{hashlib.sha256(artifact.read_bytes()).hexdigest()}  {artifact.name}")
                print(artifact.relative_to(root), flush=True)
    (output / "SHA256SUMS").write_text("\n".join(checksums) + "\n")


if __name__ == "__main__":
    main()
