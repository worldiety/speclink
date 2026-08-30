= Discussion and Limitations <sec:discussion>

The preceding sections argue that a trace link can be made a build
precondition. This section states what that buys and what it does not.

== What a green run means

A green `speclink verify` is a conjunction of the four directions
(@sec:evidence), asserted about one working tree at one commit. Every segment of every source document either yielded a
requirement or was marked `informative`. Every recognised construct that must
name a requirement names one, and the Go compiler resolved that name. Every
`Normative` requirement is named by at least one construct, and no `Abstract` or
`Superseded` requirement is. And for every normative requirement there exists a
`spec.Verified` call, with `speclink evidence` reporting separately how many of
those calls were observed to run to completion inside a passing test.

Each statement is about the *existence and shape of a link*, and none is about
the behaviour of the program. The README states the boundary in one sentence:
coverage says code was written for a requirement, and "It has never said the
code does what the requirement asks" (README §10b). Evidence narrows the gap by
one step and no more. A recorded run means that a test containing
`spec.Verified(t, R)` reached that call and then passed. It does not mean the
test asserted anything relevant to `R`, that its assertions were strong, or that
`R` was the requirement the stakeholder needed. A test whose body is empty but
for the marker produces the same evidence record as a thorough one; speclink has
no notion of test adequacy and implements no coverage measure over the code
under test.

The guarantee is therefore: *the artefacts are connected, the connections
type-check, the wordings have not moved since somebody looked at them, and a
passing test claimed each requirement.* Nothing weaker, and — this is the half
that a design paper is tempted to slide past — nothing stronger. The
verified/demonstrated pair of @sec:evidence reports an uncertainty about
*execution*, not about *adequacy*; there is no further figure, and the design
offers no path to one.

== The unchecked link at the top

The step from a human-written source segment to a `spec.Requirement` has no
formal semantics and cannot be given one, because it is a translation between a
natural-language artefact and a typed value, and no property of the pair is
decidable. Everything below it is conditional on it. If a requirement misstates
the paragraph it was extracted from, then the compiler-checked binding, the
coverage figure and the evidence record are all sound statements about a false
premise, and they will all be green.

speclink offers four mitigations, and none of them is a check on the extraction
itself. Segment coverage (`K12-SOURCE-UNCOVERED`) forces every part of every
document to be *accounted for*, so nothing is silently skipped — but a segment
is accounted for by any requirement, including a wrong one. The `informative`
marker moves the decision that a segment carries no obligation into the
document, next to the person who wrote it — but nobody checks that the marker is
deserved. Fingerprints (`K13-SOURCE-DRIFT`, `K10-REQ-CHANGED`) establish that
the segment still reads as it did when it was last frozen — which detects
*change*, never *misreading*; a requirement wrong from the first commit is
frozen wrong and stays green forever. Review recorded by
`speclink freeze -reviewer` records that a named person saw a specific
requirement wording (@sec:evidence) —
a record of an event, not a judgement about its quality, and speclink checks
neither who made the call nor whether they read anything. Mitigation is not
verification, and the four together do not amount to partial verification of a
step that admits none.

The sharpest failure mode is closure. If one model both extracts the
requirements from the source documents and writes the code that satisfies them,
then the requirement tree and the implementation share an author and an
interpretation. A misreading of the source document propagates coherently into
both sides of a link that speclink then certifies as intact. The result is a
system that is internally self-consistent, wrong with respect to what the
stakeholder wrote, and verifies perfectly. Routing the check to a second model
is not an escape, for the self-preference reason given in
@sec:introduction @panickssery2024. speclink's answer is to keep the review record outside the
code and to print the pairing of requirement and segment for a person to read.
That is an invitation to look, not a control.

== The binding is a judgement

`spec.Satisfies(R)` is checked in exactly one respect: that `R` exists and is a
requirement. That is the Go compiler's work, and it is all the compiler can do.
Whether *this* construct is the one that satisfies *this* requirement is a human
or machine judgement expressed as a well-typed term, and a wrong binding is
indistinguishable from a right one at every stage of the pipeline. Binding an
arbitrary use case to an arbitrary normative requirement removes a
`K3-REQ-UNCOVERED` finding and turns the build green. The mechanism that
prevents a trace link from decaying into a stale string does nothing whatsoever
to establish that the link was apt when it was written.

== Structural coupling

speclink enforces an architecture as much as it records a trace. It is bound to
a project layout (`<ID>.spec.go`, `<base>.annotation.go` beside `<base>.go`,
`requirements/_sources`), to a profile naming a language, a framework and a
style, and to a closed set of recognised constructs. The `K4`–`K8` families are
framework-specific by construction and live in the Go front end. A codebase that
does not organise itself this way is not partially supported; it is
unrecognised, and its constructs simply do not appear.

The cost is generality, and the traceability literature is where that cost is
felt. Link recovery, reviewed in @sec:background, exists precisely because most
systems have no links @antoniol2002 @hayes2006, and its own evaluations report
that it does not recover them completely @deLucia2007. speclink does not address
that problem at all. Its `scope` key narrows adoption package by
package, but the destination is still the layout. The answer speclink offers is
available only to projects willing to be built its way — a genuine restriction
on the class of systems the approach applies to, and one that excludes most
existing ones.

== Waivers and the incentive to stay silent

The waiver mechanism of @sec:design turns on a mandatory free-text reason, and
mandatory free text is a weak control. speclink does not check, and cannot
check, whether a reason is good; any non-empty string satisfies it. In a project
where a machine is asked to make a build green, a waiver is the cheapest
available move, and nothing in the design makes it expensive. The counter-
pressures are structural rather than semantic: a waiver appears in the generated
report and counts toward the figures, so it is visible in review; there are no
severities, so there is no quieter place to move a finding to; and a waiver
attaches to a Go construct, so it sits in the diff next to the code it excuses.
Whether these pressures actually suppress abuse is unmeasured. No waiver counts
from any project are reported in this paper, and no study of reviewer behaviour
under this mechanism has been carried out.

== Absence of empirical evaluation

This is the central limitation and it is unqualified. There is no controlled
study, no comparison against a baseline, no multi-project dataset, and no
measurement of anything. It is not shown that links maintained this way are more
accurate than links kept in a requirements-management tool, nor cheaper to
maintain, nor that they are maintained at all beyond the projects the authors
control. It is not shown that a build-time failure produces a repair rather than
a waiver. It is not shown that developers can read the diagnostics.

What such an evaluation would look like is known. The controlled experiment of
#cite(<maeder2012>, form: "prose") cited in @sec:introduction — subjects, tasks,
a control condition, an effect measured — is a fair reference point for what is
missing here. #cite(<rempel2017>, form: "prose") is a second: 24 open-source
projects and a multi-level regression relating traceability completeness to
defect rate. Nothing of either kind was attempted.

The experience behind this paper is single-subject and non-independent. The tool
has been used inside one organisation, on that organisation's own codebase, by
the people who wrote the tool. @sec:delimitation criticises a self-published
practitioner field report for being an uncontrolled single-subject experiment
with self-reported figures @tielke2026. That criticism applies to this artefact
in the same form and with the same force, and the symmetry is not rhetorical
politeness: every figure in @sec:implementation, including the constructor-naming
counts from the reference ERP, was measured by the tool's authors on their own
project with their own tool, and carries exactly the evidential weight that
arrangement allows.

== Tool trust and self-application

speclink produces evidence about other software while being unqualified software
itself. @sec:standards says so; it bears repeating, because the claim/evidence distinction
the design applies to its subjects applies to it. A finding is a claim by
speclink, and nothing external corroborates it. `go vet`'s own documentation is
explicit that its checks are heuristic guidance rather than a guarantee of
correctness @govet; speclink's rules are in the same category and have not been
reviewed by anyone outside the project.

Self-application is worse than partial. The repository contains no
`speclink.json`, no `requirements/` tree and no `speclink.lock`; the only
`.spec.go` and `.annotation.go` files in it are test fixtures under `testdata/`.
speclink is therefore not governed by speclink at any level. The tool's rules
about requirement coverage, evidence and review provenance have never been
applied to the codebase that implements them, and no
`speclink selfreport` command exists in `cmd/speclink`, so the capability to do
so does not exist — README §12 also lists it as absent, though that section is
stale on another point (@sec:implementation).
A tool that does not consume its own output is making a claim about
maintainability under its regime that it has not tested on the one project it
fully controls.

== Threats to the validity of this paper's claims

Four, briefly. First, the comparison to functional-safety and aviation standards
in @sec:standards rests on publicly visible metadata — titles, scope lists, catalogue
descriptions and Online Browsing Platform definitions — because the normative
text is paywalled. No clause content was read, and a structural affinity argued
from a table of contents is a weak argument. Automotive SPICE is the sole
exception, its PAM being freely published @aspice40. Second, every statement
about the requirements-management tool landscape derives from vendor
documentation and marketing pages, is attributed as such, and can report only
that a mechanism is *not documented* — never that it is absent. Third, the
artefact has a three-week commit history (@sec:implementation); no rule, no intermediate
representation and no lock-file format in it has survived contact with a
long-running project, which is precisely the setting the design exists for.
Fourth, this paper was largely produced by a language model under human
direction; @sec:disclosure sets out what was done by a machine, what was done by
a person and what controls were applied. That arrangement places the paper in the
same category as the artefact it describes — generated text, sampled review.
