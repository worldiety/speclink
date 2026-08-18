# Test fixtures

These are not test doubles. Each fixture is a miniature target project: its own
Go module, requiring the real framework, built by the real toolchain.

That is not a detail. speclink analyses a project by running `go list` and the
type checker **in that project's module** (`Load` sets `cfg.Dir`), so what it
sees is the project's `go.mod`, the project's framework version and the
project's build tags. A fixture that linked a hand-written stand-in would
therefore not be a smaller version of the thing under test — it would be a
different thing.

There used to be such stand-ins here, `testdata/nago` and `testdata/i18n`,
written by hand and kept in sync by hand. They hid three real defects at once:

- `nagoUIEnt` pointed at `presentation/ui/ent`, a path the framework had
  abandoned. The rule matched nothing and passed for exactly that reason.
- `evs.SeqID` had become an alias of `ndb.Seq`. The exemption that keeps a
  commit sequence from counting as data stopped matching, so every writing use
  case was silently classified as a query.
- `user.Subject` was missing `AuditResource` and `HasResourcePermission`, so the
  two authorisation checks that accept them could not be reached by any test.

Every one of them was invisible while the stand-in mirrored the mistake. All
three surfaced within minutes of building against the real module.

## The modules

| Module | Path | Role |
|---|---|---|
| `example` | `example.com/erp` | The conformant project. **Must verify with zero findings.** It is the guard against rules that nothing can satisfy. |
| `bad` | `example.com/bad` | Violates the annotation language rules, one each. |
| `arch` | `example.com/arch` | Violates the architecture linters, one each. |

`example` and `bad` carry `replace github.com/worldiety/speclink => ../..`. That
is a genuine local dependency and not a substitute: a target project imports the
`spec` package, and `spec` is where the runtime registry lives.

speclink's own `go.mod` does **not** require the framework, and must not. One
binary serves projects on different framework versions, so pinning one of them
would force it onto every analysis. The recognisers therefore couple by import
path string, which is the only version neutral coupling available.

## Upgrading the framework

A version bump is a deliberate act, and breakage is the point of it.

1. Raise the `go.wdy.de/nago` version in all three `go.mod` files.
2. `go build ./...` in each fixture. A compile error here is a real breaking
   change, stated in the terms a target project would meet it in.
3. `go test ./...` at the repository root.
4. Fix the recognisers, not the fixture, unless the fixture was genuinely wrong.

`TestNagoPathsResolve` covers the failure mode a compiler cannot: a package that
moved. A recogniser matching a path that no longer exists does not fail, it
stops matching — so that test resolves every framework path constant and fails
loudly when one goes stale.

## Cost

The suite needs the module cache or network; it is no longer hermetic. The
framework build is roughly 20 seconds cold and under two seconds warm, and is
paid once per build cache.
