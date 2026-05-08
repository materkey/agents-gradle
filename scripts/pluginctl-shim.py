#!/usr/bin/env python3
"""Compatibility shim for upstream ``pluginctl.py`` on Python 3.9 / 3.10.

``pluginctl.py`` (agent-thingz/plugin-management) imports ``tomllib`` from
stdlib, which exists only on Python 3.11+. On older interpreters this shim
auto-installs the ``tomli`` backport into the user site (with a PEP 668
``--break-system-packages`` retry for externally-managed Homebrew/Debian
Pythons), aliases it as ``tomllib``, and then hands control to the real
script via :mod:`runpy`. On Python 3.11+ the shim returns immediately.

Override the upstream script path with ``AGENTS_GRADLE_PLUGINCTL`` for tests.

Usage::

    python3 scripts/pluginctl-shim.py install agents-gradle gradle-grill --force
"""

from __future__ import annotations

import os
import runpy
import subprocess
import sys

DEFAULT_PLUGINCTL = os.path.expanduser(
    "~/.codex/plugins/cache/agent-thingz/plugin-management/0.1.0/scripts/pluginctl.py"
)


def _pip_install_tomli() -> bool:
    """Try ``pip install --user tomli``, with a PEP 668 fallback.

    Returns True on success, False otherwise. Stays quiet on success and
    surfaces the original error only if both attempts fail.
    """
    base = [sys.executable, "-m", "pip", "install", "--user", "--quiet", "tomli"]
    first = subprocess.run(base, capture_output=True, text=True)
    if first.returncode == 0:
        return True
    if "externally-managed-environment" in (first.stderr or ""):
        retry = subprocess.run(base + ["--break-system-packages"], capture_output=True, text=True)
        if retry.returncode == 0:
            return True
        sys.stderr.write(retry.stderr or first.stderr)
        return False
    sys.stderr.write(first.stderr or "")
    return False


def _ensure_tomllib() -> None:
    """Alias the ``tomli`` backport as ``tomllib`` on Python 3.9 / 3.10."""
    if sys.version_info >= (3, 11):
        return
    try:
        import tomli as _tomli  # type: ignore[import-not-found]
    except ImportError:
        if not _pip_install_tomli():
            sys.stderr.write(
                "pluginctl-shim: Python {0}.{1} detected and 'tomli' could not be "
                "installed. Install it manually, e.g.:\n"
                "  python3 -m pip install --user tomli\n"
                "  python3 -m pip install --user --break-system-packages tomli\n"
                "  python3 -m venv ~/.venv-tomli && ~/.venv-tomli/bin/pip install tomli\n"
                "Or invoke this script with a Python 3.11+ interpreter directly.\n".format(
                    sys.version_info[0], sys.version_info[1]
                )
            )
            sys.exit(2)
        import tomli as _tomli  # type: ignore[import-not-found]
    sys.modules["tomllib"] = _tomli


def main() -> None:
    _ensure_tomllib()
    pluginctl = os.environ.get("AGENTS_GRADLE_PLUGINCTL", DEFAULT_PLUGINCTL)
    if not os.path.isfile(pluginctl):
        sys.stderr.write(
            "pluginctl-shim: pluginctl.py not found at {0}\n"
            "Set AGENTS_GRADLE_PLUGINCTL to override.\n".format(pluginctl)
        )
        sys.exit(2)
    sys.argv[0] = pluginctl
    runpy.run_path(pluginctl, run_name="__main__")


if __name__ == "__main__":
    main()
