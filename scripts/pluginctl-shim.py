#!/usr/bin/env python3
"""Compatibility shim for ``pluginctl.py`` on Python 3.9–3.10.

The upstream ``pluginctl.py`` (agent-thingz/plugin-management) imports
``tomllib`` from the standard library, which only exists on Python 3.11+.
On older interpreters this shim transparently substitutes the ``tomli``
backport (PyPI package ``tomli``) under the name ``tomllib`` and then
hands control to the real script via :mod:`runpy`.

On Python 3.11+ the shim is a no-op pass-through and adds no measurable
overhead.

Usage::

    python3 scripts/pluginctl-shim.py install agents-gradle gradle-grill --force

The path to ``pluginctl.py`` can be overridden with the
``AGENTS_GRADLE_PLUGINCTL`` environment variable for tests.
"""

from __future__ import annotations

import os
import runpy
import sys

DEFAULT_PLUGINCTL = os.path.expanduser(
    "~/.codex/plugins/cache/agent-thingz/plugin-management/0.1.0/scripts/pluginctl.py"
)


def _ensure_tomllib() -> None:
    """Make ``import tomllib`` work on Python 3.9 and 3.10.

    On 3.11+ ``tomllib`` is in stdlib and we leave it alone.
    On 3.9/3.10 we try ``import tomli`` and alias it as ``tomllib``;
    if ``tomli`` is missing we abort with an actionable message.
    """
    if sys.version_info >= (3, 11):
        return
    try:
        import tomli as _tomli  # type: ignore[import-not-found]
    except ImportError:
        sys.stderr.write(
            "pluginctl-shim: Python {0}.{1} detected; the 'tomli' backport "
            "is required.\n"
            "Install it with one of:\n"
            "  python3 -m pip install --user tomli\n"
            "  pipx install tomli\n"
            "Or invoke the installer with a Python 3.11+ interpreter, e.g.:\n"
            "  /opt/homebrew/bin/python3.13 scripts/pluginctl-shim.py ...\n".format(
                sys.version_info[0], sys.version_info[1]
            )
        )
        sys.exit(2)
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
