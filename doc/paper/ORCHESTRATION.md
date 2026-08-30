# Orchestration report

Generated while assembling `00-abstract.typ` and `09-conclusion.typ`.
Word counts are raw whitespace-separated tokens of the `.typ` source (markup
included), measured on the files as they stand now.

## Per-section inventory

| File | Words | Bib keys cited |
|---|---:|---|
| `00-abstract.typ` | 224 | *(none — by design)* |
| `01-introduction.typ` | 1399 | `antoniol2002`, `deLucia2007`, `gotel2012quest`, `hayes2006`, `maeder2012`, `murphy2001`, `panickssery2024`, `perry2023`, `rempel2017`, `stulova2020` |
| `02-background.typ` | 1815 | `antoniol2002`, `antoniol2025`, `archunit`, `burdy2003entcs`, `burdy2005`, `clelandhuang2012`, `codebeamer`, `deLima2020`, `deLucia2007`, `deSilva2015`, `doorsnext`, `goanalysis`, `gobra`, `gotel1994`, `gotel2012fundamentals`, `gotel2012grand`, `gotel2012quest`, `govet`, `hayes2006`, `herold2014`, `jackson2002`, `jama`, `javaap`, `leavens1999`, `leavens2006`, `liu2018`, `maeder2012`, `maeder2015`, `menezes2021`, `meyer1992`, `murphy1995`, `murphy2001`, `oslc`, `ozkaya2023`, `polarion`, `ramesh2001`, `rempel2017`, `reqif`, `staticcheck`, `stulova2020` |
| `03-design.typ` | 2977 | *(none)* |
| `04-evidence.typ` | 2060 | `panickssery2024`, `perry2023` |
| `05-standards.typ` | 2390 | `aspice40`, `aspiceweb`, `do178c`, `do330`, `do333`, `ed12c`, `iec61508-1`, `iec61508-3`, `iec62304`, `iso26262-1`, `iso26262-6`, `iso26262-8`, `isoiec33002`, `isoiec33004` |
| `06-delimitation.typ` | 1440 | `jama`, `kiro`, `panickssery2024`, `speckit`, `tielke2026` |
| `07-implementation.typ` | 1255 | *(none)* |
| `08-discussion.typ` | 1917 | `antoniol2002`, `aspice40`, `deLucia2007`, `govet`, `hayes2006`, `maeder2012`, `panickssery2024`, `rempel2017`, `tielke2026` |
| `09-conclusion.typ` | 689 | `maeder2012`, `rempel2017` |

No key used anywhere is outside `refs.bib`/`CITATION-GUIDE.md`.

---

## Cross-section problems found (not fixed)

### A. Contradictions

**A1 — "there is an evaluation" vs "there is no evaluation".**
`06-delimitation.typ:115`: *"The evaluation presented in this paper is likewise
not a controlled experiment; it reports the behaviour of one tool on the projects
it was applied to, with the constraints set out in the discussion of
limitations."*
`08-discussion.typ:131`: *"There is no controlled study, no comparison against a
baseline, no multi-project dataset, and no measurement of anything."*
`01-introduction.typ:127`: *"the paper presents no controlled empirical
evaluation"*. §6 asserts an evaluation exists; §1 and §8 assert none does.

**A2 — build failure threshold vs the reported figure pair.**
`06-delimitation.typ:52`: *"`speclink verify` is a build step that runs after the
code exists and fails with a non-zero exit code if any of its four directions is
short of a hundred per cent"*.
`04-evidence.typ:171`: *"The summary therefore reports a pair, `100% verified,
88% demonstrated`, and the gap is the interesting number."*
`08-discussion.typ:32` repeats the same 88% pair as a normal outcome. A pair with
88% in it cannot coexist with a rule that any direction below 100% fails the
build, unless §6's "four directions" excludes the demonstrated figure — which is
nowhere stated.

**A3 — enumeration of language-independent rule families.**
`03-design.typ:311`: *"`K1`, `K3` and `K10` to `K14`, together with the
requirement-tree checks, are universal, while `K4` to `K8` belong to a profile's
style."*
`07-implementation.typ:34-43` places `K15`, `K16`, `K17`, `K18` and `K20` in
`internal/check` as well (i.e. also language-independent), and `07:159` puts
`K19` in `internal/reqtree/topic.go`. §3's list of universal families is
incomplete relative to §7.

**A4 — reference-ERP figures attributed to §7's table.**
`08-discussion.typ:154`: *"the reference-ERP figures quoted in §7 are
measurements made by the tool's authors on their own project with their own
tool"*. `@tbl:measured` in §7 measures the *speclink repository*, not the
reference ERP; the only ERP figures in §7 are `193 / 45 / 9` in the
"Constructor naming" blocker (`07:178-181`), which were not produced by
`@tbl:measured`.

**A5 — a disclosure section that does not exist.**
`08-discussion.typ:194`: *"this paper was largely produced by a language model
under human direction; the disclosure is made separately"*. There is no such
separate disclosure among the section fragments.

### B. Overlap / verbatim repetition

**B1 — `deLucia2007` quotation stated three times, verbatim.**
`01:38`, `02:66`, `08:106`, each rendering *"are still far to support a complete
semi-automatic recovery of all links"*.

**B2 — `panickssery2024` self-preference sentence stated four times.**
`01:21-24`, `04:217-220`, `06:67-69`, `08:70-73`. Three of the four also add the
same gloss ("routing the certification to another model is not an escape").

**B3 — `maeder2012` "21% faster / 60% more correct" stated three times.**
`01:12-14`, `02:37-39`, `08:140-142`.

**B4 — `antoniol2002` + `hayes2006` recovery/candidate-link pairing stated three
times.** `01:32-35`, `02:54-62`, `08:101-105`.

**B5 — the `100% verified, 88% demonstrated` figure appears three times.**
`01:97`, `04:171`, `08:32`.

**B6 — the README §10b quotation appears three times.**
`05:261`, `08:19`, and paraphrased at `01:98`: *"It has never said the code does
what the requirement asks"*.

**B7 — `murphy2001` and `stulova2020` quotations appear in both §1 and §2.**
`01:43-47` and `02:117`, `02:152-154`.

**B8 — `spec.Waive` / "only escape hatch, reason mandatory" is restated in four
sections.** `01:83`, `03:279`, `05:107`, `08:116`.

### C. Terminology inconsistency

**C1 — what "four" counts.** `04:13`: *"speclink's chain has four links"*;
`06:52`: *"any of its four directions"*; `08:8`: *"a conjunction of four
statements"*; `05:262`: *"That is why the fourth figure exists"*. Four different
nouns for what may or may not be the same enumeration; §5's "fourth figure" is
never defined in any section.

**C2 — diagnostic code rendering.** `03:301` states the code format is
`SPEC-<phase>-<number>` and the example at `03:290` uses `[SPEC-V6-056]`, but the
same section writes `V5-013`, `V5-032`, `V5-033`, `V5-035` and `SPEC-V1-001`
without a uniform prefix, and `07:149` writes *"`V6-090`, `V6-091` and `V6-092`"*
prefix-free.

**C3 — construct enumeration order.** STYLE.md fixes the order "use case,
command, event, aggregate, projection, repository, permission, query".
`01:72-73` follows it. `03:12-13` does not: *"use cases, commands, events,
aggregates, permissions, queries, projections and repositories"*.

**C4 — which command records a reviewer.** `01:81` names both
*"`speclink attest` and `speclink freeze -reviewer`"*; `04:208-210` shows
`speclink attest -reviewer "TS" SubmitQuote` **and**
`speclink freeze -reviewer "Frau Meier" ./...`; `08:57` names only
*"`speclink freeze -reviewer`"*. It is left unclear whether these are two
mechanisms or one.

**C5 — rule families named in §5 that are named nowhere else.**
`05` table row: *"[Go compiler, then `K11`–`K13`]"*. `K11` and `K12` appear
elsewhere only as `K12-SOURCE-UNCOVERED` (`07:39`, `08:48`); `K11` is never
expanded anywhere in the paper.

**C6 — "bound" / "covered" as figure names.** `02:76` refers to *"the
'bound'/'covered' figures of §2"*; §3, §4 and §8 use "coverage", "verified" and
"demonstrated" and never re-introduce "bound".

**C7 — JVM front end scope.** `01:71-72`: *"a front end for JVM bytecode that
covers Java and Kotlin"*. `03:319`: *"It also collapses three targets into one,
since Kotlin compiles to JVM bytecode and is only dexed afterwards."* The third
target (Android/dex) is never named in §1, and §7 lists only one JVM profile,
`java_springboot_ddd1`.

**C8 — causal strength of `rempel2017` differs between sections.**
`01:15-17`: *"more complete traceability is associated with a decreased expected
defect rate"* (hedged). `02:42-44`: *"more complete traceability decreases the
expected defect rate"* (unhedged). `08:144-146` uses a third form, *"relating
traceability completeness to defect rate"*. CITATION-GUIDE.md forbids *"a causal
claim beyond the authors' wording"*, so §2's form should be checked.

**C9 — cross-reference style.** `08` mixes *"Section 6 criticises"* (`08:150`)
with *"§5 says so"* (`08:160`), *"(§7)"* (`08:190`) and *"quoted in §7"*
(`08:154`); no Typst label references (`@sec:...`) are used anywhere, so section
numbers are hard-coded prose.

### D. Structural observations

**D1 —** `03-design.typ` and `07-implementation.typ` contain no citations at all.
That is defensible (both are about the artefact), but §3 makes one outside-world
claim without support: `03:220-221`, *"whether every fork is matched by exactly
one join on every path is reachability in a Petri net"*.

**D2 —** §7 records that README §12 is stale regarding the Typst backend
(`07:136-140`), while §8 still cites README §12 as authoritative for the absence
of `speclink selfreport` (`08:174`). Not a contradiction, but the two uses of
§12 have different reliability and a reader may notice.
