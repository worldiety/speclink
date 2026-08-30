= Conclusion and Future Work <sec:conclusion>

The distinction between a claim and a piece of evidence is the core of this
paper, and everything else in it is an application of that distinction. A claim
is something the source says: a binding stating that a construct satisfies a
requirement, a marker stating that a test verifies one, a name stating that
somebody reviewed something. A claim is readable statically, which is exactly
why a *missing* claim can be reported, and exactly why a *present* claim proves
nothing: a marker can sit behind a condition that never holds, and a self-
declared record of human authorship is not evidence about the author. Evidence
is a record that something happened — a test that ran to the marker and passed,
an attestation made from outside the code — and can never be re-derived from a
working tree. speclink keeps the two apart everywhere, reports them as separate
figures, and treats the gap between them as the number that carries information.

What was built is an annotation compiler with a Go front end and a JVM bytecode
front end over a shared, language-neutral requirement model. It annotates only
the one fact no analysis can infer, which requirement a construct exists for,
and makes the annotation of any inferable fact a finding; every finding is an
error and the only escape hatch is a waiver carrying a written reason. The
annotation grammar is a closed whitelist inside the ordinary build, so
references to requirements are resolved by the host language's compiler rather
than by a parser of speclink's own. Requirements are anchored to segments of
Markdown documents and to named regions of mockups; their wordings and those
segments are fingerprinted in `speclink.lock`, so a change is a diff in a pull
request rather than a silent divergence. The chain ends at recorded passing test
runs supplied by the build, and authorship and review are recorded by commands
rather than declared in the source. What was compared is narrower than it may
appear: a structural comparison against the publicly verifiable parts of ISO
26262, IEC 61508, DO-178C and their neighbours, a comparison against Automotive
SPICE that could be made against primary normative text because the model is
freely published, and a delimitation from generator-side spec-driven tooling and
from requirements-management platforms. None of these is a conformance claim, a
qualification argument, or a statement about a capability level.

The most consequential piece of future work is the one the paper does not
contain: a controlled empirical evaluation with an independent baseline —
subjects, tasks, a control condition and a measured effect, in the shape of
#cite(<maeder2012>, form: "prose"), or a multi-project regression in the shape of
#cite(<rempel2017>, form: "prose") — conducted by people who did not write the
tool, on projects the tool's authors do not control. Second, speclink does not
govern itself: the repository contains no requirement tree, no configuration and
no lock file, and the command that would produce a self-report does not exist.
Building it and applying the regime to the implementation is the cheapest
available test of whether the discipline survives a long-running project.
Third, the approach is bound to an enforced layout, a fixed set of recognised
constructs and a profile naming a language, a framework and a style; broadening
it — or characterising precisely which projects it excludes — is open. Fourth,
whether the counter-pressures on waivers actually suppress abuse is unmeasured;
waiver counts, and reviewer behaviour when a build fails, can be measured and
have not been. Fifth, the JVM front end prescribes no style rules and reads no
persisted shapes, and bringing a second language front end to parity would show
what in the design is genuinely language-independent and what merely has not
been ported. Sixth, and hardest: the step from a source segment to a requirement
has no formal semantics, and it is unclear whether any assessable quality
measure can be attached to it at all. Fingerprints detect that a segment moved;
nothing detects that it was misread. Until that question has an answer, every
figure below it is a sound statement about a premise nobody checked.
