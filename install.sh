#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLUGINS=(
  gradle-rag
  gradle-grill
  agp-sources
  gradle-sources
  kotlin-sources
  ksp-sources
)

cd "$ROOT"
make build

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

PY="$(command -v python3.13 || command -v python3.12 || command -v python3.11 || command -v python3 || true)"
if [ -z "$PY" ] || ! "$PY" -c 'import sys; sys.exit(0 if sys.version_info >= (3, 9) else 1)'; then
  echo "install.sh: need Python 3.9+ on PATH (found: ${PY:-none})" >&2
  exit 1
fi

"$PY" "$ROOT/scripts/install-codex-plugins.py" --root "$ROOT" "${PLUGINS[@]}"

claude plugin marketplace add "$ROOT"
claude plugin uninstall gradle-docs@agents-gradle --scope user --keep-data || true
claude plugin uninstall gradle@agents-gradle --scope user --keep-data || true
for plugin in "${PLUGINS[@]}"; do
  claude plugin uninstall "${plugin}@agents-gradle" --scope user --keep-data || true
  claude plugin install "${plugin}@agents-gradle" --scope user
done

"$ROOT/scripts/install-gradle-rag-bin.sh"

echo "Installed Gradle plugins:"
for plugin in "${PLUGINS[@]}"; do
  echo "  - ${plugin}@agents-gradle"
done
echo "  - gradle-rag command at ${GRADLE_RAG_INSTALL_DIR:-${HOME}/.local/bin}/gradle-rag"
