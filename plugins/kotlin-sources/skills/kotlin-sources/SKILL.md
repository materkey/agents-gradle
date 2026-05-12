---
name: kotlin-sources
description: Explore Kotlin compiler and language source code - search classes, compare versions, understand internal implementation
# EXTENDED METADATA (MANDATORY)
github_url: https://github.com/JetBrains/kotlin
github_hash: 40cdea11e577b2e180eb3a8129aaaedfb3181903
version: v2.3.0
created_at: 2026-01-24
entry_point: scripts/wrapper.py
dependencies: ["gh", "git"]
---

# Kotlin Sources Skill

Explore the Kotlin Programming Language compiler and standard library source code directly from the JetBrains/kotlin repository.

## When to Use

Use this skill when:
- Searching for Kotlin compiler implementation details
- Understanding how specific Kotlin features are implemented (coroutines, inline classes, etc.)
- Finding where a compiler error message originates
- Exploring Kotlin standard library source code
- Investigating compiler plugins (kapt, serialization support, etc.)
- Understanding Kotlin/JVM, Kotlin/JS, Kotlin/Native, or Kotlin/Wasm backends
- Looking up internal compiler APIs or IR structures

## Setup

Before first use, clone the Kotlin repository:

```bash
scripts/wrapper.py init
```

This clones to `${KOTLIN_SOURCES_DIR:-$HOME/.kotlin-sources/kotlin}`. To update or pin a version:

```bash
scripts/wrapper.py update
scripts/wrapper.py init --ref v2.3.0
```

For project-specific investigations, prefer the Kotlin version used by the target project branch. Look for `org.jetbrains.kotlin.*` plugin versions in the target project's build files, settings, or version catalog, then pass a matching Kotlin tag or commit with `--ref`.

## Repository Structure

Key directories in the Kotlin repository:

```
kotlin/
├── compiler/                    # Kotlin compiler
│   ├── frontend/               # K1 frontend (lexer, parser, resolution)
│   ├── frontend.java/          # Java interop for K1
│   ├── fir/                    # K2 frontend (FIR - Frontend IR)
│   │   ├── raw-fir/           # Raw FIR tree construction
│   │   ├── resolve/           # FIR resolution
│   │   ├── checkers/          # FIR diagnostics
│   │   └── fir2ir/            # FIR to IR lowering
│   ├── ir/                     # Intermediate Representation
│   │   ├── ir.tree/           # IR tree definitions
│   │   ├── backend.common/    # Common backend code
│   │   └── ir.interpreter/    # Compile-time evaluation
│   ├── backend.jvm/           # JVM bytecode generation
│   ├── backend.js/            # JavaScript generation
│   ├── backend.wasm/          # WebAssembly generation
│   ├── cli/                   # Command-line interface
│   └── daemon/                # Compiler daemon
├── kotlin-native/              # Kotlin/Native compiler
├── libraries/                  # Standard library and tools
│   ├── stdlib/                # Kotlin standard library
│   ├── tools/                 # Kotlin tools (kapt, etc.)
│   └── kotlinx-metadata/      # Metadata library
├── plugins/                    # Compiler plugins
│   ├── compose/               # Compose compiler plugin
│   ├── serialization/         # Serialization plugin
│   ├── kapt/                  # Annotation processing
│   ├── parcelize/             # Android Parcelable
│   └── sam-with-receiver/     # SAM conversions
├── analysis/                   # Analysis API
│   ├── analysis-api/          # Public Analysis API
│   └── analysis-api-fir/      # FIR-based implementation
├── js/                         # Kotlin/JS runtime
├── wasm/                       # Kotlin/Wasm runtime
└── idea/                       # IntelliJ IDEA plugin
```

## Common Search Patterns

### Find Compiler Error Messages
```bash
# Search for diagnostic messages
./scripts/wrapper.py search "UNRESOLVED_REFERENCE" --type kt

# Find where errors are defined
./scripts/wrapper.py search "KtDiagnosticFactory" --path compiler/fir
```

### Find IR Node Definitions
```bash
# Search for specific IR elements
./scripts/wrapper.py search "IrFunction" --path compiler/ir/ir.tree

# Find IR lowerings
./scripts/wrapper.py search "Lowering" --path compiler/ir
```

### Find Standard Library Implementations
```bash
# Search stdlib implementations
./scripts/wrapper.py search "actual fun" --path libraries/stdlib

# Find collection implementations
./scripts/wrapper.py search "ArrayList" --path libraries/stdlib
```

### Find FIR (K2) Components
```bash
# Search FIR resolution
./scripts/wrapper.py search "FirResolver" --path compiler/fir/resolve

# Find FIR checkers
./scripts/wrapper.py search "FirChecker" --path compiler/fir/checkers
```

## Usage Examples

### 1. Search for a class or function
```bash
scripts/wrapper.py search "InlineClassLowering"
```

### 2. Browse specific directory
```bash
scripts/wrapper.py browse compiler/backend.jvm
```

### 3. Read a specific file
```bash
scripts/wrapper.py read compiler/fir/resolve/src/org/jetbrains/kotlin/fir/resolve/FirTypeResolver.kt
```

### 4. Find diagnostics/errors
```bash
scripts/wrapper.py diagnostics UNRESOLVED_REFERENCE
```

### 5. Compare with a specific version
```bash
scripts/wrapper.py compare v2.0.0 v2.1.0 --path compiler/fir
```

## Quick Reference

| Component | Path | Description |
|-----------|------|-------------|
| K2 Frontend | `compiler/fir/` | New frontend with FIR |
| K1 Frontend | `compiler/frontend/` | Legacy frontend |
| JVM Backend | `compiler/backend.jvm/` | JVM bytecode generation |
| JS Backend | `compiler/backend.js/` | JavaScript generation |
| Wasm Backend | `compiler/backend.wasm/` | WebAssembly generation |
| IR | `compiler/ir/` | Intermediate representation |
| Stdlib | `libraries/stdlib/` | Standard library |
| Compose Plugin | `plugins/compose/` | Compose compiler |
| Analysis API | `analysis/` | IDE analysis infrastructure |

## Building Locally

```bash
# Clone if not present
git clone https://github.com/JetBrains/kotlin.git

# Build compiler distribution
./gradlew dist

# Run compiler tests
./gradlew compilerTest

# Run stdlib tests
./gradlew coreLibsTest
```

## Resources

- [Kotlin Site](https://kotlinlang.org/)
- [Issue Tracker (YouTrack)](https://youtrack.jetbrains.com/issues/KT)
- [Contributing Guide](https://github.com/JetBrains/kotlin/blob/master/docs/contributing.md)
- [Kotlin Blog](https://blog.jetbrains.com/kotlin/)
