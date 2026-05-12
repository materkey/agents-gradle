# agents-gradle

A marketplace of Gradle/AGP agent tooling for Claude Code and Codex.

Plugins:

- **`gradle-rag`** — local lexical search of a locally built snapshot of the official [Gradle documentation](https://docs.gradle.org/current/), backed by an embedded SQLite FTS5 index and a `gradle-rag` Go binary.
- **`gradle-grill`** — pure-skill workflow that stress-tests Gradle/AGP implementation choices against the official docs (via `gradle-rag`) and AGP source, ranks variants, and recommends the most idiomatic option with citations.
- **`agp-sources`** — Android Gradle Plugin source explorer. Downloads AGP source jars from Google Maven into `~/.agp-sources`.
- **`gradle-sources`** — Gradle Build Tool source explorer. Clones Gradle into `${GRADLE_SOURCES_DIR:-$HOME/.gradle-sources}/gradle`.
- **`kotlin-sources`** — Kotlin compiler, standard library, and plugin source explorer. Clones Kotlin into `${KOTLIN_SOURCES_DIR:-$HOME/.kotlin-sources/kotlin}`.
- **`ksp-sources`** — Google KSP source explorer. Clones KSP into `${KSP_SOURCES_DIR:-$HOME/.ksp-sources/ksp}`.

This is an independent project and is not affiliated with or endorsed by Gradle, Inc.

## How It Works

- `cmd/crawler` starts at `https://docs.gradle.org/current/userguide/userguide.html`, follows same-host links under `/current/`, extracts page sections from HTML, and builds a local snapshot at `cmd/gradle-rag/db/gradle.db`.
- `cmd/gradle-rag` embeds that database into a single binary and performs lexical FTS5 search.
- `plugins/gradle-rag/skills/gradle-rag/SKILL.md` tells an agent when and how to call the binary for Gradle-specific documentation lookups. The skill ships a `bin/gradle-rag` wrapper, and the installer copies the built binary into `${GRADLE_RAG_INSTALL_DIR:-$HOME/.local/bin}/gradle-rag` on Darwin and Linux.
- `plugins/gradle-grill/skills/gradle-grill/SKILL.md` is a pure-text skill — no binary — that orchestrates `gradle-rag`, `agp-sources`, `gradle-sources`, `kotlin-sources`, and `ksp-sources` to challenge implementation variants.
- `agp-sources`, `gradle-sources`, `kotlin-sources`, and `ksp-sources` package the source-lookup skills that Gradle workflows rely on as installable standalone plugins.
- `gradle-grill` declares plugin dependencies on `gradle-rag`, `agp-sources`, `gradle-sources`, `kotlin-sources`, and `ksp-sources` in the native plugin manifests and marketplace entries.
- `plugins/*/.claude-plugin/plugin.json` and `plugins/*/.codex-plugin/plugin.json` are versionless plugin manifests.

The crawler indexes content pages from the current User Manual, release notes, Groovy DSL, Kotlin DSL, and Java API while skipping generated navigation/search pages that would dilute search results.

The generated documentation index and built binary are intentionally not committed. `make build` crawls whichever Gradle documentation version the `current` URL resolves to at build time; use `gradle-rag info` to inspect the exact snapshot embedded in the binary. Gradle documentation content is licensed separately by Gradle.

## Requirements

- Go 1.25+
- `make`

## Development

```bash
# Fast proof that crawling, indexing, and embedding work
make crawl-docs-sample
make build-cli
./plugins/gradle-rag/skills/gradle-rag/bin/gradle-rag search "configuration cache" --limit 5

# Full current-docs crawl and binary build
make build

# Tests
make test
```

## Source Bootstrap

The source plugins do not assume a pre-populated `~/projects` checkout. They can fetch their own sources.

For project-specific investigations, use the versions from the project branch you are analyzing, not the examples below. Read the Gradle wrapper version plus AGP, Kotlin, and KSP plugin versions from that branch's build files, settings, or version catalog, then fetch or check out matching sources. Use the defaults only when the question is not tied to a specific project version.

These examples show pinned source bootstrap commands:

```bash
plugins/agp-sources/skills/agp-sources/scripts/fetch_agp_sources.py --version 8.13.0
plugins/gradle-sources/skills/gradle-sources/scripts/clone_gradle.sh v9.3.0
plugins/kotlin-sources/skills/kotlin-sources/scripts/wrapper.py init --ref v2.3.0
plugins/ksp-sources/skills/ksp-sources/scripts/explore.sh init main
```

Omitting version arguments uses the script defaults: latest AGP from Google Maven, Gradle `master`, Kotlin `master`, and KSP `main`.

Each destination can be overridden with `AGP_SOURCES_DIR`, `GRADLE_SOURCES_DIR`, `KOTLIN_SOURCES_DIR`, or `KSP_SOURCES_DIR`.

## CLI

```bash
gradle-rag search "TaskProvider register" --limit 5
gradle-rag search "configuration cache requirements" --full
gradle-rag page "https://docs.gradle.org/current/userguide/configuration_cache.html#config_cache:requirements"
gradle-rag info
```

When working from a local checkout, if `~/.local/bin` is not in `PATH`, use the skill-local wrapper directly. It falls back to the locally built binary under `references/`:

```bash
plugins/gradle-rag/skills/gradle-rag/bin/gradle-rag info
```

## Quick Start: Install The CLI Binary

The `gradle-rag` and `gradle-grill` plugins call a `gradle-rag` executable at runtime. Marketplace installation only installs the plugins; it does not build or install that binary.

Build the documentation index and install the binary into `${GRADLE_RAG_INSTALL_DIR:-$HOME/.local/bin}` first:

```bash
git clone git@github.com:materkey/agents-gradle.git
cd agents-gradle
make install
gradle-rag info
```

If you already have a compatible binary, put it on `PATH` or set `GRADLE_RAG_BIN` to its absolute path.

## Quick Start (Claude Code)

### From The Terminal

After installing the `gradle-rag` binary, install the plugins:

```bash
# Add the marketplace from Git so Claude Code can update it.
claude plugin marketplace add git@github.com:materkey/agents-gradle.git

# Install the Gradle plugins.
claude plugin install gradle-rag@agents-gradle
claude plugin install agp-sources@agents-gradle
claude plugin install gradle-sources@agents-gradle
claude plugin install kotlin-sources@agents-gradle
claude plugin install ksp-sources@agents-gradle
claude plugin install gradle-grill@agents-gradle
```

### From Inside Claude Code

The same commands are available as slash commands in a Claude Code session:

```text
/plugin marketplace add git@github.com:materkey/agents-gradle.git
/plugin install gradle-rag@agents-gradle
/plugin install agp-sources@agents-gradle
/plugin install gradle-sources@agents-gradle
/plugin install kotlin-sources@agents-gradle
/plugin install ksp-sources@agents-gradle
/plugin install gradle-grill@agents-gradle
```

### Auto-Update

Enable marketplace and plugin auto-update on each Claude Code startup:

1. `/plugins`
2. **Manage Marketplaces**
3. `agents-gradle`
4. **Enable auto-update**

Or as a one-liner:

```bash
jq '.extraKnownMarketplaces["agents-gradle"].autoUpdate=true' ~/.claude/settings.json >~/.claude/s.tmp && mv ~/.claude/s.tmp ~/.claude/settings.json
```

## Quick Start (Codex)

After installing the `gradle-rag` binary, register this repository as a remote marketplace. Codex auto-updates plugins from marketplaces registered as Git sources:

```bash
codex plugin marketplace add materkey/agents-gradle@main
```

Then install these plugin ids from Codex's plugin manager: `gradle-rag@agents-gradle`, `agp-sources@agents-gradle`, `gradle-sources@agents-gradle`, `kotlin-sources@agents-gradle`, `ksp-sources@agents-gradle`, and `gradle-grill@agents-gradle`.

The `gradle-rag` plugin expects the `gradle-rag` binary on `PATH`, or `GRADLE_RAG_BIN` pointing at the binary. The plugin wrapper intentionally resolves the runtime binary from the environment so the plugin itself can come from an auto-updating remote marketplace.

## Local Development

```bash
make crawl-docs-sample
make build-cli
make test
```

For a full local documentation snapshot and CLI install, run:

```bash
make install
```

`make install` installs the `gradle-rag` command to `${GRADLE_RAG_INSTALL_DIR:-$HOME/.local/bin}/gradle-rag` on Darwin and Linux. If that directory is not in `PATH`, the installer prints the exact zsh/bash or fish command to add it.

This repository ships versionless local plugin sources under `plugins/`: `gradle-rag`, `gradle-grill`, `agp-sources`, `gradle-sources`, `kotlin-sources`, and `ksp-sources`. The plugin manifests intentionally omit `version`.

Claude Code expands native plugin dependencies from the marketplace. For example, installing `gradle-grill` also installs the Gradle docs and source lookup plugins it orchestrates.

## Evidence Model

`gradle-rag info` reports the Gradle docs version, crawl timestamp, source URL, scheduled page count, indexed page count, and chunk count from the embedded database. Treat that as evidence for the exact documentation snapshot being searched.

## Gradle Documentation License

This project is code for crawling and searching Gradle documentation. It does not redistribute the generated full documentation index. Gradle’s User Manual and DSL Reference are licensed by Gradle under Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International License; see Gradle’s official license page: https://docs.gradle.org/current/userguide/licenses.html

If you distribute a generated index or binary that embeds Gradle documentation, handle that artifact under Gradle’s documentation license terms, including attribution, non-commercial use, and share-alike requirements.
