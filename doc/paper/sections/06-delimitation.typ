= Delimitation from Spec-Driven Development Practice <sec:delimitation>

A practitioner movement has grown around the idea that a written specification
should drive the generation of code by a language model. It travels under the
name *spec-driven development*, and it is documented almost entirely in vendor
repositories, product documentation, marketing pages and self-published video.
It is not documented in peer-reviewed literature. This section says so
explicitly, and then says why the movement still has to be addressed: it is
what a growing number of practitioners actually do, and a paper that proposes a
different place for the specification owes its reader a precise statement of
what is and what is not being claimed relative to that practice.

One methodological remark governs everything below. None of the sources cited
in this section is peer-reviewed. Two are vendor documentation, one is vendor
marketing, one is a self-published video. Consequently no statement here
supports an absolute claim about the state of the art, and every statement is
attributed to the source that makes it. Where a capability is not mentioned by
a source, that is recorded as *not documented*, never as *absent*.

== Generator-side specification

GitHub publishes Spec Kit, an MIT-licensed toolkit whose own README states that
spec-driven development makes specifications "executable", "directly generating
working implementations rather than just guiding them" @speckit. According to
that documentation, a project is taken through a fixed sequence of
slash-commands — establishing project principles, then specify, plan, tasks,
implement and converge — with `/speckit.analyze` described as cross-artifact
consistency and coverage analysis run between the task breakdown and the
implementation @speckit. The same README describes regulatory traceability and
V-model test traceability not as built-in behaviour but as something a preset or
a community extension *could* add @speckit.

Amazon's Kiro is documented by its vendor as being built and operated by a team
within AWS @kiro. Per that documentation its specs feature generates three
Markdown artefacts for each feature or bug fix — `requirements.md`, `design.md`
and `tasks.md` — capturing user stories and acceptance criteria, technical
design, and a plan of discrete trackable tasks respectively, with an
LLM-based "Analyze Requirements" step offered to catch inconsistencies,
ambiguities and gaps before design @kiro. Jama Software, a requirements-management
vendor, markets a spec-driven-development capability in which engineers and AI
engineering agents iterate in a shared context over MCP @jama; this is a
marketing statement on a product page, and is reported here only to show that
the label has been adopted on both sides of the tooling landscape.

The analytic point is one of position, and it is the whole of this section.
All three systems place the specification *in front of* the generator. The
specification is an input to code production: it is authored, it is analysed,
and it is then consumed by an agent that emits code. speclink places the
specification *behind* the compiler. A `spec.Requirement` value and the
`spec.Satisfies` binding that names it are ordinary Go terms in the ordinary
build, and `speclink verify` is a build step that runs after the code exists and exits
non-zero on any finding; the four directions defined in @sec:evidence must all
reach a hundred per cent for a run to be clean (README §2). It is indifferent to who or what wrote the code.

The two arrangements are complementary rather than competing, and a project
could reasonably use both: a generator-side workflow to get from an idea to a
first implementation, and a build-side check to keep the correspondence true
afterwards. What separates them is duty over time. A specification consumed by a
generator has discharged its role once the code has been emitted; structurally
it has no further obligation, and nothing in the build notices if the code and
the document diverge on the next commit. A specification that is a precondition
of the build has an obligation at every subsequent commit, because the check is
re-run and the failure is a red build rather than a stale document.

A second difference concerns who is trusted. Where an artefact is produced by a
generator and then judged by the same generator, the judgement carries no
independent weight, for the self-preference reason set out in
@sec:introduction @panickssery2024. This is the reason speclink takes authorship
and review records from outside the code entirely: `speclink attest -origin llm`
and `speclink attest -reviewer` are recorded by the caller, never declared in the
source, precisely because the same machine that writes the code would otherwise
write the claim about who wrote it (README §2). speclink is explicit that this
only shifts the question to who is permitted to make those calls, which is not
speclink's to decide.

It should be stated plainly what speclink does not do. It does not generate
code. It does not prompt a model. It takes no position on how the code came to
exist, and offers no workflow, template or command for producing it.

== A self-published field report

One frequently circulated practitioner account requires careful handling. The
video `Der Moment, der die Softwareentwicklung geändert hat!` was published by
David Tielke on his own YouTube channel and uploaded on 21 June 2026
@tielke2026. It is not peer-reviewed, it names no venue or event, and it has
undergone no independent review. Only the video description and its chapter list
could be retrieved; the transcript was not retrievable, and the summary that
follows is therefore derived solely from the description text.

From that description: the author reports a self-conducted, uncontrolled
experiment in which an application he had used for eighteen years was rebuilt
as an enterprise system — described as having a microservice architecture, local
AI, a workflow engine, a full specification, tests and documentation — largely
with AI agents. Per its own chapter list the work is staged through
micro-management, quality-driven and spec-driven phases, then test-driven
development from requirements, and finally idea- and voice-driven development,
closing with self-reported figures on scope, code quality, cost and speed
@tielke2026.

The methodological point follows without condescension. Self-reported figures
from an uncontrolled single-subject experiment cannot establish anything, and
this paper draws no quantitative claim from the report. What the account does
legitimately illustrate is that the practitioner community has converged on
specification-first workflows for agentic development — the same convergence
visible in the vendor documentation above. That convergence is what motivates
the question this paper addresses. Given such a workflow, what remains checkable
by a machine? The answer speclink offers is: the trace, the structure and the
evidence — that a requirement exists and is reachable from a source segment,
that a construct names one, that a test claimed one and ran. Not the intent.
Whether the code does the right thing remains a human judgement, which is why
the review record is kept and why it is kept outside the code.

Honesty requires the symmetric observation. This paper presents no controlled
experiment either — indeed no empirical evaluation at all, as @sec:discussion
states without qualification. It reports a design and an implementation built and
used inside one organisation by the people who wrote the tool. The criticism of
the video's method therefore applies, in modified form, to this artefact as well,
and no comparative claim about productivity, quality or cost is made anywhere in
this paper.

== Summary of positioning

@tbl:positioning contrasts the three arrangements along the dimensions that
distinguish them. Cells read *not documented* wherever the fetched sources are
silent, which is the only supportable phrasing for an absence in a vendor
document.

#figure(
  placement: top,
  scope: "parent",
  table(
    columns: 4,
    align: left,
    table.header(
      [*Dimension*],
      [*Generator-side spec tooling*],
      [*Requirements-management tooling*],
      [*speclink*],
    ),
    [Where the specification lives],
    [Markdown artefacts in the repository @speckit @kiro],
    [Items in the vendor's own repository @jama],
    [Go values in `<ID>.spec.go` and bindings beside the code],

    [When it is checked],
    [During the agent workflow, before implementation @speckit @kiro],
    [Not documented],
    [After the Go build, as a build step],

    [What happens on violation],
    [Not documented],
    [Not documented],
    [Non-zero exit code; every finding is an error],

    [What is checked],
    [Cross-artifact consistency and coverage, by the agent @speckit; requirements analysed for gaps and ambiguities @kiro],
    [Links between items; compliance and audit trails @jama],
    [Source segments accounted, constructs bound, requirements covered and verified],

    [Who is trusted],
    [The generating agent],
    [The tool's users and its audit trail @jama],
    [The compiler, plus records taken from outside the code],

    [Coupling to language or stack],
    [Language-independent Markdown @speckit @kiro],
    [Language-independent; source control integrated externally @jama],
    [Bound to the language: a profile per language, framework and style],

    [Legacy applicability],
    [Not documented],
    [Not documented],
    [Package by package, via `scope`],
  ),
  caption: [Positioning of speclink against generator-side spec tooling and
    requirements-management tooling. Every cell attributed to a vendor source
    reports that vendor's own documentation or marketing.],
) <tbl:positioning>
