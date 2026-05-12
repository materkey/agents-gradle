#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLUGINS=(
  gradle-rag
  agp-sources
  gradle-sources
  kotlin-sources
  ksp-sources
  gradle-grill
)

cd "$ROOT"
make build

for path in "${HOME}/.claude/skills/gradle" "${HOME}/.claude/skills/gradle-rag"; do
  if [ -L "$path" ]; then
    rm "$path"
  elif [ -e "$path" ]; then
    echo "Refusing to remove non-symlink legacy skill at $path" >&2
    exit 1
  fi
done

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
