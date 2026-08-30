= Implementation and Status <sec:implementation>

speclink is a single Go binary with no runtime dependencies. PlantUML is a
prerequisite of the environment rather than a dependency of the program:
`speclink diagrams` writes `.puml` sources and runs nothing, so a checkout
with no Java installed still executes every rule and every test.

== Internal decomposition

The tree separates *reading a project*, *representing it*, *judging it* and
*writing it out*.

Language front ends sit under `internal/lang`, behind a common interface.
`internal/lang/golang` reads Go source through the type checker and holds the
recognisers for both Go profiles: `kinds.go`, `infer.go` and `framework.go`
decide what counts as a construct, `nago.go` and `bare.go` carry the two
frameworks, `schema.go` reads persisted shapes, `endpoint.go` and
`endpoint_hapi.go` read mounted routes, `trace.go` follows a route to the use
case behind it, and `arch_context.go`, `arch_layout.go`, `arch_usecase.go`,
`layering.go` and `rules.go` implement the style rules (the K4 to K8 families,
which are language- and framework-specific and therefore live here rather than
in `internal/check`). `internal/lang/jvm` reads a JVM project from compiled
classes instead: `classfile/` is a hand-written class-file and constant-pool
reader, `spring.go` recognises Spring's annotations, and `verified.go` joins
test claims to Surefire's report.

`internal/ir` is the language-neutral intermediate representation both front
ends produce: `construct.go`, `requirement.go`, `process.go`, `topology.go`,
`endpoint.go`, `schema.go`, `entrypoint.go`, `packages.go`, `architecture.go`,
`waivers.go`, and `dialect.go`, which supplies the wording a finding needs so
that the rules can phrase a fix without knowing the language they are looking
at.

`internal/check` holds the language-independent rules, one file per family:
`structure.go` (`K1-CONSTRUCT-UNBOUND`), `fields.go` (`K1-FIELD-UNBOUND`),
`persistence.go` (`K1-PERSISTENCE-UNJUSTIFIED`), `coverage.go` (the K3
requirement-coverage rules), `draft.go` (`K9-DRAFT-REDUNDANT`),
`evolution.go` (the ten K9 schema-evolution rules against the baseline),
`drift.go` (`K10-REQ-CHANGED`, `K13-SOURCE-DRIFT`), `sources.go`
(`K12-SOURCE-UNCOVERED`), `verification.go` (the K14 evidence rules),
`lifecycle.go` (K15), `process.go` (the fourteen K16 process rules),
`topology.go` and `contract.go` (K17), `provenance.go`
(`K18-REVIEW-STALE`) and `endpoint.go` (the nine K20 endpoint and wire rules).

The remainder is supporting machinery: `internal/source` loads Markdown
documents and mockups and segments them into addressable parts
(`markdown.go`, `image.go`, `slug.go`, `standard.go`); `internal/reqtree`
orders and lays out the requirement tree and checks themes (`topic.go`);
`internal/baseline` reads and writes `speclink.lock`, the record of frozen
shapes, recorded test runs, reviews and promised addresses;
`internal/doc` renders the specification document, with `markdown.go` and
`typst.go` as the two backends; `internal/render` emits PlantUML
(`c4.go`, `packages.go`, `puml.go`); `internal/diag` is the finding type and
its set; `internal/config` reads `speclink.json`; and `internal/profile`
registers profiles and holds the `init` templates.

== Command surface

The commands are `init`, `diagrams`, `verify`, `requirements`, `freeze`,
`inventory`, `impact`, `attest`, `evidence` and `generate`. `verify` exits
with `1` if there is any finding and `0` otherwise; there are no warnings and
no severities, and `scope` is the only dial. The build order is fixed:

```text
go build ./...
speclink verify ./...
go test -json ./... | speclink evidence
```

`speclink inventory` reports what was recognised, which is the question
`verify` cannot answer because a correctly recognised construct produces no
output. `speclink impact` reports the reach of a change from a requirement, a
document anchor or a path. `speclink attest` records authorship and review
from outside the code. `speclink freeze` writes the baseline.

== Profiles and configuration

A project names one profile in `speclink.json`: `go_nago_ddd1`,
`go_bare_ddd1` or `java_springboot_ddd1`. The name has three parts because
there are three decisions — language, framework, style — and speclink refuses
to guess the third. `speclink.json` states deviations from the profile and
never the conventions themselves; every profile understands `sourceRoots`,
`scope` and `exclude`, and profile-specific keys beyond those are refused
rather than ignored. `speclink init -describe` lists profiles, templates and
their parameters, in text or JSON, because the usual caller is an agent for
which an interactive prompt is the worst possible interface.

Two styles share one language and one reader. `go_bare_ddd1` adds four K6
rules that `go_nago_ddd1` has no use for, and two terms — `spec.Persistence()`
and `spec.StoredAs[Domain]()` — which are findings under `go_nago_ddd1`,
where the repository constructor already says the same thing.

== Generated artefacts

`speclink generate` writes the specification document as Markdown or Typst;
`internal/doc/typst.go` implements the latter. This paper is typeset with
Typst, the same output format the tool emits.

== Measured facts

#figure(
  table(
    columns: 2,
    align: (left, right),
    [Go lines, tool source], [24,433],
    [Go lines, tests (`*_test.go`)], [10,234],
    [Go lines, `testdata/` fixtures], [3,222],
    [`func Test` functions], [349],
    [Go packages (`go list ./...`)], [16],
    [Distinct rule IDs in README §11], [77],
    [Git commits], [98],
    [First commit], [2026-08-10],
    [Last commit], [2026-08-30],
    [`go build ./...`], [passes],
    [`go test ./...`], [passes],
  ),
  caption: [Measured on a clean checkout of commit `030ea34`, dated
  2026-08-30. Rule IDs are the distinct `K…` identifiers in the README rule
  index; three further identifiers exist in the code but not in that index
  (see below).],
) <tbl:measured>

Both `go build ./...` and `go test ./...` pass on a clean checkout of
`030ea34`.

The history is three weeks long, which is the plainest caveat available about
the artefact's maturity: the rules, the intermediate representation and the
lock-file format have not yet been exposed to the kind of long-running
project whose evolution they exist to constrain.

== Discrepancies between the documentation and the code

Four are worth recording, because they bear on claims the tool makes about
itself.

*The Typst backend is undocumented as implemented.* README §12 lists "any
documentation backend other than the Markdown one" among the things that do
not exist, but `internal/doc/typst.go` implements one, README §2 documents
`-format markdown|typst`, and `cmd/speclink/typst_test.go` exercises it.
README §12 is stale on this point.

*Three rule identifiers exist in the code but not in the README index.*
`K1-FIELD-UNBOUND` (`internal/check/fields.go`),
`K1-PERSISTENCE-UNJUSTIFIED` (`internal/check/persistence.go`) and
`K6-CTX-PRESENTATION-NO-IMPORT` (`internal/lang/golang/describe.go`) can all
be emitted, and the index a project would consult in order to waive one does
not list them.

*Six `V6` numeric codes are assigned twice.* `SPEC-V6-090` through
`SPEC-V6-095` each name two different rules in the index. README §11 opens by stating
that every rule ID is stable and may be used in `spec.Waive`; the `K…` names
are unique, but the numeric codes that accompany them are not, and a reader
filtering a JSON diagnostic stream on `code` would conflate two families.

*The rule families are not all in `internal/check`.* K4 to K8 are implemented
in `internal/lang/golang` (`rules.go`, `layering.go`, `arch_context.go`,
`arch_layout.go`, `arch_usecase.go`, `nago.go`) because they are style rules
over a specific framework, and the two requirement-tree families K11 and K19 are
in `internal/reqtree/topic.go`. The
separation between language-independent and style-specific rules is real, but
it is a property of the tree rather than of any stated interface.

== What is not implemented

The following do not exist and must not be assumed:

- `speclink selfreport` and `verify --check-generated`;
- documentation backends other than Markdown and Typst — no HTML, no
  AsciiDoc, no JSON-LD;
- the K9 evolution rules for storage other than the JSON repository and the
  event log; a type written through some other store is not part of the
  promised set;
- any rule checking that a projection is not persisted, or that a repository
  is not reached from `ui*` beyond the existing import ban.

`java_springboot_ddd1` prescribes no style rules and has no persistence
recogniser; a run under it says so rather than reporting those families sound.

Two blockers are recorded as measured rather than suspected. *Constructor
naming*: in the reference ERP, 193 use-case constructors are a naming
convention, 45 a file convention and 9 genuine duplicates, and
`K5-UC-CONSTRUCTOR` reports all of them alike. *Event identity*: whether the
stable identifier of an event is its discriminator or a versioned type field
is unsettled, and must be decided before the first context is frozen.

== Availability

The repository is at `github.com/worldiety/speclink`. The `LICENSE` file is
not an open-source licence: it states that use requires a licence agreement
with worldiety GmbH.
