# Authoring contract for section fragments

You are writing ONE fragment of a Typst paper. Obey exactly.

## Files
- Your fragment goes into `doc/paper/sections/<name>.typ`.
- The fragment contains **only** content: headings, paragraphs, figures, tables,
  code blocks and citations. **No** `#set`, `#show`, `#import`, no page setup,
  no `#bibliography`. The parent document does all of that.

## Headings
- Your top-level heading is `= Title` (level 1). Sub-headings are `==` and `===`.
- Do not number headings manually; numbering is automatic.

## Citations
- Cite with `@key`, e.g. `@gotel1994`, or `#cite(<key>, form: "prose")` for
  prose-form ("Gotel and Finkelstein [1] show ...", use `@key[]`-free style:
  Typst syntax is `#cite(<gotel1994>, form: "prose")`).
- **Only keys that exist in `doc/paper/refs.bib` may be used.** Read that file.
- **`doc/paper/CITATION-GUIDE.md` is binding.** It states, per key, what the
  source may and may not support. Never exceed it.
- **`doc/paper/research/AUDIT.md` overrides everything else.** Anything in its
  "MUST NOT BE CITED" list is forbidden. Anything in "CITE ONLY AS ATTRIBUTED
  CLAIM" must be hedged and attributed ("X's documentation states that ...").

## The no-hallucination rule (hard)
- You may state a fact about **speclink** only if you found it in
  `README.md` or in the source under `internal/`, `cmd/`, `spec/`. Point to it.
- You may state a fact about the **outside world** only if a permitted bib key
  supports it. Otherwise do not write the sentence.
- Never invent numbers, clause numbers, version numbers, percentages, dates,
  study results or tool capabilities.
- Never write "studies show", "it is well known", "typically" as a substitute
  for a citation.
- Absence of evidence is not evidence of absence: never claim a competing tool
  *cannot* do something unless a permitted source says so.

## Style
- English, present tense, third person. No first-person plural marketing.
  "speclink does X" — lower-case `speclink` always, even at sentence start.
- Precise and sober. Prefer a shorter true sentence to a longer impressive one.
- Code blocks: use ```go / ```json / ```text fenced blocks. Keep them under
  ~14 lines and under 44 columns wide (the paper is two-column).
- Tables: `#figure(table(columns: n, ...), caption: [...])` with a `<tbl:name>`
  label. Keep tables narrow; wide tables must use
  `#figure(placement: top, scope: "parent", ...)` to span both columns.
- Figures/tables get labels `<fig:...>` / `<tbl:...>` and are referenced with
  `@fig:...`.
- Target length is given in your individual prompt. Respect it within ±20%.

## Terminology (use consistently)
- **requirement** — a `spec.Requirement` value in a `<ID>.spec.go` file.
- **construct** — a recognised program element (use case, command, event,
  aggregate, projection, repository, permission, query).
- **annotation** — a `var _ = spec.X(...)` term in a `<base>.annotation.go` file.
- **binding** — `spec.For`, `spec.ForDecl`, `spec.ForField`, `spec.ForPackage`.
- **assertion** — a pure term passed to a binding, e.g. `spec.Satisfies`.
- **finding** — a diagnostic; speclink has no warnings, every finding is an error.
- **source document** — a Markdown file or mockup under `requirements/_sources/`.
- **segment** — an addressable part of a source document (heading slug or image
  region).
- **evidence** — a recorded passing test run that claimed a requirement.
- Say "compiler-checked", not "compile-time verified"; speclink is not a verifier
  in the formal-methods sense and must never be described as one.
