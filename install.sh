#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

cd "$ROOT"
task build

for path in "${HOME}/.claude/skills/gradle" "${HOME}/.codex/skills/gradle"; do
  if [ -L "$path" ]; then
    rm "$path"
  elif [ -e "$path" ]; then
    echo "Refusing to remove non-symlink legacy skill at $path" >&2
    exit 1
  fi
done

codex plugin marketplace add "$ROOT"
python3 "${HOME}/.codex/plugins/cache/agent-thingz/plugin-management/0.1.0/scripts/pluginctl.py" install agents-gradle gradle-docs --force

claude plugin marketplace add "$ROOT"
claude plugin install gradle-docs@agents-gradle --scope user

echo "Installed Gradle docs skill:"
echo "  Plugin:       gradle-docs@agents-gradle"
