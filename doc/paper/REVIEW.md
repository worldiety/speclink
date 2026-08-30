# Review pass report

Review agent. No section of this paper was written by the reviewer. The pass
resolved cross-section contradictions, unified terminology, cut verbatim
repetition and removed unsupported claims. `typst compile main.typ` succeeds
with **0 errors and 0 warnings**.

Every defect listed in `ORCHESTRATION.md` was re-verified against the repository
before being acted on. Two of its findings were wrong and are recorded below.

---

## Facts verified against the repository before editing

| Question | Source consulted | Finding |
|---|---|---|
| What are the "four directions"? | `README.md` §2 | *accounted*, *bound*, *covered*, *verified*. "Four numbers, four directions, all of which must reach 100%." `demonstrated` is a **fifth** figure, not one of the four. |
| Is `88% demonstrated` a failing state? | `README.md` §10b | It is reported as a figure; the associated finding is `K14-VERIFICATION-STALE`. The 100%-threshold rule applies to the four directions, not to `demonstrated`. |
| Which rule families live where? | `internal/check`, `internal/lang/golang`, `internal/reqtree` (grepped rule IDs) | `internal/check`: K1, K3, K9, K10, K12–K18, K20. `internal/reqtree`: **K11, K19**. `internal/lang/golang`: K4–K8. |
| Is `K11` real? | `internal/reqtree` | Yes: `K11-REQ-UNSOURCED`, `K11-SOURCE-UNANCHORED` (README §11, codes `V5-020`/`V5-026`). **ORCHESTRATION C5 was wrong** to imply K11 does not exist. |
| `attest -reviewer` vs `freeze -reviewer` | `README.md` §2 and §10a, `cmd/speclink/attest.go`, `cmd/speclink/freeze.go` | Two distinct records. `attest -reviewer` targets **declarations** (a construct by name or a package pattern) and records that the named person read them. `freeze -reviewer` records that the named person read the **requirement wordings** as they then stood, bound to those wordings; rewriting a text discards the review. |
| JVM front-end scope | `README.md` §11a, `internal/lang/jvm/` | Reads compiled classes. Collapses three targets — Java on Spring, Kotlin on Spring, Kotlin on Android — into one, because Kotlin compiles to JVM bytecode. Exactly **one** JVM profile exists: `java_springboot_ddd1`. §1, §3 and §7 were all partially right and mutually inconsistent. |
| Construct enumeration order | `README.md` §1 rule 4 | use cases, commands, events, aggregates, permissions, queries, projections, repositories. **Note: `STYLE.md` gives a different order** (see "remaining concerns"). |
| Does `speclink selfreport` exist? | `grep -rn selfreport cmd/ internal/` | No match. §8's claim of absence is correct. |

---

## A. Contradictions fixed

**A1 — evaluation vs no evaluation.** `06-delimitation.typ` claimed "The
evaluation presented in this paper …". Rewritten to state that the paper
presents no controlled experiment and no empirical evaluation at all, pointing
at `@sec:discussion`. §1, §6 and §8 now agree.

**A2 — build-failure threshold vs `88% demonstrated`.** Reconciled across three
sections using README §2 + §10b:
- `04-evidence.typ` now defines the four directions once and states explicitly
  that `demonstrated` is a fifth figure, "a figure rather than a threshold", with
  `K14-VERIFICATION-STALE` as the accompanying finding.
- `06-delimitation.typ` no longer says a run "fails if any of its four
  directions is short of a hundred per cent"; it says `verify` exits non-zero on
  any finding, and that the four directions must reach 100% for a clean run.
- `08-discussion.typ` now refers to "the verified/demonstrated pair of
  @sec:evidence" instead of restating the numbers as a normal outcome.

**A3 — universal vs style rule families.** `03-design.typ`'s list (`K1`, `K3`,
`K10`–`K14`) was incomplete and placed the requirement-tree checks vaguely. It
now states the seam as measured in the code: `internal/check` holds K1, K3, K9,
K10, K12–K18, K20; `internal/reqtree` holds K11 and K19; only K4–K8 are style
rules and they live in `internal/lang/golang`. `07-implementation.typ` was
edited to name K11 alongside K19 as the two requirement-tree families, so the two
sections now assert the same facts.

**A4 — reference-ERP figures attributed to `@tbl:measured`.** `08-discussion.typ`
now says "every figure in @sec:implementation, including the constructor-naming
counts from the reference ERP", which is true of both `@tbl:measured` (the
speclink repository) and the `193 / 45 / 9` blocker figures, without claiming the
table produced the ERP numbers.

**A5 — missing disclosure section.** The forward reference now points at
`@sec:disclosure` (`sections/10-ai-disclosure.typ`, "Disclosure of AI
Involvement") and summarises what that section contains.

---

## B. Redundancy cut

Full statement kept exactly once; every other occurrence reduced to a clause plus
a label reference. Verified by grep after editing.

| Repeated item | Kept in full | Shortened in |
|---|---|---|
| `deLucia2007` quotation ("still far to support …") | `02-background.typ` (as instructed) | `01`, `08` |
| `panickssery2024` self-preference statement | `01-introduction.typ` (first and load-bearing occurrence) | `04`, `06`, `08` |
| `maeder2012` 21% / 60% | `01-introduction.typ` | `02`, `08` |
| `antoniol2002` + `hayes2006` pairing | `02-background.typ` | `01`, `08` |
| `100% verified, 88% demonstrated` | `04-evidence.typ` | `01`, `08` |
| README §10b quotation ("It has never said the code does what the requirement asks") | `08-discussion.typ` | `01` (removed), `05` (paraphrase + pointer) |
| `spec.Waive` "only escape hatch, reason mandatory" | `03-design.typ` | `01`, `05`, `08` (the *analysis* in §8 is preserved; only the restatement of the rule was cut) |

Additionally, the `murphy2001` and `stulova2020` quotations were duplicated
between §1 and §2 (ORCHESTRATION B7). The verbatim quotations are now in §2 only;
§1 keeps a one-clause pointer with the same two citation keys.

---

## C. Terminology unified

**C1 — one phrase for "four".** Chosen: **"four directions"**, matching README §2.
Defined once, at first substantive appearance, in `04-evidence.typ`. Changed:
- `04`: "speclink's chain has four links" → definition of the four directions,
  then "the chain those directions run along has four links".
- `06`: "any of its four directions" → "the four directions defined in
  @sec:evidence".
- `08`: "a conjunction of four statements" → "a conjunction of the four
  directions (@sec:evidence)".
- `05`: "That is why the fourth figure exists" → "That is why the *verified*
  direction exists at all" (the undefined term is gone).

**C2 — diagnostic codes.** Normalised to `SPEC-<phase>-<number>`:
`SPEC-V5-013`, `SPEC-V5-032`, `SPEC-V5-033`, `SPEC-V5-035` in `03-design.typ`;
`SPEC-V6-090`, `SPEC-V6-091`, `SPEC-V6-092` in `07-implementation.typ`.
K-family names (`K10-REQ-CHANGED` etc.) left untouched, as required.

**C3 — construct order.** `01-introduction.typ` reordered to the README §1 order
(use cases, commands, events, aggregates, permissions, queries, projections,
repositories), matching `03-design.typ`.

**C4 — reviewer recording.** A new passage in `04-evidence.typ`, at the point
where reviewer recording is first discussed, states the distinction verified
above (`attest -reviewer` = declarations read; `freeze -reviewer` = requirement
wordings read), cites README §2/§10a and both command files, and says the split
holds for every later mention. `01-introduction.typ` now points forward to it;
`08-discussion.typ`'s mention of `freeze -reviewer` now explicitly says
"requirement wording" and back-references `@sec:evidence`.

**C5 — `K11` in the §5 table.** ORCHESTRATION claimed K11 is expanded nowhere and
implied it may not exist. It does exist. Two changes: the SWE.1.BP5 row now names
`K11-REQ-UNSOURCED` and `K11-SOURCE-UNANCHORED` explicitly (this is the row K11
actually belongs to — requirement→source citation), and the SWE.3-outcome-3 row,
which was wrongly attributed to `K11`–`K13`, now names `K1-CONSTRUCT-UNBOUND` and
`K3-REQ-UNCOVERED`, which are the rules that actually hold up the source-code ↔
design direction. This was a **factual error in the paper**, not merely an
undefined term.

**C6 — "bound"/"covered" figures.** `02-background.typ`'s dangling reference to
"the 'bound'/'covered' figures of §2" now points at `@sec:evidence`, where those
directions are defined.

**C7 — JVM scope.** Rewritten once, consistently, without overstating:
- `01`: "a front end that reads JVM bytecode rather than source. Because Kotlin
  compiles to JVM bytecode, Java and Kotlin projects arrive at that front end in
  the same shape; one JVM profile, `java_springboot_ddd1`, exists (README §11a)."
- `03`: the three targets are now named (Java on Spring, Kotlin on Spring, Kotlin
  on Android) with the README §11a citation.
- `07`: unchanged; it already listed exactly one JVM profile and now agrees.

**C8 — `rempel2017` causal form.** `02-background.typ` changed from "found … that
more complete traceability **decreases** the expected defect rate" to "**report**
… that more complete traceability **is associated with** a decreased expected
defect rate", matching `01-introduction.typ` and the CITATION-GUIDE prohibition
on a causal claim beyond the authors' wording.

**C9 — cross-references.** Labels added to all ten level-1 section headings:
`sec:introduction`, `sec:background`, `sec:design`, `sec:evidence`,
`sec:standards`, `sec:delimitation`, `sec:implementation`, `sec:discussion`,
`sec:conclusion`, `sec:disclosure`. Every prose cross-reference ("Section 6",
"§5", "(§7)", "quoted in §7") converted to `@sec:…`. The only remaining `§`
occurrences in the section files are references to *README* sections
(`README §2`, `§10a`, `§10b`, `§11`, `§11a`, `§12`, `§4.2`, `§4.5`), which are
correct and now uniformly carry the `README` prefix. Verified by grep; the
document compiles with no unresolved-reference errors.

---

## D. Unsupported claims removed

1. **Petri-net reachability (`03-design.typ`).** The uncited outside-world claim
   "whether every fork is matched by exactly one join on every path is
   reachability in a Petri net" is replaced by a statement about what speclink
   does not attempt: "speclink does not attempt to decide whether every fork is
   matched by exactly one join on every path".
2. **"none of them enforces requirement-to-code links" (`02-background.typ`).**
   This was an absence-claim about six *metadata-only* citations
   (`herold2014`, `deSilva2015`, `deLima2020`, `menezes2021`, `ozkaya2023`,
   `murphy1995`), whose content the CITATION-GUIDE forbids reporting at all.
   Replaced with: "their subject is the conformance of code to an architecture
   rather than to a requirement, and no claim is made here about what any of them
   can or cannot do."
3. **README §12 as authority for an absence (`08-discussion.typ`).** §7 records
   that README §12 is stale. §8 now grounds the absence of `speclink selfreport`
   in the code (`no speclink selfreport command exists in cmd/speclink`, which the
   reviewer verified by grep) and cites README §12 only as corroboration, noting
   it is stale on another point. This resolves ORCHESTRATION D2.

**Scan for other "cannot" claims about third-party tools.** All existing
occurrences were already in the permitted form and were left alone:
`02-background.typ`'s closing paragraph ("No claim is made here that these tools
cannot enforce a trace at build time — only that no such mechanism is documented
on the vendors' product pages") and `06-delimitation.typ`'s use of *not
documented* throughout `@tbl:positioning`. No new citation was added anywhere; no
bibliography entry was added; no `#set`, `#show`, `#import` or `#bibliography`
was added to any section file.

---

## Defects from ORCHESTRATION.md NOT fixed, and why

- **ORCHESTRATION C5 as stated is not fixable because it is wrong.** It asserts
  `K11` "is never expanded anywhere in the paper", implying it may be spurious.
  `K11-REQ-UNSOURCED` and `K11-SOURCE-UNANCHORED` are real rules in
  `internal/reqtree`. The real defect was that §5's table attributed K11 to the
  wrong ASPICE row. Fixed as described under C5 above, not as ORCHESTRATION
  proposed.
- **ORCHESTRATION A3 is incomplete.** It lists K15–K18 and K20 as being in
  `internal/check` and K19 in `internal/reqtree`, but misses that **K11 is also in
  `internal/reqtree`**, not in `internal/check`. The fix reflects the code, not
  ORCHESTRATION's list.
- **ORCHESTRATION D2 was labelled "not a contradiction".** It is closer to one
  than that; it is fixed anyway (see D.3 above).

Nothing else in ORCHESTRATION.md was left unaddressed.

---

## Remaining concerns for a human

1. **`STYLE.md` and `README.md` §1 disagree on the construct enumeration order.**
   STYLE.md line 53 gives "use case, command, event, aggregate, projection,
   repository, permission, query"; README §1 rule 4 gives "use cases, commands,
   events, aggregates, permissions, queries, projections and repositories". The
   paper now uses the README order everywhere, per the review instructions.
   **`STYLE.md` should be corrected, or the decision recorded.**

2. **The `verified` direction and `K14-VERIFICATION-STALE` remain in tension.**
   README §2 requires `verified` to reach 100%, and README §10b says a
   requirement that nothing demonstrated this run "loses its record" and is then
   reported as `K14-VERIFICATION-STALE` — which, since every finding is an error,
   fails the build. If that is literally true, `88% demonstrated` cannot appear in
   a *green* run either, only in a failing one. The paper now describes
   `demonstrated` as a figure with an accompanying finding, which is what README
   §10b says, and avoids asserting that `88% demonstrated` is a passing state.
   **A human who knows the tool should confirm whether a stale verification
   actually fails `verify`, and tighten both README and paper accordingly.**

3. **The `spec.Waive` treatment in §8 was preserved almost entirely** because it
   is analysis (incentives, counter-pressures, unmeasured effectiveness) rather
   than repetition. Only the restatement of the rule itself was cut. If the
   intent was a deeper cut, that is a judgement call left to a human.

4. **Two figures in §7 are still self-measured and self-reported** (`@tbl:measured`
   and the `193 / 45 / 9` reference-ERP counts). §8 now says so explicitly, but no
   independent verification exists and none is possible from the paper.

5. **`06-delimitation.typ` and `08-discussion.typ` now both state that there is no
   evaluation**, which removes the contradiction but makes §6's closing paragraph
   partly redundant with §8's "Absence of empirical evaluation". The redundancy is
   deliberate — §6 needs the symmetry argument against the Tielke report at that
   point — but a human may wish to trim it further.

6. **The `deLucia2007` and `antoniol2002`/`hayes2006` back-references in §1 and §8
   are now bare `@key` citations** rather than prose-form attributions. This is
   permitted by STYLE.md and by the citation guide, but it makes §1 slightly
   drier. Deliberate, per the instruction to cut rather than rewrite.

---

## Compilation

```
$ cd doc/paper && typst compile main.typ /tmp/check.pdf
```

Exit status 0. **0 errors, 0 warnings.**
