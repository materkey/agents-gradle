#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

cd "$ROOT"
task build

for path in "${HOME}/.claude/skills/gradle" "${HOME}/.codex/skills/gradle" "${HOME}/.claude/skills/gradle-rag" "${HOME}/.codex/skills/gradle-rag"; do
  if [ -L "$path" ]; then
    rm "$path"
  elif [ -e "$path" ]; then
    echo "Refusing to remove non-symlink legacy skill at $path" >&2
    exit 1
  fi
done

config="${HOME}/.codex/config.toml"
if [ -f "$config" ]; then
  tmp="${config}.tmp"
  awk '
    /^\[plugins\."(gradle-docs|gradle)@agents-gradle"\]/ { skip = 1; next }
    skip && /^\[/ { skip = 0 }
    !skip { print }
  ' "$config" > "$tmp"
  mv "$tmp" "$config"
fi

# pluginctl.py upstream requires `tomllib` (Python 3.11+). To support
# Python 3.9/3.10 we run it through a shim that aliases `tomli` as
# `tomllib`. The shim is a no-op on 3.11+. We pick the newest available
# interpreter automatically; if it is older than 3.11 we make sure the
# `tomli` backport is present.
PYTHON_FOR_PLUGINCTL=""
for cmd in python3.13 python3.12 python3.11 python3; do
  if command -v "$cmd" >/dev/null 2>&1; then
    if "$cmd" -c 'import sys; sys.exit(0 if sys.version_info >= (3, 9) else 1)' 2>/dev/null; then
      PYTHON_FOR_PLUGINCTL="$cmd"
      break
    fi
  fi
done
if [ -z "$PYTHON_FOR_PLUGINCTL" ]; then
  echo "install.sh: need Python 3.9+ on PATH for pluginctl-shim.py" >&2
  exit 1
fi
if ! "$PYTHON_FOR_PLUGINCTL" -c 'import sys; sys.exit(0 if sys.version_info >= (3, 11) else 1)' 2>/dev/null; then
  if ! "$PYTHON_FOR_PLUGINCTL" -c 'import tomli' 2>/dev/null; then
    echo "install.sh: $PYTHON_FOR_PLUGINCTL is older than 3.11, installing 'tomli' backport into user site"
    "$PYTHON_FOR_PLUGINCTL" -m pip install --user --quiet tomli
  fi
fi

codex plugin marketplace add "$ROOT"
"$PYTHON_FOR_PLUGINCTL" "$ROOT/scripts/pluginctl-shim.py" install agents-gradle gradle-rag --force
"$PYTHON_FOR_PLUGINCTL" "$ROOT/scripts/pluginctl-shim.py" install agents-gradle gradle-grill --force

claude plugin marketplace add "$ROOT"
claude plugin uninstall gradle-docs@agents-gradle --scope user --keep-data || true
claude plugin uninstall gradle@agents-gradle --scope user --keep-data || true
claude plugin install gradle-rag@agents-gradle --scope user
claude plugin install gradle-grill@agents-gradle --scope user

echo "Installed Gradle plugins:"
echo "  - gradle-rag@agents-gradle"
echo "  - gradle-grill@agents-gradle"
