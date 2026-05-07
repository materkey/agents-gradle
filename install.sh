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

codex plugin marketplace add "$ROOT"
python3 "${HOME}/.codex/plugins/cache/agent-thingz/plugin-management/0.1.0/scripts/pluginctl.py" install agents-gradle gradle-rag --force

claude plugin marketplace add "$ROOT"
claude plugin uninstall gradle-docs@agents-gradle --scope user --keep-data || true
claude plugin uninstall gradle@agents-gradle --scope user --keep-data || true
claude plugin install gradle-rag@agents-gradle --scope user

echo "Installed Gradle docs skill:"
echo "  Plugin:       gradle-rag@agents-gradle"
