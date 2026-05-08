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

# pluginctl.py upstream needs `tomllib` (Python 3.11+). The shim self-bootstraps
# the `tomli` backport on 3.9/3.10, so the shell only has to find any Python ≥3.9.
PY="$(command -v python3.13 || command -v python3.12 || command -v python3.11 || command -v python3 || true)"
if [ -z "$PY" ] || ! "$PY" -c 'import sys; sys.exit(0 if sys.version_info >= (3, 9) else 1)'; then
  echo "install.sh: need Python 3.9+ on PATH (found: ${PY:-none})" >&2
  exit 1
fi

codex plugin marketplace add "$ROOT"
"$PY" "$ROOT/scripts/pluginctl-shim.py" install agents-gradle gradle-rag --force
"$PY" "$ROOT/scripts/pluginctl-shim.py" install agents-gradle gradle-grill --force

claude plugin marketplace add "$ROOT"
claude plugin uninstall gradle-docs@agents-gradle --scope user --keep-data || true
claude plugin uninstall gradle@agents-gradle --scope user --keep-data || true
claude plugin install gradle-rag@agents-gradle --scope user
claude plugin install gradle-grill@agents-gradle --scope user

echo "Installed Gradle plugins:"
echo "  - gradle-rag@agents-gradle"
echo "  - gradle-grill@agents-gradle"
