---
name: gradle-grill
description: Challenge a Gradle/AGP implementation choice against the official docs. Generates 2–4 candidate variants, cross-checks each via gradle-rag and AGP source, ranks them, and recommends the most idiomatic option with citations.
---

# Gradle Grill

Challenge a Gradle/AGP implementation choice. Generate variants, verify each against the official docs, recommend the most idiomatic one with quoted citations.

This skill exists because Gradle has accumulated several generations of APIs (eager → lazy → configuration cache, AGP DSL → AndroidComponentsExtension), and "the obvious answer" is frequently the obsolete one. The skill forces an explicit doc-grounded comparison instead of pattern-matching from memory.

## When to invoke

- The user proposes a Gradle implementation and asks "is this right?" or "is there a better way?"
- The user asks "where should I put X" — a hook callback, a task config, a precondition check, a DSL override.
- The user is about to write or edit a `Plugin<Project>`, convention plugin, custom task, or build-script `apply`/`register` block.
- A code review surfaces eager APIs (`tasks.create`, `tasks.getByName`, `afterEvaluate`, `File.exists()` in config, `configurations.X.files` in config).

Trigger phrases include: "как лучше в gradle", "правильный способ X в gradle", "идиоматично", "challenge gradle approach", "гриль gradle", "gradle-grill", "afterEvaluate vs", "tasks.register vs", "as a plugin author".

Do not invoke for pure code-search ("where is X used") — use `curiosity` or `Grep`. Do not invoke for trace analysis — use `perfetto-trace`.

## Workflow

### 1. Restate the problem in one sentence

Paraphrase the user's question into a single concrete decision. Example:
> "Where to put a fail-fast precondition that a module's `values/strings.xml` exists, when the i18n convention plugin is applied?"

If the question is too vague to ground in docs, ask exactly one clarifying question and stop. Do not generate variants from a fuzzy premise.

### 2. Generate 2–4 candidate variants

Pick distinct mechanisms, not minor stylistic variations. Lean on these axes:

| Axis | Common variants |
|---|---|
| Lifecycle hook | `apply{}` body, `afterEvaluate{}`, `AndroidComponentsExtension.finalizeDsl{}`, `onVariants{}`, `beforeVariants{}` |
| Task creation | `tasks.create(...)`, `tasks.register(...)`, `withType().configureEach{}` |
| Task lookup | `tasks.getByName(...)` / `findByName(...)` (eager), `tasks.named(...)` (lazy `TaskProvider`) |
| Validation point | configuration block of `register{}`, `doFirst{}`, `@InputFiles`+`@SkipWhenEmpty`, project `afterEvaluate`, AGP `finalizeDsl` |
| Inputs/outputs | raw `File`, `RegularFileProperty`/`DirectoryProperty`, `ConfigurableFileCollection`, `Provider<T>` chain |
| Property chains | direct `String`/`File`, `Property<T>`/`Provider<T>` with `.map{}`/`.flatMap{}` |

Each variant must include a 5–15 line code sketch — enough to be evaluated, not a full plugin.

### 3. Cross-check each variant against Gradle docs

For each variant, run **at least two** parallel `gradle-rag` searches with different angles. Phrase queries as nouns, not full sentences. Examples:

```bash
gradle-rag search "afterEvaluate restrictions plugin author" --limit 4
gradle-rag search "tasks register lazy configuration block" --limit 4
gradle-rag search "configuration cache afterEvaluate" --limit 4
```

When the topic is AGP-specific (`AndroidComponentsExtension`, `LibraryAndroidComponentsExtension`, `Variant`, `finalizeDsl`, `onVariants`, source sets), additionally search agp-sources:

```bash
# from the agp-sources skill directory
scripts/fetch_agp_sources.py --version 8.8
rg "AndroidComponentsExtension|finalizeDsl" "${AGP_SOURCES_DIR:-$HOME/.agp-sources}/8.8.0/com.android.tools.build/gradle"
```

If a Gradle-/AGP-internal mechanism is being compared, pull the actual source via `gradle-sources` / `agp-sources` to see what the API guarantees, not just what the docs say.

When the topic is KSP-specific (`com.google.devtools.ksp`, KSP1/KSP2, symbol processors, incremental processing), additionally pull the actual source via `ksp-sources`.

Capture for each variant:
- One direct quote from the docs (≤ 30 words)
- The source URL (Gradle userguide section or AGP javadoc)
- A one-line interpretation: "this means …"

Empty doc results are a signal — note explicitly that docs do not address this variant directly.

### 4. Apply the canonical principle table

These are non-negotiable principles drawn from Gradle's own best-practices guide. They override "I've seen it done this way before":

| Avoid | Prefer | Why (cite) |
|---|---|---|
| `tasks.create()` | `tasks.register()` | task configuration avoidance — task body runs only when realized |
| `tasks.getByName{}` / `findByName{}` | `tasks.named{}` | returns `TaskProvider`, stays lazy |
| `tasks.withType(T){ … }` | `tasks.withType(T).configureEach{ … }` | closure makes withType eager |
| `someTask { }` (Groovy sugar) | `tasks.named("someTask") { }` | hidden eager realization (docs name this exact pitfall) |
| `afterEvaluate{}` (in plugin code) | `AndroidComponentsExtension.finalizeDsl{}` / `onVariants{}` (AGP) or `Provider<T>` chain | docs warn: "mixing delayed configuration with the new API can cause errors that are hard to diagnose"; "if afterEvaluate is declared in a plugin then report the issue to the plugin maintainers" |
| `File.exists()` / file IO in `register{}` body | `@InputFiles` + `@SkipWhenEmpty` + `Provider<RegularFile>` chain | "always defer resolution to the execution phase by using lazy APIs" |
| `configurations.X.files` in config phase | `from(configurations.X)` (Copy spec accepts the configuration directly) | dependency resolution at config time penalises every build |
| Capturing `project` in a task action | Capture concrete value into local `val` first | configuration cache compatibility |
| `eachDependency{}` resolution | `dependencies { components.all { ... } } ` | metadata rules outlive resolution |

Mark each variant against this table. A variant that violates a principle without a stated reason is automatically demoted.

### 5. Rank and write the verdict

Output structure:

```
## Decision: <restatement>

## Variants
1. <variant name> — verdict (Recommended / Acceptable / Avoid)
2. ...

## Recommendation
<one paragraph: which one and why, in plain language>

## Evidence
- "<quote>" — <Gradle docs URL>
- "<quote>" — <Gradle docs URL>
- (AGP source pointer if relevant)

## Code
<the 5–15 line snippet for the recommended variant>

## What we considered and rejected
<one bullet per rejected variant, naming the principle it violated>
```

Keep the verdict short. The point is to surface the doc citations, not to write an essay.

### 6. Grill back if the user already picked one

If the user's prompt already states a preferred variant, **do not validate it without challenging**. Treat it as one of the variants and rank it honestly. If their choice ranks below another variant, lead with: "Your variant ranks #N — here's the principle it violates and the doc that flags it."

This is the part that makes the skill a "grill" rather than a "yes-man".

## Output discipline

- Variants must be distinct mechanisms, not formatting differences.
- Every claim must cite either a Gradle doc URL or an AGP source pointer. No memory-only assertions.
- If the docs are silent on a variant, say so explicitly — don't fabricate a citation.
- Recommendation must include a code sketch the user can paste.
- Russian or English follows the user's language. Quotes from docs stay in their original (English) form.

## Out of scope

- Performance benchmarking (use `benchmark-trace-pipeline`).
- Migration plans across major Gradle versions (use `release-management` framing or write a plan via `planning:make`).
- Reviewing existing branch diffs (use `code-review` / `pr-review-toolkit:review-pr`).
- Searching for symbols/usages (use `curiosity`, `Grep`, `gradle-sources`, `agp-sources`).

## Tools used

- `gradle-rag` — primary doc lookup (lexical search of current Gradle userguide)
- `agp-sources` — AGP class/method lookup (versioned)
- `gradle-sources` — Gradle internals (when behaviour, not just API, is in question)
- `kotlin-sources` — when Kotlin compiler/Gradle plugin behaviour matters (KSP, KAPT, kotlin-gradle-plugin)
- `ksp-sources` — when KSP API, Gradle plugin, or KSP1/KSP2 implementation details matter
- `Read`, `Grep`, `Bash` — only when the surrounding repo's existing pattern needs to be inspected before recommending
