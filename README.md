# agents-gradle

A local marketplace of Gradle/AGP agent tooling for Claude Code and Codex.

Plugins:

- **`gradle-rag`** — local lexical search of the current official [Gradle documentation](https://docs.gradle.org/current/), backed by an embedded SQLite FTS5 index and a `gradle-rag` Go binary.
- **`gradle-grill`** — pure-skill workflow that stress-tests Gradle/AGP implementation choices against the official docs (via `gradle-rag`) and AGP source, ranks variants, and recommends the most idiomatic option with citations.

This is an independent project and is not affiliated with or endorsed by Gradle, Inc.

## How It Works

- `cmd/crawler` starts at `https://docs.gradle.org/current/userguide/userguide.html`, follows same-host links under `/current/`, extracts page sections from HTML, and builds `cmd/gradle-rag/db/gradle.db`.
- `cmd/gradle-rag` embeds that database into a single binary and performs lexical FTS5 search.
- `gradle-rag/skills/gradle-rag/SKILL.md` tells an agent when and how to call the binary for Gradle-specific documentation lookups.
- `gradle-grill/skills/gradle-grill/SKILL.md` is a pure-text skill — no binary — that orchestrates `gradle-rag` and other doc tools to challenge implementation variants.
- `*/.claude-plugin/plugin.json` and `*/.codex-plugin/plugin.json` are versionless plugin manifests.

The crawler indexes content pages from the current User Manual, release notes, Groovy DSL, Kotlin DSL, and Java API while skipping generated navigation/search pages that would dilute search results.

The generated documentation index and built binary are intentionally not committed. Gradle documentation content is licensed separately by Gradle; build the index locally with `task build`.

## Requirements

- Go 1.25+
- [Task](https://taskfile.dev/) for the documented commands

## Development

```bash
# Fast proof that crawling, indexing, and embedding work
task smoke-db
task build-fast
./gradle-rag/skills/gradle-rag/references/gradle-rag search "configuration cache" --limit 5

# Full current-docs crawl and binary build
task build

# Tests
task test
```

## CLI

```bash
gradle-rag/skills/gradle-rag/references/gradle-rag search "TaskProvider register" --limit 5
gradle-rag/skills/gradle-rag/references/gradle-rag search "configuration cache requirements" --full
gradle-rag/skills/gradle-rag/references/gradle-rag page "https://docs.gradle.org/current/userguide/configuration_cache.html#config_cache:requirements"
gradle-rag/skills/gradle-rag/references/gradle-rag info
```

## Install As A Local Plugin

```bash
task build              # crawl full current docs and build the gradle-rag binary
task install-plugins    # install both gradle-rag and gradle-grill into Claude and Codex
```

This repository ships two versionless local plugin sources: `gradle-rag/` and `gradle-grill/`. The plugin manifests intentionally omit `version`; Codex installs them as `local`, while Claude Code caches them from the current source revision.

The installer removes legacy direct skill symlinks at `~/.claude/skills/gradle`, `~/.codex/skills/gradle`, `~/.claude/skills/gradle-rag`, and `~/.codex/skills/gradle-rag` so Claude and Codex expose the skills only through the plugins.

`task install-local` is kept as a compatibility alias for `task install-plugins`.

## Evidence Model

`gradle-rag info` reports the Gradle docs version, crawl timestamp, source URL, scheduled page count, indexed page count, and chunk count from the embedded database. Treat that as evidence for the exact documentation snapshot being searched.

## Gradle Documentation License

This project is code for crawling and searching Gradle documentation. It does not redistribute the generated full documentation index. Gradle’s User Manual and DSL Reference are licensed by Gradle under Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International License; see Gradle’s official license page: https://docs.gradle.org/current/userguide/licenses.html

If you distribute a generated index or binary that embeds Gradle documentation, handle that artifact under Gradle’s documentation license terms, including attribution, non-commercial use, and share-alike requirements.
