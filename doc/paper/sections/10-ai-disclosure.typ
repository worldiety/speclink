= Disclosure of AI Involvement <sec:disclosure>

This paper reports work in which a large language model was not an assistant at
the margins but the executing agent for most of the work described. The
disclosure below is deliberately specific, because a generic acknowledgement
would tell a reader nothing they could act on, and because the paper's own
argument is that a claim made by the party with an interest in it is not
evidence.

== What was done by a machine

The following were produced by a large language model operating under the
direction of the named author:

/ Research: the literature and standards survey. Every source in the
  bibliography was located, retrieved and its metadata extracted by an agent.
  No source was written down from the model's parametric memory.

/ Implementation: substantially all of the source code of speclink, including
  the front ends, the rule implementations, the intermediate representation and
  the command-line surface.

/ Tests: substantially all of the test suite and the fixtures under
  `testdata/`, including the worked example.

/ Text: the wording of this paper. Each technical section was drafted by a
  separate agent instance from the repository and the verified research notes,
  and the sections were then assembled, deduplicated and made consistent by
  further agents.

/ Measurement: the figures reported in the implementation section were produced
  by an agent running commands against a clean checkout of the repository.

== What was done by a person

The named author defined the problem, the design rules that the tool enforces
and the argument the paper makes; selected and rejected approaches; directed
each agent; and reviewed and accepted the result. The author is responsible for
every claim in this paper, including those he did not personally type.

== The controls that were applied

Because the risk in this arrangement is fabrication rather than error, the
process was structured against it rather than trusting the output:

+ Research and authoring were separated. Section authors were forbidden to
  perform their own research and could cite only from a bibliography prepared
  in advance.

+ Every source was required to be accompanied by a URL retrieved during the
  session and a verbatim quotation from the retrieved page. Anything that could
  not be retrieved was recorded as unverified and was excluded from the
  bibliography rather than silently softened.

+ An independent audit agent, which had not gathered the sources, re-fetched
  every one of them, cross-checked bibliographic metadata against at least two
  independent services where available, and string-matched every quotation
  against the retrieved page. It found no fabricated source and seventeen
  metadata or quotation defects, all of which were corrected before drafting
  began. Its report is retained in the repository alongside the research notes.

+ Section authors were bound by a written contract stating, per source, what
  that source may and may not be used to support. Paywalled standards may
  support statements about their titles, scopes, published definitions and
  publicly listed clause headings, and nothing else. Vendor documentation,
  marketing material and self-published video may support only attributed
  statements, never absolute ones.

+ A final review pass, performed by an agent that did not write any section,
  checked the assembled paper for internal contradiction and for claims
  exceeding their sources.

== What these controls do not establish

They reduce the probability of an invented citation; they do not eliminate it,
and a reader who intends to rely on a claim in this paper should check its
source. They say nothing about whether the design is sound, whether the
implementation is correct beyond what its own test suite demonstrates, or
whether the argument is the right one. The paper's central position — that a
claim made by the author of an artefact about that artefact is not evidence
about it — applies to this section as much as to any annotation in a source
file. The retained research notes, the audit report and the repository history
are offered so that the claim can be checked rather than believed.
