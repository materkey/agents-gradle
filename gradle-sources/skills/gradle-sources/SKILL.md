---
name: gradle-sources
description: Explore Gradle Build Tool source code - search classes, compare versions, understand internal implementation
# EXTENDED METADATA (MANDATORY)
github_url: https://github.com/gradle/gradle
github_hash: a7b54eb3c74093c24a8ab5b28dfc151d5b7ea79e
version: v9.3.0
created_at: 2026-01-24T12:00:00Z
entry_point: scripts/clone_gradle.sh
dependencies: ["git", "gh"]
---

# Gradle Sources Explorer

Explore Gradle Build Tool source code to understand internal implementation, search for classes, and compare versions.

## Use Cases

- **Find implementation details**: Search for how Gradle implements specific features (task execution, dependency resolution, caching)
- **Understand internal APIs**: Explore Gradle's internal classes and interfaces
- **Compare versions**: Check differences between Gradle versions
- **Debug Gradle issues**: Find source code relevant to Gradle behavior

## Setup

Before first use, clone the Gradle repository:

```bash
scripts/clone_gradle.sh
```

This clones to `${GRADLE_SOURCES_DIR:-$HOME/.gradle-sources}/gradle`.

## Workflow

### 1. Clone/Update Repository

```bash
# Initial clone or update existing
scripts/clone_gradle.sh

# Clone specific version
scripts/clone_gradle.sh v9.3.0
```

### 2. Search for Classes/Code

```bash
# Find a class by name
rg -l "class DefaultTaskExecutionGraph" "${GRADLE_SOURCES_DIR:-$HOME/.gradle-sources}/gradle"

# Search for specific implementation
rg "fun execute" --type kotlin "${GRADLE_SOURCES_DIR:-$HOME/.gradle-sources}/gradle/subprojects/core"

# Find all task-related classes
fd "Task.*\.kt$" "${GRADLE_SOURCES_DIR:-$HOME/.gradle-sources}/gradle/subprojects"
```

### 3. Key Source Directories

| Directory | Contents |
|-----------|----------|
| `subprojects/core/` | Core Gradle engine (Project, Task, Build) |
| `subprojects/core-api/` | Public API interfaces |
| `subprojects/dependency-management/` | Dependency resolution |
| `subprojects/execution/` | Task execution engine |
| `subprojects/model-core/` | Configuration model |
| `subprojects/plugins/` | Built-in plugins (Java, Application) |
| `subprojects/kotlin-dsl/` | Kotlin DSL support |
| `subprojects/tooling-api/` | IDE integration API |
| `platforms/` | Platform-specific code (JVM, Native, Software) |

### 4. Compare Versions

```bash
cd "${GRADLE_SOURCES_DIR:-$HOME/.gradle-sources}/gradle"

# List available versions
git tag | grep "^v[0-9]" | sort -V | tail -20

# Compare specific files between versions
git diff v9.2.0..v9.3.0 -- subprojects/core/src/main/java/org/gradle/api/Task.java

# Show changes in a module
git log --oneline v9.2.0..v9.3.0 -- subprojects/dependency-management/
```

### 5. Create GitHub Permalinks

When referencing Gradle source code, create permalinks:

```
https://github.com/gradle/gradle/blob/{commit_hash}/path/to/file.kt#L{line_number}
```

Example:
```
https://github.com/gradle/gradle/blob/a7b54eb3c74093c24a8ab5b28dfc151d5b7ea79e/subprojects/core/src/main/java/org/gradle/api/Task.java#L50
```

## Common Searches

### Task Execution
```bash
rg -l "TaskExecution" "${GRADLE_SOURCES_DIR:-$HOME/.gradle-sources}/gradle/subprojects/execution"
```

### Dependency Resolution
```bash
rg -l "DependencyResolver" "${GRADLE_SOURCES_DIR:-$HOME/.gradle-sources}/gradle/subprojects/dependency-management"
```

### Configuration Cache
```bash
rg -l "ConfigurationCache" "${GRADLE_SOURCES_DIR:-$HOME/.gradle-sources}/gradle/subprojects/configuration-cache"
```

### Build Cache
```bash
rg -l "BuildCache" "${GRADLE_SOURCES_DIR:-$HOME/.gradle-sources}/gradle/subprojects/build-cache"
```

### Plugin Application
```bash
rg "fun apply" --type kotlin "${GRADLE_SOURCES_DIR:-$HOME/.gradle-sources}/gradle/subprojects/plugins"
```

## Resources

- [Gradle User Manual](https://docs.gradle.org/current/userguide/userguide.html)
- [Gradle DSL Reference](https://docs.gradle.org/current/dsl/)
- [Gradle Javadoc](https://docs.gradle.org/current/javadoc/)
- [Gradle Community Slack](https://gradle.org/slack-invite)
- [Contributing Guide](https://github.com/gradle/gradle/blob/master/CONTRIBUTING.md)
