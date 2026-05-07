# agents-gradle

Local lexical search for the current official [Gradle documentation](https://docs.gradle.org/current/), packaged like `claude-code-fpf`: one crawler, one embedded SQLite FTS5 index, one `gradle-rag` binary, and one Claude Code skill.

This is an independent project and is not affiliated with or endorsed by Gradle, Inc.

## How It Works

- `cmd/crawler` starts at `https://docs.gradle.org/current/userguide/userguide.html`, follows same-host links under `/current/`, extracts page sections from HTML, and builds `cmd/gradle-rag/db/gradle.db`.
- `cmd/gradle-rag` embeds that database into a single binary and performs lexical FTS5 search.
- `skill/gradle/SKILL.md` tells an agent when and how to call the binary for Gradle-specific documentation lookups.

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
./gradle-rag search "configuration cache" --limit 5

# Full current-docs crawl and binary build
task build

# Tests
task test
```

## CLI

```bash
gradle-rag search "TaskProvider register" --limit 5
gradle-rag search "configuration cache requirements" --full
gradle-rag page "https://docs.gradle.org/current/userguide/configuration_cache.html#config_cache:requirements"
gradle-rag info
```

## Install As A Skill

```bash
task install-local
```

This installs:

- `~/.claude/skills/gradle/SKILL.md`
- `~/.claude/skills/gradle/references/gradle-rag`

## Evidence Model

`gradle-rag info` reports the Gradle docs version, crawl timestamp, source URL, scheduled page count, indexed page count, and chunk count from the embedded database. Treat that as evidence for the exact documentation snapshot being searched.

## Gradle Documentation License

This project is code for crawling and searching Gradle documentation. It does not redistribute the generated full documentation index. Gradle’s User Manual and DSL Reference are licensed by Gradle under Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International License; see Gradle’s official license page: https://docs.gradle.org/current/userguide/licenses.html

If you distribute a generated index or binary that embeds Gradle documentation, handle that artifact under Gradle’s documentation license terms, including attribution, non-commercial use, and share-alike requirements.
