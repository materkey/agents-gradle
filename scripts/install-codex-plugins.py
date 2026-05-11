#!/usr/bin/env python3
"""Install agents-gradle plugins into Codex without plugin-management."""

from __future__ import annotations

import argparse
import json
import os
import re
import shutil
import sys
import tempfile
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Optional


MARKETPLACE_MANIFESTS = (
    Path(".agents/plugins/marketplace.json"),
    Path(".claude-plugin/marketplace.json"),
)
PLUGIN_MANIFEST = Path(".codex-plugin/plugin.json")
SAFE_SEGMENT = re.compile(r"^[A-Za-z0-9._+-]+$")


class InstallError(Exception):
    pass


def fail(message: str) -> None:
    raise InstallError(message)


def load_json_object(path: Path) -> dict[str, Any]:
    try:
        with path.open() as handle:
            payload = json.load(handle)
    except FileNotFoundError:
        fail(f"Missing JSON file: {path}")
    except json.JSONDecodeError as error:
        fail(f"Invalid JSON in {path}: {error}")

    if not isinstance(payload, dict):
        fail(f"{path} must contain a JSON object.")
    return payload


def find_marketplace_manifest(root: Path) -> Path:
    for relative in MARKETPLACE_MANIFESTS:
        candidate = root / relative
        if candidate.exists():
            return candidate
    fail(
        "No marketplace manifest found under "
        f"{root}. Expected one of: "
        + ", ".join(str(path) for path in MARKETPLACE_MANIFESTS)
    )


def load_marketplace(root: Path) -> tuple[Path, dict[str, Any], str]:
    root = root.expanduser().resolve()
    manifest_path = find_marketplace_manifest(root)
    payload = load_json_object(manifest_path)
    plugins = payload.get("plugins")
    if not isinstance(plugins, list):
        fail(f"{manifest_path} field 'plugins' must be an array.")

    name = payload.get("name")
    if isinstance(name, str) and name.strip():
        marketplace = name.strip()
    else:
        marketplace = root.name
    validate_segment("marketplace name", marketplace)
    return root, payload, marketplace


def plugin_entries(payload: dict[str, Any]) -> list[dict[str, Any]]:
    entries = payload.get("plugins")
    if not isinstance(entries, list):
        fail("Marketplace field 'plugins' must be an array.")
    return [entry for entry in entries if isinstance(entry, dict)]


def find_plugin_entry(payload: dict[str, Any], plugin_name: str) -> dict[str, Any]:
    for entry in plugin_entries(payload):
        if entry.get("name") == plugin_name:
            return entry
    available = ", ".join(
        sorted(str(entry.get("name")) for entry in plugin_entries(payload) if entry.get("name"))
    )
    fail(f"Plugin '{plugin_name}' not found in marketplace. Available: {available or '<none>'}")


def expand_plugin_dependencies(
    marketplace_payload: dict[str, Any],
    selected_plugins: list[str],
    marketplace: str,
) -> list[str]:
    expanded: list[str] = []
    visiting: set[str] = set()
    visited: set[str] = set()

    def visit(plugin_name: str, chain: list[str]) -> None:
        validate_segment("plugin name", plugin_name)
        entry = find_plugin_entry(marketplace_payload, plugin_name)

        if plugin_name in visited:
            return
        if plugin_name in visiting:
            fail("Plugin dependency cycle: " + " -> ".join(chain + [plugin_name]))

        visiting.add(plugin_name)
        dependencies = entry.get("dependencies", [])
        if not isinstance(dependencies, list):
            fail(f"Plugin '{plugin_name}' dependencies must be an array.")
        for dependency in dependencies:
            dependency = normalize_dependency_ref(dependency, marketplace, plugin_name)
            visit(dependency, chain + [plugin_name])
        visiting.remove(plugin_name)
        visited.add(plugin_name)
        expanded.append(plugin_name)

    for plugin_name in selected_plugins:
        visit(plugin_name, [])

    return expanded


def normalize_dependency_ref(value: Any, marketplace: str, plugin_name: str) -> str:
    if isinstance(value, str):
        raw = value
    elif isinstance(value, dict) and isinstance(value.get("name"), str):
        raw = value["name"]
        dep_marketplace = value.get("marketplace")
        if dep_marketplace is not None:
            if not isinstance(dep_marketplace, str):
                fail(f"Plugin '{plugin_name}' dependency marketplace must be a string.")
            raw = f"{raw}@{dep_marketplace}"
    else:
        fail(f"Plugin '{plugin_name}' dependency must be a string or object with a name.")

    dep_name, _, dep_marketplace = raw.partition("@")
    validate_segment("plugin dependency", dep_name)
    if dep_marketplace:
        validate_segment("plugin dependency marketplace", dep_marketplace)
        if dep_marketplace != marketplace:
            fail(
                f"Plugin '{plugin_name}' depends on cross-marketplace plugin '{raw}', "
                "which this local installer cannot install."
            )
    return dep_name


def validate_segment(label: str, value: str) -> None:
    if not value or not SAFE_SEGMENT.match(value):
        fail(f"Unsafe {label}: {value!r}")


def resolve_local_source(marketplace_root: Path, entry: dict[str, Any]) -> tuple[Path, dict[str, Any]]:
    source = entry.get("source")
    if isinstance(source, str):
        raw_path = source
    elif isinstance(source, dict) and (source.get("source") == "local" or source.get("type") == "local"):
        raw_path = source.get("path")
        if not isinstance(raw_path, str):
            fail(f"Plugin entry '{entry.get('name')}' has no local source path.")
    else:
        fail(
            f"Plugin entry '{entry.get('name')}' must use a local source. "
            "This installer is intentionally self-contained for the local agents-gradle marketplace."
        )

    if not raw_path.startswith("./"):
        fail(f"Local plugin source path must start with './': {raw_path}")
    relative = Path(raw_path[2:])
    if not relative.parts or any(part in ("", ".", "..") for part in relative.parts):
        fail(f"Local plugin source path must stay within the marketplace root: {raw_path}")

    source_path = (marketplace_root / relative).resolve()
    try:
        source_path.relative_to(marketplace_root)
    except ValueError:
        fail(f"Plugin source escapes marketplace root: {raw_path}")
    if not source_path.is_dir():
        fail(f"Plugin source directory does not exist: {source_path}")

    return source_path, {"source": "local", "path": raw_path}


def load_plugin_manifest(plugin_root: Path) -> tuple[Path, dict[str, Any]]:
    manifest_path = plugin_root / PLUGIN_MANIFEST
    return manifest_path, load_json_object(manifest_path)


def plugin_version(manifest_path: Path, payload: dict[str, Any]) -> str:
    version = payload.get("version")
    if version is None:
        return "local"
    is_placeholder = isinstance(version, str) and version.startswith("[") and "TODO" in version
    if not isinstance(version, str) or not version.strip() or is_placeholder:
        fail(f"{manifest_path} must define a concrete string version when version is present.")
    validate_segment("plugin version", version)
    return version


def codex_home(value: Optional[str]) -> Path:
    if value:
        return Path(value).expanduser().resolve()
    return Path(os.environ.get("CODEX_HOME", "~/.codex")).expanduser().resolve()


def config_path(home: Path) -> Path:
    return home / "config.toml"


def toml_quote(value: str) -> str:
    return json.dumps(value)


def update_section(config_file: Path, header: str, entries: dict[str, str]) -> bool:
    config_file.parent.mkdir(parents=True, exist_ok=True)
    original = config_file.read_text() if config_file.exists() else ""
    lines = original.splitlines()

    index = None
    for line_index, line in enumerate(lines):
        if line.strip() == header:
            index = line_index
            break

    desired_lines = [f"{key} = {value}" for key, value in entries.items()]
    changed = False

    if index is None:
        if lines and lines[-1].strip():
            lines.append("")
        lines.append(header)
        lines.extend(desired_lines)
        changed = True
    else:
        next_section = len(lines)
        for scan in range(index + 1, len(lines)):
            stripped = lines[scan].strip()
            if stripped.startswith("[") and stripped.endswith("]"):
                next_section = scan
                break

        for key, value in entries.items():
            prefix = f"{key} ="
            desired = f"{key} = {value}"
            existing_index = None
            for scan in range(index + 1, next_section):
                if lines[scan].strip().startswith(prefix):
                    existing_index = scan
                    break
            if existing_index is None:
                lines.insert(index + 1, desired)
                next_section += 1
                changed = True
            elif lines[existing_index] != desired:
                lines[existing_index] = desired
                changed = True

    new_text = "\n".join(lines).rstrip() + "\n"
    if new_text != original:
        config_file.write_text(new_text)
        changed = True
    return changed


def update_marketplace_config(config_file: Path, marketplace: str, root: Path) -> bool:
    now = datetime.now(timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z")
    return update_section(
        config_file,
        f"[marketplaces.{marketplace}]",
        {
            "last_updated": toml_quote(now),
            "source_type": toml_quote("local"),
            "source": toml_quote(str(root)),
        },
    )


def enable_plugin(config_file: Path, plugin_id: str) -> bool:
    return update_section(config_file, f'[plugins."{plugin_id}"]', {"enabled": "true"})


def copy_plugin_source(source: Path, plugin_base: Path, version: str) -> Path:
    plugin_base.parent.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix="plugin-install-", dir=plugin_base.parent) as staging:
        staged_base = Path(staging) / plugin_base.name
        staged_version = staged_base / version
        shutil.copytree(
            source,
            staged_version,
            ignore=shutil.ignore_patterns(".git", "__pycache__", "*.pyc", ".DS_Store"),
        )
        if plugin_base.exists():
            shutil.rmtree(plugin_base)
        shutil.move(str(staged_base), str(plugin_base))
    return plugin_base / version


def install_plugin(
    home: Path,
    marketplace_root: Path,
    marketplace_payload: dict[str, Any],
    marketplace: str,
    plugin_name: str,
) -> dict[str, Any]:
    validate_segment("plugin name", plugin_name)
    entry = find_plugin_entry(marketplace_payload, plugin_name)
    plugin_source, source_info = resolve_local_source(marketplace_root, entry)
    manifest_path, manifest = load_plugin_manifest(plugin_source)

    manifest_name = manifest.get("name")
    if manifest_name != plugin_name:
        fail(f"{manifest_path} name is {manifest_name!r}, expected {plugin_name!r}.")

    version = plugin_version(manifest_path, manifest)
    install_base = home / "plugins" / "cache" / marketplace / plugin_name
    install_path = copy_plugin_source(plugin_source, install_base, version)

    plugin_id = f"{plugin_name}@{marketplace}"
    config_updated = enable_plugin(config_path(home), plugin_id)
    data_dir = home / "plugins" / "data" / f"{plugin_name}-{marketplace}"
    data_dir.mkdir(parents=True, exist_ok=True)

    return {
        "plugin_id": plugin_id,
        "marketplace": marketplace,
        "plugin": plugin_name,
        "version": version,
        "source_path": str(plugin_source),
        "source": source_info,
        "installed_path": str(install_path),
        "config_updated": config_updated,
        "enabled": True,
        "data_dir": str(data_dir),
    }


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Install local agents-gradle plugins into the Codex plugin cache."
    )
    parser.add_argument(
        "--root",
        default=str(Path(__file__).resolve().parents[1]),
        help="agents-gradle marketplace root",
    )
    parser.add_argument(
        "--codex-home",
        default=None,
        help="Codex home directory; defaults to CODEX_HOME or ~/.codex",
    )
    parser.add_argument(
        "plugins",
        nargs="*",
        help="Plugin names to install. Defaults to every plugin in the marketplace.",
    )
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    home = codex_home(args.codex_home)
    marketplace_root, marketplace_payload, marketplace = load_marketplace(Path(args.root))
    selected_plugins = args.plugins or [
        entry["name"] for entry in plugin_entries(marketplace_payload) if isinstance(entry.get("name"), str)
    ]
    if not selected_plugins:
        fail("No plugins selected.")
    expanded_plugins = expand_plugin_dependencies(marketplace_payload, selected_plugins, marketplace)

    config_file = config_path(home)
    marketplace_config_updated = update_marketplace_config(config_file, marketplace, marketplace_root)
    installed = [
        install_plugin(home, marketplace_root, marketplace_payload, marketplace, plugin_name)
        for plugin_name in expanded_plugins
    ]

    json.dump(
        {
            "config_path": str(config_file),
            "marketplace": marketplace,
            "marketplace_root": str(marketplace_root),
            "marketplace_config_updated": marketplace_config_updated,
            "requested_plugins": selected_plugins,
            "expanded_plugins": expanded_plugins,
            "installed": installed,
        },
        sys.stdout,
        indent=2,
        sort_keys=True,
    )
    sys.stdout.write("\n")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except InstallError as error:
        print(f"install-codex-plugins.py: {error}", file=sys.stderr)
        raise SystemExit(1)
