// A lightweight ACM-like (sigconf) layout for Typst.
// This is an independent re-implementation of the visual conventions of the ACM
// "acmart" sigconf template. It is not an official ACM template and must not be
// represented as one.

#let acm-paper(
  title: none,
  subtitle: none,
  authors: (),
  abstract: none,
  ccs: none,
  keywords: none,
  conference: none,
  doi: none,
  body,
) = {
  set document(title: title)

  set page(
    paper: "us-letter",
    margin: (top: 57pt, bottom: 73pt, left: 46pt, right: 46pt),
    footer: context [
      #set text(size: 8pt)
      #h(1fr) #counter(page).display("1") #h(1fr)
    ],
  )

  set text(font: ("Libertinus Serif", "New Computer Modern"), size: 9pt, lang: "en")
  set par(justify: true, leading: 0.55em, spacing: 0.55em, first-line-indent: 10pt)

  set heading(numbering: "1.1")
  show heading: it => {
    set text(font: ("Helvetica Neue", "Helvetica"))
    if it.level == 1 {
      v(10pt, weak: true)
      block(text(size: 10pt, weight: "bold", upper(it)))
      v(3pt, weak: true)
    } else if it.level == 2 {
      v(8pt, weak: true)
      block(text(size: 9.5pt, weight: "bold", it))
      v(2pt, weak: true)
    } else {
      v(6pt, weak: true)
      block(text(size: 9pt, weight: "bold", style: "italic", it))
      v(1pt, weak: true)
    }
  }

  show figure.caption: set text(size: 8pt)
  show figure: set block(breakable: false)
  set figure(placement: none)

  set table(stroke: (x, y) => (
    top: if y <= 1 { 0.6pt } else { 0pt },
    bottom: 0.6pt,
  ), inset: 4pt)
  show table.cell.where(y: 0): set text(weight: "bold", size: 8pt)
  show table: set text(size: 8pt)

  show raw.where(block: true): it => block(
    fill: rgb("#f6f6f6"),
    inset: 5pt,
    radius: 2pt,
    width: 100%,
    breakable: false,
    text(size: 7.4pt, font: ("DejaVu Sans Mono", "Menlo", "Courier New"), it),
  )
  show raw.where(block: false): it => text(size: 8.2pt, font: ("DejaVu Sans Mono", "Menlo", "Courier New"), it)

  set cite(style: "association-for-computing-machinery")
  set enum(indent: 6pt, body-indent: 5pt)
  set list(indent: 6pt, body-indent: 5pt, marker: [•])

  // ---- title block, single column ----
  place(top, float: true, scope: "parent", clearance: 12pt)[
    #set par(first-line-indent: 0pt, justify: false)
    #align(center)[
      #text(font: ("Helvetica Neue", "Helvetica"), size: 17pt, weight: "bold", title)
      #if subtitle != none [
        #linebreak()
        #v(2pt)
        #text(font: ("Helvetica Neue", "Helvetica"), size: 12pt, subtitle)
      ]
      #v(10pt)
      #grid(
        columns: authors.len(),
        column-gutter: 20pt,
        ..authors.map(a => align(center)[
          #text(size: 10pt, a.name) \
          #text(size: 8.5pt, style: "italic", a.affiliation) \
          #text(size: 8.5pt, a.location) \
          #text(size: 8.5pt, font: ("DejaVu Sans Mono", "Menlo", "Courier New"), a.email)
        ])
      )
      #v(8pt)
      #if conference != none [ #text(size: 8pt, style: "italic", conference) #v(4pt) ]
    ]

    #if abstract != none [
      #v(2pt)
      #block(width: 100%)[
        #text(font: ("Helvetica Neue", "Helvetica"), size: 10pt, weight: "bold", "ABSTRACT")
        #v(3pt)
        #set par(justify: true, first-line-indent: 0pt)
        #text(size: 9pt, abstract)
      ]
    ]
    #if ccs != none [
      #v(5pt)
      #block(width: 100%)[
        #text(font: ("Helvetica Neue", "Helvetica"), size: 10pt, weight: "bold", "CCS CONCEPTS")
        #v(3pt)
        #set par(justify: true, first-line-indent: 0pt)
        #text(size: 9pt, ccs)
      ]
    ]
    #if keywords != none [
      #v(5pt)
      #block(width: 100%)[
        #text(font: ("Helvetica Neue", "Helvetica"), size: 10pt, weight: "bold", "KEYWORDS")
        #v(3pt)
        #set par(justify: true, first-line-indent: 0pt)
        #text(size: 9pt, keywords)
      ]
    ]
    #v(6pt)
    #line(length: 100%, stroke: 0.4pt)
  ]

  show: rest => columns(2, gutter: 16pt, rest)

  body
}

// A boxed, full-width note used for the AI disclosure.
#let notebox(title, body) = block(
  width: 100%,
  stroke: 0.6pt,
  inset: 7pt,
  radius: 2pt,
  breakable: true,
)[
  #set par(first-line-indent: 0pt)
  #text(weight: "bold", size: 9pt, title)
  #v(3pt)
  #body
]
