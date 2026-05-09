#!/usr/bin/env python3
"""Download Android Gradle Plugin source jars from Google Maven."""

from __future__ import annotations

import argparse
import os
import sys
import urllib.error
import urllib.request
import xml.etree.ElementTree as ET
import zipfile
from pathlib import Path
from tempfile import NamedTemporaryFile


GOOGLE_MAVEN = "https://dl.google.com/dl/android/maven2"
DEFAULT_DEST = Path(os.environ.get("AGP_SOURCES_DIR", Path.home() / ".agp-sources"))

DEFAULT_MODULES = [
    "com.android.tools.build:gradle",
    "com.android.tools.build:gradle-api",
    "com.android.tools.build:builder",
    "com.android.tools.build:builder-model",
    "com.android.tools.build:manifest-merger",
    "com.android.tools.build:apksig",
    "com.android.tools.build:apkzlib",
    "com.android.tools:sdk-common",
    "com.android.tools:common",
    "com.android.tools.lint:lint",
    "com.android.tools.lint:lint-api",
]


def maven_path(group: str, artifact: str, version: str, file_name: str) -> str:
    return "/".join([group.replace(".", "/"), artifact, version, file_name])


def read_url(url: str) -> bytes:
    with urllib.request.urlopen(url, timeout=60) as response:
        return response.read()


def latest_agp_version() -> str:
    metadata_url = f"{GOOGLE_MAVEN}/com/android/tools/build/gradle/maven-metadata.xml"
    root = ET.fromstring(read_url(metadata_url))
    versioning = root.find("versioning")
    if versioning is None:
        raise RuntimeError("maven-metadata.xml has no <versioning>")

    for tag in ("release", "latest"):
        value = versioning.findtext(tag)
        if value:
            return value

    versions = [node.text for node in versioning.findall("versions/version") if node.text]
    if not versions:
        raise RuntimeError("maven-metadata.xml has no versions")
    return versions[-1]


def parse_module(value: str) -> tuple[str, str]:
    parts = value.split(":")
    if len(parts) != 2 or not all(parts):
        raise argparse.ArgumentTypeError(f"expected group:artifact, got {value!r}")
    return parts[0], parts[1]


def extract_sources(
    group: str,
    artifact: str,
    version: str,
    dest: Path,
    force: bool,
) -> bool:
    target = dest / version / group / artifact
    marker = target / ".agp-sources-version"
    if target.exists() and marker.exists() and not force:
        print(f"exists: {target}")
        return False

    jar_name = f"{artifact}-{version}-sources.jar"
    url = f"{GOOGLE_MAVEN}/{maven_path(group, artifact, version, jar_name)}"
    target.mkdir(parents=True, exist_ok=True)

    print(f"download: {group}:{artifact}:{version}")
    try:
        data = read_url(url)
    except urllib.error.HTTPError as exc:
        if exc.code == 404:
            print(f"skip: no sources jar for {group}:{artifact}:{version}", file=sys.stderr)
            return False
        raise

    with NamedTemporaryFile(suffix=".jar", delete=False) as tmp:
        tmp.write(data)
        tmp_path = Path(tmp.name)

    try:
        if force:
            for child in target.iterdir():
                if child.is_dir():
                    import shutil

                    shutil.rmtree(child)
                else:
                    child.unlink()

        with zipfile.ZipFile(tmp_path) as jar:
            jar.extractall(target)
        marker.write_text(version + "\n", encoding="utf-8")
    finally:
        tmp_path.unlink(missing_ok=True)

    print(f"ready: {target}")
    return True


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--version",
        default=None,
        help="AGP version to download. Defaults to the latest version from Google Maven.",
    )
    parser.add_argument(
        "--dest",
        type=Path,
        default=DEFAULT_DEST,
        help=f"Destination root. Default: {DEFAULT_DEST}",
    )
    parser.add_argument(
        "--module",
        action="append",
        type=parse_module,
        help="Module as group:artifact. Can be passed multiple times.",
    )
    parser.add_argument(
        "--force",
        action="store_true",
        help="Re-download and replace existing extracted module sources.",
    )
    args = parser.parse_args()

    version = args.version or latest_agp_version()
    modules = args.module or [parse_module(module) for module in DEFAULT_MODULES]

    args.dest.mkdir(parents=True, exist_ok=True)
    downloaded = 0
    for group, artifact in modules:
        if extract_sources(group, artifact, version, args.dest, args.force):
            downloaded += 1

    print(f"AGP {version} sources root: {args.dest / version}")
    print(f"modules downloaded: {downloaded}; modules available: {len(modules)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
