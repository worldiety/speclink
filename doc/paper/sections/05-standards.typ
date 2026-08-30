= Relation to Normative and Process-Assessment Practice <sec:standards>

The normative text of most of the documents discussed here is not publicly
available. ISO 26262, IEC 61508 and IEC 62304 are sold per part; DO-178C,
DO-330, DO-333 and their EUROCAE counterparts are sold by RTCA and EUROCAE.
What is publicly verifiable is a narrow set: part titles and scopes, the clause
*titles* that the publishers expose in free previews, the definitions published
in the ISO Online Browsing Platform, and the descriptive text the publishers
themselves write on their catalogue pages. The one exception is Automotive
SPICE: the VDA QMC publishes the complete Process Reference Model and Process
Assessment Model as a free PDF @aspice40, and that document was read directly.

This section therefore compares speclink against what can be checked, and
nothing else. No clause numbers beyond those visible in a public table of
contents, no objective numbers, no integrity-level or tool-qualification-level
scheme, and no requirement text are stated anywhere below, because none of them
could be verified from a primary source. More importantly: *no claim of
conformance is made.* speclink has not been assessed, qualified, certified or
audited against any of these documents. What follows is a structural
comparison — an argument about where a compiler-checked link sits relative to
what these regimes are organised around — not a conformance argument, and not a
tool-qualification argument.

== Functional-safety standards

ISO 26262 is positioned by its own text as a sector adaptation: "The ISO 26262
series of standards is the adaptation of IEC 61508 series of standards to
address the sector specific needs of electrical and/or electronic (E/E) systems
within road vehicles" @iso26262-8. That relationship is not incidental to the
present discussion; IEC 61508-1 states as a major objective "to facilitate the
development of product and application sector international standards by the
technical committees responsible for the product or application sector"
@iec61508-1, and ISO 26262 is precisely such a product. IEC 61508-1 also
carries the status of a basic safety publication according to IEC Guide 104
@iec61508-1.

ISO publishes ten parts of the 2018 edition as a package, from Part 1
"Vocabulary" to Part 10 "Guidelines on ISO 26262"; two further parts, on
semiconductors and on motorcycles, are cited in the bibliography of Part 8
@iso26262-8. Two parts matter here. Part 6, "Product development at the
software level", specifies requirements for product development at the software
level, covering — in the wording of its own scope — the specification of
software safety requirements, software architectural design, software unit
design and implementation, software unit verification, software integration and
verification, and testing of the embedded software; it also addresses
configurable software @iso26262-6. Part 8, "Supporting processes", lists twelve
topics in its scope, among them overall management of safety requirements,
configuration management, change management, verification, documentation
management and — the entry that concerns any tool author — *confidence in the
use of software tools* @iso26262-8. The series is based on a V-model as
reference process model and cross-references clauses in an "m-n" notation, where
"m" is the part and "n" the clause within it @iso26262-8.

Two published definitions are worth quoting because they are load-bearing for
the rest of this section. ASIL is defined as "one of four levels to specify the
item's or element's necessary ISO 26262 requirements and safety measures to
apply for avoiding an unreasonable risk, with D representing the most stringent
and A the least stringent level", with the note that "QM is not an ASIL"
@iso26262-1. And a *software tool* is "computer program used in the development
of an item or element" @iso26262-1 — a definition under which speclink plainly
falls.

A verifiable and instructive fact is that these bodies of practice cite one
another. The bibliography of ISO 26262-8:2018 lists, among others, RTCA
DO-178C, CMMI for Development, Automotive SPICE (marked as "an example of a
suitable product available commercially"), ISO/IEC/IEEE 12207, ISO/IEC/IEEE
29148 and the ISO/IEC 33000 series @iso26262-8. The domains are not sealed off
from each other; a tool argument made in one vocabulary is at least legible in
the others.

IEC 61508 itself consists of Parts 1 to 7, Edition 2.0, dated 2010-04-30, under
IEC TC 65/SC 65A @iec61508-1. Part 3, "Software requirements", is explicit about
tooling: it "provides specific requirements applicable to support tools used to
develop and configure a safety-related system" and, together with Parts 1 and 2,
"requirements for support tools such as development and design tools, language
translators, testing and debugging tools, configuration management tools"
@iec61508-3.

In civil aviation, RTCA states that "Compliance with the objectives of DO-178C
is the primary means of obtaining approval of software used in civil aviation
products" @do178c. Tool qualification is handled by a separate document: RTCA's
description of DO-330 says it "explains the process and objectives for
qualifying tools", where a tool is "a computer program or a functional part
thereof, used to help develop, transform, test, analyze, produce or modify
another program, its data or its documentation", and notes that it "may also be
used by other domains, such as automotive, space, systems, electronic hardware,
aeronautical databases and safety assessment processes" @do330. DO-333
supplements DO-178C and DO-278A for formal methods @do333. EUROCAE's own
training material describes ED-12C as "the European reference standard for
airborne software certification" and "equivalent to RTCA DO-178C" @ed12c. For
medical devices, IEC 62304, in its consolidated Edition 1.1 combining the 2006
first edition with its 2015 amendment, "Defines the life cycle requirements for
medical device software" and applies where software is itself a medical device
or an embedded part of one, while explicitly not covering validation and final
release of the device @iec62304.

The speclink-facing observation is structural. What these regimes place at the
centre — management of safety requirements, configuration management, change
management, verification, documentation management, and confidence in the use of
software tools @iso26262-8 @iec61508-3 @do330 — is exactly the set of concerns
that a tool producing evidence *about* a safety-related product inherits. If the
tool's output is to carry weight, the tool is itself an object of scrutiny.

speclink's design has affinities with that posture, and they are deliberate.
Every finding is an error: there are no warnings, no severities and no tolerance
mode, so a run either has zero findings or fails (README §1). The single waiver
mechanism described in @sec:design leaves a trace in the generated report —
which is also how the statement of applicability for an external standards
catalogue falls out, since
`"applicable": false` without a `"because"` is refused (README §4.5). Most
references are not resolved by speclink at all: `DerivedFrom`, `Supersedes` and
`spec.Satisfies` name Go identifiers, so the Go compiler proves the referent
exists before speclink runs, and `spec.Do[T]` in a process names a construct
rather than a caption (README §4.1, §4.3). And the final link in the chain is a
recorded passing test run rather than a claim in the source: `speclink evidence`
reads the test stream, because "a claim that a test verifies something is not
evidence that it did" (README §1).

This is a design affinity and nothing more. It is *not* tool qualification.
speclink has undergone no qualification activity of any kind, no qualification
plan exists for it, and none of the publicly available material cited above
would be sufficient to construct one. The most that can be said is that the tool
was built so that its own output is auditable in the same register the standards
use.

== Automotive SPICE

Because the PAM is free, this is the one comparison that can be made against
primary normative text. Automotive SPICE is published by the VDA QMC; the
current version is the Process Reference Model / Process Assessment Model
Version 4.0, authored by VDA Working Group 13, dated 2023-11-29 and marked
"Released" @aspice40 @aspiceweb. The model states that it "was developed in
accordance with the requirements of ISO/IEC 33004:2015" and that its
measurement framework is "an adaption of ISO/IEC 33020:2019", with an ISO/IEC
33003:2015 compliant measurement framework defined in its section 5 @aspice40.
The lineage back to the 15504 series is acknowledged in the model's own words —
"a PRM/PAM according to ISO/IEC 33004 (formerly ISO/IEC 15504-2)" @aspice40 —
and independently by ISO, whose catalogue page for ISO/IEC 33002:2015 lists
ISO/IEC 15504-2:2003 as its withdrawn predecessor @isoiec33002. The capability
dimension has six levels, 0 to 5, incorporating nine process attributes from
PA 1.1 Process performance to PA 5.2 Process innovation implementation; level 0
is defined as "The process is not implemented or fails to achieve its process
purpose" @aspice40. The process dimension comprises 32 processes in 3 categories
and 11 groups @aspiceweb.

The software engineering group of version 4.0 contains six processes, named in
the PAM as SWE.1 Software Requirements Analysis, SWE.2 Software Architectural
Design, SWE.3 Software Detailed Design and Unit Construction, SWE.4 Software
Unit Verification, SWE.5 Software Component Verification and Integration
Verification, and SWE.6 Software Verification @aspice40. Version 4.0 renamed
SWE.3 and SWE.5 relative to earlier versions, and Annex C of the PAM still
refers to SWE.5 by its pre-4.0 style name, "Software Integration & Integration
Test" @aspice40.

Traceability is a first-class, recurring construct in the model rather than an
afterthought: "Ensure consistency and establish bidirectional traceability" is
the *name* of a base practice in several processes, including SYS.2.BP5,
SYS.3.BP4, SYS.4.BP4, SYS.5.BP4 and SWE.1.BP5 @aspice40. SWE.1 has seven process
outcomes, of which outcomes 5 and 6 state that consistency and bidirectional
traceability are established between software requirements and the system
requirements, and between software requirements and the system architecture;
SWE.1.BP5 accordingly requires consistency and bidirectional traceability
between software requirements and *both* of those two targets @aspice40. The
PAM is unusually candid about what this buys. Note 11 to SWE.1.BP5 reads:
"Bidirectional traceability supports consistency, and facilitates impact
analysis of change requests, and demonstration of verification coverage.
Traceability alone, e.g., the existence of links, does not necessarily mean that
the information is consistent with each other." Note 9 adds: "Redundant
traceability is not intended." @aspice40 Traceability also reaches down to code:
SWE.3's outcome 3 requires bidirectional traceability between software detailed
design and software architecture, *and* between source code and software
detailed design @aspice40. SWE.4 and SWE.6 extend it to verification measures
and verification results @aspice40. Finally, "13-51 Consistency Evidence" is
listed among SWE.1's output information items @aspice40 — traceability
materialised as a work product.

The comparison is now sharp enough to state precisely. Automotive SPICE asks for
traceability and consistency between work products, and then assesses whether an
organisation's *process* is capable of producing them, on a capability scale.
speclink does something narrower and different in kind: it makes a subset of
those links a *build precondition* in one specific technology stack. @tbl:aspice
lists the correspondences.

#figure(
  placement: top,
  scope: "parent",
  table(
    columns: 3,
    align: left,
    table.header([ASPICE element @aspice40], [speclink mechanism], [held up by]),
    [SWE.1.BP5 / outcomes 5–6: consistency and bidirectional traceability for
     software requirements],
    [`Sources` on a `spec.Requirement`; `DerivedFrom` between requirements],
    [`K11-REQ-UNSOURCED` and `K11-SOURCE-UNANCHORED`, then `K13-SOURCE-DRIFT`
     and `K10-REQ-CHANGED`],

    [SWE.3 outcome 3: traceability between source code and detailed design],
    [`spec.For` / `ForDecl` / `ForField` / `ForPackage` with `spec.Satisfies`],
    [Go compiler, then `K1-CONSTRUCT-UNBOUND` and `K3-REQ-UNCOVERED`],

    [SWE.4 / SWE.6: traceability to verification measures and results],
    [`spec.Verified(t, req)` recorded by `speclink evidence` from the test
     stream],
    [a recorded passing test run],

    [Note 11: links alone do not mean the information is consistent],
    [fingerprints over requirement text and source segments],
    [`speclink.lock`, re-`freeze` on drift],

    [work product "13-51 Consistency Evidence"],
    [the generated specification document and the summary figures],
    [`speclink generate`, `speclink verify`],
  ),
  caption: [Correspondences between Automotive SPICE elements and speclink
  mechanisms. The right-hand column names what actually holds the link up; a
  correspondence is not a conformance claim.],
) <tbl:aspice>

The omissions are larger than the correspondences and should be read first.
speclink addresses nothing in requirements elicitation: it begins at a Markdown
file or mockup that somebody already wrote, and the step from a source segment
into a requirement has no formal semantics at all. It says nothing about
stakeholder requirements, nothing about system-level processes, and nothing
about hardware. It does not touch the management or supporting process groups —
project management, quality assurance, configuration management as a discipline,
supplier monitoring — and it has no notion whatsoever of the capability
dimension: there is no speclink analogue of a process attribute, a rating, or a
capability level, and a green build is not a level-anything statement. Within
the software engineering group it covers a strict subset, and only for the
languages a profile exists for.

One version question should be recorded honestly. Version numbers beyond 4.0
circulate in the assessor community, but the VDA QMC's own page refers to
version 4.0 as the current one and links the v4.0 PDF @aspiceweb, and no later
PAM document could be retrieved from a primary source. Everything stated above
is taken from version 4.0, and nothing here should be read as a statement about
any successor.

== What a compiler cannot deliver

Process assessment rates organisational capability. ISO/IEC 33002:2015 "defines
the minimum set of requirements for performing an assessment that will ensure
assessment results are objective, consistent, repeatable, and representative of
the assessed processes" @isoiec33002; ISO/IEC 33004:2015 sets the requirements
for the reference, assessment and maturity models themselves @isoiec33004. The
subject of that judgement is an organisation and its processes — not a
repository.

speclink's subject is a repository. A green run is a statement about one working
tree at one commit: that every source segment became a requirement, that every
construct carrying business meaning names one, that every normative requirement
is named by some construct, and that some test claimed and was recorded as
having demonstrated it. It says nothing about the competence of the people
involved, nothing about planning, review discipline or supplier management, and
nothing about whether the requirements are the right requirements. It cannot
even judge the one step above it: whether a requirement faithfully renders the
paragraph it came from is not decidable, and speclink only records that the
paragraph was read at the wording it then had.

The tool's own documentation is blunt about the limit that matters most:
coverage says only that code was written for a requirement, never that the code
does what the requirement asks (README §10b; quoted in full in @sec:discussion).
That is why the *verified* direction exists at all, and why the figure beside it
is a recorded test run rather than a marker in the source — and it is also the
boundary past which no compiler, speclink included, can go.
