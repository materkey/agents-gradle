#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="${HOME}/.claude/skills/gradle"

cd "$ROOT"
task build

rm -rf "$SKILL_DIR" "${HOME}/.codex/skills/gradle"
mkdir -p "${HOME}/.claude/skills" "${HOME}/.codex/skills"
ln -s "${ROOT}/gradle-docs/skills/gradle" "$SKILL_DIR"
ln -s "${ROOT}/gradle-docs/skills/gradle" "${HOME}/.codex/skills/gradle"

codex plugin marketplace add "$ROOT"
python3 "${HOME}/.codex/plugins/cache/agent-thingz/plugin-management/0.1.0/scripts/pluginctl.py" install agents-gradle gradle-docs --force

claude plugin marketplace add "$ROOT"
claude plugin install gradle-docs@agents-gradle --scope user

echo "Installed Gradle docs skill:"
echo "  Claude skill: ${SKILL_DIR}"
echo "  Codex skill:  ${HOME}/.codex/skills/gradle"
echo "  Plugin:       gradle-docs@agents-gradle"
