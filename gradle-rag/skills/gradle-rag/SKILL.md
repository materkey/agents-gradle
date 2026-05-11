---
name: gradle-rag
description: "Search the current official Gradle documentation with a local lexical index."
argument-hint: "[Gradle concept, API, DSL, plugin, error, or build behavior question]"
---

# Gradle Docs Search

Use this skill when the user asks about Gradle Build Tool behavior, Gradle APIs, the Groovy DSL, Kotlin DSL, core plugins, dependency management, configuration cache, task development, plugin development, or release/upgrade notes.

Preferred command: `gradle-rag`. The local installer copies the built binary to `${GRADLE_RAG_INSTALL_DIR:-$HOME/.local/bin}/gradle-rag` on Darwin and Linux.

Fallback command: `bin/gradle-rag` relative to this `SKILL.md`. It resolves the generated binary at `references/gradle-rag`.

If `gradle-rag` is not on `PATH`, resolve this skill's actual directory and run its `bin/gradle-rag` wrapper. Do not fail just because the bare command is unavailable.

## Workflow

1. Search first with a focused lexical query:

```bash
gradle-rag search "configuration cache requirements" --limit 5
```

2. If a result is relevant but too terse, repeat with `--full` or fetch the exact section:

```bash
gradle-rag search "TaskProvider register" --full
gradle-rag page "https://docs.gradle.org/current/userguide/configuration_cache.html#config_cache:requirements"
```

3. Check index freshness when it matters:

```bash
gradle-rag info
```

## Discipline

- Prefer direct Gradle documentation evidence over memory for version-sensitive behavior.
- Quote or cite the `Source:` URL from search results when answering.
- If the indexed docs do not cover the question, say that clearly and fall back to an explicit external lookup.
