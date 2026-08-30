= Design <sec:design>

speclink is built on one rule, and everything else in this section is a
consequence of it: *if a fact can be inferred from the code, annotating it is an
error*. The README states this as one of the five rules that must not be broken.
The emphasis belongs on *error*. A redundant annotation is not treated as
harmless duplication to be tolerated or linted away at a lower severity; it is a
finding, and speclink has no severities — every finding is an error and the run
either has zero findings or it fails.

The reason is that a duplicated fact is a fact that can disagree with itself.
speclink recognises use cases, commands, events, aggregates, permissions,
queries, projections and repositories on its own. The one thing it can never
infer is *which requirement a construct was written for*. That is what an
engineer annotates, and it is essentially the only thing. The annotation surface
is therefore minimised not for ergonomics but for soundness: a language in which
only unrecoverable facts may be stated cannot accumulate statements that drift
from the program.

The rule cuts in both directions. `spec.Persistence()` and `spec.StoredAs[D]()`
exist under `go_bare_ddd1`, where a hand-written interface says neither whether
it is a port nor what shape reaches disk. Under `go_nago_ddd1` both terms are
findings, because the repository constructors already say it — "saying it twice
is the one thing this language forbids".

== The annotation language is a closed whitelist

An annotation lives in a file `<base>.annotation.go`, next to `<base>.go` in the
same package, and it is part of the *normal build*. That coupling is the
load-bearing decision of the whole design: a broken annotation breaks the
production build, so it cannot rot unnoticed. There is no separate parser to
keep in step with the language, and no way to ship a release whose traceability
data no longer type-checks.

The grammar of such a file is a closed whitelist, not "valid Go". An annotation
file may contain only the package clause, imports, and package-level
`var _ = spec.X(…)` terms. @tbl:whitelist gives the rejected forms and the rule
code each carries.

#figure(
  table(
    columns: 2,
    align: (left, left),
    [*Forbidden*], [*Code*],
    [function definitions], [`SPEC-V1-001`],
    [type declarations], [`SPEC-V1-003`],
    [constant declarations], [`SPEC-V1-004`],
    [a `var` binding more or fewer than one name to one value], [`SPEC-V1-005`],
    [a binding not written as `var _ = …`], [`SPEC-V1-006`],
    [calls into anything other than
     `github.com/worldiety/speclink/spec`], [`SPEC-V1-007`],
    [positional fields in a struct literal (use `Field: value`)], [`SPEC-V1-009`],
    [the address-of operator `&`], [`SPEC-V1-011`],
    [function literals, binary expressions, any computation], [`SPEC-V1-010`],
  ),
  caption: [The annotation subset. Everything not in the whitelist is rejected
    in phase `V1`, before the Go compiler is consulted.],
) <tbl:whitelist>

The purpose of the closure is not austerity. A closed grammar turns the file
into a *data language that happens to be type-checked by a general-purpose
compiler*. Because no computation, no indirection and no call outside the `spec`
package is admitted, every term is a literal reference to a program element, and
that reference has already been resolved by `go build` before speclink runs. A
requirement named in an annotation is proven to exist and to be of the right
type — and speclink implements no name resolver, no import graph walker and no
scoping rules of its own to obtain that guarantee. Phase `V2` in the diagnostic
numbering is *the Go compilation itself, reported by the Go compiler and not by
speclink*: speclink's own checking simply resumes on the far side of it.

The whitelist also closes the file-level edge. An annotation file whose
neighbour `<base>.go` does not exist is an orphan and is reported as
`SPEC-V3-001`: rename it or delete it. Without that rule the sidecar convention
would be an aspiration rather than a fact, and the per-file impact analysis that
depends on it would silently answer for a file that is no longer there.

== Bindings and assertions

The language has four bindings, and they are its entire side-effecting surface.
`spec.For[T](…)` attaches to a named type: a use case func type, an event, an
aggregate, a projection state. `spec.ForDecl(ref, …)` attaches to a declared
function, variable or constant, *named by the value itself*, so the Go compiler
proves it exists. `spec.ForField[T]("Name", …)` attaches to a struct field of
`T`. `spec.ForPackage(…)` attaches to the package of the neighbouring file.

Everything else is an assertion: a pure term passed into a binding, never
standalone — `spec.Satisfies`, `spec.Transition`, `spec.External`, `spec.Help`,
`spec.Term`, `spec.Rationale`, `spec.Waive`, `spec.Draft`, `spec.Optional` and
`spec.StoredAs`. Of these, `spec.Satisfies(reqs…)` binds the construct to one or
more requirements and is "the central statement, and the only one that cannot be
inferred".

```go
package sales

import (
	"example.com/erp/requirements/fun/quote"
	"github.com/worldiety/speclink/spec"
)

var _ = spec.For[SubmitQuote](
	spec.Satisfies(quote.RQuoteSubmit),
	spec.Help(`Submit the approved quote.`),
)
```

`ForField` is the one deliberate exception, and it is worth stating plainly
rather than hiding. It takes the field name as a string, the only reference in
the whole language that the Go compiler cannot check: Go has no expression
denoting a struct field independently of a value, so the guarantee cannot be
obtained the way `ForDecl` obtains it. speclink therefore checks the name itself
against the type. This is the residue of the compiler-delegation strategy,
minimised to exactly one construct rather than pretended away. A field binding
also satisfies a requirement *for that field*, not for the type it belongs to.

== Requirements as compiled values

A requirement is a `spec.Requirement` value in a file `<ID>.spec.go`, one
requirement per file, the file named after the ID. Requirements may only appear
in the requirement tree, never in an annotation file: a requirement is owned by
the domain side and outlives any implementation.

Three fields carry the semantics the coverage rules read. `Kind` is
`Functional`, `NonFunctional`, `Constraint` or `Decision`, and a `Decision` must
carry a `Rationale`. `Discipline` is `Business`, `Technical` or `Mixed`.
`Status` is `Normative`, `Abstract`, `Planned`, `OutOfScope`, `Informative` or
`Superseded`, and it decides what coverage means: only `Normative` must be
covered; `Abstract` is a pure derivation node and must *not* be covered directly
(`K3-ABSTRACT-COVERED`); `Superseded` must no longer be covered
(`K3-SUPERSEDED-COVERED`). A normative requirement satisfied by nothing is
`K3-REQ-UNCOVERED`. Coverage is thus a status-indexed predicate rather than a
uniform one, which is what lets a tree be grown, refactored and retired without
either a false gap or a false claim appearing in the figures.

`DerivedFrom` and `Supersedes` reference other requirements by their *Go
identifier*, never by the ID string. Moving a file, or reorganising a domain
directory, therefore cannot break a reference — the compiler re-resolves it —
while a reference to a deleted requirement fails the build rather than dangling.
Cycles in `DerivedFrom` are reported (`SPEC-V5-013`), because acyclicity of a value
graph is the one property the compiler will not check.

Layout is where the design deliberately does the opposite of minimising
redundancy. Directory, ID prefix and `Kind` state the same fact three times, and
speclink checks that all three agree: `SPEC-V5-032` when the prefix contradicts
`Kind`, `SPEC-V5-033` when the directory does, `SPEC-V5-035` when the domain directory
contradicts the ID prefix. This is not a violation of the inference rule but its
mirror image. The three statements are not one fact annotated three times; they
are three independent encodings maintained by three different mechanisms — a
file system, a naming convention and a typed constant — used as a checksum on
each other. Redundancy that can be mechanically compared is a detector;
redundancy that cannot is drift.

== What is inferred

@tbl:infer gives the Go frontend's recognisers under `go_nago_ddd1`. The left
column is what an engineer writes anyway; the middle column is what speclink
derives from it by resolved type, so an alias import or a renamed identifier can
neither fool it nor help.

#figure(
  placement: top,
  scope: "parent",
  table(
    columns: 3,
    align: (left, left, left),
    [*Written*], [*Recognised*], [*Must name a requirement?*],
    [named func type, first param `auth.Subject`], [use case], [yes],
    [the same, but returning data], [query], [yes],
    [type with `Decide(auth.Subject, *Agg) ([]Evt, error)`], [command], [yes],
    [type with `Evolve(ctx, *Agg) error` *and* `Discriminator()`], [event], [yes],
    [type with `Identity() ID`, or the type an event's `Evolve` folds into], [aggregate], [no],
    [`permission.Declare…[UseCase](id, …)`], [permission, bound to that use case], [no],
    [state type of `evs.NewProjection` / `evs.NewSingleton`], [projection], [yes],
    [`type X data.Repository[E, ID]` or `= data.ReadRepository[…]`], [repository], [no],
    [`hapi.Post[In](api, hapi.Operation{Path: …})` and the chain on it], [endpoint, with its wire types], [no],
  ),
  caption: [What speclink infers, and from what.],
) <tbl:infer>

The third column is where the inference rule pays out. Aggregates, permissions
and repositories are structural: they are covered through the use case that
guards, writes or holds them, so nothing is written about them at all.
Everything else names its own requirement — one `spec.Satisfies` term, and no
second fact. A query is no exception, because reading is a promise too; a
projection is none either, because it crosses aggregates and answers a question
no single use case states.

Endpoints are read from the builder rather than declared: the method from the
name of the call, the path from the `Operation` literal, and the request and
response types from the type arguments the compiler has already resolved. The
whole chain is followed, and the use case behind it is found by type, so no
wrapper or middleware has to be recognised for the route to be seen. speclink
never guesses a wire shape — on a bare mux the columns stay empty and the
catalogue says `_not stated here_`.

== Processes and topology

A process is declared in a `<name>.process.go` file as a `spec.Process` value
with `Nodes` and `Edges`: the course of business — not what there is, but what
happens, in which order, and where it can end — and it satisfies requirements
the way a construct does.

The decision that matters is that *a step names a construct, not a caption*.
`spec.Do[sales.SubmitQuote]("abgeben")` is a type argument, so the Go compiler
establishes that the named thing exists — and then `K16-NODE-REF-UNKNOWN`
establishes that it is the right *kind* of thing, refusing an activity that names
an event and an event that names a use case. This is the two-stage pattern of
the whole tool in miniature: existence from the compiler, meaning from speclink.
A step written as prose would be a caption, and a caption cannot go stale in a
way anybody notices.

The vocabulary is nine nodes — `Start`, `End`, `Do`, `Emit`, `On`, `Fork`,
`Join`, `Choice`, `Merge` — and the omissions are deliberate: inclusive and
event-based gateways, boundary events, timers, compensation and subprocesses are
absent until a real case asks for one. Because a process is a graph rather than
nested blocks (real courses come back), the compiler cannot check the wiring;
that price is paid by `K16`, which refuses a dangling edge, a duplicated name, a
node no start reaches and a node from which no end can be reached. What is *not*
established is stated as such: speclink does not attempt to decide whether every
fork is matched by exactly one join on every path, and the rule set says
deadlock freedom is not established rather than implying otherwise.

A process lives above the bounded contexts, in a package of its own, for a
structural reason: it is precisely the thing no single context owns, and a
context that declared it would have to import its neighbour, which the layering
rules forbid.

Topology files, `<name>.topology.go`, declare `spec.Actor`, `spec.Foreign` and
`spec.Channel` values. This is the one part of the model that is declared rather
than read, and not for convenience: no Go module states that an end user exists,
that the object store is somebody else's responsibility, or that the channel to
it carries customer data under a short-lived key. That is not missing from the
code; it is knowledge the code cannot hold — the inference rule applied honestly
in the direction where inference is impossible. `Protocol`, `Data`, `Auth` and
`Crypto` are required fields, and what keeps the result from being a picture is
that both ends are enumerated: a channel naming a package that is not there is a
mistake, and an adapter no channel names is a way out that never appeared in any
interface list.

== Promised shapes and persisted data

An event struct is the wire format of a log that is replayed forever, so every
field name and every field type is a promise from the moment the first message
is stored. Two things are persisted and only two: an event, whose struct *is*
the stored form because it implements `Evolve` and carries a discriminator; and
a persistence model, the type a repository was built over.

Everything persisted is frozen by default, and `spec.Draft()` is the exception.
The inversion is deliberate: the number of drafts in a system shrinks over its
lifetime, so marking the exception costs less than marking the rule, and
forgetting fails safe — an unmarked newcomer is frozen at once, which surfaces
as an error the first time somebody changes it rather than as unreadable data a
year later. `spec.Draft()` claims one specific thing, and it is not about
deployment: that every stored message of this type may be deleted. The term
cascades over package, type and field, and marking a level the cascade already
covers states nothing new and is reported as `K9-DRAFT-REDUNDANT` in phase `V4`
— the inference rule again, applied to speclink's own terms.

Once a shape is frozen: the discriminator must never change and must be unique
in the module; a field must not be removed, nor renamed on the wire (the Go
field name may change freely, the JSON tag is what is promised), nor change its
shape incompatibly. New fields are allowed but must be marked `spec.Optional()`,
and optionality cannot be withdrawn. Two changes are free, because the record
cannot tell them apart from no change: a named type over the same underlying
type, and any change of integer width, since a promise about an integer field is
that it holds a whole number.

None of this is decidable from the current source — nothing in a working tree
says what a field used to be — so the promise is recorded in `speclink.lock`,
written by `speclink freeze` and never edited by hand. It is not a second place
to state intent; intent stays in the code, where the field type says the shape
and `spec.Draft` says the status. The lock records what has already been
committed to, the same relation `go.sum` has to `go.mod`, and `freeze` refuses
to record anything that already breaks a promise.

== Escape hatch and diagnostics

The only escape hatch is `spec.Waive(rule, reason)`, and the reason is
mandatory. A waiver leaves a trace in the report and counts toward the figures,
so suppression is visible rather than silent. What a profile deliberately cannot
do is switch a rule off; that is `spec.Waive` with a reason, or the scope, and a
third way would be severities under another name. Scope narrows package by
package rather than rule by rule, on the ground that *this package is not under
speclink yet* is a true statement while *this rule half applies here* is not.

Diagnostics are prescriptive. Each finding states what is wrong, why it is
wrong, and what to do, with the `How:` line usually the literal fix:

```text
file.go:12:6: [SPEC-V6-056] use case
  SubmitQuote has no permission of its own.
  A permission per use case is what makes
  authorisation assignable and auditable. …
  Add `PermSubmitQuote = permission.Declare
  [SubmitQuote]("…", name, description)` and
  check it in NewSubmitQuote.
```

The code is `SPEC-<phase>-<number>`, and phases run in order, each only when the
previous one is clean: `V1` the annotation subset, `V2` the Go compilation,
`V3` a binding on an illegal target, `V4` an annotation stating a fact already
established elsewhere, `V5` the requirement tree, `V6` the specification and
architecture rules. The ordering is itself a design element: a project is not
told about a coverage gap while its annotation file is not yet a legal member of
the language.

== Language independence

speclink separates a language-independent core from a per-language front end;
the Go reader lives under `internal/lang/golang` and the JVM reader under
`internal/lang/jvm`. The rule families split along the same seam. The
language-independent families live in `internal/check` — `K1`, `K3`, `K9`, `K10`
and `K12` to `K18`, and `K20` — and the requirement-tree families `K11` and
`K19` live in `internal/reqtree`. Only `K4` to `K8` belong to a profile's style,
and those are implemented in `internal/lang/golang`. The seam is a property of
the tree rather than of a stated interface; @sec:implementation records where
each family sits.

The JVM front end reads *compiled classes* rather than source, on the same
delegation argument as the Go whitelist: a class file has already been resolved,
so every type it names is fully qualified, every supertype is named outright and
every annotation has had its import worked out. It also collapses three targets
into one — Java on Spring, Kotlin on Spring and Kotlin on Android — since Kotlin
compiles to JVM bytecode and is only dexed afterwards (README §11a).

Because a Java annotation may hold only compile-time constants, a requirement is
a class, and one requirement names another by class literal
(`derivedFrom = RDecNumbering.class`); a construct binds with
`@Satisfies(RQuoteSubmit.class)`. With a string there the reference would be
unverified, "which is the one thing this design will not accept". There is no
library to depend on: a project declares these annotations itself and speclink
recognises them by fully qualified name, the same way it recognises a framework
without linking it. Verification is read from `@Test @Verifies(…)` in the
bytecode joined against Surefire's XML on the name a report gives a test.

The honest part is what the front end refuses to answer. The
`java_springboot_ddd1` profile prescribes no architecture rules and has no
persistence recogniser, and a run under it prints
`not measured: schema evolution, because this frontend reads no persisted
shapes` beside its summary rather than reporting the family sound. A direction
that was never measured must not read as one that came out clean — over an empty
set the evolution rules would report every recorded type as removed, which is a
different claim from "nobody looked". The figure is absent rather than zero for
the same reason.
