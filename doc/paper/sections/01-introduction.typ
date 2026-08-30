= Introduction <sec:introduction>

A software project in which a large share of the code is produced by a machine
changes what a review can be. When the volume of newly written code exceeds what
the responsible engineers can read, review by sampling becomes the working
model rather than a lapse from it. The traditional guarantee that a change
carries — that at least one person understood this code and judged it against
what was asked for — no longer follows from the fact that the change was merged.
It has to be established some other way, or not at all.

The evidence that this matters is not speculative. In a controlled experiment
with 52 subjects, #cite(<maeder2012>, form: "prose") found that subjects working
with traceability were on average 21% faster and produced 60% more correct
solutions than subjects without it. Across 24 open-source projects,
#cite(<rempel2017>, form: "prose") report that more complete traceability is
associated with a decreased expected defect rate for three of the four
development activities they studied. On the other side of the ledger,
#cite(<perry2023>, form: "prose") found that participants with access to an AI
assistant wrote significantly less secure code and were at the same time more
likely to believe their code was secure. Delegating the judgement to another
model is not an escape either: #cite(<panickssery2024>, form: "prose") show that
large language model evaluators exhibit self-preference, scoring their own
outputs more highly than human annotators do, correlated with their ability to
recognise those outputs as their own. A generator that certifies its own work is
not a control.

Traceability from requirement to code therefore becomes the load-bearing
artefact of such a project. It is also the artefact that is conventionally
maintained furthest from the code: in a requirements management tool, as a graph
of links between identifiers that the compiler never sees. Such a link is a
string, and strings rot. Rediscovering the links after the fact has been studied
for over two decades — information-retrieval recovery and analyst-in-the-loop
candidate-link generation, reviewed in @sec:background — and the limits are
documented by the same line of work @deLucia2007. A decade later,
#cite(<gotel2012quest>, form: "prose") observe that despite advances in
automated trace-link creation and maintenance, traceability implementation and
use is still not pervasive in industry. The drift is not confined to trace
links; the reflexion-model and documentation-consistency literature discussed in
@sec:background reports the same divergence between an artefact and the code
that is supposed to realise it @murphy2001 @stulova2020.

== The irreducible fact

speclink starts from an observation about what a static analysis can and cannot
know. Given a well-structured application, an analysis can infer a great deal:
that this function type is a use case, that this struct is an event, that this
type is an aggregate, that this interface is a repository, that this constant is
a permission. These are structural facts, and a tool that reads the program can
derive them without being told. The one fact that no analysis can ever infer is
*which requirement a construct was written for*. That is a fact about intent,
and it exists nowhere in the program text.

The design consequence is to shrink the annotation surface to exactly that one
irreducible fact, and to make annotating anything inferable an error
(README §1, rule 4). Everything the tool can derive, it derives; everything it
cannot, the author states once. Because the surface is that small, it can be
made compiler-checked rather than textual: a binding names a requirement by Go
identifier, so the Go compiler itself proves the referent exists before speclink
ever runs. A renamed or deleted requirement breaks the build instead of silently
degrading into a dangling link.

== speclink

speclink is an annotation compiler for Go, with a front end that reads JVM
bytecode rather than source. Because Kotlin compiles to JVM bytecode, Java and
Kotlin projects arrive at that front end in the same shape; one JVM profile,
`java_springboot_ddd1`, exists (README §11a). It (a) recognises constructs — use
cases, commands, events, aggregates, permissions, queries, projections and
repositories — from the program itself; (b) requires every normative requirement
to be bound to at least one construct through an annotation whose references the
Go compiler proves exist; (c) anchors each requirement to an addressable segment
of a human-written source document, a Markdown heading or a named region of a
mockup (README §10a); (d) joins that chain to actual passing test runs by
reading the test stream the build already produces (README §10b); and (e)
records authorship and human review from outside the code, through
`speclink attest` and `speclink freeze -reviewer` (the two are distinguished in
@sec:evidence), rather than letting the code declare it. Every finding is an
error; there are no warnings and no tolerance mode, and the only escape hatch is
a waiver that must carry a written reason (@sec:design).

== Claim versus evidence

The paper's central conceptual contribution is a distinction that the design
applies everywhere. A *claim* is something the source says: a binding stating
that a construct satisfies a requirement, a marker stating that a test verifies
one, a field stating that a human reviewed something. A claim is readable
statically, which is what makes a *missing* claim reportable. *Evidence* is a
record that something actually happened: a test that ran to the point of the
marker and passed, an attestation made by a party that is not the generator.
A marker in the source can sit behind a condition that never holds, or in a test
that fails long before reaching it; the source cannot tell the difference.
speclink therefore reports the two figures separately, and the gap between them
is the number that carries information (@sec:evidence).
The same distinction is why `spec.Requirement` has no `Reviewed` field: a claim
of human authorship or human review made by the author of the code is not
evidence about the author.

== Contributions

+ A design principle for annotation languages: annotate only what no analysis
  can infer — the requirement a construct was written for — and treat the
  annotation of any inferable fact as an error.
+ A compiler-checked binding mechanism in which requirements are ordinary
  program declarations and references to them are resolved by the host
  language's own compiler, so a trace link cannot decay into a stale string.
+ An explicit separation of claim from evidence, applied uniformly to
  verification, authorship and review, together with the reporting model that
  follows from it: two figures, and the gap between them.
+ An anchoring of the requirement tree to segments of the human-written source
  documents, including named regions of mockups, with fingerprint-based drift
  detection over exactly the text or pixels a requirement was extracted from.
+ An implementation covering a Go front end and a JVM bytecode front end, in
  which the same requirement model is populated from two different levels of
  program representation.
+ A discipline of measurement honesty: a direction that was not measured is
  reported as not measured rather than as measured clean, and the corresponding
  rules are not run at all.

== Scope and limits

This paper reports a design and an implementation. speclink is a single-tool,
single-organisation artefact, developed alongside one application framework, and
the paper presents no controlled empirical evaluation: there is no comparison
against a baseline, no user study, and no measurement of defect rates in
projects that adopted it. The claims made here are claims about what the tool
enforces and why it is built the way it is, not claims that it improves any
outcome. The empirical results cited above motivate the problem; they do not
validate the artefact.

The remainder of the paper first surveys traceability research, requirements
management tooling and adjacent static-analysis approaches, and positions
speclink among them. It then presents the requirement model and the annotation
language, the construct recognition that keeps the annotation surface small, the
anchoring of requirements to source documents and its drift detection, and the
evidence pipeline that joins the chain to test runs. A discussion of the JVM
front end shows what the design costs when the host language differs. The paper
closes with the limitations of the approach and the threats to the argument.
