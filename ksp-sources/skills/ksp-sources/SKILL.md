---
name: ksp-sources
description: Explore Google KSP (Kotlin Symbol Processing) source code - search classes, understand annotation processing API, debug KSP processors, compare KSP1 vs KSP2 implementations
# EXTENDED METADATA (MANDATORY)
github_url: https://github.com/google/ksp
github_hash: 190feab990d690ec9862990da63c530bd9f0b17f
version: 0.1.0
created_at: 2026-01-26
entry_point: scripts/explore.sh
dependencies: ["git", "grep", "find"]
---

# KSP Sources Explorer

Explore and understand Google's Kotlin Symbol Processing (KSP) source code directly.

## What is KSP?

Kotlin Symbol Processing (KSP) is a lightweight compiler plugin API that leverages Kotlin's power while keeping the learning curve minimal. Compared to KAPT, annotation processors using KSP run up to 2x faster.

**Key Points:**
- KSP2 is the default since 2025 (enabled by default in KSP 2.0.0+)
- KSP1 is deprecated and won't support Kotlin 2.3.0+ or AGP 9.0+
- Switch between versions: `ksp.useKSP2=false` or `ksp { useKsp2 = false }`

## When to Use This Skill

- Debugging KSP processor issues
- Understanding KSP API internals
- Finding how specific KSP features are implemented
- Comparing KSP1 vs KSP2 implementations
- Learning how to write KSP processors by studying examples

## Repository Structure

Key directories in google/ksp:
- `api/` - KSP API interfaces (`KSProcessor`, `Resolver`, `KSNode`, etc.)
- `compiler-plugin/` - KSP1 compiler plugin implementation
- `kotlin-analysis-api/` - KSP2 implementation using Kotlin Analysis API
- `common-util/` - Shared utilities
- `gradle-plugin/` - Gradle integration
- `symbol-processing/` - Core symbol processing logic
- `test-utils/` - Testing utilities for KSP processors

## Common Tasks

### Find a KSP API class
```bash
# Find where KSClassDeclaration is defined
grep -rn "interface KSClassDeclaration" "${KSP_SOURCES_DIR:-$HOME/.ksp-sources/ksp}/"

# Find all visitor implementations
grep -rn "class.*Visitor" "${KSP_SOURCES_DIR:-$HOME/.ksp-sources/ksp}/api/"
```

### Understand Resolver API
```bash
# Find Resolver interface
grep -rn "interface Resolver" "${KSP_SOURCES_DIR:-$HOME/.ksp-sources/ksp}/api/"

# Find symbol resolution logic
grep -rn "getSymbolsWithAnnotation" "${KSP_SOURCES_DIR:-$HOME/.ksp-sources/ksp}/"
```

### Compare KSP1 vs KSP2
```bash
# KSP1 implementation
ls "${KSP_SOURCES_DIR:-$HOME/.ksp-sources/ksp}/compiler-plugin/src/"

# KSP2 implementation
ls "${KSP_SOURCES_DIR:-$HOME/.ksp-sources/ksp}/kotlin-analysis-api/src/"
```

### Debug incremental processing
```bash
grep -rn "incremental" "${KSP_SOURCES_DIR:-$HOME/.ksp-sources/ksp}/gradle-plugin/"
grep -rn "IncrementalProcessor" "${KSP_SOURCES_DIR:-$HOME/.ksp-sources/ksp}/"
```

## Workflow

1. **Clone if needed**: The skill clones the repo to `${KSP_SOURCES_DIR:-$HOME/.ksp-sources/ksp}` if not present
2. **Search**: Use grep/find to locate relevant code
3. **Read**: Examine source files to understand implementation
4. **Compare**: Look at both KSP1 and KSP2 implementations when relevant

## Key Classes to Know

| Class | Purpose |
|-------|---------|
| `SymbolProcessor` | Main interface for KSP processors |
| `SymbolProcessorProvider` | Factory for creating processors |
| `Resolver` | Access to symbols during processing |
| `KSClassDeclaration` | Represents a class in source |
| `KSFunctionDeclaration` | Represents a function |
| `KSPropertyDeclaration` | Represents a property |
| `KSAnnotation` | Represents an annotation |
| `KSVisitor` | Visitor pattern for traversing symbols |
| `CodeGenerator` | Generates output files |

## Documentation Links

- [KSP Overview](https://kotlinlang.org/docs/ksp-overview.html)
- [Quickstart](https://kotlinlang.org/docs/ksp-quickstart.html)
- [KSP Examples](https://kotlinlang.org/docs/ksp-examples.html)
- [Incremental Processing](https://kotlinlang.org/docs/ksp-incremental.html)
- [Multi-round Processing](https://kotlinlang.org/docs/ksp-multi-round.html)
- [KSP2 Introduction](https://github.com/google/ksp/blob/main/docs/ksp2.md)

## Example: Finding How Annotations Are Resolved

```bash
# Step 1: Clone/update repo
scripts/explore.sh init

# Step 2: Find annotation handling
grep -rn "getSymbolsWithAnnotation" "${KSP_SOURCES_DIR:-$HOME/.ksp-sources/ksp}/" --include="*.kt"

# Step 3: Find the Resolver implementation
grep -rn "override fun getSymbolsWithAnnotation" "${KSP_SOURCES_DIR:-$HOME/.ksp-sources/ksp}/"
```
