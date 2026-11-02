# agents-gradle

**agents-gradle** is a plugin marketplace for Claude Code and Codex. It gives your coding agent a local search over the official Gradle docs, the source of AGP (Android Gradle Plugin), Gradle, Kotlin and KSP (Kotlin Symbol Processing), and `gradle-grill`, a review that makes the agent defend a build choice with citations that a second agent re-checks blind.

Gradle has accumulated several generations of APIs, so the answer an agent remembers is often the obsolete one.

<sub>Independent project, not affiliated with or endorsed by Gradle, Inc.</sub>

## What's inside

| Plugin | What your agent gets | Needs |
| --- | --- | --- |
| [`gradle-grill`](plugins/gradle-grill/skills/gradle-grill/SKILL.md) | A ranked verdict on a Gradle or AGP choice, with verified citations | The five below |
| [`gradle-rag`](plugins/gradle-rag/skills/gradle-rag/SKILL.md) | Search over a local snapshot of docs.gradle.org | The `gradle-rag` binary |
| [`agp-sources`](plugins/agp-sources/skills/agp-sources/SKILL.md) | AGP source for each mirrored release, from the third-party [jrodbx/agp-sources](https://github.com/jrodbx/agp-sources) | `git`, about 4 GB of disk |
| [`gradle-sources`](plugins/gradle-sources/skills/gradle-sources/SKILL.md) | Gradle Build Tool source | `git` |
| [`kotlin-sources`](plugins/kotlin-sources/skills/kotlin-sources/SKILL.md) | Kotlin compiler, standard library and plugin source | `git`, `python3` |
| [`ksp-sources`](plugins/ksp-sources/skills/ksp-sources/SKILL.md) | KSP source | `git` |

`gradle-rag` indexes docs.gradle.org only: User Manual, release notes, Groovy DSL, Kotlin DSL and Java API. It does not cover Android's documentation, and no plugin here does: AGP API answers rest on `agp-sources`, plus any developer.android.com page the agent fetches with its own web access. On 2026-11-02 the `agp-sources` mirror held AGP 8.0.0 to 9.1.0 without 8.13.2, and had nothing newer.

## How `gradle-grill` works

A *variant* here is one candidate approach, not an Android build variant. A single context that invents the variants, gathers their citations and ranks them grades its own homework, so the evidence work goes to separate subagents:

- **Orchestrator:** your main session. It frames the decision, lists the variants and ranks them, first against the skill's built-in table of avoid/prefer rules (the principle table), then on the evidence.
- **Advocate:** one subagent per variant. It gathers quotes for that variant from docs and source, and is told to report any counter-evidence it finds.
- **Verifier:** one subagent per variant. It gets the claims but not the advocate's argument, and re-reads each source. `CONFIRMED`: the quote is there and means what the advocate says. `PLAUSIBLE`: the quote is there but the reading stretches it, or the docs are silent. `REFUTED`: the quote is not there or is out of context, or the docs say the opposite.

`CONFIRMED` judges the quote, not the variant, and the skill tells the orchestrator to keep `REFUTED` quotes out of its verdict's Evidence section.

| Level | What runs | Subagents |
| --- | --- | ---: |
| `low` | Checks your approach against a principle table, at most one docs search | 0 |
| `medium` | 2 to 4 variants, one advocate each, blind verify, rank | 4 to 8 |
| `high` | 3 to 5 variants, including mechanisms you did not ask about | 6 to 10 |
| `xhigh`, `max` | As `high`, plus a sweep agent hunting for a missed mechanism, at most 6 variants | 7 to 13 |

With no level given it reuses the last one you typed in the session, otherwise `medium`.

Start with `low` for a quick check; it has no advocates and no verifier. Without a subagent tool the skill runs every role in one context, and the check is no longer blind.

## Install

**1. Build the `gradle-rag` binary.** The `gradle-rag` and `gradle-grill` plugins call it at runtime. If you only want source plugins, skip to step 2. This takes macOS or Linux, Go 1.25+, `make` and `bash`; on any other system `make install` fails at the end, after the crawl. There are no prebuilt binaries.

```bash
git clone https://github.com/materkey/agents-gradle.git
cd agents-gradle
make install
gradle-rag info
```

`make install` crawls about 11,000 pages of docs.gradle.org, embeds them in a single binary and copies it to `${GRADLE_RAG_INSTALL_DIR:-$HOME/.local/bin}`. The crawl prints nothing until it ends (about 6 minutes on the last measured run), and a few `HTTP 404` warnings at the end are normal. If that directory is not on `PATH`, the installer prints the line to add. Keep the clone: running `make install` again is how you refresh the docs, and it crawls again each time.

**2. Add the plugins.**

Claude Code. Installing `gradle-grill` also installs the five plugins it depends on:

```bash
claude plugin marketplace add materkey/agents-gradle
claude plugin install gradle-grill@agents-gradle
```

To take a single plugin instead, install it by name, for example `claude plugin install agp-sources@agents-gradle`. For updates on startup, open `/plugin`, go to the **Marketplaces** tab, select `agents-gradle` and choose **Enable auto-update**.

Codex has no plugin dependencies, so each plugin is added by name. Drop any you do not want, but `gradle-grill` needs the other five:

```bash
codex plugin marketplace add materkey/agents-gradle
for p in gradle-rag agp-sources gradle-sources kotlin-sources ksp-sources gradle-grill; do
  codex plugin add "$p@agents-gradle"
done
```

Codex updates plugins from Git marketplaces on its own. In both clients, plugin updates change only skills and scripts. The docs snapshot changes only when you re-run `make install`.

**3. Use it.** Ask a Gradle question in plain words and the agent can pick the skill up by itself, or call the review directly:

```text
# Claude Code
/gradle-grill:gradle-grill medium Where should my convention plugin wire generated sources?

# Codex
$gradle-grill:gradle-grill medium Where should my convention plugin wire generated sources?
```

On first use each source plugin clones its repository, so the first AGP question also writes about 4 GB to disk, which took about a minute when measured. A Codex sandbox blocks that clone, so run the scripts under **Source plugins** below first.

## Reference

<details>
<summary><b><code>gradle-rag</code> CLI</b></summary>

<br>

```bash
gradle-rag search "TaskProvider register" --limit 5
gradle-rag search "configuration cache requirements" --full
gradle-rag page "https://docs.gradle.org/current/userguide/configuration_cache_requirements.html"
gradle-rag info
```

- Search is lexical (SQLite FTS5, a full-text index). Despite the name, there are no embeddings, no LLM call and no API key, and it works offline. Only the crawl needs the network.
- A hit needs any term, not all of them, and the default limit is 10. Broad queries often return API reference pages or duplicate sections from the single-page User Manual, so phrase queries narrowly.
- `page` takes the exact URL from a `Source:` line, or an exact section heading, and prints one section.
- There is no `--help`: it would search for the word "help". Run `gradle-rag` with no arguments to see usage.
- Use `plugins/gradle-rag/skills/gradle-rag/bin/gradle-rag` from a checkout when the binary is not on `PATH`. This wrapper tries `GRADLE_RAG_BIN`, then `PATH`, then the local build. Agents prefer a bare `gradle-rag` on `PATH`, so set `GRADLE_RAG_BIN` only when no other copy is there.

**Which docs am I searching?** `gradle-rag info` reports the exact snapshot, and that is the evidence for any citation. `Source:` URLs always point at `/current/`, so a cited page can be newer than your snapshot.

```console
$ gradle-rag info
gradle-rag version: v0.2026.0819
gradle_version: 9.7.0
docs_source: https://docs.gradle.org/current/userguide/userguide.html
scope_path: /current/
crawled_at: 2026-08-19T13:17:01Z
page_count: 11019
scheduled_count: 11052
chunk_count: 49618
```

</details>

<details>
<summary><b>Source plugins: where they fetch from and how to pick a version</b></summary>

<br>

Outside a Codex sandbox you can skip this if you only use the plugins: each skill tells the agent to fetch its sources on first use. For a specific project, match its versions: Gradle from `gradle/wrapper/gradle-wrapper.properties`, AGP, Kotlin and KSP from the build files, settings or version catalog.

| Plugin | Fetches | Default directory | Override | Pick a version |
| --- | --- | --- | --- | --- |
| `agp-sources` | `jrodbx/agp-sources` | `~/projects/agp-sources` | `AGP_SOURCES_DIR` | Open `<dir>/<version>/`; there is no version argument |
| `gradle-sources` | `gradle/gradle` | `~/.gradle-sources/gradle` | `GRADLE_SOURCES_DIR` (parent directory) | `clone_gradle.sh v9.3.0` |
| `kotlin-sources` | `JetBrains/kotlin` | `~/.kotlin-sources/kotlin` | `KOTLIN_SOURCES_DIR` | `wrapper.py init --ref <full commit SHA>` |
| `ksp-sources` | `google/ksp` | `~/.ksp-sources/ksp` | `KSP_SOURCES_DIR` | `explore.sh init 2.3.13` |

The same scripts, run by hand from a checkout of this repository:

```bash
plugins/agp-sources/skills/agp-sources/scripts/clone_agp_sources.sh
plugins/gradle-sources/skills/gradle-sources/scripts/clone_gradle.sh v9.3.0
plugins/kotlin-sources/skills/kotlin-sources/scripts/wrapper.py init \
  --ref f95cb2f7693a15966fd8fd9b1d78d440545e7b66
plugins/ksp-sources/skills/ksp-sources/scripts/explore.sh init 2.3.13
```

That Kotlin SHA is tag `v2.3.0`. With no version, you get the tip of the AGP mirror (every mirrored release, one directory each), Gradle `master`, and on a fresh clone Kotlin `master` and KSP `main`.

- `kotlin-sources` reuses an existing checkout if it finds one: the current directory, then `~/projects/kotlin`, `~/dev/kotlin`, `~/kotlin`. It leaves a reused checkout untouched and ignores `--ref`. A checkout already at `KOTLIN_SOURCES_DIR` (default `~/.kotlin-sources/kotlin`) is different: `--ref` fetches that commit and moves its working tree, and it is preferred over the home-directory fallbacks. If `KOTLIN_SOURCES_DIR` does not exist yet and a fallback holds a checkout, the fallback wins and nothing is cloned. Tag and branch names currently fail on `--ref`, so pass a full commit SHA, for example from `git ls-remote https://github.com/JetBrains/kotlin.git refs/tags/v2.3.0`.
- `gradle-sources` and `ksp-sources` move the working tree of their clone to the version you ask for. Re-running `explore.sh init` with no version fails with `Not possible to fast-forward` once KSP's `main` has moved; pass a tag instead.
- In a Codex sandbox the source plugins cannot clone on first use: the network is off by default and their default directories are outside the writable workspace, so allowing network access alone still fails with `Operation not permitted`. Run the scripts above beforehand, outside the sandbox. To clone from inside it instead, allow network access, create the directory that will contain the clone (`~/.ksp-sources`, `~/.gradle-sources`, `~/.kotlin-sources`, or `~/projects` for AGP) and add it to `sandbox_workspace_write.writable_roots`. Adding the clone directory itself fails, because Codex keeps `.git` read-only at the top of a writable root. In `read-only` mode `gradle-rag` fails too, because it copies its index to a temp directory. Checked with codex-cli 0.166.0 on macOS.

</details>

<details>
<summary><b>Development</b></summary>

<br>

```bash
make crawl-docs-sample
make build-cli
./plugins/gradle-rag/skills/gradle-rag/references/gradle-rag info
make test
```

- The sample crawl takes 80 pages, seconds instead of minutes, and `info` should print `page_count: 80`. `make crawl-docs-sample` replaces `cmd/gradle-rag/db/gradle.db`, and the next `make build-cli` embeds only those 80 pages. Getting the full index back takes a full crawl. `make build` is that crawl plus the binary; `make install` also installs it.
- `cmd/crawler` follows same-host links under `/current/` from the User Manual and builds the index. `cmd/gradle-rag` embeds it in the binary. Run the crawler through `make`, because its own default path is not the one that gets embedded.
- The generated index and the built binary are deliberately not committed. Plugin manifests under `plugins/` deliberately omit `version`.

</details>

## Licenses

The code in this repository is under the [MIT License](LICENSE). It does not redistribute the generated documentation index, and the source plugins fetch their sources at runtime.

Gradle's User Manual and DSL Reference are licensed by Gradle under the Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International License; see [Gradle's license page](https://docs.gradle.org/current/userguide/licenses.html). If you distribute a generated index or a binary that embeds Gradle documentation, handle that artifact under Gradle's documentation license terms, including attribution, non-commercial use and share-alike requirements.
