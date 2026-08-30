#par(justify: true)[
  Machine-written code changes what a review can be: once new code exceeds what
  the responsible engineers can read, review by sampling becomes the
  working model, and the trace from a requirement to the code written for it
  becomes the load-bearing artefact. That trace is conventionally maintained
  outside the compiler, as links between identifiers that no build ever resolves.
  Such a link is a string, and strings rot; links recovered after the fact are
  heuristic and, by the recovery literature's own account, incomplete. speclink
  takes the opposite route. It annotates only the one fact no analysis can infer —
  which requirement a construct exists for — in a closed whitelist grammar that is
  part of the ordinary build, so the host language's compiler proves that
  every reference resolves. Requirements are anchored to
  addressable segments of the human-written source documents; drift in those
  segments and in the requirement wordings is caught by fingerprints; the
  chain is joined to recorded passing test runs rather than to markers in the
  source; and authorship and review are recorded from outside the code. The
  paper reports a design and an implementation, a structural comparison against
  the publicly verifiable parts of ISO 26262, IEC 61508, DO-178C and Automotive
  SPICE, and a delimitation from generator-side spec-driven tooling. No controlled
  evaluation, no conformance assessment and no behavioural verification is
  claimed.
]
