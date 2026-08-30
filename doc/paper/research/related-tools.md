# Related Tools — Practitioner / Industrial Landscape

Research notes for the speclink paper. **All facts below were fetched during a single
research session on 2026-08-30.** Nothing is written from memory. Each entry carries a
source-type label. Vendor and community sources may only be used for *attributed, hedged*
statements in the paper ("X's documentation states that ...").

Source-type labels used:

- **peer-reviewed** — published in a refereed venue
- **official standard** — published by a recognised SDO (OMG, OASIS, Oracle/JCP, Go project)
- **vendor documentation** — technical reference material published by the vendor
- **vendor marketing** — product/landing pages published by the vendor
- **community content** — self-published talks, blogs, videos, OSS project READMEs

---

## A) The YouTube talk (https://www.youtube.com/watch?v=eLDHrqKplVI)

### A.1 Identification

- **Title:** `Der Moment, der die Softwareentwicklung geändert hat!`
- **Author / channel:** David Tielke (`https://www.youtube.com/@DavidTielke`)
- **Publish / upload date:** `2026-06-21T09:34:52-07:00`
- **View count at time of fetch:** 170,523
- **URL fetched:** `https://www.youtube.com/oembed?url=https://www.youtube.com/watch?v=eLDHrqKplVI&format=json`
  and the watch page `https://www.youtube.com/watch?v=eLDHrqKplVI` (video metadata / `shortDescription`
  extracted from the page payload).
- **Source type:** **community content** (self-published YouTube video)

Verbatim quote (oEmbed JSON):

> `{"title":"Der Moment, der die Softwareentwicklung geändert hat!","author_name":"David Tielke","author_url":"https://www.youtube.com/@DavidTielke","type":"video",...,"provider_name":"YouTube"}`

**Affiliation — partially verified.** The video description links to `https://david-tielke.de/#workshops`
and offers workshops; the description text is verbatim:

> `🎓 Meine Workshops (neu auf KI ausgerichtet): „KI für Entwickler", „Anforderungen mit KI" und „KI-Strategie für Unternehmen" 👉 https://david-tielke.de/#workshops`

That establishes that the author operates commercially under `david-tielke.de` as a trainer/consultant.
No employer/university affiliation could be verified from the fetched material.

### A.2 Event / venue — NOT a conference talk

**This is not a conference talk.** The fetched metadata identifies it as a video published on the
author's own YouTube channel. No event, conference, or programme committee is named anywhere in the
fetched material. Consequently:

- It is **not peer-reviewed**.
- It is **not** a talk given at a refereed or even a named industrial venue, as far as the fetched
  evidence shows.
- In the paper it can only be cited as **self-published practitioner commentary**, with an explicit
  statement that it carries no peer review and no methodological control.

### A.3 Thesis — from the fetched description only

The full transcript was **not** reachable in this session; only the video description (`shortDescription`
field of the watch page payload) and the chapter list were fetched. The summary below is derived
**solely** from that text.

Verbatim opening of the description:

> `Kann KI Mitte 2026 wirklich große Enterprise-Software bauen – mit echter Architektur, echten Anforderungen und kompromissloser Qualität? Oder ist das alles nur Vibe-Coding-Hype aus kurzen YouTube-Demos?`

> `Ich habe dazu ein Experiment gemacht, das mich seit Wochen nicht schlafen lässt. Über mehrere Monate habe ich – meistens nachts neben Kundenprojekten – eine Anwendung, die ich seit 18 Jahren nutze, komplett neu entwickelt: als Enterprise-System, mit Microservice-Architektur, lokaler KI, Workflow-Engine, vollständiger Spezifikation, Tests und Dokumentation.`

Verbatim chapter markers relevant to spec-driven development:

> `[14:59] Das Vorgehen & Phase 1: Mikro-Management`
> `[15:35] Phase 2 & 3: Quality Driven & Spec Driven mit Harness`
> `[18:43] Phase 4: Test Driven Development aus Anforderungen`
> `[20:04] Phase 5: Idea Driven & Voice Driven Development`
> `[23:31] Die Ergebnisse: User Stories, Services & Lines of Code`
> `[25:45] Strukturelle Schulden & Code-Qualität im Check`
> `[26:45] Was kostet KI-Entwicklung wirklich?`
> `[30:13] KI vs. Entwicklerteam: Der Qualitätsvergleich`
> `[31:11] Der Geschwindigkeitsfaktor – das überraschende Ergebnis`

**Factual description (1–3 sentences):** The video is a first-person report of a self-conducted,
uncontrolled experiment in which the author re-implemented an application he had used for 18 years as
an "enterprise system" (microservice architecture, local AI, workflow engine, full specification, tests,
documentation) largely using AI agents. Per its own chapter list, the process is staged from
"micro-management" through "quality driven" and "spec driven" phases to test-, idea- and voice-driven
development, and closes with self-reported numbers on scope, code quality, cost, token consumption and
speed. **The claimed results are self-reported and not independently verifiable from the fetched material.**

**Caveat for the paper:** any use must be phrased as e.g. "In a self-published video, Tielke reports
that ..." — never as evidence for a general claim about AI-assisted enterprise development.

---

## B) "Spec-driven development" tooling

### B.1 Spec Kit (`github/spec-kit`)

- **Vendor / author:** GitHub, Inc. (repository under the `github` organisation); README credits
  "the work and research of John Lam".
- **URL fetched:** `https://github.com/github/spec-kit`
- **Source type:** **vendor documentation** (project README published by the vendor)
- **Licence / scale at fetch time:** MIT; 132.3k stars, 11.9k forks, 1,872 commits, version 1.0.0.

Verbatim quotes:

> `💫 Toolkit to help you get started with Spec-Driven Development`

> `Spec-Driven Development **flips the script** on traditional software development. For decades, code has been king — specifications were just scaffolding we built and discarded once the "real work" of coding began. Spec-Driven Development changes this: **specifications become executable**, directly generating working implementations rather than just guiding them.`

> `0. **Establish** your project principles once (`/speckit-constitution`). This is a one-time step per project. 1. **Specify** what you want to build (`/speckit-specify`). 2. **Plan** how you will build it (`/speckit-plan`). 3. **Break down** the plan into actionable tasks (`/speckit-tasks`). 4. **Implement** the tasks (`/speckit-implement`). 5. **Converge** the implementation against the spec, plan, and tasks (`/speckit-converge`).`

> `/speckit.analyze | speckit-analyze | Cross-artifact consistency & coverage analysis (run after /speckit.tasks, before /speckit.implement)`

> `For example, presets could restructure spec templates to require regulatory traceability`

> `For example, extensions could add Jira integration, post-implementation code review, V-Model test traceability, or project health diagnostics.`

**Factual description:** Spec Kit is a GitHub-published, MIT-licensed Python CLI (`specify`) plus a set
of prompt/slash-command templates that drive an AI coding agent through a fixed
constitution→specify→plan→tasks→implement→converge workflow, producing Markdown artefacts in a
`.specify/` tree. Its consistency checking (`/speckit.analyze`, `/speckit.converge`) is performed by the
LLM agent over Markdown artefacts; the README describes traceability only as something a *preset* or
*community extension* could add, not as a built-in, compiler-enforced mechanism.

**Delimitation note for the paper:** Spec Kit's artefacts are natural-language Markdown and the
"analysis" is agentic, not a static check on source code. There is no evidence in the fetched README of
any mechanism that fails a build when a requirement has no corresponding code.

### B.2 Kiro

- **Vendor:** AWS (Amazon Web Services). Verified from the vendor's own About page.
- **URLs fetched:** `https://kiro.dev/docs/specs/` and `https://kiro.dev/about/`
- **Source type:** **vendor documentation** (docs pages) / **vendor marketing** (About page)

Verbatim quotes (About page):

> `Kiro is built and operated by a small, opinionated team within AWS.`

> `Through Kiro, we reinvented how developers work with AI agents. We pioneered spec-driven development, where Kiro turns your prompt into structured requirements, design, and tasks that are then implemented by agents.`

Verbatim quotes (docs/specs, page updated August 27, 2026):

> `Specs or specifications are structured artifacts that formalize the development process for features and bug fixes in your application. They provide a systematic approach to transform high-level ideas into detailed implementation plans with clear tracking and accountability.`

> `Every spec generates three key files that form the foundation of your specification: **requirements.md** (or **bugfix.md**) - Captures user stories, acceptance criteria, or bug analysis in structured notation - **design.md** - Documents technical architecture, sequence diagrams, and implementation considerations - **tasks.md** - Provides a detailed implementation plan with discrete, trackable tasks`

> `**Break down requirements** into user stories with acceptance criteria`

> `[Analyze Requirements](/docs/specs/analyze-requirements/) — Catch inconsistencies, ambiguities, and gaps in your requirements before design.`

**Factual description:** Kiro is an AWS-built agentic development environment whose "specs" feature
generates and maintains three Markdown artefacts (`requirements.md`, `design.md`, `tasks.md`) and drives
agents through a requirements→design→tasks workflow with task-level status tracking. Per its own
documentation the traceability unit is a *task checkbox* in `tasks.md`, plus an LLM-based
"Analyze Requirements" step; the docs do not describe a compile-time link between an individual
requirement identifier and a source-code construct.

**Note:** Kiro's own About page claims it "pioneered spec-driven development". This is a
**vendor marketing claim** and must not be repeated as fact; the priority claim is not independently
verified here.

### B.3 Jama Connect (marketed under a "Spec-Driven Development" heading)

- **Vendor:** Jama Software
- **URL fetched:** `https://www.jamasoftware.com/platform/jama-connect/`
- **Source type:** **vendor marketing**

Verbatim quote:

> `**Spec-Driven Development** – Engineers and AI engineering agents iterate in a shared context via MCP`

**Note:** relevant here because it shows a *requirements-management vendor* (see section C) has adopted
the "spec-driven development" label for agent workflows — i.e. the term is being used by both the AI-tooling
and the ALM camp. See C.3 for the substantive Jama entry.

---

## C) Requirements management / ALM traceability tools, and OSLC

### C.1 IBM Engineering Requirements Management DOORS / DOORS Next

- **Vendor:** IBM
- **URL fetched:** `https://www.ibm.com/products/requirements-management-doors-next`
- **Source type:** **vendor marketing** (product page; features section)

Verbatim quotes:

> `DOORS is a proven requirements management solution that has been successfully used by teams in complex, high-compliance systems engineering programs across all industrial sectors for several decades. It offers mature capabilities, including structured requirements specification modules, round-trip data import and export, electronic signatures, baselines and customizable requirements views with multi-level traceability.`

> `#### Traceability — Link artifacts for alignment and use a graphical explorer to visualize project relationships.`

> `[DOORS Next] enables you to capture, trace, analyze and manage changes to requirements while maintaining compliance with regulations and standards.`

> `Support standards such as ASPICE, ISO 26262 and DO178C while reducing the effort needed to demonstrate compliance.`

**Factual description:** IBM markets DOORS (classic) and DOORS Next as requirements-management products
for high-compliance systems engineering, offering requirement modules, baselines, electronic signatures
and multi-level traceability between artefacts held in the tool's own repository. The traceability unit
per the vendor page is a *link between artefacts inside the RM tool*; nothing on the fetched page
describes traceability into a source-code compiler.

### C.2 Siemens Polarion (REQUIREMENTS / ALM)

- **Vendor:** Siemens (Siemens Digital Industries Software)
- **URL fetched:** `https://www.siemens.com/en-us/products/polarion/requirements/`
  (redirect target of `https://polarion.plm.automation.siemens.com/products/polarion-requirements`)
- **Source type:** **vendor marketing**

Verbatim quotes:

> `Collaboration, traceability and workflow are the three core principles built into our DNA.`

> `Pass any audit, compliance or regulatory inspection with traceability that is easily implemented and guaranteed via automatic change control of every requirement`

> `Verify that the final delivered software has all of the planned enhancements that are supposed to be included in the release by full traceability of every source code modification up to the change request`

> `An exclusive innovation, Polarion LiveDocs, enables you to collaborate concurrently and securely on specification documents while having every single paragraph uniquely identifiable and traceable`

> `Built-in ReqIF enables lossless requirements and test case specifications exchange with customers and suppliers`

> `Support SVN and GIT out of the box, and other software (for example: Perforce; Plastic SCM) via add-ons`

**Factual description:** Polarion is Siemens' requirements-management/ALM product; per its own marketing
it makes each document paragraph uniquely identifiable and traceable ("LiveDocs"), supports ReqIF
import/export, and connects requirements to *version-control commits/change requests* via SVN/Git
integration. The code-side traceability granularity claimed on the page is the **source-code
modification (commit) linked up to a change request**, not an annotation on a code construct.

### C.3 Jama Connect

- **Vendor:** Jama Software
- **URL fetched:** `https://www.jamasoftware.com/platform/jama-connect/`
- **Source type:** **vendor marketing**

Verbatim quotes:

> `Jama Connect eliminates manual compliance with AI-generated requirements, test cases and traceability that adhere to industry-specific standards.`

> `**Scale** – Projects scale to 10 million items and instances to 100 million items`

> `**Compliance** – All AI governance and industry standard compliance met with approvals and audit trails.`

> `**LLM inference quality & token efficiency** – A semantic product graph spans all disciplines via MCP`

The site also exposes a dedicated navigation entry `Requirements Traceability`
(`https://www.jamasoftware.com/solutions/requirements-traceability/`), and comparison pages against
`IBM DOORS`, `IBM DOORS Next`, `Polarion` and `PTC Codebeamer and Integrity/RV&S` — useful evidence that
these four products form one competitive segment.

**Factual description:** Jama Connect is a commercial requirements-management/engineering-management
platform whose marketing emphasises requirements traceability, industry-standard compliance
(ISO 26262, ASPICE, DO-178C, IEC 62304-adjacent, ISO 13485, FDA 21 CFR Part 11 are named on the page),
audit trails, and — recently — MCP-based interaction with AI agents. Its traceability model is a graph
of items held in Jama's repository.

### C.4 PTC Codebeamer

- **Vendor:** PTC
- **URL fetched:** `https://www.ptc.com/en/products/codebeamer`
- **Source type:** **vendor marketing**

Verbatim quotes:

> `Codebeamer is an ALM platform for advanced product and software development. The open platform extends ALM functionalities with product line configuration capabilities, and provides unique configurability for complex processes.`

> `End-to-end traceability throughout the software engineering lifecycle, from the requirements stage all the way through testing and release. Codebeamer's centralized data repository guarantees complete traceability across all work items. Codebeamer also integrates with the PTC engineering digital thread, using OSLC technology.`

> `Codebeamer connects requirements, risks, tests, and changes across projects and tools—providing a complete digital thread from concept to release.`

> `Codebeamer supports a wide range of standards including ISO 26262, ASPICE, IEC 62304, DO-178C, ISO 14971, and others commonly required in safety-critical industries.`

> `Codebeamer integrates with PLM, CAD, modeling, development, and test tools, including Windchill, Jira, Git, IBM Rhapsody, and many others.`

**Factual description:** Codebeamer is PTC's ALM platform for regulated/safety-critical development; per
its own pages it provides end-to-end traceability across work items in a centralised repository,
integrates with Git and other engineering tools, and uses OSLC to link into PTC's wider "digital thread".
Traceability is again between *work items*, with source control integrated as an external tool.

### C.5 OSLC (Open Services for Lifecycle Collaboration)

- **Publisher:** OASIS Open (as an **OASIS Open Project**). *Note: the task brief said "OMG/OASIS";
  the fetched site attributes OSLC to OASIS, not OMG.*
- **URL fetched:** `https://open-services.net/`
- **Source type:** **official standard** (standards-body project site)

Verbatim quotes:

> `Open Services for Lifecycle Collaboration — Creating standard REST APIs to connect data`

> `Open Services for Lifecycle Collaboration © is an [OASIS Open Project](https://oasis-open-projects.org/). All Rights Reserved.`

> `The OSLC Core Specification is a Hypermedia API standard currently mainly adopted in software and systems engineering domains, but with the potential to provide value to any domain with data integration challenges. The OSLC Core specifications expands on the W3C LDP capabilities, to define the essential and common technical elements of OSLC domain specifications and offers guidance on common concerns for creating, updating, retrieving, and linking to lifecycle resources.`

> `OSLC domain-specific specifications define the equivalent of schemas in RDF for enabling data interoperability. They consist of RDF vocabularies and OSLC resource shapes.`

> `OASIS is pleased to announce that the Call for Consent has closed and, effective 26 May 2021, OSLC Change Management Version 3.0 is an OASIS Standard according …` (news item, May 27, 2021)

> `OSLC Configuration Management Version 1.0 published as an OASIS Standard` (news item, Jul 23, 2023)

**Factual description:** OSLC is an OASIS Open Project defining REST/Linked-Data (RDF, W3C LDP-based)
specifications for linking lifecycle resources across engineering tools; several domain specifications
(e.g. Change Management 3.0, PROMCODE 1.0, Configuration Management 1.0) have reached OASIS Standard
status. It standardises *inter-tool linking of resources by URL*, i.e. integration at the repository/API
level — orthogonal to any in-language, compile-time enforcement.

---

## D) Code-level enforcement tools

### D.1 ArchUnit

- **Author / maintainer:** Peter Gafert (site copyright); supported by TNG Technology Consulting GmbH;
  repository `TNG/ArchUnit`.
- **URL fetched:** `https://www.archunit.org/`
- **Source type:** **community content** (OSS project site) — technically authoritative for the project
  but not peer-reviewed.
- **Version visible at fetch:** v1.5.0 API docs linked; latest news item "Apr 18, 2026 – New release of ArchUnit (v1.4.2)".

Verbatim quotes:

> `Unit test your Java architecture — Start enforcing your architecture within 30 minutes using the test setup you already have.`

> `ArchUnit is a free, simple and extensible library for checking the architecture of your Java code using any plain Java unit test framework. That is, ArchUnit can check dependencies between packages and classes, layers and slices, check for cyclic dependencies and more. It does so by analyzing given Java bytecode, importing all classes into a Java code structure.`

> `There also exists a port for .NET/C#, which you can find here.`

**Factual description:** ArchUnit is a Java library that expresses architectural rules as ordinary unit
tests and evaluates them by importing and analysing Java **bytecode**. Its enforcement point is the test
suite (and thus CI), and its rule vocabulary is structural (packages, classes, layers, slices, cycles) —
it enforces *architecture* constraints, not requirement↔code correspondence.

### D.2 `go vet`

- **Publisher:** The Go project (Google); part of the Go standard distribution (`cmd/vet`).
- **URL fetched:** `https://pkg.go.dev/cmd/vet` (documented against go1.27.0)
- **Source type:** **official standard / vendor documentation** (official Go toolchain documentation)

Verbatim quotes:

> `Vet examines Go source code and reports suspicious constructs, such as Printf calls whose arguments do not align with the format string. Vet uses heuristics that do not guarantee all reports are genuine problems, but it can find errors not caught by the compilers.`

> `Vet's exit code is non-zero for erroneous invocation of the tool or if a problem was reported, and 0 otherwise. Note that the tool does not check every possible problem and depends on unreliable heuristics, so it should be used as guidance only, not as a firm indicator of program correctness.`

> `For information on writing a new check, see golang.org/x/tools/go/analysis.`

The page lists the built-in checks, e.g.:

> `printf           check consistency of Printf format strings and arguments`
> `structtag        check that struct field tags conform to reflect.StructTag.Get`
> `unusedresult     check for unused results of calls to some functions`

**Factual description:** `go vet` is the Go toolchain's built-in static checker; it runs a fixed set of
analyzers over Go source and exits non-zero when a problem is reported, which makes it usable as a CI
gate. Its official documentation explicitly frames it as heuristic guidance, not a correctness
guarantee, and points to `golang.org/x/tools/go/analysis` as the extension point for new checks.

### D.3 Staticcheck

- **Author:** Dominik Honnef (`github.com/dominikh/go-tools`)
- **URL fetched:** `https://staticcheck.dev/docs/`
- **Source type:** **community content** (OSS project documentation)

Verbatim quotes:

> `Staticcheck is a state of the art linter for the Go programming language. Using static analysis, it finds bugs and performance issues, offers simplifications, and enforces style rules.`

> `Each of the 150+ checks has been designed to be fast, precise and useful. When Staticcheck flags code, you can be sure that it isn’t wasting your time with unactionable warnings. Unlike many other linters, Staticcheck focuses on checks that produce few to no false positives. It’s the ideal candidate for running in CI without risking spurious failures.`

> `Staticcheck aims to be trivial to adopt. It behaves just like the official go tool and requires no learning to get started with. Just run staticcheck ./... on your code in addition to go vet ./....`

**Factual description:** Staticcheck is a third-party Go linter with 150+ checks, run as a CLI, in CI, or
via gopls in the editor. Its check catalogue is about bugs, performance, simplifications and style — it
has no notion of requirements or traceability.

### D.4 `golang.org/x/tools/go/analysis`

- **Publisher:** The Go project (Google), `golang.org/x/tools`, BSD-3-Clause.
- **URL fetched:** `https://pkg.go.dev/golang.org/x/tools/go/analysis` (documented at v0.49.0,
  published Aug 13, 2026; "Imported by: 6,564")
- **Source type:** **official standard / vendor documentation** (official Go sub-repository docs)

Verbatim quotes:

> `Package analysis defines the interface between a modular static analysis and an analysis driver program.`

> `A static analysis is a function that inspects a package of Go code and reports a set of diagnostics (typically mistakes in the code), and perhaps produces other results as well, such as suggested refactorings or other facts. An analysis that reports mistakes is informally called a "checker".`

> `A "modular" analysis is one that inspects one package at a time but can save information from a lower-level package and use it when inspecting a higher-level package, analogous to separate compilation in a toolchain.`

> `By implementing a common interface, checkers from a variety of sources can be easily selected, incorporated, and reused in a wide range of driver programs including command-line tools (such as vet), text editors and IDEs, build and test systems (such as go build, Bazel, or Buck), test frameworks, code review tools, code-base indexers (such as SourceGraph), documentation viewers (such as godoc), batch pipelines for large code bases, and so on.`

> `A Fact is an intermediate fact produced during analysis. Each fact is associated with a named declaration (a types.Object) or with a package as a whole. ... Facts may be produced in one analysis pass and consumed by another analysis pass even if these are in different address spaces. If package P imports Q, all facts about Q produced during analysis of that package will be available during later analysis of P. Facts are analogous to type export data in a build system: just as export data enables separate compilation of several passes, facts enable "separate analysis".`

> `The analysistest subpackage provides utilities for testing an Analyzer. ... Expectations are expressed using "// want ..." comments in the input code.`

> `The singlechecker package provides the main function for a command that runs one analyzer.`

**Factual description:** `go/analysis` is the official Go framework in which a static analysis is a
declarative `Analyzer` value (name, docs, flags, `Requires` dependencies, `FactTypes`) with a `Run`
function receiving a `Pass` (AST, type info, file set, package) and emitting `Diagnostic`s with optional
`SuggestedFix` text edits. Its distinguishing mechanism is **serialisable Facts** propagated along the
import graph, enabling modular, separately-cacheable cross-package analysis; `singlechecker`/`multichecker`
turn analyzers into standalone commands, and `analysistest` supports golden-file testing via `// want`
comments.

**Delimitation relevance:** this is the mechanism layer a Go annotation compiler such as speclink can build
on. The framework provides *no* requirement/traceability semantics; those must be supplied by an Analyzer.

### D.5 Java annotation processing (JSR 269 / `javax.annotation.processing`)

- **Publisher:** Oracle (Java SE 21 / JDK 21 API specification, module `java.compiler`).
- **URL fetched:** `https://docs.oracle.com/en/java/javase/21/docs/api/java.compiler/javax/annotation/processing/Processor.html`
- **Source type:** **official standard** (Java SE API specification)
- **Since:** 1.6

Verbatim quotes:

> `public interface Processor — The interface for an annotation processor.`

> `Annotation processing happens in a sequence of rounds. On each round, a processor may be asked to process a subset of the annotations found on the source and class files produced by a prior round. The inputs to the first round of processing are the initial inputs to a run of the tool; these initial inputs can be regarded as the output of a virtual zeroth round of processing.`

> `The tool uses a discovery process to find annotation processors and decide whether or not they should be run. ... the list of candidate processors to run can be set directly or controlled by a search path used for a service-style lookup.`

> `Processes a set of annotation interfaces on root elements originating from the prior round and returns whether or not these annotation interfaces are claimed by this processor. If true is returned, the annotation interfaces are claimed and subsequent processors will not be asked to process them`

> `If a processor raises an error, the current round will run to completion and the subsequent round will indicate an error was raised.`

> `To be robust when running in different tool implementations, an annotation processor should have the following properties: 1. The result of processing a given input is not a function of the presence or absence of other inputs (orthogonality). 2. Processing the same input produces the same output (consistency). 3. Processing input A followed by processing input B is equivalent to processing B then A (commutativity) 4. Processing an input does not rely on the presence of the output of other annotation processors (independence)`

**Factual description:** Java's standard annotation-processing SPI runs `Processor` implementations
inside `javac` over multiple *rounds*, discovered via a service-style lookup on the annotation processor
path; processors inspect annotated `Element`s, may generate new sources via the `Filer`, may *claim*
annotation interfaces, and may raise errors that fail the compilation. This is the closest existing
mainstream precedent for "annotations checked/expanded by the compiler", but the specification carries no
requirements-traceability semantics.

**Note on naming:** the Oracle page does not use the string "JSR 269"; it documents the API as
`javax.annotation.processing` in module `java.compiler`, `Since: 1.6`. The "JSR 269" designation is
therefore listed as UNVERIFIED below.

---

## E) Formal / design-by-contract for Go

### E.1 Gobra

- **Authors / project:** Viper project, ETH Zürich (Programming Methodology group).
- **URLs fetched:** `https://github.com/viperproject/gobra`
  (project site `https://gobra.ethz.ch` is linked from the repo's About section but was **not** fetched;
  `https://www.pegasus.ethz.ch/gobra.html` returned an ETH webarchive 404 and is **not** a valid source).
- **Source type:** **community content** (OSS project README, authored by an academic group)
- **Licence:** Mozilla Public License 2.0. 182 stars, 2,332 commits at fetch time.

Verbatim quotes (GitHub About + README):

> `Gobra is an automated, modular verifier for Go programs, based on the Viper verification infrastructure.`

> `Gobra is a prototype verifier for Go programs, based on the Viper verification infrastructure.`

> `We call annotated Go programs Gobra programs and use the file extension `.gobra` for them.`

> `Install Z3 and Boogie. ... Gobra uses the Silicon verification backend by default.`

> `## Projects verified with Gobra — [VerifiedSCION](https://github.com/viperproject/VerifiedSCION) — [Security of protocol implementations via refinement w.r.t. a Tamarin model] ... In particular, implementations of the signed Diffie-Hellman and WireGuard protocols have been verified.`

**Factual description:** Gobra is an automated, modular deductive verifier for Go, built on the Viper
infrastructure (Silicon/Carbon backends, Z3/Boogie), in which developers write **annotated Go programs**
(specifications embedded as annotations, `.gobra` files) that are discharged by an SMT solver. It is the
clearest existing example of "annotations on Go code checked by a tool", but the annotations are
*functional-correctness contracts*, not requirement identifiers, and the README self-describes it as a
prototype.

**Relevance for delimitation:** Gobra shows the annotation-on-Go-code design point already exists in the
verification space; speclink must be positioned against it on the axis of *what the annotation denotes*
(a requirement ID and its coverage obligation vs. a logical pre/postcondition) and *what the checker
guarantees*.

---

## F) ReqIF

- **Publisher:** Object Management Group (OMG). Founded 1989; described on the page as
  "an international, open membership, not-for-profit technology standards consortium."
- **URL fetched:** `https://www.omg.org/reqif/`
- **Source type:** **official standard** (SDO specification page)
- **Specification download pointer:** `https://www.omg.org/spec/ReqIF/About-ReqIF/`

Verbatim quotes:

> `Requirements Interchange Format™ (ReqIF™)`

> `The Requirements Interchange Format™ specification provides standards-based exchange of requirements authored in different Requirements Management (RM) tools; almost all RM and SysML modeling tools today support ReqIF import and export.`

> `For technical and organizational reasons, two companies in the manufacturing industry are rarely able to work on the same requirements repository and sometimes do not work with the same requirements authoring tools. A generic, nonproprietary format for requirements information is required to cross the chasm and to satisfy the urgent industry need for exchanging requirement information between different companies without losing the advantage of requirements management at the organizations' borders.`

> `Requirements Interchange Format (ReqIF) defines such an open, non-proprietary exchange format. Requirement information is exchanged by transferring XML documents that comply to the ReqIF format.`

> `The Object Management Group® (OMG®) is an international, open membership, not-for-profit technology standards consortium.`

**Factual description:** ReqIF is an OMG specification defining an open, non-proprietary XML format for
exchanging requirements between different requirements-management tools, motivated primarily by
supply-chain collaboration in manufacturing (automotive, aerospace, medical, defence). It standardises
*interchange of requirement information*, not any relationship to source code.

**Cross-check:** Siemens' Polarion page independently states `Built-in ReqIF enables lossless requirements
and test case specifications exchange with customers and suppliers` — vendor evidence for the OMG page's
claim of broad tool support. (Note the OMG page's own "almost all RM and SysML modeling tools today
support ReqIF" is an unquantified claim by the standards body and should be hedged if cited.)

---

## Summary: verified vs. unverified

### Verified in this session (URL fetched + verbatim quote captured)

| # | Item | Source type |
|---|------|-------------|
| A | YouTube video title, channel/author, upload date 2026-06-21, view count, full description & chapter list | community content |
| B.1 | GitHub Spec Kit: vendor (GitHub), MIT, workflow commands, "specifications become executable" thesis, extensions/presets model | vendor documentation |
| B.2 | Kiro: built by a team within AWS, three-file spec structure, three-phase workflow, Analyze Requirements | vendor doc + marketing |
| B.3 | Jama Software markets a "Spec-Driven Development" capability via MCP | vendor marketing |
| C.1 | IBM DOORS / DOORS Next: vendor IBM, multi-level traceability, baselines, e-signatures, ASPICE/ISO 26262/DO-178C claim | vendor marketing |
| C.2 | Siemens Polarion: LiveDocs paragraph-level identifiability, built-in ReqIF, SVN/Git traceability to change requests | vendor marketing |
| C.3 | Jama Connect: vendor Jama Software, traceability + compliance positioning, scale claims | vendor marketing |
| C.4 | PTC Codebeamer: vendor PTC, end-to-end traceability, OSLC-based digital thread, standards list, Git integration | vendor marketing |
| C.5 | OSLC published by **OASIS** (Open Project); RDF/LDP-based REST specs; several OASIS Standards (CM 3.0, PROMCODE 1.0, Config Mgmt 1.0) | official standard |
| D.1 | ArchUnit: Java library, bytecode analysis, rules as unit tests, TNG-supported, © Peter Gafert | community content |
| D.2 | `go vet`: Go toolchain built-in, check list, non-zero exit, explicitly heuristic | official Go docs |
| D.3 | Staticcheck: Dominik Honnef, 150+ checks, CI-oriented, complements `go vet` | community content |
| D.4 | `go/analysis`: Analyzer/Pass/Diagnostic/Fact model, modular cross-package Facts, singlechecker/multichecker, analysistest `// want` | official Go docs |
| D.5 | `javax.annotation.processing.Processor`: rounds, service-style discovery, claiming, error raising, robustness properties, Since 1.6 | official standard (Oracle) |
| E.1 | Gobra: Viper/ETH, automated modular verifier for Go, annotated `.gobra` programs, Silicon/Carbon + Z3/Boogie, VerifiedSCION | community content (academic OSS) |
| F | ReqIF published by OMG; open non-proprietary XML interchange format for requirements | official standard |

### UNVERIFIED — DO NOT USE

The following could **not** be confirmed from any source fetched in this session. Do not state them in
the paper without further verification.

1. **Any content of the YouTube video beyond its description.** The transcript and the audio/video body
   were not fetched. No claim about what is actually said, shown, or measured in the video may be made.
2. **The video being a conference talk.** No event, venue, or date-of-presentation was found. It appears
   to be a self-published channel video. Do not attribute it to a conference.
3. **David Tielke's institutional affiliation.** Only a commercial personal site (`david-tielke.de`) with
   workshop offerings was evidenced. No employer/university verified.
4. **The identifier "JSR 269".** The Oracle Java SE 21 API page documents `javax.annotation.processing`
   without using the string "JSR 269". The JCP JSR page was not fetched.
5. **OSLC being an OMG standard.** The task brief said "OMG/OASIS"; the fetched OSLC site attributes it
   solely to **OASIS Open**. No OMG involvement was verified. Attribute OSLC to OASIS only.
6. **Any peer-reviewed publication about Gobra.** Only the GitHub README was fetched. The Gobra papers
   (and `https://gobra.ethz.ch`) were NOT fetched; do not cite a Gobra paper until its DOI/venue is
   verified.
7. **IBM DOORS technical documentation.** Only the marketing product page was fetched
   (`ibm.com/products/...`). The IBM Docs pages (`ibm.com/docs/en/engineering-lifecycle-management-suite/doors-next`)
   were **not** fetched; no statement about DOORS' actual link-type model, OSLC support, or ReqIF support
   may be made.
8. **Jama Connect's and Codebeamer's actual traceability data models.** Only marketing pages were
   fetched, not product documentation. All statements must be hedged as vendor claims.
9. **Comparative/market claims** appearing on vendor pages — e.g. PTC's "Why PTC Is #1 in ALM" /
   QKS Spark Matrix leadership, Jama's G2 "Leader" badges, Kiro's "We pioneered spec-driven development",
   Polarion's "An exclusive innovation you won't find elsewhere", and OMG's "almost all RM and SysML
   modeling tools today support ReqIF". These are **vendor marketing claims**; do not repeat as fact.
10. **ArchUnit's current latest version.** The site simultaneously links v1.5.0 API docs and lists
    v1.4.2 (Apr 18, 2026) as the newest news item. Do not state a version number.
11. **Whether any of the B/C tools performs compile-time or build-failing enforcement of
    requirement↔code links.** No fetched source describes such a mechanism — but absence of evidence in
    marketing pages is not evidence of absence. Phrase as "no such mechanism is documented on the
    vendor's product pages", not as "tool X cannot do this".
12. **Any "Go contracts" research proposal.** Not searched or fetched in this session. Nothing may be
    said about it.

### Suggested follow-up fetches (to close gaps)

- `https://www.jcp.org/en/jsr/detail?id=269` — to confirm the JSR 269 designation.
- `https://gobra.ethz.ch` and the Gobra paper venue/DOI (likely a formal-methods conference) — for a
  peer-reviewed citation.
- `https://www.ibm.com/docs/en/engineering-lifecycle-management-suite/doors-next` — for actual DOORS Next
  link-type / OSLC documentation rather than marketing.
- `https://www.omg.org/spec/ReqIF/About-ReqIF/` — for the exact current ReqIF version number and date.
- `https://open-services.net/specifications` — for the exact list and status of OSLC domain specs.
- Peer-reviewed literature on requirements traceability (Gotel & Finkelstein; Cleland-Huang et al.) —
  none of section C's claims currently rest on academic sources.
