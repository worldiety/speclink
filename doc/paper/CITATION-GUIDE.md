# Citation Guide

Companion to `doc/paper/refs.bib`. Authority for every judgement below is
`doc/paper/research/AUDIT.md` (independent citation audit, re-verified 2026-08-30).
Every entry in `refs.bib` was confirmed by that audit; sources and claims the audit
rejected are **absent** from `refs.bib` and listed below.

---

## ⛔ NOT CITABLE AT ALL — do not write these claims, with or without a citation

These failed verification or are paywalled with no public text. No entry in `refs.bib`
supports them. Adding a citation to any of them is a fabrication.

1. **ISO 26262-8 Tool Impact (TI) / Tool Error Detection (TD) / Tool Confidence Level (TCL)
   clause numbers, the TCL1–TCL3 classification, or the qualification-method tables.**
   Paywalled; the public Part-1 vocabulary contains neither "tool confidence level" nor
   "tool impact".
2. **Any DO-178C objective-table content or objective numbers (e.g. Table A-2/A-7),
   DAL A–E definitions, or DO-330 TQL-1…TQL-5 tool qualification levels.** Paywalled;
   RTCA product pages carry none of it.
3. **IEC 62304 software safety classes A/B/C**, **IEC 61508 SIL 1–4 definitions**, and any
   **SIL↔ASIL mapping table**. Paywalled, no public text.
4. **Any ISO/IEC/IEEE 29148 or 12207 statement about traceability or requirement
   characteristics.** Not in the public abstracts. Only the *attributed, indirect* ASPICE
   SWE.1.BP1 Note 1/2 formulation (via `aspice40`) may be used.
5. **Any ISO 26262-6 clause number beyond 10.2** and **any ISO 26262-8 clause number beyond
   10.2**; the public tables of contents are truncated there.
6. **The word "traceability" as appearing in ISO 26262-8.** It does not occur in the public
   portion; do not claim it does or does not occur in the normative text.
7. **Any content of the David Tielke video beyond its description text and chapter list**
   (`tielke2026`). Nothing said, shown or measured in the video body may be reported. Do not
   call it a conference talk, do not attribute an institutional affiliation to its author,
   and do not quote its view count as a fixed figure.
8. **"Software Integration & Integration Verification" as an ASPICE PAM v4.0 quotation.**
   That string does not exist in the document. The Annex C wording is
   *"Software Integration & Integration Test"*.
9. **ISBN 9782889109852 as the ISBN of IEC 61508 Parts 1–7.** It identifies the IEC 61508:2010
   *Commented Version* (CMV), a value-added product. Cite individual parts instead.
10. **The spliced Polarion LiveDocs "quotation"** ("An exclusive innovation, Polarion LiveDocs,
    enables you to collaborate … uniquely identifiable and traceable"). Not a contiguous string
    on the page. Split into two quotes or paraphrase.
11. **Kelly & Weaver, "The Goal Structuring Notation — A Safety Argument Notation".**
    No venue/year/pages/DOI verifiable. Use `gsn`, `chelouati2023` or `wu2007` instead.
12. **Jackson & Wing, "Lightweight Formal Methods" (IEEE Computer).** Not separately indexed
    anywhere. Use `jackson2002`.
13. **Any Gobra publication.** No venue/DOI verified. Cite only `gobra` (the repository).
14. **ArchUnit as a peer-reviewed source**, and **any ArchUnit version number** (the site
    simultaneously advertises v1.5.0 and v1.4.2).
15. **"Automotive SPICE 4.1" as a released PAM version.** VDA QMC still publishes 4.0.
16. **Any CMMI version number.** None is published on the ISACA/CMMI Institute pages.
17. **The identifier "JSR 269"** in connection with the cited Oracle page — the string does not
    appear there.
18. **OSLC as an OMG standard.** It is an OASIS Open Project; "OMG" is absent from the site.
19. **Precision/recall numbers attributed to `antoniol2002` or `hayes2006`.** Confirmed absent
    from both abstracts.
20. **Findings, results or paraphrased content of any abstract-less work** — see the
    "metadata only" rows in the table below.
21. **Comparative claims about the internal traceability data models of DOORS / DOORS Next,
    Jama Connect, Polarion or Codebeamer**, and **any statement that one of these tools
    "cannot" do compile-time requirement↔code enforcement.** Only vendor marketing pages were
    fetched. The supportable form is: *"no such mechanism is documented on the vendor's
    product pages."*
22. **Vendor superlatives repeated as fact:** PTC "#1 in ALM" / Spark Matrix, Jama's G2 badges,
    Kiro's "We pioneered spec-driven development", Polarion's "An exclusive innovation you
    won't find elsewhere", OMG's "almost all RM and SysML modeling tools today support ReqIF".

---

## Citation table

| bib key | Source type | May be used to support (exactly this) | Must NOT be used to support |
|---|---|---|---|
| `gotel1994` | peer-reviewed | That requirements traceability is a multifaceted problem with no all-encompassing solution, based on an empirical study of over 100 practitioners, and the pre-RS/post-RS distinction. | Any tool comparison, any numeric traceability metric, or the "C.W. Finkelstein" author rendering. |
| `ramesh2001` | peer-reviewed | That reference models for traceability were derived empirically from focus groups and interviews in 26 organisations, distinguishing low-end from high-end traceability users. | Any claim about compile-time enforcement or about specific tool implementations today. |
| `gotel2012quest` | peer-reviewed | That despite advances in automated trace-link creation and maintenance, traceability implementation and use is "still not pervasive in industry". | Any specific measurement of industrial adoption rates. |
| `gotel2012grand` | peer-reviewed (metadata only) | The existence and authorship of the community's grand-challenge roadmap chapter. | Any of its content, findings or definitions — no abstract was retrievable. |
| `gotel2012fundamentals` | peer-reviewed (metadata only) | The existence of a community chapter on traceability fundamentals. | Any traceability definition or terminology attributed to it. |
| `clelandhuang2012` | peer-reviewed (book, metadata only) | The existence of the edited volume as the container of the two chapters above. | Any content claim. |
| `antoniol2002` | peer-reviewed | That IR-based (probabilistic and vector-space) recovery of links between code and free text was proposed, premised on meaningful identifier names, and that the authors themselves discuss its limitations. | Any precision/recall number — none appears in the abstract. |
| `antoniol2025` | peer-reviewed (metadata only) | That a 2025 TSE retrospective on the 2002 paper exists. | Any of its content or conclusions. |
| `hayes2006` | peer-reviewed | That candidate-link generation for requirements tracing is analyst-in-the-loop, with goals/measures defined and the RETRO prototype presented. | Any numeric accuracy figure — none appears in the abstract. |
| `deLucia2007` | peer-reviewed | The verbatim finding that IR-based tools "are still far to support a complete semi-automatic recovery of all links", evaluated over seventeen projects with about 150 students. | Generalisation of that finding to today's LLM-based recovery. |
| `maeder2012` | peer-reviewed | That in a controlled experiment with 52 subjects, subjects with traceability were on average 21% faster and produced 60% more correct solutions. | Attaching those numbers to `maeder2015`. |
| `maeder2015` | peer-reviewed (metadata only) | That a journal version of the traceability-benefit experiment exists in EMSE 20(2). | Any finding or number — its abstract was not retrievable. |
| `rempel2017` | peer-reviewed | That across 24 open-source projects, multi-level Poisson regression showed more complete traceability decreases the expected defect rate, for three of the four studied activities. | A causal claim beyond the authors' wording, or the "Parick Mader" name form. |
| `chelouati2023` | peer-reviewed (metadata only) | That GSN-based graphical safety assurance cases are an active peer-reviewed research topic (autonomous trains). | Any of its content — its abstract was not retrievable. |
| `gsn` | community/SDO-adjacent standard, non-peer-reviewed | That a Goal Structuring Notation Community Standard exists and is the authoritative published definition of GSN. | Attribution to Kelly & Weaver, or any claim about GSN's origin/history. |
| `wu2007` | peer-reviewed (metadata only) | That GSN has been combined with Bayesian belief networks for architectural safety reasoning. | Any result or evaluation figure. |
| `denney2015` | peer-reviewed (metadata only) | That "dynamic safety cases" for through-life assurance have been proposed. | Any result or evaluation figure. |
| `palin2010` | peer-reviewed (metadata only) | That safety-case approaches have been applied to automotive safety assurance. | Any result or evaluation figure. |
| `meyer1992` | peer-reviewed | That Design by Contract provides methodological guidelines for reliability, realised in Eiffel, with assertions as the mechanism. | Any claim about contract checking in Go or about modern verifiers. |
| `leavens2006` | peer-reviewed | That JML embeds pre-/postconditions and assertions **intermixed with Java source code**, using Java expressions so working engineers can write them. | Any adoption, tooling-maturity or industrial-uptake claim. |
| `leavens1999` | peer-reviewed (metadata only) | That an earlier chapter version of the JML notation exists. | Any content claim. |
| `burdy2005` | peer-reviewed (metadata only) | That a journal overview of JML tools and applications exists (STTT 7(3), 2005). | Any list of tools, capabilities or evaluation results — its abstract was not retrievable. |
| `burdy2003entcs` | peer-reviewed (metadata only) | That an earlier version appeared in ENTCS vol. 80, the FMICS 2003 proceedings issue. | Citing it as a bare "FMICS 2003" conference paper, or any content claim. |
| `jackson2002` | peer-reviewed | That Alloy is a small declarative language for structural properties whose syntax is "amenable to a fully automatic semantic analysis". | The term "lightweight formal methods" attributed to Jackson & Wing. |
| `murphy2001` | peer-reviewed | The verbatim statement that "the artifacts constituting a software system often drift apart over time", and that reflexion models summarise consistency between a high-level artefact and source (applied to ~1 MLOC Microsoft Excel). | Any claim that reflexion models enforce requirement↔code links. |
| `murphy1995` | peer-reviewed (metadata only) | That the reflexion-model technique was first published at FSE'95 (proceedings DOI, not the SIGSOFT SEN DOI). | Any content claim — prefer `murphy2001` for content. |
| `herold2014` | peer-reviewed (metadata only) | That rule-based architecture conformance checking has been proposed as a quality-management measure. | Any result or evaluation figure. |
| `deLima2020` | peer-reviewed (metadata only) | That architecture conformance checking has been implemented for Python (ArchPython). | Any result or evaluation figure. |
| `deSilva2015` | peer-reviewed (metadata only) | That static architecture conformance checking has been proposed against architecture erosion. | Any result or evaluation figure. |
| `menezes2021` | peer-reviewed (metadata only) | That model checking has been applied to architecture conformance checking. | Any result or evaluation figure. |
| `ozkaya2023` | peer-reviewed (metadata only) | That architecture conformance checking is discussed for infrastructure as code in practitioner-facing venues. | Any result or evaluation figure. |
| `liu2018` | peer-reviewed | That heuristic detection of outdated comments achieves 74.6% detection and 77.2% precision using 64 features. | Extrapolating those figures to any other drift-detection technique. |
| `stulova2020` | peer-reviewed | The verbatim statement that "none of these tools checks for consistency of the documentation accompanying the code", and that upDoc's evaluation is self-described as preliminary. | A claim that *no tool whatsoever* can check documentation consistency today. |
| `pearce2022` | peer-reviewed | That across 89 scenarios Copilot produced 1,689 programs of which approximately 40% were vulnerable. | Generalising to current-generation models or to non-security defects. |
| `pearce2025cacm` | peer-reviewed (metadata only) | That a CACM version of the Copilot security study exists. | Any figure not already sourced from `pearce2022`. |
| `perry2023` | peer-reviewed | That participants with access to an AI assistant wrote significantly less secure code *and* were more likely to believe it was secure (overconfidence). | Any quantified effect size. |
| `panickssery2024` | peer-reviewed | That LLM evaluators exhibit self-preference — scoring their own outputs higher than human annotators do — correlated with self-recognition capability. | A claim that LLM self-assessment is useless in all settings. |
| `zheng2023` | peer-reviewed (metadata only) | That LLM-as-a-judge evaluation was formalised at NeurIPS 2023. | Any of its findings — its abstract was not fetched. |
| `chen2021codex` | preprint (not peer-reviewed) | Attributed statements only: "Chen et al. report that Codex solves 28.8% of HumanEval problems at one sample and 70.2% with 100 samples." | Any statement of those figures as established fact, or as a benchmark of current models. |
| `iso26262-1` | international standard | The ASIL definition (term 3.6), the "software tool" definition (term 3.158), the title/edition/date/committee, and that neither "tool confidence level" nor "tool impact" appears in the public vocabulary. | Any normative requirement text. |
| `iso26262-6` | international standard | The title/edition/page count, the scope topic list (including "testing of the embedded software" and configurable software), and the publicly visible clause titles 5–10. | Any clause number beyond 10.2, or any normative requirement. |
| `iso26262-8` | international standard | The title/edition, the 12-item supporting-process scope list (including "confidence in the use of software tools"), the IEC 61508-adaptation sentence, the V-model statement, the "m-n" clause notation, and the bibliography entries (DO-178C, CMMI-DEV, Automotive SPICE®, 12207, 29148, ISO/IEC 33000). | TI/TD/TCL clause numbers or content; anything about traceability requirements. |
| `iec61508-1` | international standard | The title/edition/date, that it covers E/E/PE safety functions, its objective "to facilitate the development of product and application sector international standards", and its IEC Guide 104 basic-safety status. | SIL definitions or SIL↔ASIL mapping. |
| `iec61508-3` | international standard | That Part 3 provides specific requirements applicable to support tools (development/design tools, language translators, testing and debugging tools, configuration management tools). | The actual normative tool requirements, or SIL definitions. |
| `iec62304` | international standard | The title, edition 1.1 (2006 + AMD1:2015), page count, ISBN, committee, and that it defines medical device software life cycle requirements and excludes final device validation. | Software safety classes A/B/C. |
| `do178c` | vendor doc (SDO product page, attributed) | Attributed only: RTCA states DO-178C is "the primary means of obtaining approval of software used in civil aviation products"; plus title, committee SC-205 and issue date 2011-12-13. | Any objective number, objective-table content, or DAL definition. |
| `do330` | vendor doc (SDO product page, attributed) | Attributed only: RTCA's tool definition, that the document "explains the process and objectives for qualifying tools", and its cross-domain applicability sentence. | TQL-1…TQL-5 levels or DO-178C tool criteria 1/2/3. |
| `do333` | vendor doc (SDO product page, attributed) | Attributed only: that DO-333 supplements DO-178C/DO-278A for formal methods, and RTCA's definition of formal methods. | Any modified objective or its numbering. |
| `ed12c` | vendor marketing (SDO training/catalogue page) | Attributed only: EUROCAE's own wording that ED-12C "is the European reference standard for airborne software certification and is equivalent to RTCA DO-178C", and the existence of ED-215/216/217/218/ED-94C. | Any regulatory recognition claim (FAA/EASA), or any normative content of ED-12C. |
| `aspice40` | international standard (freely published normative model) | That "Ensure consistency and establish bidirectional traceability" is a base-practice name (SYS.2.BP5, SYS.3.BP4, SYS.4.BP4, SYS.5.BP4, SWE.1.BP5); SWE.1's 7 outcomes incl. 5 and 6; Notes 9 and 11 to SWE.1.BP5; SWE.3 outcome 3 reaching source code ↔ detailed design; SWE.4/SWE.6 outcome 4; the SWE.1–SWE.6 process names; the six capability levels and nine process attributes; the ISO/IEC 33004/33003/33020 conformance statements; and the attributed SWE.1.BP1 Note on requirement characteristics. | The phrase "Software Integration & Integration Verification"; the "13-51 Consistency Evidence → outcomes 5 and 6" mapping (unresolved); the claim that SWE.1.BP5 requires traceability "in both directions" rather than to two targets. |
| `aspiceweb` | vendor doc (model owner page) | The 32 processes / 3 categories / 11 groups counts, the 2005 first publication, the SPICE expansion, and that VDA QMC currently publishes version 4.0. | The self-reported assessor/country/language figures as stable facts; any claim that version 4.1 is released; the enumeration of all 11 groups. |
| `isoiec33002` | international standard | That it defines the minimum requirements for performing a process assessment, and that ISO/IEC 15504-2:2003 is its withdrawn predecessor (the 15504→330xx lineage). | Any assessment method detail. |
| `isoiec33004` | international standard | That it sets requirements for process reference, process assessment and maturity models, and relates them to activities/tasks defined in ISO/IEC 12207 and 15288. | Any conformance criterion detail. |
| `iso12207` | international standard | The title, publication status, and the abstract's framing of a common framework for software life cycle processes. | Anything about traceability. |
| `iso12207-2017` | international standard | That the 2017 edition is withdrawn (stage 95.99) and revised by the 2026 edition. | Anything about traceability. |
| `iso29148` | international standard | The title/edition/date/page count and the four abstract bullets (required processes, guidelines relative to 15288/12207, required information items and their contents). | Requirement characteristics (unambiguous, verifiable, traceable, …) or anything about traceability — use the attributed `aspice40` formulation instead. |
| `cmmi` | vendor marketing | Attributed only: that ISACA owns and operates CMMI and describes it as a set of global best practices originally created for the U.S. DoD that has expanded beyond software engineering. | Any technical model content, practice-area detail, or version number. |
| `reqif` | international standard (SDO page) | That ReqIF is an OMG specification defining an open, non-proprietary XML interchange format for requirements, motivated by cross-company supply-chain exchange. | The unquantified "almost all RM and SysML modeling tools today support ReqIF" claim; the current ReqIF version number. |
| `oslc` | international standard (SDO page) | That OSLC is an OASIS Open Project defining REST/Linked-Data (RDF, W3C LDP-based) specifications for linking lifecycle resources, with CM 3.0 (effective 26 May 2021) and Configuration Management 1.0 as OASIS Standards. | Any attribution of OSLC to OMG; any claim about compile-time enforcement. |
| `speckit` | vendor doc | That Spec Kit is a GitHub-published, MIT-licensed toolkit whose README states "specifications become executable", the 0–5 constitution→converge workflow, `/speckit.analyze` as cross-artifact consistency analysis, and that regulatory/V-model traceability is described only as something presets or extensions *could* add. | The star/fork counts as stable figures; any claim that Spec Kit *cannot* fail a build. |
| `kiro` | vendor doc | That Kiro is built by a team within AWS and that its specs feature generates `requirements.md`, `design.md` and `tasks.md`, with an LLM-based "Analyze Requirements" step to catch inconsistencies, ambiguities and gaps. | The "We pioneered spec-driven development" priority claim; any claim that Kiro *cannot* link a requirement ID to a code construct. |
| `doorsnext` | vendor marketing | Attributed only: IBM markets DOORS/DOORS Next with structured requirement modules, baselines, electronic signatures, multi-level traceability and a graphical link explorer, and claims support for ASPICE, ISO 26262 and DO-178C. | DOORS' actual link-type model, OSLC or ReqIF support; any "cannot" statement. |
| `polarion` | vendor marketing | Attributed only: Siemens claims traceability via automatic change control, "full traceability of every source code modification up to the change request", paragraph-level identifiability in LiveDocs, built-in ReqIF, and SVN/Git support. | The spliced LiveDocs quotation; Polarion's internal data model; any "cannot" statement. |
| `jama` | vendor marketing | Attributed only: Jama markets requirements traceability, standards compliance (ISO 26262, ASPICE, DO-178C, ISO 13485, 21 CFR Part 11), audit trails, and a "Spec-Driven Development" capability via MCP. | Scale figures (10M/100M items) as verified facts; G2 badges; Jama's internal traceability data model. |
| `codebeamer` | vendor marketing | Attributed only: PTC claims end-to-end traceability across work items in a centralised repository, OSLC-based digital-thread integration, a standards list (ISO 26262, ASPICE, IEC 62304, DO-178C, ISO 14971) and Git/Jira/Windchill/Rhapsody integrations. | "#1 in ALM" / Spark Matrix leadership; Codebeamer's internal data model; any "cannot" statement. |
| `archunit` | community, non-peer-reviewed | That ArchUnit is a Java library evaluating structural architecture rules (packages, classes, layers, slices, cycles) as ordinary unit tests by analysing bytecode, with a .NET/C# port. | Any version number; any claim of requirement↔code traceability support; citation as peer-reviewed work. |
| `govet` | vendor doc (official toolchain documentation) | That `go vet` ships with the Go toolchain, exits non-zero when a problem is reported (usable as a CI gate), lists checks such as `printf`/`structtag`/`unusedresult`, and is explicitly documented as heuristic guidance rather than a correctness guarantee. | Any claim about requirements or traceability semantics. |
| `staticcheck` | community, non-peer-reviewed | That Staticcheck is a Go linter with 150+ checks aimed at bugs, performance, simplifications and style, designed for low false positives and CI use, complementing `go vet`. | Attribution to a named individual on the cited docs page; any traceability capability. |
| `goanalysis` | vendor doc (official toolchain documentation) | The `Analyzer`/`Pass`/`Diagnostic`/`Fact` model, modular per-package analysis with facts propagated along the import graph ("separate analysis"), `singlechecker`/`multichecker` drivers, and `analysistest`'s `// want` comments. | Any requirement or traceability semantics — the framework provides none. |
| `javaap` | international standard (Java SE API specification) | That annotation processing runs `Processor` implementations inside the compiler over rounds, discovered by service-style lookup, able to claim annotations and to raise errors that fail compilation, with the four documented robustness properties. | The identifier "JSR 269"; any traceability semantics. |
| `gobra` | community, non-peer-reviewed | That Gobra is a self-described prototype, automated modular verifier for Go built on Viper (Silicon/Carbon, Z3/Boogie) in which specifications are written as annotations in `.gobra` programs, applied to VerifiedSCION and WireGuard. | Any peer-reviewed Gobra publication; the star count as a stable figure. |
| `tielke2026` | community, non-peer-reviewed (self-published video) | Attributed only: that a practitioner has publicly reported a self-conducted, uncontrolled experiment re-implementing an application as an AI-built "enterprise system", staged through micro-management, quality-driven, spec-driven, test-driven and idea/voice-driven phases (per its own description and chapter list). | Anything said or measured in the video body; that it is a conference talk; the author's institutional affiliation; the view count; any general claim about AI-assisted enterprise development. |

---

## Phrasing rules

- **Peer-reviewed, abstract-verified** entries may be stated as plain facts.
- **"metadata only"** entries may be cited as *pointers to existing work* only:
  "see also [key]". Never paraphrase their findings.
- **International standards** may be stated as facts only for the exact publicly verified
  strings listed above; everything normative is paywalled.
- **Vendor doc / vendor marketing** must always be phrased as
  *"According to X's own documentation …"* or *"X markets Y as …"*.
- **Community and preprint** sources must be phrased as
  *"X reports that …"*, with an explicit note that the source is not peer-reviewed.
- For any absence of a capability in a third-party tool, the only supportable phrasing is
  *"no such mechanism is documented on the vendor's product pages"*.
