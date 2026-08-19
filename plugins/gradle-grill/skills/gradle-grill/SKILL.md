---
name: gradle-grill
description: Challenge a Gradle/AGP implementation choice against the official docs. Generates candidate variants, has an independent advocate build the case for each, blind-verifies every citation, ranks the variants, and recommends the most idiomatic option with quoted citations. Accepts an effort level (low/medium/high/xhigh/max).
---

# Gradle Grill

Challenge a Gradle/AGP implementation choice. Generate variants, build the case
for each with an independent advocate, verify every citation blind, rank, and
recommend the most idiomatic one with quoted citations.

This skill exists because Gradle has accumulated several generations of APIs
(eager → lazy → configuration cache, AGP DSL → AndroidComponentsExtension), and
"the obvious answer" is frequently the obsolete one. The skill forces an
explicit doc-grounded comparison instead of pattern-matching from memory — and
separates generation from evaluation, because a single context that invents the
variants, gathers their citations, and ranks them grades its own homework.

## When to invoke

- The user proposes a Gradle implementation and asks "is this right?" or "is there a better way?"
- The user asks "where should I put X" — a hook callback, a task config, a precondition check, a DSL override.
- The user is about to write or edit a `Plugin<Project>`, convention plugin, custom task, or build-script `apply`/`register` block.
- A code review surfaces eager APIs (`tasks.create`, `tasks.getByName`, `afterEvaluate`, `File.exists()` in config, `configurations.X.files` in config).

Trigger phrases include: "как лучше в gradle", "правильный способ X в gradle", "идиоматично", "challenge gradle approach", "гриль gradle", "gradle-grill", "afterEvaluate vs", "tasks.register vs", "as a plugin author".

Do not invoke for pure code-search ("where is X used") — use `curiosity` or `Grep`. Do not invoke for trace analysis — use `perfetto-trace`.

## Effort level

Parse the effort level from the invocation arguments: `low`, `medium`, `high`,
`xhigh`, or `max`. With no level given, reuse the level the user typed last in
this session; if none (including non-interactive runs), use `medium`.

| Level | Scheme |
|---|---|
| `low` | no variants, no subagents → canonical-table check of the user's approach → verdict |
| `medium` | 2–4 variants → 1 advocate each → blind citation verify → rank |
| `high` | 3–5 variants (recall-biased: include the non-obvious mechanisms) → advocates → verify → rank |
| `xhigh` | as high → plus a sweep agent hunting for a missed mechanism (≤6 variants total) |
| `max` | as xhigh at maximum effort |

At `medium` you are grilling for **precision**: every variant must be a
mechanism a plugin author would seriously consider. At `high` and above you
are grilling for **recall**: the missed variant is usually the idiomatic one —
enumerate mechanisms the asker didn't know existed.

## Phase 0 — Restate the problem

Paraphrase the user's question into a single concrete decision. Example:
> "Where to put a fail-fast precondition that a module's `values/strings.xml` exists, when the i18n convention plugin is applied?"

If the question is too vague to ground in docs, ask exactly one clarifying
question and stop. Do not generate variants from a fuzzy premise.

If the user's prompt already states a preferred variant, it is always
**variant #1** and gets grilled like the rest — do not validate it without
challenging. This is what makes the skill a "grill" rather than a "yes-man".

## Low effort — canonical-table check

At `low`, take the user's stated approach (or the single obvious one), check
it row by row against the canonical principle table below, run at most one
`gradle-rag` search to confirm the one principle in doubt, and give the
verdict: fine as-is, or "swap X for Y" with the table's citation. No variant
fan-out, no subagents. If the question has no clear single approach, say the
comparison needs `medium` and stop.

## Phase 1 — Enumerate variants (medium and above)

Pick 2–4 distinct mechanisms (up to 5 at `high`+) — different mechanisms, not
minor stylistic variations. Lean on these axes:

| Axis | Common variants |
|---|---|
| Lifecycle hook | `apply{}` body, `afterEvaluate{}`, `AndroidComponentsExtension.finalizeDsl{}`, `onVariants{}`, `beforeVariants{}` |
| Task creation | `tasks.create(...)`, `tasks.register(...)`, `withType().configureEach{}` |
| Task lookup | `tasks.getByName(...)` / `findByName(...)` (eager), `tasks.named(...)` (lazy `TaskProvider`) |
| Validation point | configuration block of `register{}`, `doFirst{}`, `@InputFiles`+`@SkipWhenEmpty`, project `afterEvaluate`, AGP `finalizeDsl` |
| Inputs/outputs | raw `File`, `RegularFileProperty`/`DirectoryProperty`, `ConfigurableFileCollection`, `Provider<T>` chain |
| Property chains | direct `String`/`File`, `Property<T>`/`Provider<T>` with `.map{}`/`.flatMap{}` |

## Phase 2 — Advocates (one per variant, blind to competitors)

Run **one advocate agent per variant** via the Agent tool. Each advocate
receives the restated decision, its ONE variant, and the tool instructions
below — never the other variants; an advocate that sees the field hedges
toward the consensus instead of building the strongest honest case. If the
Agent tool is not available, do not error — run each advocate (and each
verification) yourself, sequentially, in this context. If a launched agent
never returns — a practical trigger: still pending after all its peers have
returned — run its role inline.

Each advocate must:

1. Write a 5–15 line code sketch — enough to be evaluated, not a full plugin.
2. Run **at least two** `gradle-rag` searches with different angles, queries
   phrased as nouns:

```bash
gradle-rag search "afterEvaluate restrictions plugin author" --limit 4
gradle-rag search "tasks register lazy configuration block" --limit 4
```

   Before spawning anyone, you (the orchestrator) resolve the working
   `gradle-rag` command — the bare binary on `PATH`, or the absolute path to
   the `gradle-rag` skill's `bin/gradle-rag` wrapper — and pass that exact
   command into every advocate and verifier prompt; a fresh subagent cannot
   locate the wrapper on its own. Do not treat a missing `PATH` entry as
   missing documentation.
3. When the topic is AGP-specific (`AndroidComponentsExtension`, `Variant`,
   `finalizeDsl`, `onVariants`, source sets), additionally check `agp-sources`;
   Gradle internals via `gradle-sources` when behaviour, not just API, is in
   question; KSP topics via `ksp-sources`; Kotlin plugin behaviour via
   `kotlin-sources`.
4. Return: the sketch; for each claim one direct quote (≤30 words) with its
   source URL (Gradle userguide section or AGP javadoc/source pointer) and a
   one-line interpretation ("this means …"); and — non-negotiable — any
   counter-evidence found: a doc warning against the variant is part of an
   honest case, not something to omit. Empty doc results are a signal: state
   explicitly that docs do not address this variant directly.

## Phase 3 — Verify citations (blind, 1-vote, 3-state)

For each variant, run **one verifier** via the Agent tool — one verifier per
variant, its verdict is final (if an inline fallback for a lost verifier is
underway when the agent returns, the returned agent's verdict wins and the
inline duplicate is discarded). The verifier receives the variant, its sketch,
and its claims — each claim's quote, URL, and one-line interpretation, since
the interpretation is what CONFIRMED judges — but never the advocate's
narrative case or the other variants. Counter-evidence quotes are claims too
and go through the same verification. It re-runs the searches (or fetches the
cited sections) and returns, per claim:

- **CONFIRMED** — the quote exists at the cited source, in context, and the
  interpretation holds.
- **PLAUSIBLE** — the quote exists but the interpretation stretches it, or the
  docs are genuinely silent and the claim rests on API signatures. State what
  the docs actually establish.
- **REFUTED** — the quote does not exist at the source, is out of context, or
  the docs say the opposite. Quote the evidence.

A variant with a REFUTED core claim is not auto-eliminated — it re-enters
ranking marked "docs do not support the advocate's case", which in practice
demotes it. Never present a REFUTED quote in the final Evidence section. Never
anticipate a pending verifier's verdict — a verdict exists only once returned.

## Sweep (xhigh and max)

Run **one more agent** as a fresh reviewer who has the variant list and the
restated decision. Its only job is gaps: name a distinct mechanism NOT on the
list that a Gradle/AGP plugin author would consider (search the axes table's
blind spots — newer AGP Variant APIs, `ValueSource`, shared build services,
artifact transforms). If it finds one, run it through Phases 2–3. If nothing
new, an empty sweep is the correct answer — do not pad.

## Phase 4 — Rank and write the verdict

Ranking is yours (the orchestrator's) — no advocate ranks, no advocate sees
the comparison. Apply, in order:

1. **The canonical principle table.** These are non-negotiable principles from
   Gradle's own best-practices guide and override "I've seen it done this way
   before". A variant that violates a row without a stated reason is
   automatically demoted:

| Avoid | Prefer | Why (cite) |
|---|---|---|
| `tasks.create()` | `tasks.register()` | task configuration avoidance — task body runs only when realized |
| `tasks.getByName{}` / `findByName{}` | `tasks.named{}` | returns `TaskProvider`, stays lazy |
| `tasks.withType(T){ … }` | `tasks.withType(T).configureEach{ … }` | closure makes withType eager |
| `someTask { }` (Groovy sugar) | `tasks.named("someTask") { }` | hidden eager realization (docs name this exact pitfall) |
| `afterEvaluate{}` (in plugin code) | `AndroidComponentsExtension.finalizeDsl{}` / `onVariants{}` (AGP) or `Provider<T>` chain | AGP docs: "use specially made extension points instead of registering the typical Gradle lifecycle callbacks (such as afterEvaluate())" (developer.android.com/build/extend-agp) |
| `File.exists()` / file IO in `register{}` body | `@get:InputFile` `RegularFileProperty` set from a `Provider<RegularFile>` chain — Gradle's built-in input validation fails on a missing file before the action runs | "always defer resolution to the execution phase by using lazy APIs"; note `@InputFiles` + `@SkipWhenEmpty` is the opposite tool — it turns a missing input into a silent NO_SOURCE skip, wrong for a precondition |
| `configurations.X.files` in config phase | `from(configurations.X)` (Copy spec accepts the configuration directly) | dependency resolution at config time penalises every build |
| Capturing `project` in a task action | Capture concrete value into local `val` first | configuration cache compatibility |
| `eachDependency{}` resolution | `dependencies { components.all { ... } } ` | metadata rules outlive resolution |

2. **Verified evidence.** CONFIRMED citations outweigh PLAUSIBLE; a variant
   whose case is REFUTED ranks on the principle table alone.
3. **Config-cache and laziness posture** as the tie-breaker.

Output structure:

```
## Decision: <restatement>   (effort level and scheme on the next line)

## Variants
1. <variant name> — verdict (Recommended / Acceptable / Avoid), citations: N CONFIRMED / M PLAUSIBLE / K REFUTED
2. ...

## Recommendation
<one paragraph: which one and why, in plain language>

## Evidence
- "<quote>" — <Gradle docs URL>   (verified citations only)
- (AGP source pointer if relevant)

## Code
<the 5–15 line snippet for the recommended variant>

## What we considered and rejected
<one bullet per rejected variant, naming the principle it violated or the verdict its evidence got>
```

Keep the verdict short. The point is to surface the verified citations, not to
write an essay. If the user's variant (always #1) ranks below another, lead
with: "Your variant ranks #N — here's the principle it violates and the doc
that flags it."

## Output discipline

- Variants must be distinct mechanisms, not formatting differences.
- Every claim must cite either a Gradle doc URL or an AGP source pointer, and
  survive Phase 3 — no memory-only assertions, no unverified quotes in the
  final Evidence.
- If the docs are silent on a variant, say so explicitly — don't fabricate a citation.
- Recommendation must include a code sketch the user can paste.
- Russian or English follows the user's language. Quotes from docs stay in their original (English) form.

## Out of scope

- Performance benchmarking (use a trace/benchmark skill such as `perfetto-trace`).
- Migration plans across major Gradle versions (write a plan via a planning skill).
- Reviewing existing branch diffs (use `code-review`).
- Searching for symbols/usages (use `curiosity`, `Grep`, `gradle-sources`, `agp-sources`).

## Tools used

- `gradle-rag` — primary doc lookup (lexical search of current Gradle userguide)
- `agp-sources` — AGP class/method lookup (versioned)
- `gradle-sources` — Gradle internals (when behaviour, not just API, is in question)
- `kotlin-sources` — when Kotlin compiler/Gradle plugin behaviour matters (KSP, KAPT, kotlin-gradle-plugin)
- `ksp-sources` — when KSP API, Gradle plugin, or KSP1/KSP2 implementation details matter
- `Read`, `Grep`, `Bash` — only when the surrounding repo's existing pattern needs to be inspected before recommending
