# Verified Literature: Requirements Traceability and Related Topics

All entries below were verified by fetching a URL during this research session
(dblp publication API, Crossref API, OpenAlex API, Semantic Scholar Graph API,
arXiv API, or the DOI resolver). Verbatim proof snippets are quoted from the
fetched responses. Nothing in the "verified" sections is written from memory.

Verification endpoints used:

- `https://dblp.org/search/publ/api?q=<terms>&format=json` (bibliographic metadata)
- `https://api.crossref.org/works/<doi>` (metadata, sometimes abstract)
- `https://api.openalex.org/works/doi:<doi>` (metadata + reconstructed abstract)
- `https://api.semanticscholar.org/graph/v1/paper/DOI:<doi>` (abstract)
- `https://export.arxiv.org/api/query?id_list=<id>` (preprints)
- `https://doi.org/<doi>` (DOI resolution / HTTP status check)

---

## 1. Requirements Traceability Foundations

### 1.1 Gotel & Finkelstein — the traceability problem

- **Authors:** O. C. Z. Gotel; Anthony Finkelstein
- **Exact title:** An analysis of the requirements traceability problem
- **Venue:** Proceedings of IEEE International Conference on Requirements
  Engineering (ICRE 1994)
- **Year:** 1994
- **Pages:** 94–101
- **DOI:** 10.1109/ICRE.1994.292398
- **Verification URLs fetched:**
  - `https://dblp.org/search?q=Gotel+Finkelstein+traceability+problem`
  - `https://api.semanticscholar.org/graph/v1/paper/DOI:10.1109/ICRE.1994.292398?fields=title,year,venue,abstract,externalIds,authors`

**Verbatim proof snippet (dblp):**

> "O. C. Z. Gotel, Anthony Finkelstein:
> An analysis of the requirements traceability problem. ICRE 1994: 94-101"

**Verbatim proof snippet (Semantic Scholar):**

> T: An analysis of the requirements traceability problem
> Y: 1994 V: Proceedings of IEEE International Conference on Requirements Engineering
> A: ['O. Gotel', 'A. Finkelstein']

**Verbatim abstract snippet (Semantic Scholar):**

> "Investigates and discusses the underlying nature of the requirements
> traceability problem. Our work is based on empirical studies, involving over
> 100 practitioners, and an evaluation of current support. We introduce the
> distinction between pre-requirements specification (pre-RS) traceability and
> post-requirements specification (post-RS) traceability to demonstrate why an
> all-encompassing solution to the problem is unlikely, and to provide a
> framework through which to understand its multifaceted nature. We report how
> the majority of the problems attributed to poor requirements traceability are
> due to inadequate pre-RS traceability and show the fundamental need for
> improvements."

**What the work claims (from the fetched abstract only):** Based on empirical
study of over 100 practitioners, the paper argues the traceability problem is
multifaceted and that an all-encompassing solution is unlikely. It introduces
the pre-RS / post-RS traceability distinction and attributes the majority of
reported traceability problems to inadequate pre-RS traceability.

**Relevance:** Peer-reviewed, canonical framing of *why* traceability is hard;
useful for motivating a mechanised alternative.

---

### 1.2 Ramesh & Jarke — reference models for requirements traceability

- **Authors:** Balasubramaniam Ramesh; Matthias Jarke
- **Exact title:** Toward reference models for requirements traceability
  (dblp renders the title with title-case: "Toward Reference Models of
  Requirements Traceability"; OpenAlex/IEEE render "Toward reference models
  **for** requirements traceability" — see note below)
- **Venue:** IEEE Transactions on Software Engineering
- **Year:** 2001
- **Volume/Issue/Pages:** 27(1), 58–93
- **DOI:** 10.1109/32.895989
- **Verification URLs fetched:**
  - `https://dblp.org/search/publ/api?q=Ramesh+Jarke+reference+models+requirements+traceability&format=json`
  - `https://api.openalex.org/works/doi:10.1109/32.895989`

**Verbatim proof snippet (dblp API):**

> - Toward Reference Models of Requirements Traceability.
>    authors: Balasubramaniam Ramesh; Matthias Jarke
>    venue: IEEE Trans. Software Eng. | year: 2001 | type: Journal Articles | pages: 58-93
>    doi: 10.1109/32.895989

**Verbatim proof snippet (OpenAlex):**

> T: Toward reference models for requirements traceability | Y: 2001
> SRC: IEEE Transactions on Software Engineering | biblio: {'volume': '27', 'issue': '1', 'first_page': '58', 'last_page': '93'}

**Verbatim abstract snippet (OpenAlex):**

> "Requirements traceability is intended to ensure continued alignment between
> stakeholder requirements and various outputs of the system development
> process. To be useful, traces must be organized according to some modeling
> framework. Indeed, several such frameworks have been proposed, mostly based on
> theoretical considerations or analysis of other literature. This paper, in
> contrast, follows an empirical approach. Focus groups and interviews conducted
> in 26 major software development organizations demonstrate a wide range of
> traceability practices with distinct low-end and high-end users of
> traceability."

**What the work claims (from the fetched abstract only):** Empirically derived
(focus groups and interviews in 26 organisations) reference models of
traceability link types, distinguishing low-end from high-end traceability
users. Four kinds of traceability link types are identified, and the models were
validated in case studies and incorporated into traceability tools.

> **Note on title variance:** dblp records the title as "Toward Reference Models
> **of** Requirements Traceability", OpenAlex as "Toward reference models **for**
> requirements traceability". Both were fetched this session. Confirm against the
> IEEE Xplore PDF before final submission.

---

## 2. The Grand Challenge of Traceability / CoEST Body of Work

### 2.1 Gotel et al. — traceability research roadmap (RE 2012)

- **Authors:** Orlena Gotel; Jane Cleland-Huang; Jane Huffman Hayes; Andrea
  Zisman; Alexander Egyed; Paul Grünbacher; Giuliano Antoniol
- **Exact title:** The quest for Ubiquity: A roadmap for software and systems
  traceability research
- **Venue:** RE (IEEE International Requirements Engineering Conference) 2012
- **Year:** 2012
- **Pages:** 71–80
- **DOI:** 10.1109/RE.2012.6345841
- **Verification URLs fetched:**
  - `https://dblp.org/search/publ/api?q=Cleland-Huang+Gotel+Zisman+Software+and+Systems+Traceability&format=json`
  - `https://api.openalex.org/works/doi:10.1109/RE.2012.6345841`

**Verbatim proof snippet (dblp API):**

> - The quest for Ubiquity: A roadmap for software and systems traceability research.
>    authors: Orlena Gotel; Jane Cleland-Huang; Jane Huffman Hayes; Andrea Zisman; Alexander Egyed; Paul Grünbacher; Giuliano Antoniol
>    venue: RE | year: 2012 | type: Conference and Workshop Papers | pages: 71-80
>    doi: 10.1109/RE.2012.6345841

**Verbatim abstract snippet (OpenAlex):**

> "Traceability underlies many important software and systems engineering
> activities, such as change impact analysis and regression testing. Despite
> important research advances, as in the automated creation and maintenance of
> trace links, traceability implementation and use is still not pervasive in
> industry. A community of traceability researchers and practitioners has been
> collaborating to understand the hurdles to making traceability ubiquitous. ...
> We present a brief view of the state of the art in traceability, the grand
> challenge for traceability and future directions for the field."

**What the work claims (from the fetched abstract only):** Despite research
advances in *automated* trace link creation and maintenance, traceability
implementation and use is "still not pervasive in industry". The paper presents
a community-derived research roadmap and states the grand challenge for
traceability.

**Relevance:** Directly citable evidence that automated recovery has not made
traceability ubiquitous in practice — the motivating gap for a compiler-checked
approach.

---

### 2.2 Gotel et al. — The Grand Challenge of Traceability (v1.0)

- **Authors:** Orlena Gotel; Jane Cleland-Huang; Jane Huffman Hayes; Andrea
  Zisman; Alexander Egyed; Paul Grünbacher; Alex Dekhtyar; Giuliano Antoniol;
  Jonathan I. Maletic
- **Exact title:** The Grand Challenge of Traceability (v1.0)
- **Venue:** Book chapter in *Software and Systems Traceability* (Springer)
- **Year:** 2012 (dblp); OpenAlex records publication year 2011 — see note
- **Pages:** 343–409
- **DOI:** 10.1007/978-1-4471-2239-5_16
- **Verification URLs fetched:**
  - `https://dblp.org/search/publ/api?q=Cleland-Huang+Gotel+Zisman+Software+and+Systems+Traceability&format=json`
  - `https://api.openalex.org/works/doi:10.1007/978-1-4471-2239-5_16`

**Verbatim proof snippet (dblp API):**

> - The Grand Challenge of Traceability (v1.0).
>    authors: Orlena Gotel; Jane Cleland-Huang; Jane Huffman Hayes; Andrea Zisman; Alexander Egyed; Paul Grünbacher; Alex Dekhtyar; Giuliano Antoniol; Jonathan I. Maletic
>    venue: Software and Systems Traceability | year: 2012 | type: Parts in Books or Collections | pages: 343-409
>    doi: 10.1007/978-1-4471-2239-5_16

**Verbatim proof snippet (OpenAlex):**

> T: The Grand Challenge of Traceability (v1.0) | Y: 2011 | type: book-chapter
> biblio: {'first_page': '343', 'last_page': '409'}

**What the work claims:** **No abstract was retrievable** from OpenAlex
(`ABS: (none)`) or Crossref this session. Metadata (authors, title, pages, DOI)
is verified; **do not paraphrase its content** without reading the chapter.

> **Note:** dblp says year 2012, OpenAlex says 2011. The parent book is dated
> 2012 (see 2.4). Use 2012 and confirm on the Springer page.

---

### 2.3 Gotel et al. — Traceability Fundamentals

- **Authors:** Orlena Gotel; Jane Cleland-Huang; Jane Huffman Hayes; Andrea
  Zisman; Alexander Egyed; Paul Grünbacher; Alex Dekhtyar; Giuliano Antoniol;
  Jonathan I. Maletic; Patrick Mäder
- **Exact title:** Traceability Fundamentals
- **Venue:** Book chapter in *Software and Systems Traceability* (Springer)
- **Year:** 2012 (dblp); OpenAlex records 2011
- **Pages:** 3–22
- **DOI:** 10.1007/978-1-4471-2239-5_1
- **Verification URLs fetched:**
  - `https://dblp.org/search/publ/api?q=Cleland-Huang+Gotel+Zisman+Software+and+Systems+Traceability&format=json`
  - `https://api.openalex.org/works/doi:10.1007/978-1-4471-2239-5_1`

**Verbatim proof snippet (dblp API):**

> - Traceability Fundamentals.
>    authors: Orlena Gotel; Jane Cleland-Huang; Jane Huffman Hayes; Andrea Zisman; Alexander Egyed; Paul Grünbacher; Alex Dekhtyar; Giuliano Antoniol; Jonathan I. Maletic; Patrick Mäder
>    venue: Software and Systems Traceability | year: 2012 | type: Parts in Books or Collections | pages: 3-22
>    doi: 10.1007/978-1-4471-2239-5_1

**What the work claims:** **No abstract retrievable** (`ABS: (none)` from
OpenAlex). This is the chapter that defines the community's traceability
terminology, but that characterisation is *not* verified from a fetched
abstract — read the chapter before citing it for a definition.

---

### 2.4 Cleland-Huang, Gotel & Zisman (eds.) — Software and Systems Traceability

- **Editors:** Jane Cleland-Huang; Olly Gotel; Andrea Zisman
- **Exact title:** Software and Systems Traceability
- **Venue:** Springer (book / editorship)
- **Year:** 2012
- **DOI:** 10.1007/978-1-4471-2239-5
- **Verification URLs fetched:**
  - `https://dblp.org/search/publ/api?q=Cleland-Huang+Gotel+Zisman+Software+and+Systems+Traceability&format=json`
  - `https://api.openalex.org/works/doi:10.1007/978-1-4471-2239-5`
  - `https://doi.org/10.1007/978-1-4471-2239-5` (HTTP status check)

**Verbatim proof snippet (dblp API):**

> - Software and Systems Traceability.
>    authors: Jane Cleland-Huang; Olly Gotel; Andrea Zisman
>    venue: None | year: 2012 | type: Editorship
>    doi: 10.1007/978-1-4471-2239-5

**Verbatim proof snippet (OpenAlex):**

> T: Software and Systems Traceability | Y: 2012 | book

**Verbatim proof snippet (DOI resolver):**

> `302 https://link.springer.com/10.1007/978-1-4471-2239-5`

**What the work claims:** Edited volume; **no abstract fetched**. Cite as an
edited book only.

---

## 3. Automated Trace Link Recovery Using Information Retrieval

This section matters because it documents the *reported accuracy limits* of
heuristic recovery, in contrast to a compiler-checked approach.

### 3.1 Antoniol et al. — IR-based trace recovery (TSE 2002)

- **Authors:** Giuliano Antoniol; Gerardo Canfora; Gerardo Casazza; Andrea De
  Lucia; Ettore Merlo
- **Exact title:** Recovering traceability links between code and documentation
- **Venue:** IEEE Transactions on Software Engineering
- **Year:** 2002
- **Volume/Issue/Pages:** 28(10), 970–983
- **DOI:** 10.1109/TSE.2002.1041053
- **Verification URLs fetched:**
  - `https://dblp.org/search/publ/api?q=Antoniol+recovering+traceability+links+between+code+and+documentation&format=json`
  - `https://api.crossref.org/works/10.1109/TSE.2002.1041053`
  - `https://api.openalex.org/works/doi:10.1109/TSE.2002.1041053`

**Verbatim proof snippet (Crossref):**

> TITLE: ['Recovering traceability links between code and documentation']
> AUTH: ['G. Antoniol', 'G. Canfora', 'G. Casazza', 'A. De Lucia', 'E. Merlo']
> VENUE: ['IEEE Transactions on Software Engineering'] | VOL 28 ISSUE 10 | PAGE 970-983
> YEAR: [[2002, 10]]

**Verbatim abstract snippet (OpenAlex):**

> "We propose a method based on information retrieval to recover traceability
> links between source code and free text documents. A premise of our work is
> that programmers use meaningful names for program items, such as functions,
> variables, types, classes, and methods. ... We apply both a probabilistic and
> a vector space information retrieval model in two case studies to trace C++
> source code onto manual pages and Java code to functional requirements. We
> compare the results of applying the two models, discuss the benefits and
> **limitations**, and describe directions for improvements."

**What the work claims (from the fetched abstract only):** Proposes IR-based
(probabilistic and vector-space) recovery of links between source code and free
text, premised on programmers using meaningful identifier names. Evaluated in
two case studies; the abstract explicitly says it discusses "benefits and
limitations" and "directions for improvements" — i.e. the authors themselves
frame the accuracy as not settled.

> **Caution:** The fetched abstract states *no* specific precision/recall
> numbers. Do not attribute numeric accuracy figures to this abstract.

Related retrospective (metadata verified via dblp only, abstract not fetched):
Giulio Antoniol; Gerardo Canfora; Gerardo Casazza; Andrea De Lucia; Ettore
Merlo, "Recovering Traceability Links Between Code and Documentation: A
Retrospective", IEEE Trans. Software Eng., 2025, pp. 825–832,
DOI 10.1109/TSE.2025.3534027.

### 3.2 Hayes, Dekhtyar & Sundaram — candidate link generation (TSE 2006)

- **Authors:** Jane Huffman Hayes; Alex Dekhtyar; Senthil Karthikeyan Sundaram
- **Exact title:** Advancing candidate link generation for requirements tracing:
  the study of methods
- **Venue:** IEEE Transactions on Software Engineering
- **Year:** 2006
- **Volume/Issue/Pages:** 32(1), 4–19
- **DOI:** 10.1109/TSE.2006.3
- **Verification URLs fetched:**
  - `https://dblp.org/search/publ/api?q=Hayes+Dekhtyar+Sundaram+advancing+candidate+link+generation+requirements+tracing&format=json`
  - `https://api.crossref.org/works/10.1109/TSE.2006.3`
  - `https://api.openalex.org/works/doi:10.1109/TSE.2006.3`

**Verbatim proof snippet (Crossref):**

> TITLE: ['Advancing candidate link generation for requirements tracing: the study of methods']
> AUTH: ['J.H. Hayes', 'A. Dekhtyar', 'S.K. Sundaram']
> VENUE: ['IEEE Transactions on Software Engineering'] | VOL 32 ISSUE 1 | PAGE 4-19
> YEAR: [[2006, 1]]

**Verbatim abstract snippet (OpenAlex):**

> "This paper addresses the issues related to improving the overall quality of
> the dynamic candidate link generation for the requirements tracing process for
> verification and validation and independent verification and validation
> analysts. The contribution of the paper is four-fold: we define goals for a
> tracing tool based on analyst responsibilities in the tracing process, we
> introduce several new measures for validating that the goals have been
> satisfied, we implement analyst feedback in the tracing process, and we
> present a prototype tool that we built, RETRO (REquirements TRacing
> On-target), to address these goals."

**What the work claims (from the fetched abstract only):** Defines goals and new
measures for candidate-link-generation tracing tools, incorporates analyst
feedback into the tracing process, and presents the RETRO prototype. Note the
framing throughout is *candidate* links requiring an **analyst in the loop** —
recovery is explicitly not autonomous.

### 3.3 De Lucia et al. — LSI-based recovery (TOSEM 2007) — **key limitation quote**

- **Authors:** Andrea De Lucia; Fausto Fasano; Rocco Oliveto; Genoveffa Tortora
- **Exact title:** Recovering traceability links in software artifact management
  systems using information retrieval methods
- **Venue:** ACM Transactions on Software Engineering and Methodology (TOSEM)
- **Year:** 2007
- **Volume/Issue/Article:** 16(4), Article 13
- **DOI:** 10.1145/1276933.1276934
- **Verification URLs fetched:**
  - `https://dblp.org/search/publ/api?q=De+Lucia+recovering+traceability+links+artifact+management&format=json`
  - `https://api.crossref.org/works/10.1145/1276933.1276934`

**Verbatim proof snippet (Crossref):**

> TITLE: ['Recovering traceability links in software artifact management systems using information retrieval methods']
> AUTH: ['Andrea De Lucia', 'Fausto Fasano', 'Rocco Oliveto', 'Genoveffa Tortora']
> VENUE: ['ACM Transactions on Software Engineering and Methodology'] | VOL 16 ISSUE 4 | PAGE 13
> YEAR: [[2007, 9]]

**Verbatim abstract snippet (Crossref — full abstract available):**

> "We have improved an artifact management system with a traceability recovery
> tool based on Latent Semantic Indexing (LSI), an information retrieval
> technique. We have assessed LSI to identify strengths and limitations of using
> information retrieval techniques for traceability recovery and devised the need
> for an incremental approach. The method and the tool have been evaluated during
> the development of seventeen software projects involving about 150 students.
> We observed that although tools based on information retrieval provide a useful
> support for the identification of traceability links during software
> development, **they are still far to support a complete semi-automatic recovery
> of all links.** The results of our experience have also shown that such tools
> can help to identify quality problems in the textual description of traced
> artifacts."

**What the work claims (from the fetched abstract only):** LSI-based recovery
was evaluated across seventeen student software projects (~150 students). The
authors conclude IR tools give useful support but are "still far to support a
complete semi-automatic recovery of all links", and that they additionally
surface quality problems in artefact text.

**Relevance:** This is the single strongest verified, directly quotable
statement of the accuracy ceiling of heuristic IR-based trace recovery — the
core contrast for a compiler-checked design.

---

## 4. Traceability in Safety-Critical / Empirical Industrial Contexts

### 4.1 Mäder & Egyed — controlled experiment, does traceability help? (ICSM 2012)

- **Authors:** Patrick Mäder; Alexander Egyed
- **Exact title:** Assessing the effect of requirements traceability for software
  maintenance
- **Venue:** ICSM (IEEE International Conference on Software Maintenance) 2012
- **Year:** 2012
- **Pages:** 171–180
- **DOI:** 10.1109/ICSM.2012.6405269
- **Verification URLs fetched:**
  - `https://dblp.org/search/publ/api?q=Mader+Egyed+assessing+effect+requirements+traceability+maintenance&format=json`
  - `https://api.openalex.org/works/doi:10.1109/ICSM.2012.6405269`

**Verbatim proof snippet (dblp API):**

> - Assessing the effect of requirements traceability for software maintenance.
>    au: Patrick Mäder; Alexander Egyed
>    venue: ICSM | y: 2012 | type: Conference and Workshop Papers | pp: 171-180 | doi: 10.1109/ICSM.2012.6405269

**Verbatim abstract snippet (OpenAlex):**

> "However, despite its growing popularity, there exists no published evaluation
> about the usefulness of requirements traceability. ... We thus conducted a
> controlled experiment with 52 subjects performing real maintenance tasks on two
> third-party development projects: half of the tasks with and the other half
> without traceability. Our findings show that subjects with traceability
> performed on average **21% faster** on a task and created on average **60% more
> correct solutions** — suggesting that traceability not only saves downstream
> cost but can profoundly improve software maintenance quality."

**What the work claims (from the fetched abstract only):** A controlled
experiment with 52 subjects on real maintenance tasks in two third-party
projects. Subjects with traceability were on average 21% faster and produced 60%
more correct solutions. The paper also attempts an initial cost-benefit
estimate.

> **Important:** The 21% / 60% figures belong to this **ICSM 2012** paper's
> abstract as fetched. Do **not** transfer these numbers to the 2015 EMSE journal
> article below — its abstract could not be retrieved this session.

### 4.2 Mäder & Egyed — journal version (EMSE 2015)

- **Authors:** Patrick Mäder; Alexander Egyed
- **Exact title:** Do developers benefit from requirements traceability when
  evolving and maintaining a software system?
- **Venue:** Empirical Software Engineering
- **Year:** 2015 (dblp); Crossref/OpenAlex issue date 2014, volume 20(2)
- **Volume/Issue/Pages:** 20(2), 413–441
- **DOI:** 10.1007/s10664-014-9314-z
- **Verification URLs fetched:**
  - `https://dblp.org/search/publ/api?q=Mader+Egyed+do+developers+benefit+traceability&format=json`
  - `https://api.crossref.org/works/10.1007/S10664-014-9314-Z`
  - `https://api.semanticscholar.org/graph/v1/paper/DOI:10.1007/s10664-014-9314-z?fields=title,year,venue,abstract,authors,externalIds`
  - `https://doi.org/10.1007/s10664-014-9314-z`

**Verbatim proof snippet (Crossref):**

> TITLE: ['Do developers benefit from requirements traceability when evolving and maintaining a software system?']
> AUTH: ['Patrick Mäder', 'Alexander Egyed']
> VENUE: ['Empirical Software Engineering'] | VOL 20 ISSUE 2 | PAGE 413-441
> YEAR: [[2014, 6, 22]]

**Verbatim proof snippet (DOI resolver, formatted citation):**

> "Mäder, P., & Egyed, A. (2014). Do developers benefit from requirements
> traceability when evolving and maintaining a software system? Empirical
> Software Engineering, 20(2), 413–441. https://doi.org/10.1007/s10664-014-9314-z"

**What the work claims:** **Abstract NOT retrievable** this session (Crossref,
OpenAlex, and Semantic Scholar all returned no abstract; `link.springer.com`
returned a 3 KB bot-block page). Metadata is fully verified; **content claims
are not**. Cite for metadata, or read the PDF before making claims.

> **Note on year:** dblp says 2015, Crossref issue date says 2014-06-22, volume
> is 20(2). Verify the canonical year on the Springer page before submission.

### 4.3 Rempel & Mäder — traceability completeness and defect rate (TSE 2017)

- **Authors:** Patrick Rempel; Patrick Mäder
- **Exact title:** Preventing Defects: The Impact of Requirements Traceability
  Completeness on Software Quality
- **Venue:** IEEE Transactions on Software Engineering
- **Year:** 2017 (dblp/Crossref issue); OpenAlex publication year 2016
- **Volume/Issue/Pages:** 43(8), 777–797
- **DOI:** 10.1109/TSE.2016.2622264
- **Verification URLs fetched:**
  - `https://dblp.org/search/publ/api?q=Rempel+Mader+traceability+quality+defects&format=json`
  - `https://api.crossref.org/works/10.1109/TSE.2016.2622264`
  - `https://api.openalex.org/works/doi:10.1109/TSE.2016.2622264`

**Verbatim proof snippet (Crossref):**

> TITLE: ['Preventing Defects: The Impact of Requirements Traceability Completeness on Software Quality']
> AUTH: ['Patrick Rempel', 'Parick Mader']
> VENUE: ['IEEE Transactions on Software Engineering'] | VOL 43 ISSUE 8 | PAGE 777-797
> YEAR: [[2017, 8, 1]]

> (Crossref contains a typo in the second author's given name, "Parick Mader";
> dblp gives the correct "Patrick Mäder".)

**Verbatim abstract snippet (OpenAlex):**

> "Among stakeholders, traceability is often unpopular due to the unclear
> benefits. In fact, little evidence exists regarding the expected traceability
> benefits. ... we selected 24 medium to large-scale open-source projects. ... We
> analyzed that data in a multi-level Poisson regression analysis. We found that
> the degree of traceability completeness for three of the studied activities
> significantly affects software quality, which we quantified as defect rate. Our
> results provide for the first time empirical evidence that **more complete
> traceability decreases the expected defect rate** in the developed software."

**What the work claims (from the fetched abstract only):** Across 24 medium-to-
large open-source projects, using multi-level Poisson regression, traceability
completeness for three of four studied activities significantly affects defect
rate. The authors claim this is the first empirical evidence that more complete
traceability decreases expected defect rate, and argue traceability has
practical value even when not mandated by a standard or regulation.

**Relevance:** Verified empirical justification for *completeness* of
traceability — a strong argument for enforced/total link coverage rather than
best-effort recovery.

### 4.4 Chelouati et al. — GSN assurance cases (RESS 2023)

- **Authors:** Mohammed Chelouati; Abderraouf Boussif; Julie Beugin;
  El-Miloudi El-Koursi
- **Exact title:** Graphical safety assurance case using Goal Structuring
  Notation (GSN) — challenges, opportunities and a framework for autonomous
  trains
- **Venue:** Reliability Engineering & System Safety
- **Year:** 2023 (dblp); OpenAlex publication year 2022
- **Volume/Article:** 230, 108933
- **DOI:** 10.1016/j.ress.2022.108933
- **Verification URLs fetched:**
  - `https://dblp.org/search/publ/api?q=goal+structuring+notation&format=json`
  - `https://api.openalex.org/works/doi:10.1016/J.RESS.2022.108933`

**Verbatim proof snippet (dblp API):**

> - Graphical safety assurance case using Goal Structuring Notation (GSN) - challenges, opportunities and a framework for autonomous trains.
>    au: Mohammed Chelouati; Abderraouf Boussif; Julie Beugin; El-Miloudi El-Koursi
>    venue: Reliab. Eng. Syst. Saf. | y: 2023 | type: Journal Articles | pp: 108933 | doi: 10.1016/J.RESS.2022.108933

**Verbatim proof snippet (OpenAlex):**

> T: Graphical safety assurance case using Goal Structuring Notation (GSN) — challenges, opportunities and a framework for autonomous trains | Y: 2022 | type: article
> SRC: Reliability Engineering & System Safety | biblio: {'volume': '230', 'first_page': '108933'}

**What the work claims:** **Abstract NOT retrievable** (`ABS: (none)`). Metadata
verified. Usable as a peer-reviewed *pointer* to GSN-based assurance cases, but
do not summarise its content without reading it. For the original GSN definition
see UNVERIFIED section.

Also verified (metadata only, via the same dblp query; abstracts not fetched):

- Weihang Wu; Tim Kelly, "Combining Bayesian Belief Networks and the Goal
  Structuring Notation to Support Architectural Reasoning About Safety",
  SAFECOMP 2007, pp. 172–186, DOI 10.1007/978-3-540-75101-4_17
- Ewen Denney; Ganesh J. Pai; Ibrahim Habli, "Dynamic Safety Cases for
  Through-Life Safety Assurance", ICSE 2015, pp. 587–590,
  DOI 10.1109/ICSE.2015.199
- Robert Palin; Ibrahim Habli, "Assurance of Automotive Safety - A Safety Case
  Approach", SAFECOMP 2010, pp. 82–96, DOI 10.1007/978-3-642-15651-9_7

---

## 5. Annotation-Based / Code-Embedded Specification and Lightweight Formal Methods

### 5.1 Meyer — Design by Contract

- **Author:** Bertrand Meyer
- **Exact title:** Applying "Design by Contract"
- **Venue:** Computer (IEEE Computer)
- **Year:** 1992
- **Volume/Issue/Pages:** 25(10), 40–51
- **DOI:** 10.1109/2.161279
- **Verification URLs fetched:**
  - `https://dblp.org/search/publ/api?q=Meyer+applying+design+by+contract&format=json`
  - `https://api.openalex.org/works/doi:10.1109/2.161279`

**Verbatim proof snippet (dblp API):**

> - Applying &quot;Design by Contract&quot;.
>    au: Bertrand Meyer 0001
>    venue: Computer | y: 1992 | type: Journal Articles | pp: 40-51 | doi: 10.1109/2.161279

**Verbatim abstract snippet (OpenAlex):**

> "Methodological guidelines for object-oriented software construction that
> improve the reliability of the resulting software systems are presented. It is
> shown that the object-oriented techniques rely on the theory of design by
> contract, which underlies the design of the Eiffel analysis, design, and
> programming language and of the supporting libraries, from which a number of
> examples are drawn. The theory of contract design and the role of assertions in
> that theory are discussed."

**What the work claims (from the fetched abstract only):** Presents
methodological guidelines for OO construction that improve reliability, grounded
in the theory of design by contract as realised in Eiffel, and discusses the
role of assertions in that theory.

### 5.2 Leavens, Baker & Ruby — JML

- **Authors:** Gary T. Leavens; Albert L. Baker; Clyde Ruby
- **Exact title:** Preliminary design of JML: a behavioral interface
  specification language for Java
- **Venue:** ACM SIGSOFT Software Engineering Notes
- **Year:** 2006
- **Volume/Issue/Pages:** 31(3), 1–38
- **DOI:** 10.1145/1127878.1127884
- **Verification URLs fetched:**
  - `https://dblp.org/search/publ/api?q=Leavens+Baker+Ruby+preliminary+design+JML&format=json`
  - `https://api.openalex.org/works/doi:10.1145/1127878.1127884`

**Verbatim proof snippet (dblp API):**

> - Preliminary design of JML: a behavioral interface specification language for java.
>    au: Gary T. Leavens; Albert L. Baker; Clyde Ruby
>    venue: ACM SIGSOFT Softw. Eng. Notes | y: 2006 | type: Journal Articles | pp: 1-38 | doi: 10.1145/1127878.1127884

**Verbatim abstract snippet (OpenAlex):**

> "JML is a behavioral interface specification language tailored to Java(TM).
> Besides pre- and postconditions, it also allows assertions to be intermixed
> with Java code; these aid verification and debugging. JML is designed to be
> used by working software engineers; to do this it follows Eiffel in using Java
> expressions in assertions. JML combines this idea from Eiffel with the
> model-based approach to specifications, typified by VDM and Larch, which
> results in greater expressiveness."

**What the work claims (from the fetched abstract only):** JML is a behavioural
interface specification language for Java allowing pre-/postconditions and
assertions **intermixed with Java code**, designed for working engineers by
using Java expressions in assertions, and combining Eiffel's approach with
model-based specification (VDM, Larch).

**Relevance:** The canonical precedent for specifications *embedded in source
code* rather than held in an external document — directly comparable to an
annotation-based traceability design.

Earlier chapter version (metadata verified via dblp only, abstract not fetched):
Gary T. Leavens; Albert L. Baker; Clyde Ruby, "JML: A Notation for Detailed
Design", in *Behavioral Specifications of Businesses and Systems*, 1999,
pp. 175–188, DOI 10.1007/978-1-4615-5229-1_12.

### 5.3 Burdy et al. — JML tools and applications

- **Authors:** Lilian Burdy; Yoonsik Cheon; David R. Cok; Michael D. Ernst;
  Joseph R. Kiniry; Gary T. Leavens; K. Rustan M. Leino; Erik Poll
- **Exact title:** An overview of JML tools and applications
- **Venue:** International Journal on Software Tools for Technology Transfer
  (STTT)
- **Year:** 2005 (dblp); OpenAlex publication year 2004
- **Volume/Issue/Pages:** 7(3), 212–232
- **DOI:** 10.1007/s10009-004-0167-4
- **Verification URLs fetched:**
  - `https://dblp.org/search/publ/api?q=Burdy+Cheon+Cok+overview+JML+tools+applications&format=json`
  - `https://api.openalex.org/works/doi:10.1007/S10009-004-0167-4`

**Verbatim proof snippet (dblp API):**

> - An overview of JML tools and applications.
>    au: Lilian Burdy; Yoonsik Cheon; David R. Cok; Michael D. Ernst; Joseph R. Kiniry; Gary T. Leavens; K. Rustan M. Leino; Erik Poll
>    venue: Int. J. Softw. Tools Technol. Transf. | y: 2005 | type: Journal Articles | pp: 212-232 | doi: 10.1007/S10009-004-0167-4

**What the work claims:** **Abstract NOT retrievable** (`ABS:(none)`). Metadata
verified only. A conference version also exists: FMICS 2003, pp. 75–91,
DOI 10.1016/S1571-0661(04)80810-7 (verified via the same dblp query).

### 5.4 Jackson — Alloy

- **Author:** Daniel Jackson
- **Exact title:** Alloy: a lightweight object modelling notation
- **Venue:** ACM Transactions on Software Engineering and Methodology (TOSEM)
- **Year:** 2002
- **Volume/Issue/Pages:** 11(2), 256–290
- **DOI:** 10.1145/505145.505149
- **Verification URLs fetched:**
  - `https://dblp.org/search/publ/api?q=Jackson+Alloy+lightweight+object+modelling+notation&format=json`
  - `https://api.openalex.org/works/doi:10.1145/505145.505149`

**Verbatim proof snippet (dblp API):**

> - Alloy: a lightweight object modelling notation.
>    au: Daniel Jackson 0001
>    venue: ACM Trans. Softw. Eng. Methodol. | y: 2002 | type: Journal Articles | pp: 256-290 | doi: 10.1145/505145.505149

**Verbatim abstract snippet (OpenAlex):**

> "Alloy is a little language for describing structural properties. It offers a
> declaration syntax compatible with graphical object models, and a set-based
> formula syntax powerful enough to express complex constraints and yet amenable
> to a fully automatic semantic analysis. Its meaning is given by translation to
> an even smaller (formally defined) kernel. This paper presents the language in
> its entirety, and explains its motivation, contributions and deficiencies."

**What the work claims (from the fetched abstract only):** Alloy is a small
language for structural properties whose set-based formula syntax is expressive
yet "amenable to a fully automatic semantic analysis"; semantics given by
translation to a smaller formally defined kernel.

---

## 6. Architecture Conformance Checking / Dependency Rule Enforcement

### 6.1 Murphy, Notkin & Sullivan — Software Reflexion Models (TSE 2001)

- **Authors:** Gail C. Murphy; David Notkin; Kevin J. Sullivan
- **Exact title:** Software reflexion models: bridging the gap between design and
  implementation
- **Venue:** IEEE Transactions on Software Engineering
- **Year:** 2001
- **Volume/Issue/Pages:** 27(4), 364–380
- **DOI:** 10.1109/32.917525
- **Verification URLs fetched:**
  - `https://dblp.org/search/publ/api?q=Murphy+Notkin+Sullivan+software+reflexion+models&format=json`
  - `https://api.openalex.org/works/doi:10.1109/32.917525`

**Verbatim proof snippet (dblp API):**

> - Software Reflexion Models: Bridging the Gap between Design and Implementation.
>    au: Gail C. Murphy; David Notkin; Kevin J. Sullivan
>    venue: IEEE Trans. Software Eng. | y: 2001 | type: Journal Articles | pp: 364-380 | doi: 10.1109/32.917525

**Verbatim abstract snippet (OpenAlex):**

> "**The artifacts constituting a software system often drift apart over time.**
> We have developed the software reflexion model technique to help engineers
> perform various software engineering tasks by exploiting, rather than removing,
> the drift between design and implementation. More specifically, the technique
> helps an engineer compare artifacts by summarizing where one artifact (such as
> a design) is consistent with and inconsistent with another artifact (such as
> source). ... The software reflexion model technique has been applied to support
> a variety of tasks, including design conformance, change assessment, and an
> experimental reengineering of the million-lines-of-code Microsoft Excel
> product."

**What the work claims (from the fetched abstract only):** Software artefacts
drift apart over time; the reflexion model technique summarises where a
high-level artefact is consistent and inconsistent with source, and has been
applied to design conformance, change assessment, and a reengineering of
Microsoft Excel (~1 MLOC).

**Relevance:** Verified, quotable statement of the artefact-drift problem, plus
the classic conformance-checking technique.

### 6.2 Murphy, Notkin & Sullivan — original FSE 1995 paper

- **Authors:** Gail C. Murphy; David Notkin; Kevin Sullivan
- **Exact title:** Software reflexion models: bridging the gap between source and
  high-level models
- **Venue:** SIGSOFT FSE 1995 (Proceedings of the 3rd ACM SIGSOFT Symposium on
  Foundations of Software Engineering); also appears in ACM SIGSOFT Software
  Engineering Notes 20(4)
- **Year:** 1995
- **Pages:** 18–28
- **DOI:** 10.1145/222124.222136 (proceedings)
- **Verification URLs fetched:**
  - `https://dblp.org/search/publ/api?q=Murphy+Notkin+Sullivan+software+reflexion+models&format=json`
  - `https://doi.org/10.1145/222124.222136` (HTTP status check)
  - `https://api.crossref.org/works/10.1145/222124.222136`
  - `https://api.crossref.org/works/10.1145/222132.222136`

**Verbatim proof snippet (dblp API):**

> - Software Reflexion Models: Bridging the Gap Between Source and High-Level Models.
>    au: Gail C. Murphy; David Notkin; Kevin J. Sullivan
>    venue: SIGSOFT FSE | y: 1995 | type: Conference and Workshop Papers | pp: 18-28 | doi: 10.1145/222124.222136

**Verbatim proof snippet (DOI resolver):**

> `302 -> https://dl.acm.org/doi/10.1145/222124.222136`

**Verbatim proof snippet (Crossref, 10.1145/222124.222136):**

> ['Software reflexion models'] ['Proceedings of the 3rd ACM SIGSOFT symposium on Foundations of software engineering'] 18-28 {'date-parts': [[1995, 10]]}

**Verbatim proof snippet (Crossref, 10.1145/222132.222136):**

> ['Software reflexion models'] ['ACM SIGSOFT Software Engineering Notes'] 18-28 {'date-parts': [[1995, 10]]}

**DOI disambiguation (resolved this session):** Two distinct DOIs exist and both
resolve. `10.1145/222124.222136` is the **FSE proceedings** record;
`10.1145/222132.222136` is the **SIGSOFT Software Engineering Notes** issue
record. Use the proceedings DOI when citing FSE'95. **No abstract** was
retrievable for this version (OpenAlex returned ACM boilerplate, not an
abstract). Prefer the TSE 2001 version (6.1) when citing content.

### 6.3 Other verified conformance-checking work (metadata only)

Verified via `https://dblp.org/search/publ/api?q=architecture+conformance+checking&format=json`;
abstracts were **not** fetched, so cite metadata only:

- Sebastian Herold; Andreas Rausch, "A Rule-Based Approach to Architecture
  Conformance Checking as a Quality Management Measure", in *Relating System
  Quality and Software Architecture*, 2014, pp. 181–207,
  DOI 10.1016/B978-0-12-417009-4.00007-7
- Eduardo F. de Lima; Ricardo Terra, "ArchPython: architecture conformance
  checking for Python systems", SBES 2020, pp. 772–777,
  DOI 10.1145/3422392.3422505
- Mahesh De Silva; Indika Perera, "Preventing software architecture erosion
  through static architecture conformance checking", ICIIS 2015, pp. 43–48,
  DOI 10.1109/ICIINFS.2015.7398983
- Bruno Menezes; Ana Teresa C. Martins; Thiago Alves Rocha, "A Two-Level
  Approach Based on Model Checking to Support Architecture Conformance
  Checking", SBMF 2021, pp. 1–16, DOI 10.1007/978-3-030-92137-8_1
- Ipek Ozkaya, "Infrastructure as Code and Software Architecture Conformance
  Checking", IEEE Software, 2023, pp. 4–8, DOI 10.1109/MS.2022.3213880

**ArchUnit:** A dblp publication search for `ArchUnit` returned **zero results**
this session. Treat ArchUnit as a **tool / non-peer-reviewed artefact**; do not
cite it as a peer-reviewed source. Use 6.3 (esp. ArchPython, Herold & Rausch)
for the peer-reviewed framing of rule-based dependency enforcement.

---

## 7. Documentation–Code Drift

Note: a dblp search for `living documentation` returned no relevant software
engineering results (hits were about clinical/HCI documentation and one 2026
arXiv preprint). The drift theme is instead supported by the code-comment
inconsistency literature below, plus Murphy et al. (6.1) on artefact drift.

### 7.1 Liu et al. — automatic detection of outdated comments

- **Authors:** Zhiyong Liu; Huanchao Chen; Xiangping Chen; Xiaonan Luo; Fan Zhou
- **Exact title:** Automatic Detection of Outdated Comments During Code Changes
- **Venue:** COMPSAC 2018
- **Year:** 2018
- **Pages:** 154–163
- **DOI:** 10.1109/COMPSAC.2018.00028
- **Verification URLs fetched:**
  - `https://dblp.org/search/publ/api?q=outdated+comments+detection&format=json`
  - `https://api.openalex.org/works/doi:10.1109/COMPSAC.2018.00028`

**Verbatim proof snippet (dblp API):**

> - Automatic Detection of Outdated Comments During Code Changes.
>    au: Zhiyong Liu; Huanchao Chen; Xiangping Chen; Xiaonan Luo; Fan Zhou 0001
>    venue: COMPSAC | y: 2018 | type: Conference and Workshop Papers | pp: 154-163 | doi: 10.1109/COMPSAC.2018.00028

**Verbatim abstract snippet (OpenAlex):**

> "Comments are used as standard practice in software development to increase the
> readability of code and to express programmers' intentions in a more explicit
> manner. Nevertheless, **keeping comments up-to-date is often neglected for
> programmers.** In this paper, we proposed a machine learning based method for
> detecting the comments that should be changed during code changes. We utilized
> 64 features ... Experimental results show that **74.6% of outdated comments can
> be detected** using our method, and **77.2% of our detected outdated comments
> are real comments** which require to be updated."

**What the work claims (from the fetched abstract only):** Keeping comments
up-to-date is often neglected. A 64-feature ML method detects 74.6% of outdated
comments, with 77.2% of its detections being genuinely outdated.

**Relevance:** Verified, quotable recall (74.6%) and precision (77.2%) figures
for *heuristic* doc-drift detection — a concrete accuracy ceiling to contrast
with a compiler-enforced guarantee.

### 7.2 Stulova et al. — inconsistent comments in Java

- **Authors:** Nataliia Stulova; Arianna Blasi; Alessandra Gorla;
  Oscar Nierstrasz
- **Exact title:** Towards Detecting Inconsistent Comments in Java Source Code
  Automatically
- **Venue:** SCAM 2020 (IEEE Int. Working Conf. on Source Code Analysis and
  Manipulation)
- **Year:** 2020
- **Pages:** 65–69
- **DOI:** 10.1109/SCAM51674.2020.00012
- **Verification URLs fetched:**
  - `https://dblp.org/search/publ/api?q=detecting+inconsistent+code+comments&format=json`
  - `https://api.openalex.org/works/doi:10.1109/SCAM51674.2020.00012`

**Verbatim proof snippet (dblp API):**

> - Towards Detecting Inconsistent Comments in Java Source Code Automatically.
>    au: Nataliia Stulova; Arianna Blasi; Alessandra Gorla; Oscar Nierstrasz
>    venue: SCAM | y: 2020 | type: Conference and Workshop Papers | pp: 65-69 | doi: 10.1109/SCAM51674.2020.00012

**Verbatim abstract snippet (OpenAlex):**

> "A number of tools are available to software developers to check consistency of
> source code during software evolution. **However, none of these tools checks for
> consistency of the documentation accompanying the code. As a result, code and
> documentation often diverge, hindering program comprehension.** This leads to
> errors in how developers use source code, especially in the case of APIs of
> reusable libraries. We propose a technique and a tool, upDoc, to automatically
> detect code-comment inconsistency during code evolution."

**What the work claims (from the fetched abstract only):** Existing consistency
tools do not check documentation, so code and documentation diverge, hindering
comprehension and causing API misuse. Proposes upDoc, which maps code to
documentation and checks that code changes are matched by documentation changes;
the evaluation is explicitly described as **preliminary**.

**Relevance:** Directly quotable statement that no existing tooling
*compiler-checks* documentation consistency — the gap a compiler-checked
approach fills.

---

## 8. LLM-Generated Code Correctness and Trust in AI Artefacts

### 8.1 Pearce et al. — security of Copilot's code contributions (PEER-REVIEWED)

- **Authors:** Hammond Pearce; Baleegh Ahmad; Benjamin Tan; Brendan Dolan-Gavitt;
  Ramesh Karri
- **Exact title:** Asleep at the Keyboard? Assessing the Security of GitHub
  Copilot's Code Contributions
- **Venue:** IEEE Symposium on Security and Privacy (SP) 2022
- **Year:** 2022
- **Pages:** 754–768
- **DOI:** 10.1109/SP46214.2022.9833571
- **Verification URLs fetched:**
  - `https://dblp.org/search/publ/api?q=asleep+at+the+keyboard+security+code+contributions+Copilot&format=json`
  - `https://api.openalex.org/works/doi:10.1109/SP46214.2022.9833571`

**Verbatim proof snippet (dblp API):**

> - Asleep at the Keyboard? Assessing the Security of GitHub Copilot&apos;s Code Contributions.
>    au: Hammond Pearce; Baleegh Ahmad; Benjamin Tan 0001; Brendan Dolan-Gavitt; Ramesh Karri
>    venue: SP | y: 2022 | type: Conference and Workshop Papers | pp: 754-768 | doi: 10.1109/SP46214.2022.9833571

**Verbatim abstract snippet (OpenAlex):**

> "...we systematically investigate the prevalence and conditions that can cause
> GitHub Copilot to recommend insecure code. To perform this analysis we prompt
> Copilot to generate code in scenarios relevant to high-risk cybersecurity
> weaknesses, e.g. those from MITRE's "Top 25" Common Weakness Enumeration (CWE)
> list. ... In total, we produce 89 different scenarios for Copilot to complete,
> producing **1,689 programs. Of these, we found approximately 40% to be
> vulnerable.**"

**What the work claims (from the fetched abstract only):** Across 89 scenarios
targeting MITRE Top-25 CWEs, Copilot produced 1,689 programs of which
approximately 40% were vulnerable.

A later CACM version is also recorded in dblp (metadata verified, abstract not
separately fetched): Commun. ACM, 2025, pp. 96–105, DOI 10.1145/3610721.

### 8.2 Perry et al. — do users write more insecure code with AI assistants? (PEER-REVIEWED)

- **Authors:** Neil Perry; Megha Srivastava; Deepak Kumar; Dan Boneh
- **Exact title:** Do Users Write More Insecure Code with AI Assistants?
- **Venue:** ACM CCS 2023 (Conference on Computer and Communications Security)
- **Year:** 2023
- **Pages:** 2785–2799
- **DOI:** 10.1145/3576915.3623157
- **Verification URLs fetched:**
  - `https://dblp.org/search/publ/api?q=Do+users+write+more+insecure+code+with+AI+assistants&format=json`
  - `https://api.openalex.org/works/doi:10.1145/3576915.3623157`

**Verbatim proof snippet (dblp API):**

> - Do Users Write More Insecure Code with AI Assistants?
>    au: Neil Perry; Megha Srivastava; Deepak Kumar 0006; Dan Boneh
>    venue: CCS | y: 2023 | type: Conference and Workshop Papers | pp: 2785-2799 | doi: 10.1145/3576915.3623157

**Verbatim abstract snippet (OpenAlex):**

> "...we conduct a user study to examine how users interact with AI code
> assistants to solve a variety of security related tasks. Overall, we find that
> **participants who had access to an AI assistant wrote significantly less secure
> code than those without access to an assistant. Participants with access to an
> AI assistant were also more likely to believe they wrote secure code,
> suggesting that such tools may lead users to be overconfident about security
> flaws in their code.**"

**What the work claims (from the fetched abstract only):** In a user study,
participants with an AI assistant wrote significantly less secure code, *and*
were more likely to believe their code was secure — an explicit finding of
overconfidence in AI-assisted output.

**Relevance:** The strongest verified, peer-reviewed evidence that human trust in
AI-generated code is miscalibrated — motivating mechanical rather than
self-asserted verification of AI-produced artefacts.

### 8.3 Panickssery, Bowman & Feng — LLM self-preference bias (PEER-REVIEWED, NeurIPS)

- **Authors:** Arjun Panickssery; Samuel R. Bowman; Shi Feng
- **Exact title:** LLM Evaluators Recognize and Favor Their Own Generations
- **Venue:** NeurIPS 2024 (Neural Information Processing Systems)
- **Year:** 2024
- **arXiv (preprint of same work):** arXiv:2404.13076,
  DOI 10.48550/arXiv.2404.13076
- **Verification URLs fetched:**
  - `https://dblp.org/search/publ/api?q=LLM+evaluators+recognize+favor+their+own+generations&format=json`
  - `https://api.semanticscholar.org/graph/v1/paper/DOI:10.48550/ARXIV.2404.13076?fields=title,year,venue,abstract,authors`

**Verbatim proof snippet (dblp API):**

> - LLM Evaluators Recognize and Favor Their Own Generations.
>    au: Arjun Panickssery; Samuel R. Bowman; Shi Feng 0005
>    venue: NeurIPS | y: 2024 | type: Conference and Workshop Papers | pp: None | doi: None

**Verbatim proof snippet (Semantic Scholar):**

> LLM Evaluators Recognize and Favor Their Own Generations 2024 Neural Information Processing Systems
> ['Arjun Panickssery', 'Samuel R. Bowman', 'Shi Feng']

**Verbatim abstract snippet (Semantic Scholar):**

> "But new biases are introduced due to the same LLM acting as both the evaluator
> and the evaluatee. One such bias is **self-preference, where an LLM evaluator
> scores its own outputs higher than others' while human annotators consider them
> of equal quality.** ... We discover that, out of the box, LLMs such as GPT-4 and
> Llama 2 have non-trivial accuracy at distinguishing themselves from other LLMs
> and humans. By fine-tuning LLMs, we discover a linear correlation between
> self-recognition capability and the strength of self-preference bias... We
> discuss how self-recognition can interfere with unbiased evaluations and AI
> safety more generally."

**What the work claims (from the fetched abstract only):** When the same LLM is
both evaluator and evaluatee, a self-preference bias arises: the model scores its
own outputs higher than human annotators do. The paper finds a linear correlation
between self-recognition capability and self-preference strength, argued causal
via controlled experiments.

**Relevance:** Peer-reviewed evidence that LLM **self-certification is
systematically biased** — a direct argument against letting an LLM attest to its
own conformance, and for external/compiler-checked verification.

### 8.4 Zheng et al. — LLM-as-a-Judge (PEER-REVIEWED, NeurIPS)

- **Authors:** Lianmin Zheng; Wei-Lin Chiang; Ying Sheng; Siyuan Zhuang;
  Zhanghao Wu; Yonghao Zhuang; Zi Lin; Zhuohan Li; Dacheng Li; Eric P. Xing;
  Hao Zhang; Joseph E. Gonzalez; Ion Stoica
- **Exact title:** Judging LLM-as-a-Judge with MT-Bench and Chatbot Arena
- **Venue:** NeurIPS 2023
- **Year:** 2023
- **arXiv (preprint of same work):** arXiv:2306.05685,
  DOI 10.48550/arXiv.2306.05685
- **Verification URL fetched:**
  `https://dblp.org/search/publ/api?q=Judging+LLM-as-a-Judge+MT-Bench+Chatbot+Arena&format=json`

**Verbatim proof snippet (dblp API):**

> - Judging LLM-as-a-Judge with MT-Bench and Chatbot Arena.
>    au: Lianmin Zheng; Wei-Lin Chiang; Ying Sheng 0007; Siyuan Zhuang; Zhanghao Wu; Yonghao Zhuang 0001; Zi Lin; Zhuohan Li 0001; Dacheng Li; Eric P. Xing; Hao Zhang 0025; Joseph E. Gonzalez; Ion Stoica
>    venue: NeurIPS | y: 2023 | type: Conference and Workshop Papers | pp: None | doi: None

**What the work claims:** **Abstract NOT fetched** this session. Metadata
(authors, title, NeurIPS 2023) verified via dblp. Do not summarise its findings
without fetching the abstract.

### 8.5 Chen et al. — Codex / HumanEval — **PREPRINT (arXiv only)**

> **STATUS: PREPRINT.** dblp classifies this as "CoRR / Informal and Other
> Publications". It is **not** peer-reviewed. May only support attributed,
> non-absolute statements (e.g. "Chen et al. report that…").

- **Authors:** Mark Chen; Jerry Tworek; Heewoo Jun; Qiming Yuan;
  Henrique Ponde de Oliveira Pinto; Jared Kaplan; *et al.* (58 authors total,
  per the fetched arXiv record)
- **Exact title:** Evaluating Large Language Models Trained on Code
- **Venue:** arXiv (CoRR) — preprint
- **Year:** 2021 (arXiv published 2021-07-07)
- **arXiv ID:** arXiv:2107.03374
- **Verification URLs fetched:**
  - `https://dblp.org/search/publ/api?q=Chen+evaluating+large+language+models+trained+on+code&format=json`
  - `https://export.arxiv.org/api/query?id_list=2107.03374&max_results=1`

**Verbatim proof snippet (dblp API):**

> - Evaluating Large Language Models Trained on Code.
>    au: Mark Chen 0003; Jerry Tworek; ...
>    venue: CoRR | y: 2021 | type: Informal and Other Publications | pp: None | doi: None

**Verbatim proof snippet (arXiv API):**

> T: Evaluating Large Language Models Trained on Code
> PUB: 2021-07-07T17:41:24Z
> AUTH: ['Mark Chen', 'Jerry Tworek', 'Heewoo Jun', 'Qiming Yuan', 'Henrique Ponde de Oliveira Pinto', 'Jared Kaplan'] total 58

**Verbatim abstract snippet (arXiv API):**

> "We introduce Codex, a GPT language model fine-tuned on publicly available code
> from GitHub, and study its Python code-writing capabilities. A distinct
> production version of Codex powers GitHub Copilot. **On HumanEval, a new
> evaluation set we release to measure functional correctness for synthesizing
> programs from docstrings, our model solves 28.8% of the problems**, while GPT-3
> solves 0% and GPT-J solves 11.4%. ... Using this method, we solve 70.2% of our
> problems with 100 samples per problem. Careful investigation of our model
> reveals its limitations, including difficulty with docstrings describing long
> chains of operations and with binding operations to variables."

**What the work claims (from the fetched abstract only):** Codex solves 28.8% of
HumanEval problems at one sample, rising to 70.2% with 100 samples per problem.
The authors report limitations with docstrings describing long operation chains
and with binding operations to variables.

**Relevance (with preprint caveat):** Establishes that functional correctness of
LLM-generated code from a natural-language specification is far from guaranteed,
and that the *specification-to-code* step is precisely where LLMs struggle.

---

## UNVERIFIED — DO NOT USE

The following were sought but **could not be verified** in this session. Do not
cite any of these until independently verified.

1. **Kelly & Weaver, "The Goal Structuring Notation — A Safety Argument
   Notation"** (the original GSN definition paper).
   - Tried: dblp publication API queries `Kelly Weaver goal structuring
     notation`, `goal structuring notation`, `assurance case safety`.
   - Result: the exact Kelly & Weaver GSN paper did not appear in any dblp
     result set. Only later GSN applications were found (see §4.4).
   - **Do not guess the venue, year, or pages.** Use §4.4 (Chelouati et al.) or
     the verified Wu & Kelly SAFECOMP 2007 entry as a GSN pointer instead.

2. **Jackson & Wing, "Lightweight Formal Methods"** (IEEE Computer).
   - Tried: dblp queries `Jackson Wing lightweight formal methods Computer`,
     `Jackson Wing lightweight formal`, `lightweight formal methods`.
   - Result: zero matching hits; only unrelated recent lightweight-formal-methods
     papers were returned. **No DOI, venue, or year verified.**
   - Use Jackson's Alloy TOSEM 2002 paper (§5.4, verified) for the "lightweight
     formal methods" idea instead.

3. **ArchUnit** as a peer-reviewed citation.
   - Tried: dblp publication API query `ArchUnit`.
   - Result: **zero results.** Classify as a **tool / non-peer-reviewed
     artefact**. If cited, cite the project/documentation, not a paper.

4. **"Living documentation"** as a peer-reviewed software engineering concept.
   - Tried: dblp query `living documentation`.
   - Result: no relevant peer-reviewed SE source. Hits were clinical/HCI
     documentation papers, a 2015 SKY workshop paper on wiki-based living
     documentation (Yagel, DOI 10.5220/0005643700220026 — metadata seen but
     relevance and quality not assessed, abstract not fetched), and a 2026 arXiv
     preprint (CODENS, arXiv 2607.18356 — preprint, not peer-reviewed).
   - Use §7 (Liu et al.; Stulova et al.) for doc-code drift instead.

5. **Abstracts not retrievable** (metadata verified, content claims unverified —
   do not paraphrase these works' contents):
   - Gotel et al., "The Grand Challenge of Traceability (v1.0)" (§2.2)
   - Gotel et al., "Traceability Fundamentals" (§2.3)
   - Cleland-Huang, Gotel & Zisman (eds.), *Software and Systems Traceability*
     book (§2.4)
   - Mäder & Egyed, EMSE 2015 journal article (§4.2) — Springer blocked
     scraping; Crossref/OpenAlex/S2 all returned no abstract
   - Chelouati et al., RESS (§4.4)
   - Burdy et al., "An overview of JML tools and applications" (§5.3)
   - Murphy et al., FSE 1995 version (§6.2)
   - Zheng et al., "Judging LLM-as-a-Judge" (§8.4)
   - Antoniol et al. 2025 TSE retrospective
   - Leavens et al. 1999 "JML: A Notation for Detailed Design" chapter
   - All entries listed in §4.4 (secondary list) and §6.3

6. **Precision/recall figures for Antoniol et al. (2002) and Hayes et al.
   (2006).** The fetched abstracts contain **no numeric accuracy results**. Do
   not attribute specific precision/recall numbers to these two papers. Verified
   numeric limitation claims are available only from De Lucia et al. (§3.3,
   qualitative: "still far to support a complete semi-automatic recovery of all
   links") and Liu et al. (§7.1, 74.6% / 77.2%).

7. **Year/title discrepancies to resolve before submission** (both variants were
   observed in fetched sources this session):
   - Ramesh & Jarke title: "of" (dblp) vs "for" (OpenAlex) — §1.2
   - Mäder & Egyed EMSE: 2015 (dblp) vs 2014 (Crossref) — §4.2
   - Rempel & Mäder TSE: 2017 (dblp/Crossref) vs 2016 (OpenAlex) — §4.3
   - Grand Challenge / Traceability Fundamentals chapters: 2012 (dblp) vs 2011
     (OpenAlex) — §2.2, §2.3
   - Burdy et al. STTT: 2005 (dblp) vs 2004 (OpenAlex) — §5.3
   - Chelouati et al. RESS: 2023 (dblp) vs 2022 (OpenAlex) — §4.4

---

## BibTeX-Ready List

Only entries whose metadata was verified this session. Page ranges and DOIs are
copied verbatim from fetched API responses; none are guessed. Where a year was
ambiguous across sources, the dblp value is used and the alternative noted above.

```bibtex
@inproceedings{gotel1994analysis,
  author    = {Gotel, Orlena C. Z. and Finkelstein, Anthony},
  title     = {An Analysis of the Requirements Traceability Problem},
  booktitle = {Proceedings of the IEEE International Conference on Requirements Engineering (ICRE)},
  year      = {1994},
  pages     = {94--101},
  doi       = {10.1109/ICRE.1994.292398}
}

@article{ramesh2001toward,
  author  = {Ramesh, Balasubramaniam and Jarke, Matthias},
  title   = {Toward Reference Models for Requirements Traceability},
  journal = {IEEE Transactions on Software Engineering},
  year    = {2001},
  volume  = {27},
  number  = {1},
  pages   = {58--93},
  doi     = {10.1109/32.895989}
}

@inproceedings{gotel2012quest,
  author    = {Gotel, Orlena and Cleland-Huang, Jane and Hayes, Jane Huffman and
               Zisman, Andrea and Egyed, Alexander and Gr{\"u}nbacher, Paul and
               Antoniol, Giuliano},
  title     = {The Quest for Ubiquity: A Roadmap for Software and Systems Traceability Research},
  booktitle = {IEEE International Requirements Engineering Conference (RE)},
  year      = {2012},
  pages     = {71--80},
  doi       = {10.1109/RE.2012.6345841}
}

@incollection{gotel2012grand,
  author    = {Gotel, Orlena and Cleland-Huang, Jane and Hayes, Jane Huffman and
               Zisman, Andrea and Egyed, Alexander and Gr{\"u}nbacher, Paul and
               Dekhtyar, Alex and Antoniol, Giuliano and Maletic, Jonathan I.},
  title     = {The Grand Challenge of Traceability (v1.0)},
  booktitle = {Software and Systems Traceability},
  publisher = {Springer},
  year      = {2012},
  pages     = {343--409},
  doi       = {10.1007/978-1-4471-2239-5_16}
}

@incollection{gotel2012fundamentals,
  author    = {Gotel, Orlena and Cleland-Huang, Jane and Hayes, Jane Huffman and
               Zisman, Andrea and Egyed, Alexander and Gr{\"u}nbacher, Paul and
               Dekhtyar, Alex and Antoniol, Giuliano and Maletic, Jonathan I. and
               M{\"a}der, Patrick},
  title     = {Traceability Fundamentals},
  booktitle = {Software and Systems Traceability},
  publisher = {Springer},
  year      = {2012},
  pages     = {3--22},
  doi       = {10.1007/978-1-4471-2239-5_1}
}

@book{clelandhuang2012sst,
  editor    = {Cleland-Huang, Jane and Gotel, Orlena and Zisman, Andrea},
  title     = {Software and Systems Traceability},
  publisher = {Springer},
  year      = {2012},
  doi       = {10.1007/978-1-4471-2239-5}
}

@article{antoniol2002recovering,
  author  = {Antoniol, Giuliano and Canfora, Gerardo and Casazza, Gerardo and
             De Lucia, Andrea and Merlo, Ettore},
  title   = {Recovering Traceability Links between Code and Documentation},
  journal = {IEEE Transactions on Software Engineering},
  year    = {2002},
  volume  = {28},
  number  = {10},
  pages   = {970--983},
  doi     = {10.1109/TSE.2002.1041053}
}

@article{hayes2006advancing,
  author  = {Hayes, Jane Huffman and Dekhtyar, Alex and Sundaram, Senthil Karthikeyan},
  title   = {Advancing Candidate Link Generation for Requirements Tracing: The Study of Methods},
  journal = {IEEE Transactions on Software Engineering},
  year    = {2006},
  volume  = {32},
  number  = {1},
  pages   = {4--19},
  doi     = {10.1109/TSE.2006.3}
}

@article{delucia2007recovering,
  author  = {De Lucia, Andrea and Fasano, Fausto and Oliveto, Rocco and Tortora, Genoveffa},
  title   = {Recovering Traceability Links in Software Artifact Management Systems
             Using Information Retrieval Methods},
  journal = {ACM Transactions on Software Engineering and Methodology},
  year    = {2007},
  volume  = {16},
  number  = {4},
  articleno = {13},
  doi     = {10.1145/1276933.1276934}
}

@inproceedings{mader2012assessing,
  author    = {M{\"a}der, Patrick and Egyed, Alexander},
  title     = {Assessing the Effect of Requirements Traceability for Software Maintenance},
  booktitle = {IEEE International Conference on Software Maintenance (ICSM)},
  year      = {2012},
  pages     = {171--180},
  doi       = {10.1109/ICSM.2012.6405269}
}

@article{mader2015developers,
  author  = {M{\"a}der, Patrick and Egyed, Alexander},
  title   = {Do Developers Benefit from Requirements Traceability When Evolving
             and Maintaining a Software System?},
  journal = {Empirical Software Engineering},
  year    = {2015},
  volume  = {20},
  number  = {2},
  pages   = {413--441},
  doi     = {10.1007/s10664-014-9314-z}
}

@article{rempel2017preventing,
  author  = {Rempel, Patrick and M{\"a}der, Patrick},
  title   = {Preventing Defects: The Impact of Requirements Traceability
             Completeness on Software Quality},
  journal = {IEEE Transactions on Software Engineering},
  year    = {2017},
  volume  = {43},
  number  = {8},
  pages   = {777--797},
  doi     = {10.1109/TSE.2016.2622264}
}

@article{chelouati2023gsn,
  author  = {Chelouati, Mohammed and Boussif, Abderraouf and Beugin, Julie and
             El-Koursi, El-Miloudi},
  title   = {Graphical Safety Assurance Case Using Goal Structuring Notation (GSN):
             Challenges, Opportunities and a Framework for Autonomous Trains},
  journal = {Reliability Engineering \& System Safety},
  year    = {2023},
  volume  = {230},
  pages   = {108933},
  doi     = {10.1016/j.ress.2022.108933}
}

@inproceedings{wu2007combining,
  author    = {Wu, Weihang and Kelly, Tim},
  title     = {Combining Bayesian Belief Networks and the Goal Structuring Notation
               to Support Architectural Reasoning About Safety},
  booktitle = {SAFECOMP},
  year      = {2007},
  pages     = {172--186},
  doi       = {10.1007/978-3-540-75101-4_17}
}

@inproceedings{denney2015dynamic,
  author    = {Denney, Ewen and Pai, Ganesh J. and Habli, Ibrahim},
  title     = {Dynamic Safety Cases for Through-Life Safety Assurance},
  booktitle = {International Conference on Software Engineering (ICSE)},
  year      = {2015},
  pages     = {587--590},
  doi       = {10.1109/ICSE.2015.199}
}

@article{meyer1992applying,
  author  = {Meyer, Bertrand},
  title   = {Applying ``Design by Contract''},
  journal = {Computer},
  year    = {1992},
  volume  = {25},
  number  = {10},
  pages   = {40--51},
  doi     = {10.1109/2.161279}
}

@article{leavens2006preliminary,
  author  = {Leavens, Gary T. and Baker, Albert L. and Ruby, Clyde},
  title   = {Preliminary Design of {JML}: A Behavioral Interface Specification
             Language for Java},
  journal = {ACM SIGSOFT Software Engineering Notes},
  year    = {2006},
  volume  = {31},
  number  = {3},
  pages   = {1--38},
  doi     = {10.1145/1127878.1127884}
}

@article{burdy2005overview,
  author  = {Burdy, Lilian and Cheon, Yoonsik and Cok, David R. and Ernst, Michael D. and
             Kiniry, Joseph R. and Leavens, Gary T. and Leino, K. Rustan M. and Poll, Erik},
  title   = {An Overview of {JML} Tools and Applications},
  journal = {International Journal on Software Tools for Technology Transfer},
  year    = {2005},
  volume  = {7},
  number  = {3},
  pages   = {212--232},
  doi     = {10.1007/s10009-004-0167-4}
}

@article{jackson2002alloy,
  author  = {Jackson, Daniel},
  title   = {Alloy: A Lightweight Object Modelling Notation},
  journal = {ACM Transactions on Software Engineering and Methodology},
  year    = {2002},
  volume  = {11},
  number  = {2},
  pages   = {256--290},
  doi     = {10.1145/505145.505149}
}

@article{murphy2001reflexion,
  author  = {Murphy, Gail C. and Notkin, David and Sullivan, Kevin J.},
  title   = {Software Reflexion Models: Bridging the Gap between Design and Implementation},
  journal = {IEEE Transactions on Software Engineering},
  year    = {2001},
  volume  = {27},
  number  = {4},
  pages   = {364--380},
  doi     = {10.1109/32.917525}
}

@inproceedings{murphy1995reflexion,
  author    = {Murphy, Gail C. and Notkin, David and Sullivan, Kevin J.},
  title     = {Software Reflexion Models: Bridging the Gap Between Source and
               High-Level Models},
  booktitle = {Proceedings of the 3rd ACM SIGSOFT Symposium on Foundations of
               Software Engineering (FSE)},
  year      = {1995},
  pages     = {18--28},
  doi       = {10.1145/222124.222136}
}

@incollection{herold2014rule,
  author    = {Herold, Sebastian and Rausch, Andreas},
  title     = {A Rule-Based Approach to Architecture Conformance Checking as a
               Quality Management Measure},
  booktitle = {Relating System Quality and Software Architecture},
  year      = {2014},
  pages     = {181--207},
  doi       = {10.1016/B978-0-12-417009-4.00007-7}
}

@inproceedings{lima2020archpython,
  author    = {de Lima, Eduardo F. and Terra, Ricardo},
  title     = {{ArchPython}: Architecture Conformance Checking for Python Systems},
  booktitle = {Brazilian Symposium on Software Engineering (SBES)},
  year      = {2020},
  pages     = {772--777},
  doi       = {10.1145/3422392.3422505}
}

@inproceedings{liu2018automatic,
  author    = {Liu, Zhiyong and Chen, Huanchao and Chen, Xiangping and Luo, Xiaonan
               and Zhou, Fan},
  title     = {Automatic Detection of Outdated Comments During Code Changes},
  booktitle = {IEEE Annual Computer Software and Applications Conference (COMPSAC)},
  year      = {2018},
  pages     = {154--163},
  doi       = {10.1109/COMPSAC.2018.00028}
}

@inproceedings{stulova2020towards,
  author    = {Stulova, Nataliia and Blasi, Arianna and Gorla, Alessandra and
               Nierstrasz, Oscar},
  title     = {Towards Detecting Inconsistent Comments in Java Source Code Automatically},
  booktitle = {IEEE International Working Conference on Source Code Analysis and
               Manipulation (SCAM)},
  year      = {2020},
  pages     = {65--69},
  doi       = {10.1109/SCAM51674.2020.00012}
}

@inproceedings{pearce2022asleep,
  author    = {Pearce, Hammond and Ahmad, Baleegh and Tan, Benjamin and
               Dolan-Gavitt, Brendan and Karri, Ramesh},
  title     = {Asleep at the Keyboard? Assessing the Security of {GitHub Copilot}'s
               Code Contributions},
  booktitle = {IEEE Symposium on Security and Privacy (SP)},
  year      = {2022},
  pages     = {754--768},
  doi       = {10.1109/SP46214.2022.9833571}
}

@inproceedings{perry2023insecure,
  author    = {Perry, Neil and Srivastava, Megha and Kumar, Deepak and Boneh, Dan},
  title     = {Do Users Write More Insecure Code with {AI} Assistants?},
  booktitle = {ACM SIGSAC Conference on Computer and Communications Security (CCS)},
  year      = {2023},
  pages     = {2785--2799},
  doi       = {10.1145/3576915.3623157}
}

@inproceedings{panickssery2024llm,
  author    = {Panickssery, Arjun and Bowman, Samuel R. and Feng, Shi},
  title     = {{LLM} Evaluators Recognize and Favor Their Own Generations},
  booktitle = {Advances in Neural Information Processing Systems (NeurIPS)},
  year      = {2024},
  note      = {Preprint: arXiv:2404.13076, doi:10.48550/arXiv.2404.13076}
}

@inproceedings{zheng2023judging,
  author    = {Zheng, Lianmin and Chiang, Wei-Lin and Sheng, Ying and Zhuang, Siyuan
               and Wu, Zhanghao and Zhuang, Yonghao and Lin, Zi and Li, Zhuohan and
               Li, Dacheng and Xing, Eric P. and Zhang, Hao and Gonzalez, Joseph E.
               and Stoica, Ion},
  title     = {Judging {LLM-as-a-Judge} with {MT-Bench} and {Chatbot Arena}},
  booktitle = {Advances in Neural Information Processing Systems (NeurIPS)},
  year      = {2023},
  note      = {Preprint: arXiv:2306.05685, doi:10.48550/arXiv.2306.05685}
}

% PREPRINT — NOT PEER REVIEWED. Use only for attributed, non-absolute statements.
@misc{chen2021codex,
  author       = {Chen, Mark and Tworek, Jerry and Jun, Heewoo and Yuan, Qiming and
                  Pinto, Henrique Ponde de Oliveira and Kaplan, Jared and others},
  title        = {Evaluating Large Language Models Trained on Code},
  year         = {2021},
  eprint       = {2107.03374},
  archivePrefix = {arXiv},
  primaryClass = {cs.LG},
  note         = {arXiv preprint; not peer reviewed}
}
```

---

## Summary: Verified vs. Unverified

### Fully verified (metadata **and** abstract fetched this session) — 15

| # | Work | Type |
|---|------|------|
| 1 | Gotel & Finkelstein, ICRE 1994 | peer-reviewed |
| 2 | Ramesh & Jarke, TSE 2001 | peer-reviewed |
| 3 | Gotel et al., RE 2012 (roadmap) | peer-reviewed |
| 4 | Antoniol et al., TSE 2002 | peer-reviewed |
| 5 | Hayes et al., TSE 2006 | peer-reviewed |
| 6 | De Lucia et al., TOSEM 2007 | peer-reviewed |
| 7 | Mäder & Egyed, ICSM 2012 | peer-reviewed |
| 8 | Rempel & Mäder, TSE 2017 | peer-reviewed |
| 9 | Meyer, IEEE Computer 1992 | peer-reviewed |
| 10 | Leavens et al., SIGSOFT SEN 2006 (JML) | peer-reviewed |
| 11 | Jackson, TOSEM 2002 (Alloy) | peer-reviewed |
| 12 | Murphy et al., TSE 2001 (reflexion) | peer-reviewed |
| 13 | Liu et al., COMPSAC 2018 | peer-reviewed |
| 14 | Stulova et al., SCAM 2020 | peer-reviewed |
| 15 | Pearce et al., IEEE S&P 2022 | peer-reviewed |
| 16 | Perry et al., CCS 2023 | peer-reviewed |
| 17 | Panickssery et al., NeurIPS 2024 | peer-reviewed |
| 18 | Chen et al., arXiv 2021 (Codex) | **PREPRINT** |

(18 rows; 17 peer-reviewed + 1 explicitly labelled preprint.)

### Metadata verified, abstract NOT retrievable — 11+

Grand Challenge chapter; Traceability Fundamentals chapter; the Springer
*Software and Systems Traceability* book; Mäder & Egyed EMSE 2015; Chelouati et
al. RESS; Burdy et al. STTT; Murphy et al. FSE 1995; Zheng et al. NeurIPS 2023;
Antoniol et al. TSE 2025 retrospective; Leavens et al. 1999 chapter; plus the
secondary lists in §4.4 and §6.3. **Cite metadata only — do not paraphrase
content.**

### Not verified at all — 4

1. Kelly & Weaver original GSN paper — not found in dblp.
2. Jackson & Wing "Lightweight Formal Methods" — not found in dblp.
3. ArchUnit peer-reviewed source — zero dblp results; treat as a tool.
4. "Living documentation" peer-reviewed SE source — no relevant result.

### Key usable evidence for a compiler-checked traceability argument

- **Heuristic recovery has a real ceiling:** De Lucia et al. (§3.3) — IR tools
  are "still far to support a complete semi-automatic recovery of all links";
  Liu et al. (§7.1) — 74.6% recall / 77.2% precision for outdated-comment
  detection.
- **Traceability is not ubiquitous despite automation research:** Gotel et al.
  RE 2012 (§2.1).
- **Traceability demonstrably pays off:** Mäder & Egyed ICSM 2012 (21% faster,
  60% more correct solutions, §4.1); Rempel & Mäder TSE 2017 (completeness
  lowers defect rate, §4.3).
- **Nothing currently compiler-checks documentation:** Stulova et al. (§7.2).
- **Artefacts drift:** Murphy et al. TSE 2001 (§6.1).
- **AI self-assessment is untrustworthy:** Perry et al. (users overconfident,
  §8.2); Panickssery et al. (LLM self-preference bias, §8.3); Pearce et al.
  (~40% of Copilot programs vulnerable, §8.1).
