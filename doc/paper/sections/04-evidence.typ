= From Claim to Evidence <sec:evidence>

A marker in the source is a *claim*. A line written while the program ran is
*evidence*. The two are not degrees of the same thing: a claim is a statement
that somebody wrote something down, and a static analysis can read it, count it
and report its absence, but it cannot report its emptiness. A binding can name a
requirement that the code does not honour. A verification marker can sit behind
a condition that never holds, or in a test that fails long before reaching it,
and the source cannot tell that apart from a working test by any amount of
analysis (`cmd/speclink/evidence.go`). Evidence is the opposite trade: it can
only be produced by execution, it can never be re-derived from the working tree,
and for exactly that reason it has to be written down.

A green run is a statement about *four directions*, and the term is used in that
sense throughout this paper. `speclink verify` prints one figure for each and all
four must reach 100% (README §2): *accounted*, that every segment of every source
document became a requirement; *bound*, that every construct carrying business
meaning names a requirement; *covered*, that every normative requirement is named
by at least one construct; and *verified*, that something in the source claims to
demonstrate it. A fifth figure, *demonstrated*, reports how many of those claims
were observed to run; it is a figure rather than a threshold, and the finding
that accompanies it is `K14-VERIFICATION-STALE`.

The chain those directions run along has four links, held up by three different
kinds of check.

/ Source segment $arrow$ requirement: the step with no formal semantics. The
  anchor is resolved mechanically — a missing heading slug or region name is a
  finding — but whether the requirement says what the segment means is not
  decidable, and speclink does not pretend otherwise. What *is* mechanised is
  that the segment was read at the wording it then had: fingerprint-checked,
  `K13-SOURCE-DRIFT`.
/ Requirement text $arrow$ its own past: fingerprint-checked against
  `speclink.lock`, `K10-REQ-CHANGED`.
/ Requirement $arrow$ construct: compiler-checked. A binding names a requirement
  by Go identifier, so the referent is proved to exist by the Go compiler before
  speclink runs.
/ Requirement $arrow$ passing test: run-time evidence, recorded by
  `speclink evidence` from the test stream the build already produces. The claim
  side of the same link — that a marker exists at all — is static, and its
  absence is `K14-REQ-UNVERIFIED`.

Provenance and human review sit beside the chain rather than inside it, and are
recorded from outside the code entirely.

#figure(
  table(
    columns: 3,
    align: left,
    table.header([link], [held up by], [rule]),
    [segment $arrow$ requirement], [fingerprint], [`K13-SOURCE-DRIFT`],
    [requirement wording], [fingerprint], [`K10-REQ-CHANGED`],
    [requirement $arrow$ construct], [Go compiler], [—],
    [claim of a test], [static read], [`K14-REQ-UNVERIFIED`],
    [passing test run], [run-time record], [`K14-VERIFICATION-STALE`],
    [human review], [fingerprint], [`K18-REVIEW-STALE`],
  ),
  caption: [The chain from a source segment to a recorded test run, and what
  holds each link up.],
) <tbl:chain>

== The step with no formal semantics

Above the requirement tree are the documents people actually wrote. speclink
recognises two kinds of source and no others (README §10a). A Markdown file is
segmented by its headings, and the anchor is the heading slug. An image — `.png`
or `.jpg` — is segmented by a sidecar manifest `<image>.speclink.json` that
declares named regions as pixel rectangles with origin at the top left
(`internal/source/image.go`). From the requirement side there was never a
difference between pointing at a section and pointing at part of a screen; both
are anchors into a sequence of segments.

PDF is deliberately unsupported. The instruction is to convert it to Markdown,
which makes the conversion a visible step and the result diffable in the pull
request — as opposed to a conversion hidden inside the tool, where nobody
reviews it.

Regions are declared, not detected. An image is not decomposable by any
deterministic rule, and a model inventing regions would only move the
invented-requirement problem one level down: instead of requirements with no
source, there would be sources with no author.

Every segment must yield at least one requirement. Some genuinely carry no
obligation — a title, an introduction, a glossary, a chrome element — and that
is stated where the segment is written: the line
`<!-- speclink:informative -->` in the Markdown
(`internal/source/markdown.go`, `InformativeMarker`), or `"informative": true`
on the region in the manifest.

This is not `spec.Waive` and cannot be. A waiver attaches to a Go construct, and
a document section has none, so a waiver narrowed to one section could not be
written down at all. Putting the statement in the document also puts the
decision with the person who wrote the section, and keeps the fact in one place
rather than splitting it between a paragraph and a package.

This step should be read as the honest weak point of the design. It is the one
link in the whole chain that a compiler cannot hold up. Everything below the
requirement tree — the binding, the construct recognition, the coverage figure —
is only as meaningful as the extraction that produced the tree, and no rule in
speclink can establish that a requirement faithfully renders the paragraph it
was taken from. What the tool can do is make the pairing readable: a single
output places each requirement next to the segment it came from, so a person can
judge the extraction, and record that a person did.

== Drift as a fingerprint, not a heuristic

A requirement text and a source segment are the same kind of edge as a persisted
wire format: the compiler cannot re-derive them, so `speclink.lock` records what
they used to say (`internal/baseline`, `internal/check/drift.go`).

Rewrite the text of a requirement and its identifier is unchanged, every
`spec.Satisfies` still compiles, and the coverage figure stays at 100%.
`K10-REQ-CHANGED` is the only thing that notices. Rewrite the section it was
extracted from and the anchor still resolves; `K13-SOURCE-DRIFT` is the only
thing that notices.

For images the fingerprint covers the pixels inside the declared rectangle, and
the coordinates are what make the report specific. Recolouring one button
reports the requirements of that button and nothing else. A layout shift moves
everything below it and reports drift across the board — coarse, but visible,
and resolved by one `freeze`, which is preferable to a report so blunt that it
gets ignored (`internal/source/image.go`).

Reformatting is not drift and reflowing is not a rewrite; both cases are pinned
by tests, because a rule that fires on a formatter is a rule that gets waived out
of habit and then stays waived.

Both findings are answered the same way: re-read what the finding names, change
what has to change, then run `speclink freeze`. The diff of `speclink.lock` *is*
the review. That is the reason for a recorded fingerprint rather than a status
field somebody maintains: a moment of review that appears as a diff in a pull
request survives contact with a project, where a discipline does not
(`cmd/speclink/freeze.go`).

== Verification: the two lives of one call

At the end of a test, the test says what it demonstrated:

```go
func TestSubmitDrawsAGaplessNumber(t *testing.T) {
	// … exercise the use case, assert …
	spec.Verified(t, quote.RQuoteSubmit)
}
```

Position matters here and nowhere else in the language. `spec.Verified` writes a
line when it runs, so putting it at the end says the test got there; putting it
at the top says only that the test started.

The call has two lives. It is read statically, which is what makes a *missing*
call reportable as `K14-REQ-UNVERIFIED`. It writes a line when it executes,
which is what makes a *present* call believable.

#figure(
  table(
    columns: 3,
    align: left,
    table.header([in the source], [recorded], [meaning]),
    [no call], [–], [nothing claims to verify it — `K14-REQ-UNVERIFIED`],
    [a call], [nothing, or against an older wording], [claimed, not shown — `K14-VERIFICATION-STALE`],
    [a call], [matching, from a passing test], [demonstrated],
  ),
  caption: [The three states of one requirement under verification (README §10b).],
) <tbl:verified>

The second half runs as the fourth step of the build order:

```text
go test -json ./... | speclink evidence
```

Only passing tests are recorded: a test that claimed something and then failed
showed nothing, and recording it would make the failure invisible. A run is the
whole truth — a requirement that nothing demonstrated this time loses its
record. `K14-VERIFICATION-STALE` then reports it, and the same finding covers
three distinct mistakes that are indistinguishable from the source and are all
fixed the same way: the call sits behind a condition that never holds; the test
fails before reaching the end; or the requirement was rewritten after the last
run, so the evidence that exists was evidence for other words.

The summary therefore reports a pair, `100% verified, 88% demonstrated`, and the
gap is the interesting number. It says that a test exists, compiles and claims
something, and has not been seen doing it. A single figure would have to choose
between over- and under-reporting; the pair reports the uncertainty instead of
resolving it by fiat.

speclink deliberately does not run the tests itself. The build order is
compiler, then speclink, then tests, and a command that invoked the suite would
either violate that order or duplicate it. It also makes the evidence something
CI hands over rather than something speclink produces, which is the right way
round for evidence: the party that asserts a result should not also be the party
that grades it.

Some requirements cannot be tested at all. A structural decision — *customer
data is stored as state, not as facts* — is discharged by the type existing, and
a test could only assert that the code compiles as written. Such a requirement
is waived on a construct that satisfies it:

```go
var _ = spec.For[Customer](
	spec.Satisfies(dec.RDecCustomerState),
	spec.Waive("K14-REQ-UNVERIFIED", "…"),
)
```

The waiver attaches to a construct, not to a requirement, because `spec.Waive`
attaches to a Go construct and a requirement declaration is not one. The finding
names a construct that the waiver can be put on, so the escape hatch is always
anchored somewhere the compiler can see it, and always carries a written reason.

== Provenance and review, recorded from outside

Authorship and human review are recorded by commands, never declared in the
code (README §2, §10a, `cmd/speclink/attest.go`, `cmd/speclink/freeze.go`). Two
commands carry a `-reviewer` flag and they record two different facts.
`speclink attest -reviewer` names one or more *declarations* — a construct, by
name or by package pattern — and records that the named person read them;
`speclink freeze -reviewer` records that the named person read the *requirement
wordings* as they then stood, bound to those wordings, so that rewriting a
requirement text discards the review. Every later mention in this paper follows
that split.

```text
speclink attest -origin llm ./app/sales/...
speclink attest -reviewer "TS" SubmitQuote
speclink freeze -reviewer "Frau Meier" ./...
```

There is no `Reviewed` field on `spec.Requirement` and no authorship field in
the code. The reason is the claim/evidence distinction applied to the author:
the same machine that writes the code writes the annotation, so a self-declared
claim of human authorship or human review is not evidence about the author.
Routing the certification to another model is not an escape either, for the
self-preference reason given in @sec:introduction @panickssery2024. The human
end of the same failure is documented too: #cite(<perry2023>, form: "prose")
found that participants with access to
an AI assistant wrote significantly less secure code and were at the same time
more likely to believe it secure, which is the human end of the same failure:
the confidence attached to generated code is not calibrated to it.

Reviews are targeted rather than run-wide. A reviewer is usually specialised and
reads a few declarations at a time; recording a whole run as read because
somebody looked at one use case is the fastest way to make the figure
meaningless. A use case's fingerprint therefore covers its constructor and not
only the named func type: the signature is a line, the logic a reviewer actually
reads lives in the `New…` beside it, and a fingerprint over the signature alone
would survive any rewrite of the body and go on claiming the review still holds.
When the subject does change under a recorded review, `K18-REVIEW-STALE` reports
it — somebody's name attached to text they never saw, which nothing else in the
record can detect.

There is deliberately no finding for unread code. In a project whose code a
machine writes and a person samples, a build that stays red until everything has
been read is never green, and a signal that is always on is not a signal. It is
reported as a figure instead:

```text
8 declarations (8 machine written, 2 read by a person)
9 requirements (8 normative, 3 reviewed)
```

Unattested is neither: a declaration nothing has said anything about counts as
neither machine written nor read, because silence must not be able to pass for
handwork. And speclink writes down what it is told and checks none of it. If
whatever drives the generator may also invoke `-reviewer`, the record is worth
nothing; what keeps it honest is who is permitted to make which call, which is
an organisational control and not speclink's to decide. Saying so plainly is
better than a mechanism that implies a guarantee it has not got.
