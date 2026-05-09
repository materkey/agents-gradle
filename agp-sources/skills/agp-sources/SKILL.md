---
name: agp-sources
description: Explore Android Gradle Plugin (AGP) source code - search classes, compare versions, understand internal implementation
entry_point: scripts/fetch_agp_sources.py
dependencies: ["python3"]
---

# AGP Sources Exploration Skill

Search and explore Android Gradle Plugin source code from local repository.

## Repository Location

```
${AGP_SOURCES_DIR:-$HOME/.agp-sources}
```

## Setup

Before first use, download AGP source jars from Google Maven:

```bash
scripts/fetch_agp_sources.py --version 8.13.0
```

Without `--version`, the script downloads the latest AGP version listed in Google Maven:

```bash
scripts/fetch_agp_sources.py
```

## Downloaded Versions

The local cache contains only versions that have been downloaded with `scripts/fetch_agp_sources.py`.
To list downloaded versions:
```bash
ls -d "${AGP_SOURCES_DIR:-$HOME/.agp-sources}"/8.* | xargs -n1 basename | sort -V
```

## Repository Structure

Each version contains these modules:

| Module | Description |
|--------|-------------|
| `com.android.tools.build/gradle` | **Main AGP plugin code** - tasks, extensions, variants |
| `com.android.tools.build/gradle-api` | Public API for build scripts |
| `com.android.tools.build/builder` | Build operations (dex, resources, signing) |
| `com.android.tools.build/builder-model` | Model classes for IDE integration |
| `com.android.tools.build/manifest-merger` | AndroidManifest.xml merging |
| `com.android.tools.build/apksig` | APK signing implementation |
| `com.android.tools.build/apkzlib` | APK/ZIP manipulation |
| `com.android.tools.lint` | Lint checks and API |
| `com.android.tools/sdk-common` | SDK utilities |
| `com.android.tools/common` | Shared utilities |

## Key Directories

Main plugin implementation:
```
{version}/com.android.tools.build/gradle/com/android/build/gradle/
├── internal/          # Internal implementation
│   ├── tasks/         # Gradle tasks (compile, package, etc.)
│   ├── variant/       # Variant configuration
│   ├── dsl/           # DSL implementation
│   └── cxx/           # Native build support
├── api/               # Public API interfaces
├── tasks/             # Public task classes
└── options/           # Build options
```

## Common Search Patterns

### Find a class by name

```bash
# Find class in latest version
find "${AGP_SOURCES_DIR:-$HOME/.agp-sources}"/8.13.* -name "*.kt" -o -name "*.java" | xargs grep -l "class YourClassName"

# Find in specific version
find "${AGP_SOURCES_DIR:-$HOME/.agp-sources}"/8.7.0 -name "*.kt" -o -name "*.java" | xargs grep -l "class YourClassName"
```

### Search for task implementation

```bash
# Find task by name pattern
grep -r "class.*Task" "${AGP_SOURCES_DIR:-$HOME/.agp-sources}"/8.7.0/com.android.tools.build/gradle/com/android/build/gradle/internal/tasks/ --include="*.kt" | head -20
```

### Find DSL property

```bash
# Search for DSL property
grep -r "propertyName" "${AGP_SOURCES_DIR:-$HOME/.agp-sources}"/8.7.0/com.android.tools.build/gradle/com/android/build/api/ --include="*.kt"
```

## Compare Versions

### Diff specific file between versions

```bash
diff "${AGP_SOURCES_DIR:-$HOME/.agp-sources}"/8.6.0/com.android.tools.build/gradle/com/android/build/gradle/internal/tasks/SomeTask.kt \
     "${AGP_SOURCES_DIR:-$HOME/.agp-sources}"/8.7.0/com.android.tools.build/gradle/com/android/build/gradle/internal/tasks/SomeTask.kt
```

### Find when class was added/changed

```bash
for v in "${AGP_SOURCES_DIR:-$HOME/.agp-sources}"/8.*; do
  if [ -f "$v/com.android.tools.build/gradle/com/android/build/gradle/SomeClass.kt" ]; then
    echo "$(basename $v): exists"
  fi
done
```

## Important Classes

### Plugins
- `AppPlugin.kt` - Application plugin (`com.android.application`)
- `LibraryPlugin.kt` - Library plugin (`com.android.library`)
- `DynamicFeaturePlugin.kt` - Dynamic feature module

### Extensions (DSL)
- `BaseExtension.kt` - Base android {} block
- `ApplicationExtension.kt` - app-specific DSL
- `LibraryExtension.kt` - library-specific DSL
- `CommonExtension.kt` - Shared DSL interface

### Tasks
- `internal/tasks/` - All task implementations
- `MergeResources.kt` - Resource merging
- `PackageApplication.kt` - APK packaging
- `CompileLibraryClassesTask.kt` - Compilation

### Variants
- `internal/variant/` - Variant system
- `VariantBuilder.kt` - Variant configuration
- `ComponentIdentity.kt` - Variant identity

## Usage Examples

### Example 1: Find how minSdk is processed

```bash
grep -rn "minSdk" "${AGP_SOURCES_DIR:-$HOME/.agp-sources}"/8.7.0/com.android.tools.build/gradle/ --include="*.kt" | head -30
```

### Example 2: Understand R8/ProGuard integration

```bash
ls "${AGP_SOURCES_DIR:-$HOME/.agp-sources}"/8.7.0/com.android.tools.build/gradle/com/android/build/gradle/internal/tasks/*R8*.kt
cat "${AGP_SOURCES_DIR:-$HOME/.agp-sources}"/8.7.0/com.android.tools.build/gradle/com/android/build/gradle/internal/tasks/R8Task.kt
```

### Example 3: Find all Gradle properties

```bash
grep -r "GradleProperty\|Property<" "${AGP_SOURCES_DIR:-$HOME/.agp-sources}"/8.7.0/com.android.tools.build/gradle-api/ --include="*.kt"
```

## Trigger Patterns

Use this skill when user asks about:
- AGP source code / "как работает AGP" / "покажи код AGP"
- Internal AGP implementation / "как реализовано в AGP"
- AGP task implementation / "как работает task"
- Comparing AGP versions / "что изменилось в AGP"
- DSL implementation / "как работает android {} блок"
- Variant system / "как работают варианты"
- Build types/flavors implementation

## Tips

1. **Start with gradle-api** for public interfaces, then look at `internal/` for implementation
2. **Use 8.7.0 as reference** - stable, well-documented version
3. **Check internal/tasks/** for task implementations
4. **Check internal/dsl/** for DSL parsing
5. **Use glob for faster file finding**: `ls {version}/*/gradle/**/SomeClass.kt`

## Notes

- Sources extracted from official Google Maven repository
- Only Java/Kotlin sources included (no compiled classes)
- Some generated code may be missing (proto files, etc.)
