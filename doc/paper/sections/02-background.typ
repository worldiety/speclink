= Background and Related Work <sec:background>

speclink sits at the intersection of four lines of work that have so far stayed
apart: requirements traceability, specifications embedded in program text,
architecture conformance checking, and documentation-drift detection. This
section reviews each and states, for each, what speclink borrows and where it
deliberately differs.

== Requirements traceability

The modern framing of the problem is due to
#cite(<gotel1994>, form: "prose"), who analysed requirements traceability on
the basis of an empirical study of over 100 practitioners and concluded that it
is a multifaceted problem for which no all-encompassing solution exists. Their
separation of pre-requirements-specification from
post-requirements-specification traceability is the distinction that matters
here: speclink is concerned exclusively with the post-RS direction, from a
recorded requirement to the program construct written for it.

#cite(<ramesh2001>, form: "prose") derived reference models for traceability
empirically, from focus groups and interviews across 26 organisations, and
distinguished low-end from high-end traceability users. Their work establishes
that traceability practice is not uniform and that the useful unit of
description is a model of what is linked to what, rather than a tool feature
list.

The community consolidated this line in an edited volume @clelandhuang2012,
which contains both a chapter on traceability fundamentals
@gotel2012fundamentals and the community's grand-challenge roadmap
@gotel2012grand; both are cited here as pointers to that programme rather than
for any specific content. The accompanying conference paper is more direct:
#cite(<gotel2012quest>, form: "prose") report that, despite advances in
automated trace-link creation and maintenance, traceability implementation and
use is "still not pervasive in industry".

That gap is worth closing because the benefit is measurable: the controlled
experiment of #cite(<maeder2012>, form: "prose") quoted in @sec:introduction
found subjects working with traceability both faster and more often correct; a
journal version of the experiment exists @maeder2015. Traceability *quality*, not merely its
existence, also matters: #cite(<rempel2017>, form: "prose") studied 24
open-source projects and report, using multi-level Poisson regression, that
more complete traceability is associated with a decreased expected defect rate
for three of the four studied requirements-implementation supporting
activities.

speclink takes these results as its motivation, and takes the "not pervasive"
observation as a design constraint rather than an exhortation. Its response is
that a trace which is optional will be omitted, so the trace is made a
precondition of a successful build.

== Recovering trace links versus declaring them

A large part of the traceability literature is concerned with *recovering*
links that were never recorded. #cite(<antoniol2002>, form: "prose") proposed
information-retrieval based recovery — probabilistic and vector-space models —
of links between source code and free text, premised on programmers using
meaningful identifier names, and themselves discuss the limitations of the
approach; a retrospective on that paper appeared in 2025 @antoniol2025.
#cite(<hayes2006>, form: "prose") formulated requirements tracing as
candidate-link generation with an analyst in the loop, defining goals and
measures and presenting the RETRO prototype. The analyst is not incidental
there: the tool proposes, a human disposes.

#cite(<deLucia2007>, form: "prose"), evaluating IR-based traceability recovery
over seventeen software projects involving about 150 students, state the
limitation plainly: such tools "are still far to support a complete
semi-automatic recovery of all links". That sentence is the anchor for
speclink's positioning.

speclink recovers nothing. It does not infer which requirement a construct was
written for, and it does not offer candidate links for an analyst to accept. It
requires the link to be declared, in the code, as a term the Go compiler
already type-checks — `spec.Satisfies(...)` inside a
`<base>.annotation.go` sidecar file — and it fails the build when a construct
that carries business meaning has no such declaration (README §4.2); the
corresponding directions are named in @sec:evidence. The trade is explicit and unfavourable in one
direction: an approach that refuses to guess cannot be pointed at a legacy
system to obtain a trace matrix, which is precisely what recovery techniques
are for. What it buys is that every link in the resulting record was written
down by whoever wrote the code, and that a missing link is a build failure
rather than a low-ranked candidate.

== Specifications embedded in code

Placing specifications next to the code they describe is an old idea.
#cite(<meyer1992>, form: "prose") introduced Design by Contract as
methodological guidelines for reliability, realised in Eiffel with assertions
as the mechanism. The Java Modeling Language continues that line:
#cite(<leavens2006>, form: "prose") embed pre-conditions, post-conditions and
assertions intermixed with Java source code, written in Java expressions so
that working engineers can write them; an earlier version of the notation
exists @leavens1999, as does a journal overview of the associated tooling
@burdy2005 @burdy2003entcs. For Go, the Gobra project describes itself as a
prototype automated modular verifier built on the Viper infrastructure, in
which specifications are written as annotations in `.gobra` programs and which
has been applied to VerifiedSCION and WireGuard @gobra; the project is not a
peer-reviewed source. In a different style, #cite(<jackson2002>, form: "prose")
shows that a small declarative language for structural properties can be given
a syntax "amenable to a fully automatic semantic analysis".

speclink shares the placement and nothing else. Contracts, JML and Gobra
specify *behaviour* and aim at (partial) formal verification of that behaviour.
speclink specifies *provenance*: which requirement a construct exists for. Its
annotations are ordinary Go terms, and what it checks is structure — that the
named requirement exists, that the construct is of the kind claimed, that the
binding targets a declaration the compiler has resolved, that a promised wire
shape has not changed incompatibly. It does not check that the code implements
the requirement, and it is not a verifier. Where behavioural evidence enters at
all, it enters from outside: a passing test run is fed back through
`speclink evidence`, because a test's claim to verify a requirement is not
itself evidence that it did (README §1, §2).

== Architecture conformance and static analysis tooling

The reflexion-model technique @murphy1995 compares a high-level model with the
source that is supposed to realise it. #cite(<murphy2001>, form: "prose") state
the underlying observation that "the artifacts constituting a software system
often drift apart over time", and summarise consistency between a high-level
artefact and the source, in an application to roughly one million lines of
Microsoft Excel. A body of work follows on rule-based conformance checking as a
quality-management measure @herold2014, static conformance checking against
architecture erosion @deSilva2015, conformance checking for Python
@deLima2020, model checking applied to conformance @menezes2021, and
conformance checking for infrastructure as code @ozkaya2023. These are cited
only as pointers to an active area; their subject is the conformance of code to
an architecture rather than to a requirement, and no claim is made here about
what any of them can or cannot do.

At the practical layer, ArchUnit is a Java library that evaluates structural
architecture rules over packages, classes, layers, slices and cycles as
ordinary unit tests by analysing bytecode, with a #box[.NET]/C\# port; it is a
community project, not peer-reviewed work @archunit. In the Go ecosystem,
`go vet` ships with the toolchain, exits non-zero when a problem is reported
and is therefore usable as a CI gate, but its own documentation describes it as
heuristic guidance rather than a correctness guarantee @govet. Staticcheck adds
a linter with over 150 checks aimed at bugs, performance, simplifications and
style, designed for low false-positive rates and CI use, complementing
`go vet` @staticcheck. Both build on the `go/analysis` framework, whose
`Analyzer`/`Pass`/`Diagnostic` model performs modular per-package analysis and
whose `Fact` mechanism propagates information along the import graph, enabling
"separate analysis"; the framework carries no requirement or traceability
semantics of its own @goanalysis. The Java analogue at compile time is
annotation processing, in which `Processor` implementations run inside the
compiler over rounds, are discovered by service-style lookup, may claim
annotations and may raise errors that fail compilation @javaap.

speclink is an analysis of this practical kind — it recognises constructs and
reports findings over a Go program, and, like annotation processing, its
findings stop the build. What it adds is the subject matter: the rules are
about requirements and their bindings, which none of these frameworks supplies.

== Documentation drift

#cite(<stulova2020>, form: "prose") observe of existing analysis tooling that
"none of these tools checks for consistency of the documentation accompanying
the code", and present upDoc with an evaluation the authors themselves describe
as preliminary. Detection approaches in this space are typically heuristic:
#cite(<liu2018>, form: "prose") report heuristic detection of outdated comments
achieving 74.6% detection and 77.2% precision using 64 features.

speclink attacks the same drift, but not by inference. A requirement text and
the source-document segment it was derived from are edges the compiler cannot
re-derive, so they are recorded in `speclink.lock`, a file written by
`speclink freeze` and never edited by hand, standing to the sources as
`go.sum` stands to `go.mod` (README §6). Each recorded segment carries a
fingerprint — over the text of a Markdown heading section, or over the pixels
inside a declared rectangle of a mockup image. When the underlying text or
image changes, the fingerprint no longer matches and the affected requirements
are reported (`K13-SOURCE-DRIFT`); the same mechanism reports a review that
outlived the declaration it covered (`K18-REVIEW-STALE`). The result is
deterministic rather than heuristic: there is no detection rate, because the
comparison is an equality test on a recorded value, and the reviewable artefact
is the diff of `speclink.lock`.

== Requirements management tools and interchange

Industrial practice is dominated by dedicated requirements-management
platforms. IBM markets DOORS and DOORS Next with structured requirement
modules, baselines, electronic signatures, multi-level traceability and a
graphical link explorer, and claims support for Automotive SPICE, ISO 26262 and
DO-178C @doorsnext. Siemens claims for Polarion traceability via automatic
change control, "full traceability of every source code modification up to the
change request", paragraph-level identifiability in LiveDocs, built-in ReqIF
and out-of-the-box SVN and Git support @polarion. Jama markets requirements
traceability, standards compliance and audit trails, together with a
spec-driven-development capability exposed over MCP @jama. PTC claims for
codebeamer end-to-end traceability across work items in a centralised
repository, OSLC-based digital-thread integration and a range of standards and
tool integrations @codebeamer. All four statements are the vendors' own
marketing claims.

Interchange between such repositories is standardised. ReqIF is an OMG
specification defining an open, non-proprietary XML interchange format for
requirements, motivated by cross-company supply-chain exchange @reqif. OSLC is
an OASIS Open Project — not an OMG standard — defining REST and Linked-Data
specifications, based on RDF and the W3C Linked Data Platform, for linking
lifecycle resources; its Change Management 3.0 and Configuration Management 1.0
specifications are OASIS Standards @oslc.

The positioning point is one of location, not capability. These platforms hold
the trace in a repository beside the code and connect to it through
integrations and interchange formats; speclink holds the trace in the code, as
compiler-checked Go terms, and makes an incomplete trace a failing build. No
claim is made here that these tools cannot enforce a trace at build time — only
that no such mechanism is documented on the vendors' product pages, and that
the artefact under review differs: a repository record on one side, a source
diff and a non-zero exit code on the other.
