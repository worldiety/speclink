#import "acmart.typ": acm-paper, notebox

#show: acm-paper.with(
  title: [Annotate Only What Cannot Be Inferred],
  subtitle: [A Compiler-Checked Chain from Source Document to Test Evidence,
             and What It Is Worth When a Machine Writes the Code],
  authors: (
    (
      name: [Torben Schinke],
      affiliation: [worldiety GmbH],
      location: [Oldenburg, Germany],
      email: "torben.schinke@worldiety.de",
    ),
  ),
  abstract: include "sections/00-abstract.typ",
  ccs: [
    Software and its engineering → Software verification and validation;
    Requirements analysis; Automated static analysis; Documentation.
  ],
  keywords: [
    requirements traceability, static analysis, annotation compiler,
    functional safety, Automotive SPICE, AI-assisted software engineering,
    evidence, Go
  ],
  conference: [Preprint. Not submitted to, reviewed by, or published at any
               venue. Typeset in an independent, ACM-like layout; this is not
               an official ACM template and this work is not an ACM
               publication.],
)

#include "sections/01-introduction.typ"
#include "sections/02-background.typ"
#include "sections/03-design.typ"
#include "sections/04-evidence.typ"
#include "sections/05-standards.typ"
#include "sections/06-delimitation.typ"
#include "sections/07-implementation.typ"
#include "sections/08-discussion.typ"
#include "sections/09-conclusion.typ"
#include "sections/10-ai-disclosure.typ"

= Availability

The tool, the worked example, the research notes gathered for this paper, the
independent source audit and the sources of this document are in the repository
at #link("https://github.com/worldiety/speclink")[`github.com/worldiety/speclink`].
The paper's sources are under `doc/paper/`.

#bibliography("refs.bib", title: [References], style: "association-for-computing-machinery")
