# speclink

**Read this before writing any code in this repository.**

This document is written for a coding agent that starts with no prior knowledge
of speclink or of the nago framework. Everything you need is stated here; you do
not need to read any other document to work correctly.

speclink is an annotation compiler. It verifies that the implementation and the
requirements agree, and it derives documentation from that single source. It
also enforces the architecture of a nago project.

---

## 1. The rules you must not break

1. **Build order is fixed.** `go build ./...` → `speclink verify ./...` →
   `go test -json ./... | speclink evidence`. If the Go build is broken,
   speclink refuses to run and tells you so; fix the compile error first. The
   last step hands the test results back, because a claim that a test verifies
   something is not evidence that it did.
2. **Every finding is an error.** There are no warnings, no severities and no
   tolerance mode. The run either has zero findings or it fails.
3. **The only escape hatch is `spec.Waive(rule, reason)`**, and the reason is
   mandatory. A waiver leaves a trace in the report. Do not reach for it to make
   a finding disappear; reach for it when the rule genuinely cannot hold and you
   can say why in one sentence.
4. **If a fact can be inferred from the code, annotating it is an error.**
   speclink recognises use cases, commands, events, aggregates, permissions,
   queries, projections and repositories on its own. The one thing it can never
   infer is *which requirement a construct was written for*. That is what you
   annotate, and essentially the only thing.
5. **Read the `How:` line of a finding.** Diagnostics are prescriptive: they
   state what is wrong, why it is wrong, and what to do. The `How:` line is
   usually the literal fix.

---

## 2. Running the tool

```
speclink init         [flags]
speclink verify       [flags] [packages]
speclink requirements [flags] [packages]
speclink freeze       [flags] [packages]
speclink inventory    [flags] [packages]
speclink impact       [flags] <requirement|doc.md#anchor|path>...

  -format text|json   text is the default
  -root <dir>         repository root, default "."
  -config <file>      layout configuration; defaults to speclink.json in the root
  -n                  freeze only: report what would be recorded, write nothing
  -reviewer <who>     freeze only: record that this person read the requirements
  -in <file>          evidence only: read the test stream from a file
  -out <file>         generate only: write here instead of standard output
  -profile <name>     overrides the profile from speclink.json
  -kind <kind>        inventory only: restrict to one kind, e.g. event
  -template <name>    init only: which starting point of the profile to write
  -dir <dir>          init only: where to write, default "."; must be empty
  -module, -context   init only: the parameters of the template
  -describe           init only: list profiles, templates and their parameters
```

### Starting a project

`speclink init` writes a starting point. It asks nothing interactively: the
caller is usually an agent, for which a prompt is the worst possible interface,
so every missing answer is a refusal that names the alternatives.

```
speclink init                                   lists the profiles
speclink init -profile go_bare_ddd1             lists that profile's templates
speclink init -describe -format json            the same catalogue, machine readable

speclink init -profile go_bare_ddd1 -template full \
  -module example.com/erp -context sales
```

There is no default template even where a profile has only one, so that adding
a second is never a breaking change. The binary name is derived from the module
path rather than asked for, because a separate answer could contradict the
import paths.

The generated project carries an `AGENTS.md` with its own working instructions,
and no `speclink.lock`. The lock records what happened — tests that ran, shapes
somebody approved — and a templated one would assert both without grounds. So
the first `verify` of a new project reports findings, and clearing them is the
first run of the loop:

```
go mod tidy
go build ./...
go test -json ./... | speclink evidence
speclink freeze
speclink verify
```

Typical invocation:

```
speclink verify -root . ./...
```

Exit code is `1` if there is any finding, `0` otherwise. The summary goes to
stderr:

```
7 source segments (100% accounted), 19 constructs (100% bound), 8 normative requirements (100% covered, 100% verified), 30 bindings, 0 findings
```

Four numbers, four directions, all of which must reach 100%.

*Accounted* is the direction above the requirement tree: did every part of the
documents people actually wrote become a requirement? *Bound* is the forward
direction below it: does every construct that carries business meaning name a
requirement? *Covered* is the backward direction: is every normative
requirement satisfied by at least one construct? *Verified* is the question the
other three never ask: does anything demonstrate that the code does what the
requirement says?

Read them in that order, because the first decides what the middle two are
worth and the last decides what any of them are worth. Bound and covered
measure the tree against the code, and the Go compiler is already holding most
of that edge up. Accounted measures the tree against what was asked for, and
nothing else in the chain does — a tree can be internally perfect, every other
figure at a hundred percent, and a whole section of the specification simply
absent. Verified is the one that stops the whole thing from being a very
thorough account of code that nobody ever ran.

### Checking the requirement tree on its own

```
speclink requirements -root . ./requirements/...
```

This asks a different question: is the tree sound in itself? It checks identity,
the derivation graph, the layout and the outer edge to the source documents, and
it reads no annotation, infers no construct and measures no coverage.

Use it while a tree is being built. Until the last requirement is in place there
is no point asking whether the code covers it — `verify` would drown the tree's
own defects under one finding per unbound construct, and the tree is what has to
be right first. Only the packages you name have to compile, so the tree can be
grown while the implementation around it is still in pieces.

```
343 requirements (298 normative), 0 findings
```

### Deriving the specification document

```
speclink generate -root . -out SPECIFICATION.md ./...
```

Every requirement with the words it was given, where those words came from, what
implements them, what demonstrated them, who has read them — and a gap list of
everything missing, each accepted gap carrying the reason somebody had to write
for it.

This is the part that decides whether the tool is worth having. As long as
speclink exists *beside* a hand written specification, it makes the situation
worse: one more thing to keep in step. Nothing can be removed until the document
comes out of here.

Markdown, and deliberately nothing cleverer. It renders everywhere and it diffs,
and a diff is the form in which this document is actually reviewed.

It renders a project with open findings, because that is exactly when somebody
wants to read it.

### Tracing what a change reaches

```
speclink impact -root . R-QUOTE-SUBMIT
speclink impact -root . requirements/_sources/sales/quoteflow.md#8-abgabe
speclink impact -root . app/sales/uc_submit_quote.go
```

The chain runs source segment → requirement → derived requirement → construct.
Every one of those edges was already being computed for a percentage and thrown
away; this walks them.

It answers the two questions the loop keeps asking. Somebody edits a paragraph
and has to know which code that reaches. An agent is handed a diff and has to
know which requirements it touches. Neither is answerable by reading the code,
because the chain runs through the tree, and neither by reading the tree,
because it runs through the document.

The file direction matches per file, not per package. The sidecar convention is
what makes that exact: `<base>.annotation.go` names precisely the constructs
declared in `<base>.go`. Matching on the package instead would return every
requirement the package has, which is not an answer to anything.

This command reports rather than judges. There are no findings, and reaching
nothing is an answer, not an error.

### Listing what was recognised

```
speclink inventory -root . ./...
+ event        example.com/erp/app/sales.QuoteSubmitted
  repository   example.com/erp/app/sales.CustomerRepository

event           2  2 bound
repository      2  0 bound
```

`verify` reports what it objects to, which is right for a build step and wrong
for the question "does the tool see the same system the specification
describes?" — a construct recognised correctly produces no output at all. The
inventory answers that one. `+` marks a construct that names a requirement.

### Diagnostic format

```
path/to/file.go:12:6: [SPEC-V6-056] use case SubmitQuote has no permission of its own.
    A permission per use case is what makes authorisation assignable and auditable. …
    Add `PermSubmitQuote = permission.Declare[SubmitQuote]("…", name, description)` and check it in NewSubmitQuote.
```

The code is `SPEC-<phase>-<number>`. Phases run in order and each only runs when
the previous one is clean:

| Phase | Meaning |
|---|---|
| `V1` | The annotation file uses a Go construct that the subset forbids |
| `V2` | The Go compilation itself (reported by the Go compiler, not by speclink) |
| `V3` | A binding attaches to an illegal target |
| `V4` | An annotation states a fact that is already established elsewhere |
| `V5` | Requirement tree: sources, anchors, IDs, directory layout, the DAG |
| `V6` | The specification and architecture rules |

`-format json` emits the same findings with `"version": 1` and the fields
`code`, `phase`, `what`, `why`, `how`.

### The profile

Every project names one, in `speclink.json`:

```json
{ "profile": "go_nago_ddd1" }
```

| profile | |
|---|---|
| `go_nago_ddd1` | Go on nago, DDD in three layers with a functional core |
| `go_bare_ddd1` | Go with no framework, over a hand written foundation |
| `java_springboot_ddd1` | Java on Spring Boot, same architecture; no rules yet |

The name has three parts because there are three decisions, and they are a
chain rather than a product: the language decides the reader, the framework
decides what counts as a use case, and the style decides the rules — which are
written in terms of those roles. There is no `go × springboot` cell that nobody
filled in; there is no such cell.

The third part is a **style**, not an organisation. Nothing in nago says a use
case lives in `uc_<name>.go` or that a context bundles its use cases in a
`UseCases` struct. That is an architecture, and it is selectable because a
project can reasonably follow a different one. `ddd1` and a later `ddd2` are two
styles rather than two versions of one, so they are numbered rather than
versioned.

A style carries two things. Its **rules** are what K4 to K8 check. Its
**conventions** are how the answers are spelled — that a use case file is
`uc_submit_quote.go` and its constructor `NewSubmitQuote` — and the rules ask
rather than assume. "One file per use case, named after it" is architectural;
`uc_` is not, and separating them is what lets two projects follow one
architecture and disagree about the spelling.

What a style deliberately cannot do is switch a rule off. That is `spec.Waive`
with a reason, or the scope, and a third way would be severities under another
name.

### Two styles over one language

`go_nago_ddd1` and `go_bare_ddd1` share a language and a reader and agree about
little else.

```
app/<bc>/              domain: use case types, models, ports, the UseCases bundle
app/<bc>/rest/         package rest<ctx> — this context's routes
app/<bc>/cli/          package cli<ctx> — this context's commands
app/<bc>/adapter/fs/   package fs — an implementation of a port
foundation/            auth, permission, data, rest, flag
cmd/<bin>/             the entry point, and the only place that names an adapter
```

A use case has the same shape as under nago — `func(subject auth.Subject, cmd In)
(Out, error)` — because a subject as a parameter forces the caller to decide who
is calling, where a context can be passed without deciding anything.

Four roles rather than eight. Command, event and projection are the vocabulary
of event sourcing and name nothing here; query is absent because nothing tells
it from a use case once every one of them returns `(Out, error)`.

Three rules the other style has no use for:

| | |
|---|---|
| `K6-ADAPTER-WIRED-IN-CMD` | only `cmd/` may import `app/<bc>/adapter/**` |
| `K6-PRESENTATION-NO-BUNDLE` | a handler takes the use cases it calls, not the bundle |
| `K6-CTX-NO-PRESENTATION-IMPORT` | a context does not import another's presentation |
| `K6-CTX-PRESENTATION-PKG` | the package is named `rest<ctx>`, `cli<ctx>` |

And two terms. `spec.Persistence()` marks an interface as a port or a struct as
a stored shape, because a hand written interface says neither.
`spec.StoredAs[Domain]()` marks a struct as the written form of a domain type,
which moves the promise onto it and leaves the domain type free to change.
Under `go_nago_ddd1` both are findings: `data.Repository` and the repository
constructors already say it, and saying it twice is the one thing this language
forbids.

**speclink will not guess it.** A missing profile stops the run with the list
above. Language could be worked out from a `go.mod` and framework from an
import, but style cannot be worked out from anything, and guessing it wrongly
reports dozens of findings about a convention the project never meant to
follow — which teaches the reader that the tool is wrong rather than that the
project is.

### Configuration

`speclink.json` states **deviations from the profile**, never the conventions
themselves. A project that follows its style writes only the profile.

```json
{
  "profile": "go_nago_ddd1",
  "cmdRoot": "tools",
  "scope": ["app/sales"]
}
```

Every profile understands `sourceRoots`, `scope` and `exclude`. Beyond those:

| profile | also understands |
|---|---|
| `go_nago_ddd1` | `contextRoot`, `cmdRoot`, `infraRoots` |
| `go_bare_ddd1` | the same, plus `foundationRoot` |
| `java_springboot_ddd1` | `classRoots`, `sourceCode`, `reportRoots`, `specPackage` |

Anything else is **refused, not ignored**. `classRoots` under a Go profile is
not a harmless leftover — it is somebody expecting an effect that will not
happen, almost always a profile that changed without the configuration
following.

Path patterns support `dir/**`.

`-config <file>` reads the layout from somewhere else, which is how a project
can be measured without being modified. A file named this way must exist; only
the conventional location may be absent.

### Scope is the only dial

There are no warnings, no severities and no tolerance mode, and `scope` is the
single thing that can be turned. A package is measured or it is not.

The reasons compound. The Go compiler behaves the same way. Softening would be
incoherent anyway, because the annotations sit in the ordinary build and a
compile error cannot be downgraded to a warning. Warnings meant for a migration
become a permanent excuse — not a guess, but the standing behaviour of every
codebase with a warning backlog. And the reader is a model, which iterates
until green: commented-out code is visibly unfinished, a suppressed warning is
invisibly unfinished.

So a codebase is brought in **package by package, not rule by rule**. The
difference is the whole point: *this package is not under speclink yet* is a
true statement, *this rule half applies here* is not one.

Three things follow, and all three are load-bearing:

- **The scope decides what is measured, never what is loaded.** Rules that
  resolve across packages — the permission i18n check follows a helper one step
  — keep seeing the whole module. A rule that changed its verdict on untouched
  code because of a configuration setting would be worse than one switched off.
- **The requirement tree is always read in full.** An in-scope construct may
  bind to any requirement. What the scope decides is what is *asked of* a
  requirement, not whether it exists: a requirement declared outside the scope
  is not reported as uncovered by code nobody claimed to have brought in yet.
- **A restricted run says so.** `2 packages outside the configured scope and
  not measured` follows the summary. Without it a hundred percent would be true
  of what was looked at and silent about what was not.

---

## 3. Project layout

```
cmd/<name>/main.go                          the entry point, exactly here
app/<context>/                              a bounded context (the domain)
app/<context>/usecases.go                   the UseCases bundle
app/<context>/uc_<use_case>.go              one use case per file
app/<context>/uc_<use_case>.annotation.go   its binding to a requirement
app/<context>/ui/                           package ui<context>, the views
app/<context>/cfg/                          wiring; the only place ui and domain meet
pkg/, foundation/                           infrastructure, domain free

requirements/dec/R-DEC-<NAME>.spec.go      Kind: Decision
requirements/nfr/R-NFR-<NAME>.spec.go      Kind: NonFunctional
requirements/cst/R-CST-<NAME>.spec.go      Kind: Constraint
requirements/fun/<domain>/R-<DOMAIN>-<NAME>.spec.go   Kind: Functional
requirements/fun/<domain>/R-<DOMAIN>-<NAME>/          attachments, no .go
requirements/_sources/                     the raw source documents
requirements/_material/                    material shared between requirements
```

Directory, ID prefix and `Kind` are the same fact stated three times, and
speclink checks that they agree. Cross-cutting kinds are grouped by kind because
they have no domain home; functional requirements are grouped by domain.

What is enforced and what is not: the kind level `dec/`, `nfr/`, `cst/`, `fun/`
is fixed, and so is `<ID>.spec.go` as the file name. The name of the directory
above it is not — speclink finds the kind level anywhere in the path, and a
requirement outside such a tree is simply not layout checked. `requirements/`
and `_sources/` are therefore convention, and the diagnostics name them so that
a project does not have to invent its own. `Source.Doc` is only required to be a
repository-relative path to a file that exists; putting it under `_sources/` is
convention too.

---

## 4. The three kinds of file you write

### 4.1 Requirement files: `<ID>.spec.go`

One requirement per file, the file named after the ID. Requirements may only
appear in the requirement tree, never in an annotation file: a requirement is
owned by the domain side and outlives any implementation.

```go
// Package quote holds the requirements of the quotation domain.
package quote

import "github.com/worldiety/speclink/spec"

var RQuoteSubmit = spec.Requirement{
	ID:         "R-QUOTE-SUBMIT",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Quote number on submission",
	Text:       "On submitting an approved quote a sequential, duplicate free quote number MUST be drawn.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/sales/quoteflow.md", Anchor: "8-abgabe"},
	},
}
```

Field rules:

- `Text` is normative and short — one sentence. It appears in lists, matrices
  and diagnostics. Long form belongs in the Markdown file named by `Detail`.
- `Kind`: `Functional`, `NonFunctional`, `Constraint`, `Decision`.
  A `Decision` **must** carry a `Rationale`.
- `Discipline`: `Business`, `Technical`, `Mixed`.
- `Status`: `Normative`, `Abstract`, `Planned`, `OutOfScope`, `Informative`,
  `Superseded`. **Only `Normative` must be covered.** `Abstract` is a pure
  derivation node and must *not* be covered directly; `Superseded` must no
  longer be covered.
- `Sources` is mandatory for `Normative`. Exactly one of `Doc` and `Extern` per
  source. `Doc` is a repository-relative path below `requirements/_sources/`
  and the file must exist. `Anchor` is the slug of a heading in that file:
  lower case, spaces and dashes become `-`, punctuation dropped. `## 8. Abgabe`
  yields `8-abgabe`. speclink verifies the heading exists. `Extern` carries laws
  and standards with no document in the repository, e.g. `"HGB §§ 383 ff."`.
- `DerivedFrom` and `Supersedes` reference other requirements by their **Go
  identifier**, never by the ID string, so moving a file never breaks a
  reference. Cycles in `DerivedFrom` are reported.

### 4.2 Annotation files: `<base>.annotation.go`

Sits next to `<base>.go`, in the same package, and is part of the **normal
build**. That coupling is deliberate: a broken annotation breaks the production
build, so it cannot rot unnoticed.

```go
package sales

import (
	"example.com/erp/requirements/fun/quote"
	"github.com/worldiety/speclink/spec"
)

var _ = spec.For[SubmitQuote](
	spec.Satisfies(quote.RQuoteSubmit),
	spec.Help(`Submit the approved quote. The system draws the next quote
number from the central registry.`),
)
```

**The grammar is a closed whitelist, not "valid Go".** An annotation file may
contain only the package clause, imports, and package-level `var _ = spec.X(…)`
terms. Everything below is rejected:

| Forbidden | Code |
|---|---|
| function definitions | `SPEC-V1-001` |
| type declarations | `SPEC-V1-003` |
| constant declarations | `SPEC-V1-004` |
| a `var` binding more or fewer than one name to one value | `SPEC-V1-005` |
| a binding not written as `var _ = …` | `SPEC-V1-006` |
| calls into anything other than `github.com/worldiety/speclink/spec` | `SPEC-V1-007` |
| positional fields in a struct literal (use `Field: value`) | `SPEC-V1-009` |
| the address-of operator `&` | `SPEC-V1-011` |
| function literals, binary expressions, any computation | `SPEC-V1-010` |

An annotation file whose neighbour `<base>.go` does not exist is an orphan
(`SPEC-V3-001`): rename it or delete it.

#### Binding functions

These four are the entire side-effecting surface of the language.

| Binding | Target |
|---|---|
| `spec.For[T](…)` | a named type: use case func type, event, aggregate, projection state |
| `spec.ForDecl(ref, …)` | a declared function, variable or constant, named by the value itself |
| `spec.ForField[T]("Name", …)` | a struct field of `T` |
| `spec.ForPackage(…)` | the package of the neighbouring file |

`ForDecl` takes the declaration as a value, so the Go compiler proves it exists.
`ForField` takes the field name as a string — the only reference in the whole
language the compiler cannot check, so speclink checks it against the type.

#### Assertions

Assertions are pure and are passed into a binding. They are never standalone.

| Assertion | Meaning |
|---|---|
| `spec.Satisfies(reqs…)` | binds the construct to one or more requirements. **The central statement, and the only one that cannot be inferred.** |
| `spec.Transition[T](state)` | after event `T` the aggregate assumes this lifecycle state; every event needs one, see `K15` |
| `spec.External()` | this event arrives from outside, so nothing here produces it |
| `spec.Help(text)` | end-user instruction for documentation and help. Use a raw string for several lines |
| `spec.Term(g)` | anchors a glossary entry at the construct that defines it |
| `spec.Rationale(text)` | justifies a decision at the construct implementing it |
| `spec.Waive(rule, reason)` | suspends one rule here; the reason is mandatory |
| `spec.Draft()` | this persisted shape is not promised yet; see section 6 |
| `spec.Optional()` | this field may be absent from stored data; see section 6 |
| `spec.StoredAs[D]()` | this struct is the written form of domain type `D`; `go_bare_ddd1` only, see section 11a |

A field binding satisfies a requirement for that field, not for the type it
belongs to — annotating one field does not make the whole command covered.

---

### 4.3 Process files: `<name>.process.go`

A process is the course of business: not what there is, but what happens, in
which order, and where it can end. It answers the question a building block view
cannot, and it satisfies requirements the way a construct does.

```go
var PQuoteDecision = spec.Process{
	ID:        "P-QUOTE-DECISION",
	Title:     "Angebot bis zur Entscheidung",
	Purpose:   "Ein abgegebenes Angebot wird freigegeben, zurückgezogen oder zurückgereicht.",
	Satisfies: []spec.Requirement{quote.RQuoteSubmit, quote.RQuoteApprove},
	Nodes: []spec.Node{
		spec.Start("entwurf", "Angebot ist entworfen"),
		spec.Do[sales.SubmitQuote]("abgeben"),
		spec.Emit[sales.QuoteSubmitted]("abgegeben"),
		spec.Choice("pruefen"),
		spec.Do[sales.ApproveQuote]("freigeben"),
		spec.End("freigegeben", "freigegeben"),
		spec.End("verworfen", "zurückgezogen"),
	},
	Edges: []spec.Edge{
		{From: "entwurf", To: "abgeben"},
		{From: "abgeben", To: "abgegeben"},
		{From: "abgegeben", To: "pruefen"},
		{From: "pruefen", To: "freigeben", When: "angenommen"},
		{From: "pruefen", To: "verworfen", When: "abgelehnt"},
		{From: "freigeben", To: "freigegeben"},
	},
}
```

**It lives above the contexts**, in a package of its own, because a process is
precisely the thing no single context owns. A context that declared it would
have to import its neighbour, which the layering rules forbid.

**A step names a construct, not a caption.** `spec.Do[sales.SubmitQuote]` is
checked by the compiler and then by `K16-NODE-REF-UNKNOWN`, which refuses an
activity that names an event and an event that names a use case. A step written
as prose would be a caption, and a caption cannot go stale in a way anybody
notices.

**The vocabulary is nine nodes**, and what is missing is missing on purpose:

| Node | Means |
|---|---|
| `spec.Start(id, label)` | a beginning; more than one is allowed |
| `spec.End(id, label)` | an outcome; several are the point, not a defect |
| `spec.Do[UseCase](id)` | a step somebody takes |
| `spec.Emit[Event](id)` | a fact is recorded here |
| `spec.On[Event](id)` | the course waits for a fact |
| `spec.Fork(id)` / `spec.Join(id)` | branches that all run, and the wait for them |
| `spec.Choice(id)` / `spec.Merge(id)` | exactly one branch, and the rejoining |

Inclusive and event based gateways, boundary events, timers, compensation and
subprocesses are deliberately absent. Each is expressible the moment a real case
asks for it, and each one added before then would be another vocabulary nobody
reads.

**Why a graph and not nested blocks.** Real processes come back: a quote sent
for rework returns to the drafting step. A nested form can express everything
except the jump backwards, which is the one thing that occurs constantly. The
price is that the compiler no longer checks the wiring — and that price is paid
by `K16`, which refuses a dangling edge, a duplicated name, a node no start
reaches and a node from which no end can be reached.

**What is not proved.** Whether every fork is matched by exactly one join on
every path is reachability in a Petri net and is not cheaply decidable. The
degree rules make the common shapes right and the tracer finds loops with no
exit; deadlock freedom is not established, and the rule set says so rather than
implying otherwise.

## 5. nago in one page

nago is the application framework this project is built on. speclink knows the
framework, never your project: everything below is recognised by resolved type,
so an alias import or a renamed identifier cannot fool it and cannot help you
either.

### What speclink recognises, and from what

| You write | speclink infers | Must name a requirement? |
|---|---|---|
| named func type, first param `auth.Subject` | **use case** | yes |
| the same, but returning data | **query** | yes |
| type with `Decide(auth.Subject, *Agg) ([]Evt, error)` | **command** | yes |
| type with `Evolve(ctx, *Agg) error` **and** `Discriminator()` | **event** | yes |
| type with `Identity() ID`, or the type an event's `Evolve` folds into | **aggregate** | no |
| `permission.Declare…[UseCase](id, …)` | **permission**, bound to that use case | no |
| state type of `evs.NewProjection` / `evs.NewSingleton` | **projection** | yes |
| `type X data.Repository[E, ID]` or `= data.ReadRepository[…]` | **repository** | no |

Aggregates, permissions and repositories are structural: they are covered
through the use case that guards, writes or holds them. Everything else names
its own requirement.

A query is no exception. Reading is a promise too — that someone may see a
thing — and every architecture rule already treats a query as a use case: its
own `uc_` file, its constructor, its permission, its place in the bundle. A
projection likewise: it crosses aggregates and answers a question no single use
case states, so nothing covers it transitively.

### Two type aliases that will bite you

- `auth.Subject` is an alias of `user.Subject`. The resolved type reports the
  `user` package.
- `evs.SeqID` is an alias of `ndb.Seq`. A use case returning `(evs.SeqID, error)`
  is a *write*, not a query.

### Packages

```
go.wdy.de/nago/auth                      auth.Subject
go.wdy.de/nago/application/user          user.Subject, user.PermissionDeniedErr
go.wdy.de/nago/application/permission    permission.ID, Declare…, Auditable
go.wdy.de/nago/application/rebac         rebac.Namespace, rebac.Instance
go.wdy.de/nago/application/evs           Cmd, Evt, Handler, Projection
go.wdy.de/nago/pkg/data                  Repository, ReadRepository, Aggregate
go.wdy.de/nago/pkg/ndb                   the message store (Seq, Followable)
go.wdy.de/nago/pkg/cloner                Cloner, required by a projection state
```

---

## 6. Promising a shape

An event struct is the wire format. It is written into a log that is replayed
forever, so every field name and every field type is a promise from the moment
the first message is stored.

Two things are persisted, and only two:

| | |
|---|---|
| **an event** | it implements `Evolve` and carries a discriminator, so the struct *is* the stored form |
| **a persistence model** | the type a repository was built over, wherever that happened |

The second is not visible in the declaration. Nothing about a struct says it is
stored; the construction says it, and the framework makes the choice explicit:

```go
// Two models, mapped. Only CustomerEntity is promised; Customer stays free.
json.NewJSONRepository[Customer, CustomerID, CustomerEntity, CustomerID](
	store, intoDomain, intoPersistence)

// One model. The domain type IS the stored form, and is promised as it stands.
json.NewSloppyJSONRepository[Ledger, LedgerID](store)
```

Prefer the first for anything that will outlive a prototype. The framework's own
documentation calls the sloppy form a shorthand for throw-away code where
neither model has been stabilised — and it is exactly that: from the moment you
choose it, every rename in your domain is a change to stored data.

**Everything persisted is frozen by default.** `spec.Draft()` is the exception,
and committing to a shape is not an act of writing anything — it is deleting
that term and recording what remains.

The inversion is deliberate. The number of drafts in a system shrinks over its
lifetime, so marking the exception costs less than marking the rule. And
forgetting fails safe: an unmarked newcomer is frozen at once, which surfaces as
an error the first time somebody changes it, rather than as unreadable data a
year later.

### When do I remove the term?

`spec.Draft()` claims one thing, and it is not about deployment: **we are
willing to delete every stored message of this type.** The framework can do
exactly that — `ndb.DeleteType` removes them all — which is what makes the state
real rather than aspirational.

So remove it the moment nobody would purge any more. That moment has nothing to
do with going live. A development database can hold data somebody minds losing,
and from then on the shape is promised whatever the source says; a deployed
system whose data may be thrown away at will is still a draft. As long as you
would call `DeleteType` without hesitating, the term is honest.

Removing it is a two-step act: delete the term, then run `speclink freeze`. The
diff of `speclink.lock` is what you review.

### The cascade

```go
var _ = spec.ForPackage(spec.Draft())          // every persisted type in it
var _ = spec.For[QuoteWithdrawn](spec.Draft()) // every field of the type
var _ = spec.ForField[Quote]("Note", spec.Draft())
```

Attach it at the level that is actually true. Marking a level that is already
covered from above states nothing new and is reported (`K9-DRAFT-REDUNDANT`,
phase V4) — a field term only means something once the type itself is frozen.

Working on a new context? One `spec.ForPackage(spec.Draft())` covers it, and
it is deleted once when the context goes live.

### What you may still change

While something is a draft: anything. Add fields, remove them, retype them,
delete the event outright. The framework can purge every message of a type, so
nothing is stranded.

Give the tag a namespace. `"sales.quote.submitted.v1"` cannot collide with
another context and survives a rename of the Go type; a bare `"QuoteSubmitted"`
carries no package, so two contexts naming a type alike end up in one stream,
and renaming the type for clarity orphans everything written so far.

Once it is frozen:

| | |
|---|---|
| **The discriminator must never change** | It is the key stored messages are decoded by. Changing it does not rename anything, it orphans every message written under the old tag. |
| **and it must be unique in the module** | It is used directly as the store's type id, so it is global. Two types sharing one write into the same stream and read each other's messages. Checked for drafts too, because that is corruption rather than a broken promise. |
| A field must not be **removed** | Readers cannot tell an absent value from one that was never written. |
| A field must not be **renamed on the wire** | The Go field name may change freely; the json tag is what is promised. |
| A field's **shape** must not change incompatibly | `string` to a named type with underlying `string` is fine, and so is any change between integer widths. `int` to `string`, or a scalar to a slice, is not. |
| **New fields are allowed** | but must be marked `spec.Optional()`, because messages written before they existed do not carry them. |
| **Optionality cannot be withdrawn** | those messages still lack the field, and no release can reach them. |

### speclink.lock

None of this is decidable from the current source. Nothing in a working tree
says what a field used to be, so the promise is recorded in `speclink.lock`.

The file is written by `speclink freeze` and **never edited by hand**. It is not
a second place to state intent — intent stays in the code, where the field type
says the shape and `spec.Draft` says the status. It records what has already
been committed to, the same relation `go.sum` has to `go.mod`.

```
speclink freeze -n ./...    # what would be recorded
speclink freeze ./...       # record it
```

`freeze` refuses to record anything that already breaks a promise. Otherwise it
would be the one command that makes a break official, and every rule reading the
file would be worthless.

### Growing a promised type

```go
type QuoteSubmitted struct {
	QuoteID     string
	QuoteNumber string
	Channel     string // added later
}
```

```go
var _ = spec.ForField[QuoteSubmitted]("Channel",
	spec.Optional(),
)
```

Then `speclink freeze` to record it. Without the term the field is reported
(`K9-FIELD-ADDED-REQUIRED`): the shape was promised before it existed, so every
message stored until then lacks it, and nothing the writer does from here on
changes that.

Two shape changes are free and need no term at all, because the record cannot
tell them apart from no change:

```go
QuoteID string   →  QuoteID QuoteRef   // named type, underlying string
Count   int32    →  Count   int64      // any integer width
```

The fingerprint records the underlying structure and collapses every integer
width into one token. A promise about an integer field is that it holds a whole
number, not that it holds thirty-two bits of one. Floating point is not
collapsed: narrowing a `float64` to a `float32` loses precision silently.

A frozen type that has never been recorded is reported
(`K9-BASELINE-MISSING`). That finding is the point at which somebody has to
decide: promise it, or mark it a draft. And once a shape is recorded, marking
it a draft again is refused — a promise cannot be taken back by editing
source, because the stored messages do not read it.

---

## 7. Choosing a persistence pattern

This is a decision you make **per aggregate**, driven by the requirements — not
once for the project, and not by habit. The three patterns below are meant to be
combined.

### 7.1 Decide / Evolve — event sourcing

The write side. A command decides, an event records, the aggregate folds.

```go
// The aggregate. Must be a pointer type; Evolve mutates it in place.
type QuoteAggregate struct {
	ID     string
	Status string
}

func (q QuoteAggregate) Identity() string { return q.ID }

// The command. Decide authorises and validates, and returns facts.
// It must not mutate: the aggregate it receives is the state to decide against.
type SubmitQuoteCmd struct {
	QuoteID string
	Title   string
}

func (c SubmitQuoteCmd) Decide(s auth.Subject, q *QuoteAggregate) ([]QuoteSubmitted, error) {
	if err := s.Audit(PermSubmitQuote); err != nil {
		return nil, err
	}
	return []QuoteSubmitted{{QuoteID: c.QuoteID}}, nil
}

// The event. Evolve folds it in; the discriminator is the serialisation tag.
type QuoteSubmitted struct {
	QuoteID     string
	QuoteNumber string
}

func (e QuoteSubmitted) Evolve(_ context.Context, q *QuoteAggregate) error {
	q.Status = "submitted"
	return nil
}

func (e QuoteSubmitted) Discriminator() evs.Discriminator { return "sales.quote.submitted.v1" }
```

Wire it with `evs.NewHandler[*QuoteAggregate, Evt, Primary](backend, aggID, register)`,
then `RegisterEvents(…)` before any other call.

- **The discriminator is versioned and permanent.** The log is replayed forever;
  changing the tag of an existing event type orphans every message already
  written under it. Add `.v2`, never edit `.v1`.
- The aggregate type must be a pointer and implement `Clone()` and
  `IsDeleted()`. Once `IsDeleted` reports true the handler drops it, and it
  stays gone on replay.
- **Choose this when** history or auditability is itself a requirement, when
  decisions must be reconstructible, or when invariants are per-aggregate.
- **Cost:** every schema change becomes an event version, and the whole log is
  replayed at start.

### 7.2 Projection — the read side of 7.1

`Evolve` targets exactly one aggregate. A read model must fold the same event
into an arbitrary target, possibly under a different key. That is why a
projection registers its fold **at the target**, not on the event.

> `evs.Projection` requires nago `v0.0.0-20260806113855` or newer. On an older
> version the type does not exist and this whole section does not apply; fold
> into a repository instead and state in the decision requirement why.

```go
type QuoteOverview struct {
	Submitted int
	LastQuote string
}

// Required: readers are handed a deep clone, so an accidental mutation on a
// returned value can never corrupt the folded state.
func (o *QuoteOverview) Clone() *QuoteOverview {
	if o == nil {
		return nil
	}
	c := *o
	return &c
}

func newQuoteOverview(src evs.Source) *evs.Singleton[*QuoteOverview] {
	p := evs.NewSingleton[*QuoteOverview](src, evs.ProjectionOptions{})
	evs.Project(p,
		func(QuoteSubmitted) evs.Unit { return evs.TheUnit() },
		func(s *QuoteOverview, e QuoteSubmitted) {
			s.Submitted++
			s.LastQuote = e.QuoteNumber
		},
	)
	return p
}
```

Use `evs.NewProjection[K, *S]` when there is a key, `evs.NewSingleton[*S]` when
the whole log folds into one row. Call `Run()` once — it replays the history and
then follows live writes through the same code path — and read with `Get(k)` or
`All()`.

- **A projection keeps no persistence of its own.** It is rebuilt by
  constructing it again. Persisting one turns a derived view into a second
  truth; do not do it.
- **It is not read-your-write.** Folding is asynchronous. When a caller must
  observe its own write, use the sequence the write returned:

  ```go
  env, _ := backend.Append(subject, evt)
  _ = view.WaitFor(ctx, ndb.Seq(env.Sequence))
  v, _ := view.Get(key) // now guaranteed to include env
  ```

  For an interactive screen this is irrelevant; a re-render follows the event.
- `fold` is deliberately errorless. Report and skip exceptional events through
  `ProjectionOptions.OnError`; a panic inside `fold` is recovered and routed
  there too, so one bad event never stalls the projection.
- **Choose this when** a query crosses aggregates, needs a different key, or
  feeds a list or dashboard. Never answer a query by replaying the log inside
  the request.

If you chose 7.1 you will almost always need 7.2 as well. They are one pattern
in two halves.

### 7.3 Repository — document and relational storage

Current-state storage of an aggregate. The framework idiom is to name the
instantiation once:

```go
type QuoteRepository data.Repository[QuoteAggregate, string]
```

`E` must satisfy `data.Aggregate[ID]` (`Identity() ID`); `ID` is constrained by
`data.IDType` (`~int | ~int64 | ~int32 | ~string`). Depend on
`data.ReadRepository[E, ID]` when a use case only reads — the lack of write
access then shows in the signature instead of in a comment.

Things that are easy to get wrong:

- `FindByID(id) (option.Opt[E], error)` returns an **option**, not a zero value
  and a bool. Unwrap it explicitly.
- Traversals are `iter.Seq2[E, error]`: `All`, `FindAllByPrefix`, `Identifiers`,
  `IdentifiersByPrefix`, `FindAllByID`. Handle the per-element error.
- **Never call back into the same repository from inside a `yield`.** Most
  implementations deadlock.
- Prefix iteration is lexicographic. Integer keys have no leading zeros, so
  prefix queries on them do not mean what they look like.
- `Delete(predicate)` is always O(n); returning `data.SkipAll` stops the
  traversal without error.

**Choose this when** the aggregate is current-state, history is not a
requirement, or the data is unbounded or owned by someone else.

### 7.4 Mixing them

Legitimate, and common in one project:

- an event-sourced core aggregate, because its history is a requirement,
- repositories for reference and master data, where history is noise,
- projections for every screen that spans aggregates.

Not legitimate:

- the same aggregate written both through a repository and through an event log
  — that is two truths and they will diverge,
- a projection given its own persistence — it must stay rebuildable,
- repository or handler access from a `ui*` package — that inverts the
  dependency (`K6-CTX-NO-UI-IMPORT`).

Record the choice. A persistence pattern is a `Kind: spec.Decision` requirement
with a `Rationale`, bound at the construct that implements it with
`spec.Rationale(…)`. That is how the choice stops being tribal knowledge.

---

## 8. Writing a use case

Every rule below is checked. This template satisfies all of them.

**File `app/sales/uc_submit_quote.go`** — one use case per file, named after it,
type and constructor together:

```go
package sales

import (
	"go.wdy.de/nago/application/evs"
	"go.wdy.de/nago/auth"
)

// SubmitQuote submits an approved quote and draws a quote number.
type SubmitQuote func(subject auth.Subject, cmd SubmitQuoteCmd) (evs.SeqID, error)

// NewSubmitQuote builds the submission use case.
func NewSubmitQuote(numbers NumberRegistry) SubmitQuote {
	return func(subject auth.Subject, cmd SubmitQuoteCmd) (evs.SeqID, error) {
		if err := subject.Audit(PermSubmitQuote); err != nil {
			return 0, err
		}
		if _, err := numbers.Next(); err != nil {
			return 0, err
		}
		return 1, nil
	}
}
```

**File `app/sales/perm.go`** — one permission per use case, texts from the
translation catalogue:

```go
var PermSubmitQuote = permission.DeclareCreate[SubmitQuote]("sales.quote.submit", "Quote")
```

**File `app/sales/usecases.go`** — the bundle:

```go
type UseCases struct {
	SubmitQuote SubmitQuote
}

func NewUseCases(numbers NumberRegistry) UseCases {
	return UseCases{SubmitQuote: NewSubmitQuote(numbers)}
}
```

**File `app/sales/uc_submit_quote.annotation.go`** — the requirement:

```go
var _ = spec.For[SubmitQuote](
	spec.Satisfies(quote.RQuoteSubmit),
)
```

### The checklist

- Signature: `auth.Subject` first, `error` last. Always. Every use case can fail
  authorisation, so the error is not optional.
- The type lives in `uc_<snake_case_name>.go` and so does `New<Name>`, which
  returns `<Name>`.
- The implementation consults the subject. Any of these counts:
  `subject.Audit(…)`, `subject.AuditResource(…)`, `subject.HasPermission`,
  `HasResourcePermission`, `HasRole`, `HasGroup`, returning an error wrapping
  `user.PermissionDeniedErr`, or passing the subject on to another use case that
  checks.
- At least one permission is declared with the use case as type parameter, and
  the implementation actually references it. A declared but unchecked permission
  is worse than none: it appears in the role editor and promises a protection
  that does not exist.
- Permission texts come from i18n. Use a `permission.Declare<Verb>` helper
  (`DeclareCreate`, `DeclareFindByID`, `DeclareFindAll`, `DeclareUpdate`,
  `DeclareDeleteByID`, …), or wrap your own texts in `i18n.MustString(…)`. These
  strings appear in the role editor, where a non-developer decides who may do
  what.
- Dependencies enter through the constructor and are captured in the closure.
  Never read a package-level variable — permission IDs and constants excepted.
- The use case is a field of `UseCases` and is set in `NewUseCases`.

### Use `AuditResource` when the instance decides

```go
func NewFindQuoteOverview(view QuoteOverviewReader) FindQuoteOverview {
	return func(subject auth.Subject, customer string) (QuoteOverview, error) {
		if err := subject.AuditResource(Namespace, rebac.Instance(customer), PermFindQuoteOverview); err != nil {
			return QuoteOverview{}, err
		}
		o, _ := view.Get(customer)
		return o, nil
	}
}
```

---

## 9. Bounded contexts and layering

- A bounded context under `app/<ctx>/` defines what the system does. The user
  interface is one way of reaching it, and there may be others.
- **The domain must not import anything under `/presentation/` or any package
  whose name starts with `ui`.** Importing it inverts the ordering and makes the
  context untestable without a renderer.
- `app/<ctx>/ui/` declares package `ui<ctx>` — the directory is always `ui`, so
  the package name is what tells a reader which context's interface this is.
- `app/<ctx>/cfg/` is the wiring layer and is exempt: it is the one place views
  and use cases are supposed to meet.
- Every use case of a context is a field of `UseCases`, built by `NewUseCases`
  in `usecases.go`. Callers depend on the bundle, not on the internals.
- `pkg/` and `foundation/` are infrastructure. They must not import a bounded
  context and must not declare a use case. Infrastructure is what the domain
  builds on, not the other way round.
- Exactly one `main` package, under `cmd/`. A module without one is a library
  and must waive `K8-MAIN-EXISTS`.

---

## 10. Generic CRUD is not available

These are rejected by `K4-NO-GENERIC-CRUD`:

- `cfgent.Enable(…)`, `cfgent.EnableUseCases(…)`
- `ent.DeclarePermissions(…)`
- `ent.NewUseCases(…)`
- importing `go.wdy.de/nago/application/ent/ui` (`uient`)

**The ban is on the wiring, not on the design.** These helpers derive
permissions, use cases and routes at run time from a prefix, so those facts
exist only while the program runs and no requirement can ever be traced to the
code that satisfies it. That is the problem, and it is the only problem.

The screens `uient` renders are the reference for how a nago user interface
should look and behave. Take that style — lift it into your own components or a
template if it helps — and write the screen explicitly. What you may not do is
adopt the generic route.

Write the use cases by hand, declare one permission each, and bind them with
`spec.Satisfies(…)`. It is more code and it is the point: the module is not
wrong, it is merely too inflexible to carry a specification.

---

## 10a. The source documents

The requirement tree is not the top of the chain. Above it are the documents
people wrote — Markdown and mockups — and the step from those into the tree is
the only one with no formal semantics. Everything below the tree is held up by
the Go compiler; this is the part that decides whether the coverage figures
mean anything.

### What a source is

Two kinds, and no others. `sourceRoots` in `speclink.json` says where they
live, defaulting to `requirements/_sources`.

| Kind | Segmented by | Anchor is |
|---|---|---|
| `.md` | its headings | the heading slug |
| `.png`, `.jpg` | a sidecar manifest | the region name |

PDF is deliberately not supported. Convert it to Markdown: the conversion is
then a visible step, and the result is diffable in the pull request.

A document is a sequence of **segments**, not an atom, and `Source.Anchor`
addresses one of them. An image region is an anchor like any other — there was
never a difference from the requirement side between pointing at a section and
pointing at part of a screen.

### Mockups

Regions are declared next to the image in `<image>.speclink.json`:

```json
{
  "version": 1,
  "regions": [
    { "name": "kopfleiste", "rect": { "x": 0, "y": 0, "w": 240, "h": 28 }, "informative": true },
    { "name": "abgabeknopf", "rect": { "x": 148, "y": 128, "w": 80, "h": 22 } }
  ]
}
```

They are declared rather than found because an image is not decomposable by any
deterministic rule, and a model inventing regions would only move the
invented-requirement problem one level down.

Coordinates are what make drift specific: the fingerprint covers the pixels
inside the rectangle, so recolouring one button reports the requirements of
that button and nothing else. A layout shift moves everything below it and
reports drift across the board — visible, resolved by one `freeze`, and
preferable to a report so coarse that it gets ignored.

### Sections that carry no obligation

Every segment must produce at least one requirement. A title, an introduction,
a glossary or a chrome element genuinely does not, and says so where it is
written:

```markdown
# Einleitung

<!-- speclink:informative -->
```

For a region, set `"informative": true` in the manifest.

This is not `spec.Waive` and cannot be. A waiver attaches to a Go construct and
a section has none, so a waiver narrowed to one section could not be written
down at all. Stating it in the document puts the decision with the person who
wrote the section and keeps the fact in one place.

### Reviews

The people who own the source documents do not read Go and never see a
`.spec.go` file. They edit Markdown and mockups, read the requirements that were
extracted from them, and say whether those say what they meant.

That "yes" is recorded, not declared:

```
speclink freeze -reviewer "Frau Meier" ./...
```

There is deliberately no `Reviewed` field on `spec.Requirement`. It would be
written by the same model that wrote the requirement — a generator certifying
its own output, which is worth nothing. `freeze -reviewer` is run by a person,
or by a surface acting for a named person, which is a different claim.

A review is bound to the wording it was given for. Rewrite the text and the
review is gone, because what was read is no longer what is there. Running
`freeze` without `-reviewer` still records the text — the drift rules need
something to compare against — and records no review, because none happened.
That is what CI does.

Reviews are counted, never required. `speclink requirements` reports
`9 requirements (8 normative, 3 reviewed)`. Nothing fails for being unreviewed:
spot checks are the working model, and a rule demanding a hundred percent would
be answered by a script.

### The read surface

```
speclink requirements -format json -root . ./requirements/... > tree.json
```

`requirements` asks whether the tree is sound, so its machine readable answer
is the tree: every requirement with its text, its origin, its status and its
review state, every source segment with the requirements extracted from it, and
the findings as one field among the rest.

That last pairing is what makes a review possible at all. Judging whether an
extraction is faithful means reading the requirement next to the paragraph or
the part of the screen it came from, and this is the only output that puts the
two together.

It carries no satisfiers, because this command reads no annotation — that is
what lets the tree be worked on while the code around it is in pieces, and the
name of the Go type implementing a requirement tells a reviewer nothing anyway.

`speclink verify -format json` is unchanged and stays a findings list. Its
reader is an agent, which needs to know what is broken.

### Drift

A requirement text and a source segment are the same kind of edge as a wire
format: the compiler cannot re-derive them, so `speclink.lock` records what
they used to say.

Rewrite the text of a requirement and its identifier is unchanged, every
`spec.Satisfies` still compiles and the coverage stays at 100%.
`K10-REQ-CHANGED` is the only thing that notices. Rewrite the section it came
from and the anchor still resolves; `K13-SOURCE-DRIFT` is the only thing that
notices.

Both are answered the same way: re-read what the finding names, change what has
to change, then `speclink freeze`. The diff of `speclink.lock` is the review.

Reformatting is not drift and reflowing is not a rewrite. Both are pinned by
tests, because a rule that fires on a formatter is one that gets waived by
habit.

---

## 10b. Verification

Coverage says code was written for a requirement. It has never said the code
does what the requirement asks, and where the implementation and the tests come
from the same place and people review by sampling, nothing else does either.

At the end of a test, say what it demonstrated:

```go
func TestSubmitQuoteDrawsAGaplessNumber(t *testing.T) {
	// … exercise the use case, assert the number sequence …

	spec.Verified(t, quote.RQuoteSubmit)
}
```

**Position matters here, and nowhere else in this language.** `spec.Verified`
writes a line when it runs, so putting it at the end says the test got there.
Putting it at the top says only that the test started.

That is the point of it. A marker read from the source is a claim, and a claim
is not evidence: it can sit behind a condition that never holds, or in a test
that fails long before reaching it. So the call has two lives. It is read
statically, which is what makes a *missing* one reportable. It writes a line
when it executes, which is what makes a *present* one believable.

| in the source | recorded | meaning |
|---|---|---|
| no call | – | nothing claims to verify it — `K14-REQ-UNVERIFIED` |
| a call | nothing, or against an older wording | claimed, not shown — `K14-VERIFICATION-STALE` |
| a call | matching, from a passing test | demonstrated |

### Handing the results back

```
go test -json ./... | speclink evidence
```

This is the fourth step of the build order, and the only one that writes
evidence rather than reading it. It records which tests passed while claiming
which requirements, bound to the wording the requirement had at the time.

speclink does not run the tests itself. The build order puts them after it, and
a command that invoked the suite would either break that or duplicate it. It
also makes the evidence something CI hands over rather than something speclink
produces, which is the right way round.

Only passing tests are recorded, and a run is the whole truth: a requirement
nothing demonstrated this time loses its record. `K14-VERIFICATION-STALE` then
reports it, and the same finding covers three mistakes that look identical from
the source and are all fixed the same way:

- the call sits behind a condition that never holds
- the test fails before reaching the end
- the requirement was rewritten after the last run, so the evidence was for
  other words

The summary reports both figures, and the gap between them is the interesting
number: `100% verified, 88% demonstrated` means a test exists, compiles, claims
something, and has not been seen doing it.

This is the one place in speclink where a fact is produced at run time, and it
is not the exception it looks like. P9 bans constructs that turn *static* facts
into dynamic ones. A test result was never a static fact; no analysis can
derive it.

### When a requirement cannot be tested

Some cannot. A structural decision — *customer data is stored as state, not as
facts* — is discharged by the type existing at all, and a test for it could only
assert that the code compiles as written. Waive it on a construct that
satisfies it:

```go
var _ = spec.For[Customer](
	spec.Satisfies(dec.RDecCustomerState),
	spec.Waive("K14-REQ-UNVERIFIED", "The ruling is that a customer is stored as state rather than as facts, …"),
)
```

The waiver goes on the construct, not on the requirement: `spec.Waive` attaches
to a Go construct and a requirement declaration is not one. The finding names a
construct you can put it on.

---

## 11. Rule index

Every rule ID is stable and may be used in `spec.Waive`.

Not every rule runs everywhere. The **K1, K3, K10 to K14** families and the
requirement tree checks are universal: they are about requirements, sources,
coverage and evidence, and know nothing about a language. **K4 to K8** belong to
a profile's style, and a run that does not have them says so:

```
not measured: architecture, because profile java_springboot_ddd1 prescribes no rules yet
```

The **K16** family runs wherever a project declares a process. None declared
means no figure about processes is printed at all, because a share of nothing is
no claim.

The **K15** family runs wherever the frontend recognises events. An architecture
with none — `go_bare_ddd1` is not event sourced — has no lifecycles to check and
no figure that claims otherwise, so nothing is reported either way.

The **K9** family is neither. Its rules are language independent, but they have
nothing to run over until the framework's recognisers can point at a persisted
type — and recognising one is not something a language does. Each framework
states it differently:

| Profile | What says a type is stored |
|---|---|
| `go_nago_ddd1` | the repository constructor: `NewJSONRepository` names a persistence model distinct from the domain model, `NewSloppyJSONRepository` names none and ties the domain type to the wire |
| `go_bare_ddd1` | the element type of a repository port, `type R data.Repository[E, ID]`; an adapter that keeps a shape of its own says so with `spec.StoredAs` |
| `java_springboot_ddd1` | nothing yet |

A profile with no such recogniser says so rather than reporting the family
sound:

```
not measured: schema evolution, because profile java_springboot_ddd1 has no persistence recogniser
```

That line matters more than it looks. Without it such a run reports `0 findings`
and a summary reading 100% in every column, which is a clean bill of health for
stored data that no rule guarded. The distinction cannot come from the frontend:
both Go profiles share one reader, and it does read schemas. What has or has not
a notion of persistence is the framework, which is a property of the profile.

### Which shape is promised

Under `go_bare_ddd1` a repository names the domain type it stores, and where
that is all there is, the domain type is what ends up on disk: every rename in
it is a change to stored data. An adapter may instead keep a shape of its own
and map between the two, which is what buys the freedom to restructure the
domain without touching a byte. Nothing about the two structs says which
arrangement is in force, so it is stated:

```go
var _ = spec.For[QuoteStore](
	spec.StoredAs[sales.Quote](),
)
```

The promise moves onto `QuoteStore`, and `sales.Quote` is released. Without it
the promise stays on the domain type, which is the stricter reading and the
safe default. `spec.StoredAs` is a term of this style only; under
`go_nago_ddd1` the constructor already says it, and a second source of one fact
is refused rather than accepted.

| Rule | Codes | Meaning |
|---|---|---|
| `K1-CONSTRUCT-UNBOUND` | `V6-020` | a use case, command, event or projection names no requirement |
| `K3-REQ-UNCOVERED` | `V6-001` | a normative requirement is satisfied by nothing |
| `K3-ABSTRACT-COVERED` | `V6-002` | an abstract requirement was satisfied directly |
| `K3-SUPERSEDED-COVERED` | `V6-003` | a superseded requirement is still being satisfied |
| `K4-NO-GENERIC-CRUD` *(style)* | `V6-010`, `V6-011` | generic CRUD factory or its user interface |
| `K5-UC-FILE` *(style)* | `V6-050` | use case not in the file the style names |
| `K5-UC-SIGNATURE` *(style)* | `V6-051` | use case does not return `error` last |
| `K5-UC-CONSTRUCTOR` *(style)* | `V6-052`, `V6-053`, `V6-054` | constructor missing, misplaced, or returns the wrong type; the name is the style's |
| `K5-UC-AUTHZ` *(style)* | `V6-055` | nothing in the implementation looks like an authorisation check |
| `K5-UC-PERMISSION` *(style)* | `V6-056`, `V6-057` | no permission of its own, or one declared but never used |
| `K5-UC-DEPS` *(style)* | `V6-058` | use case reads a package-level variable |
| `K5-UC-PERMISSION-I18N` *(style)* | `V6-059` | permission carries hardcoded texts |
| `K6-CTX-UI-PKG` *(style)* | `V6-040` | the `ui` directory does not declare `ui<ctx>` |
| `K6-CTX-NO-UI-IMPORT` *(style)* | `V6-041` | a domain package imports the user interface |
| `K6-CTX-USECASES` *(style)* | `V6-042`, `V6-043`, `V6-044` | missing `UseCases`, missing `NewUseCases`, or a use case not in the bundle |
| `K6-CTX-NO-PRESENTATION-IMPORT` *(style)* | `V6-045` | a context imports another context's presentation |
| `K6-ADAPTER-WIRED-IN-CMD` *(style)* | `V6-046` | something outside `cmd/` imports an adapter |
| `K6-PRESENTATION-NO-BUNDLE` *(style)* | `V6-047` | a handler takes the whole `UseCases` bundle |
| `K6-CTX-PRESENTATION-PKG` *(style)* | `V6-048` | a presentation package is not named `rest<ctx>` or `cli<ctx>` |
| `K7-INFRA-DOMAIN-FREE` *(style)* | `V6-032`, `V6-033` | infrastructure imports a context or declares a use case |
| `K8-MAIN-LOCATION` *(style)* | `V6-030` | a `main` package outside `cmd/` |
| `K8-MAIN-EXISTS` *(style)* | `V6-031` | the module has no entry point |
| `K9-DRAFT-REDUNDANT` | `V4-001`, `V4-002` | a draft term at a level the cascade already covers |
| `K9-BASELINE-MISSING` | `V6-090` | a frozen shape that was never recorded |
| `K9-DISCRIMINATOR-FROZEN` | `V6-091` | the serialisation tag of a promised type changed |
| `K9-FIELD-REMOVED` | `V6-092` | a promised field is gone |
| `K9-FIELD-RENAMED` | `V6-093` | a promised field changed its stored name |
| `K9-TYPE-REMOVED` | `V6-094` | a promised type is gone; waive it on the package |
| `K9-DRAFT-FROZEN` | `V6-095` | a promise taken back by marking it a draft |
| `K9-FIELD-SHAPE` | `V6-096` | a promised field changed its stored shape |
| `K9-OPTIONAL-REVOKED` | `V6-097` | a field stopped being optional |
| `K9-FIELD-ADDED-REQUIRED` | `V6-098` | a field added to a promised type without `spec.Optional()` |
| `K9-DISCRIMINATOR-COLLISION` | `V6-099` | two persisted types claim the same serialisation tag |
| `K10-REQ-CHANGED` | `V6-110` | a requirement's text was rewritten under its satisfiers |
| `K11-REQ-UNSOURCED` | `V5-020` | a normative requirement cites no source |
| `K11-SOURCE-UNANCHORED` | `V5-026` | a citation names a document but no part of it |
| `K12-SOURCE-UNCOVERED` | `V6-100` | a section or region became no requirement |
| `K13-SOURCE-DRIFT` | `V6-111` | a source segment was rewritten under the requirements derived from it |
| `K16-PROCESS-NO-START` | `V6-070` | a process has no beginning |
| `K16-PROCESS-NO-END` | `V6-071` | a process has no way to finish |
| `K16-PROCESS-NO-ACTIVITY` | `V6-072` | a process routes and finishes but performs nothing |
| `K16-NODE-DUPLICATE` | `V6-073` | two nodes of one process share a name |
| `K16-EDGE-DANGLING` | `V6-074` | an edge names a node that does not exist |
| `K16-NODE-DEGREE` | `V6-075` | the wiring at a node cannot mean anything |
| `K16-CHOICE-UNCONDITIONAL` | `V6-076` | an alternative states no condition |
| `K16-NODE-UNREACHABLE` | `V6-077` | no start reaches a node |
| `K16-NODE-TRAPPED` | `V6-078` | from a node no end can be reached |
| `K16-NODE-REF-UNKNOWN` | `V6-079` | a step names something that is not what it performs |
| `K16-PROCESS-UNBOUND` | `V6-080` | a process names no requirement |
| `K16-PROCESS-DUPLICATE` | `V6-081` | two processes claim one ID |
| `K15-EVENT-NO-TRANSITION` | `V6-060` | an event does not say which state it leaves the aggregate in |
| `K15-TRANSITION-UNKNOWN` | `V6-061` | a transition names something that folds nothing |
| `K14-REQ-UNVERIFIED` | `V6-120` | no test demonstrates a normative requirement |
| `K14-VERIFICATION-STALE` | `V6-121` | a test claims a requirement but no run has shown it |

`K12-SOURCE-UNCOVERED` has no per-segment waiver; mark the segment informative
in the document instead. The undirected `spec.Waive` switches the whole rule
off, which is how a project adopts the rest of speclink before its documents
are in shape.

Requirement-tree findings (`V5`) are not waivable per construct, because they
concern the requirement, not the code. The two exceptions are
`K11-REQ-UNSOURCED` and `K11-SOURCE-UNANCHORED`, which are waivable because
both can genuinely fail to hold. `V5-001` missing ID, `V5-002` missing
Kind, `V5-003` missing Status, `V5-004` decision without rationale, `V5-005`
missing Text, `V5-013` cycle in `DerivedFrom`, `V5-020` normative requirement
without a source, `V5-021`/`V5-022` a source naming neither or both of
`Doc`/`Extern`, `V5-023` source document missing or unsegmentable, `V5-025` the
anchor names no segment of that document, `V5-027` a configured source root is
missing, `V5-030` file name ≠ ID, `V5-031` malformed ID, `V5-032` prefix contradicts
Kind, `V5-033` directory contradicts Kind, `V5-034` functional requirement
directly in `fun/`, `V5-035` domain directory contradicts the ID prefix.

---

## 11a. Other languages

`speclink requirements` reads a JVM project too, from compiled classes rather
than from source:

```
speclink verify -root . -profile java_springboot_ddd1
```

Normally the profile is in `speclink.json` and the flag is for trying one.

**Why bytecode.** A class file has already been resolved. Every type it names is
fully qualified, every supertype is named outright, every annotation has had its
import worked out — so there are no wildcards, no implicit `java.lang`, and
inheritance does not stop at the project boundary. It also collapses three
targets into one: Kotlin compiles to JVM bytecode and is only dexed afterwards,
so Java on Spring, Kotlin on Spring and Kotlin on Android arrive in the same
shape.

**The carrier is a class.** A Java annotation may hold only compile time
constants, so a requirement is a class and one requirement names another by
class literal:

```java
@Requirement(id = "R-QUOTE-SUBMIT", kind = FUNCTIONAL, status = NORMATIVE,
             text = "…", derivedFrom = RDecNumbering.class,
             sources = @Source(doc = "…", anchor = "8-abgabe"))
public final class RQuoteSubmit {}
```
```java
@Satisfies(RQuoteSubmit.class)
public interface SubmitQuote { … }
```

With a string there the reference would be unverified, which is the one thing
this design will not accept.

**There is no library to depend on.** A project declares these annotations
itself — some forty lines — and speclink recognises them by fully qualified
name, the same way it recognises a framework without linking it. Nothing to
publish, nothing to pin.

**It infers, too.** Spring declares its architecture in annotations, and
annotations are exactly what a declaration level reader sees:

| written | recognised | needs a requirement |
|---|---|---|
| method of a `@RestController` with a mapping | endpoint | yes |
| public method of a `@Service` | service operation | yes |
| `@Entity` | entity | no, but its fields do |
| `@Repository`, or extending Spring Data | repository | no |

**Verification costs no runtime.** A test says what it demonstrates and the
build's own report says whether it passed:

```java
@Test @Verifies(RQuoteSubmit.class)
void submitDrawsANumber() { … }
```
```
mvn test
speclink evidence -root .
```

The claim is read from the bytecode, the result from Surefire's XML, and the two
are joined on the name a report gives a test. Nothing runs, nothing is
installed. The Go form has to write a line from inside the test to prove control
reached it; here the annotation is on the method, and a method that passed ran
to its end.

```
speclink verify -root .

2 source segments (100% accounted), 5 constructs (100% bound),
3 normative requirements (100% covered, 100% verified, 100% demonstrated),
5 bindings, 0 findings
not measured: schema evolution, because this frontend reads no persisted shapes
not measured: architecture, because profile java_springboot_ddd1 prescribes no rules yet
```

**What it cannot do it says**, rather than answering anyway. That line is not
decoration: a direction that was never measured must not read as one that came
out clean, so the rules for it are not run at all — over an empty set they would
report every recorded type as removed, which is a different claim from "nobody
looked". The figure is absent rather than zero for the same reason.

---

## 12. Not implemented yet

Do not invoke or assume these; they do not exist:

- `speclink selfreport` and `verify --check-generated`
- any documentation backend other than the Markdown one: no HTML, no AsciiDoc,
  no JSON-LD
- the evolution rules for storage other than the JSON repository and the event
  log. A type written through some other store is not part of the promised set.
- any rule that checks a projection is not persisted, or that a repository is
  not reached from `ui*` beyond the existing import ban

### Known blockers

These are measured, not suspected. They are recorded here because forgetting
them would mean measuring them again.

Generated code used to head this list: 226 of the reference ERP's 644
requirement-bearing constructs live in `*_gen.go` files carrying `DO NOT EDIT`,
and both K1 and K14 ask for something that cannot be written into one. It is
not a blocker, because those files come from the specification generator that
speclink and hand written code replace. The constructs stop being generated
rather than learning to carry annotations — which is the point of the tool
rather than an obstacle to it.

- **Constructor naming.** The same project names use case constructors with a
  `UC` suffix rather than `New<Type>`: 193 are a naming convention, 45 a file
  convention, 9 are genuine duplicates. K5-UC-CONSTRUCTOR reports all of them
  alike.
- **Event identity.** Whether the stable identifier of an event is its
  discriminator or a versioned type field is open, and it has to be settled
  before the first context is frozen rather than after.

---

## 13. Worked example

A complete, verifying slice lives in `testdata/example`. It is the reference to
imitate and it is guaranteed to be correct, because the test suite requires it
to verify with zero findings:

```
testdata/example/
  cmd/erp/main.go
  app/sales/
    model.go                            aggregate, command, event
    model.annotation.go
    perm.go                             one permission per use case
    usecases.go                         the bundle
    uc_submit_quote.go / .annotation.go        write use case
    uc_approve_quote.go / .annotation.go       write use case
    uc_find_quote_overview.go / .annotation.go query, AuditResource
    quoteoverview.go / .annotation.go   the projection
    repository.go                       the repository
  requirements/
    dec/R-DEC-NUMBERING.spec.go
    fun/quote/R-QUOTE-SUBMIT.spec.go
    fun/quote/R-QUOTE-APPROVE.spec.go
    fun/quote/R-QUOTE-OVERVIEW.spec.go
    _sources/sales/quoteflow.md
```

`testdata/bad` and `testdata/arch` are the negative fixtures: each violates one
rule on purpose. Read them when a diagnostic is unclear.

---

This README is the contract. There is no second document.
