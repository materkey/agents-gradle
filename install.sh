#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="${HOME}/.claude/skills/gradle"
REF_DIR="${SKILL_DIR}/references"

cd "$ROOT"
task build

mkdir -p "$REF_DIR"
cp skill/gradle/SKILL.md "${SKILL_DIR}/SKILL.md"
cp gradle-rag "${REF_DIR}/gradle-rag"

echo "Installed Gradle docs skill:"
echo "  Skill:  ${SKILL_DIR}/SKILL.md"
echo "  Binary: ${REF_DIR}/gradle-rag"

