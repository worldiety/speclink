# Verified facts on functional-safety and software-process standards

Research log for the paper. **All statements below are backed by a page fetched during
the research session of 2026-08-30.** Verbatim supporting quotes are in the Sources table
at the end. Anything that could not be verified is listed in
"UNVERIFIED — DO NOT USE".

Date accessed for every source: **2026-08-30**.

---

## 1. ISO 26262 — Road vehicles — Functional safety (2nd edition, 2018)

### Verified facts

- ISO 26262-1:2018 is titled "Road vehicles — Functional safety — Part 1: Vocabulary",
  Edition 2, published 2018-12, ISO/TC 22/SC 32, 33 pages. [S1]
- The current lifecycle stage of ISO 26262-1:2018 is 90.92 ("International Standard to be
  revised"); a successor ISO/DIS 26262-1 is under development. [S1]
- ISO publishes a package containing 10 parts of the 2018 edition; their exact titles are: [S2]
  - Part 1: Vocabulary
  - Part 2: Management of functional safety
  - Part 3: Concept phase
  - Part 4: Product development at the system level
  - Part 5: Product development at the hardware level
  - Part 6: Product development at the software level
  - Part 7: Production, operation, service and decommissioning
  - Part 8: Supporting processes
  - Part 9: Automotive safety integrity level (ASIL)-oriented and safety-oriented analyses
  - Part 10: Guidelines on ISO 26262
- Two further parts exist beyond that package and are cited in the bibliography of
  ISO 26262-8:2018: ISO 26262-11:2018 "Guideline on application of ISO 26262 to
  semiconductors" and ISO 26262-12:2018 "Adaptation of ISO 26262 for motorcycles". [S6]
- ISO 26262 is explicitly positioned as a sector adaptation of IEC 61508: "The ISO 26262
  series of standards is the adaptation of IEC 61508 series of standards to address the
  sector specific needs of electrical and/or electronic (E/E) systems within road
  vehicles." [S6]
- The series uses a V-model as reference process model, and cross-references clauses in the
  notation "m-n" where m is the part number and n the clause number (e.g. "2-6" = ISO
  26262-2:2018, Clause 6). [S6]
- ASIL is normatively defined in ISO 26262-1:2018, term 3.6, as "one of four levels to
  specify the item's or element's necessary ISO 26262 requirements and safety measures to
  apply for avoiding an unreasonable risk, with D representing the most stringent and A the
  least stringent level"; "QM is not an ASIL." [S5]
- "software tool" is defined in ISO 26262-1:2018, term 3.158 as "computer program used in
  the development of an item or element". [S5]

#### Part 6 — Product development at the software level

- ISO 26262-6:2018, Edition 2, 2018-12, 57 pages, ISO/TC 22/SC 32. [S3]
- Scope (verbatim from the ISO abstract): it "specifies the requirements for product
  development at the software level for automotive applications", covering general topics,
  specification of the software safety requirements, software architectural design,
  software unit design and implementation, software unit verification, software integration
  and verification, and testing of the embedded software; it also specifies requirements
  associated with the use of configurable software. [S3]
- Clause structure (from the publicly visible ISO OBP table of contents, which is truncated
  after 10.2): 4 Requirements for compliance; 5 General topics for the product development
  at the software level; 6 Specification of software safety requirements; 7 Software
  architectural design; 8 Software unit design and implementation; 9 Software unit
  verification; 10 Software integration and verification. [S8]

#### Part 8 — Supporting processes

- ISO 26262-8:2018, Edition 2, 2018-12, 60 pages, ISO/TC 22/SC 32. [S4]
- Scope: it "specifies the requirements for supporting processes", explicitly listing:
  interfaces within distributed developments; overall management of safety requirements;
  configuration management; change management; verification; documentation management;
  **confidence in the use of software tools**; qualification of software components;
  evaluation of hardware elements; proven in use argument; interfacing an application out of
  scope of ISO 26262; and integration of safety-related systems not developed according to
  ISO 26262. [S4][S6]
- Clause numbering that is publicly visible on ISO OBP (TOC truncated after 10.2):
  5 Interfaces within distributed developments; 6 Specification and management of safety
  requirements; 7 Configuration management; 8 Change management; 9 Verification;
  10 Documentation management. [S6]
- The bibliography of ISO 26262-8:2018 cites, among others, RTCA DO-178C, CMMI for
  Development (CMMI-DEV, Carnegie Mellon University SEI), Automotive SPICE®,
  ISO/IEC/IEEE 12207, ISO/IEC/IEEE 29148 and the ISO/IEC 33000 series. Automotive SPICE® is
  marked as "an example of a suitable product available commercially". [S6]

> **Paywall note.** The normative requirement text of ISO 26262-6 and ISO 26262-8 is behind
> a paywall (CHF 225 per part). Only the foreword, introduction, scope, normative
> references and a truncated table of contents are publicly readable on ISO OBP. Therefore
> **no normative requirement of ISO 26262 is paraphrased in this file.** The specific
> clause numbers for "Confidence in the use of software tools" (TI/TCL) and the notion of
> Tool Impact / Tool Confidence Level are listed under UNVERIFIED below.

---

## 2. IEC 61508 — Functional safety of E/E/PE safety-related systems

### Verified facts

- The series title is "Functional safety of electrical/electronic/programmable electronic
  safety-related systems"; the current edition consists of **Parts 1 to 7**, Edition 2.0,
  publication date 2010-04-30, IEC TC 65/SC 65A, ISBN 9782889109852. [S9]
- Part titles (all :2010, Edition 2.0): [S10]–[S16]
  - Part 1: General requirements
  - Part 2: Requirements for electrical/electronic/programmable electronic safety-related systems
  - Part 3: Software requirements
  - Part 4: Definitions and abbreviations
  - Part 5: Examples of methods for the determination of safety integrity levels
  - Part 6: Guidelines on the application of IEC 61508-2 and IEC 61508-3
  - Part 7: Overview of techniques and measures
- IEC 61508-1:2010 "covers those aspects to be considered when electrical/electronic/
  programmable electronic (E/E/PE) systems are used to carry out safety functions". A major
  stated objective is "to facilitate the development of product and application sector
  international standards by the technical committees responsible for the product or
  application sector" — this is exactly the mechanism by which ISO 26262 exists. [S10]
- IEC 61508-1:2010 "has the status of a basic safety publication according to IEC Guide
  104". [S10]
- IEC 61508-3:2010 (software) explicitly addresses tools: it "provides specific requirements
  applicable to support tools used to develop and configure a safety-related system" and
  "provides, in conjunction with IEC 61508-1 and IEC 61508-2, requirements for support tools
  such as development and design tools, language translators, testing and debugging tools,
  configuration management tools". [S12]
- Relation to ISO 26262 is verified from the ISO side: ISO 26262 is "the adaptation of
  IEC 61508 series of standards" to road vehicles, and ISO 26262-8:2018 lists
  "IEC 61508 (all parts)" in its bibliography. [S6]

---

## 3. DO-178C / ED-12C and supplements (RTCA / EUROCAE)

### Verified facts

- **DO-178C** — "Software Considerations in Airborne Systems and Equipment Certification",
  publisher RTCA, committee SC-205, issue date 2011-12-13. RTCA states: "Compliance with the
  objectives of DO-178C is the primary means of obtaining approval of software used in civil
  aviation products." Errata exist against DO-178C. [S17]
- **DO-330** — "Software Tool Qualification Considerations", RTCA, committee SC-205, issue
  date 2011-12-13. It "explains the process and objectives for qualifying tools"; a tool is
  defined there as "a computer program or a functional part thereof, used to help develop,
  transform, test, analyze, produce or modify another program, its data or its
  documentation". RTCA notes it "provides guidance for airborne and ground-based software"
  and "may also be used by other domains, such as automotive, space, systems, electronic
  hardware, aeronautical databases and safety assessment processes". [S18]
- **DO-333** — "Formal Methods Supplement to DO-178C and DO-278A", RTCA, committee SC-205,
  issue date 2011-12-13. It "identifies the additions, modifications and substitutions to
  DO-178C and DO-278A objectives when formal methods are used as part of a software life
  cycle". RTCA defines formal methods there as "mathematically-based techniques for the
  specification, development and verification of software aspects of digital systems". [S19]
- **ED-12C** — "Software considerations in airborne systems and equipment certification",
  publisher EUROCAE. EUROCAE's own training page states ED-12C "is the European reference
  standard for airborne software certification and is equivalent to RTCA DO-178C". [S20]
- EUROCAE counterparts of the DO-178C supplements: **ED-216** "Formal methods supplement to
  ED-12C and ED-109A"; **ED-217** object-oriented technology supplement; **ED-218**
  model-based development and verification supplement; **ED-94C** "Supporting Information
  for ED-12C and ED-109A". [S20]
- **ED-215** — "Software Tool Qualification Considerations" (EUROCAE counterpart of DO-330);
  its purpose "is to provide tool qualification guidance", with clarification material in
  the form of FAQs. A Corrigendum 1 exists. [S21]
- ISO 26262-8:2018 cites "RTCA DO-178C, Software Considerations in Airborne Systems and
  Equipment Certification" in its bibliography — a verified cross-domain link. [S6]

> **Paywall note.** DO-178C, DO-330, DO-333, ED-12C and ED-215 are paid documents
> (DO-178C electronic USD 525; DO-330/DO-333 hard copy USD 335.40). The normative objective
> tables (including the traceability objectives between requirements, source code and
> executable object code) were **not** accessible. See UNVERIFIED.

---

## 4. IEC 62304 — Medical device software

### Verified facts

- IEC 62304 is titled "Medical device software — Software life cycle processes", published
  by IEC, technical committee TC 62/SC 62A. The current consolidated version is
  **IEC 62304:2006+AMD1:2015 CSV**, Edition 1.1, publication date 2015-06-26, 170 pages,
  ISBN 9782832227657, ICS 11.040.01. [S22]
- It "Defines the life cycle requirements for medical device software. The set of processes,
  activities, and tasks described in this standard establishes a common framework for
  medical device software life cycle processes." It applies "when software is itself a
  medical device or when software is an embedded or integral part of the final medical
  device", and does "not cover validation and final release of the medical device". [S22]
- The consolidated version "consists of the first edition (2006) and its amendment 1
  (2015)". [S22]

---

## 5. Automotive SPICE (ASPICE)

The full normative PAM/PRM document is **freely downloadable** from VDA QMC, so the facts
below are quoted from the primary source, not from an abstract.

### Verified facts

- Publisher/owner: VDA QMC (Quality Management Center of the German Association of the
  Automotive Industry, VDA e.V., Behrenstraße 35, 10117 Berlin). "Automotive SPICE® is a
  registered trademark of the VDA-QMC." [S23][S26]
- Current published version: **Automotive SPICE® Process Reference Model / Process
  Assessment Model, Version 4.0**, author "VDA Working Group 13", dated **2023-11-29**,
  status "Released". It is a revision of version 3.1. [S24]
- Automotive SPICE was first published in 2005 by the "Automotive Special Interest Group";
  "SPICE" stands for "Systems/Software Process Improvement and Capability DEtermination"
  (VDA QMC wording; the intacs wording is "Software Process Improvement and Capability
  dEtermination"). [S23][S26]
- ASPICE consists of two dimensions: a **process dimension** and a **capability dimension**.
  VDA QMC states the process dimension "describes the 32 processes of the Automotive
  SPICE® model, which are divided into 3 categories and 11 groups". [S23]
- Relation to the ISO/IEC 330xx family: "The Automotive SPICE process reference model and
  process assessment model are conformant with the ISO/IEC 33004:2015 and can be used as
  the basis for conducting an assessment of process capability. An ISO/IEC 33003:2015
  compliant Measurement Framework is defined in section 5." The PAM "was developed in
  accordance with the requirements of ISO/IEC 33004:2015". [S24]
- The measurement framework "is an adaption of ISO/IEC 33020:2019"; the rating scale "is
  identical to ISO/IEC 33020:2019"; rating and aggregation methods are taken from
  ISO/IEC 33020:2019. The PAM also reproduces material from ISO/IEC 15504-5:2006. [S24]
- ASPICE 4.0 explicitly notes the 15504→33004 lineage: "a PRM/PAM according to ISO/IEC 33004
  (formerly ISO/IEC 15504-2)". [S24]
- Six capability levels (0–5) with nine process attributes: [S24]
  - Level 0 Incomplete process — "The process is not implemented or fails to achieve its process purpose."
  - Level 1 Performed process (PA 1.1 Process performance)
  - Level 2 Managed process (PA 2.1 Performance management, PA 2.2 Work product management)
  - Level 3 Established process (PA 3.1 Process definition, PA 3.2 Process deployment)
  - Level 4 Predictable process (PA 4.1 Quantitative analysis, PA 4.2 Quantitative control)
  - Level 5 Innovating process (PA 5.1 Process innovation, PA 5.2 Process innovation implementation)
- The SWE (Software Engineering) process group of ASPICE 4.0 contains exactly six
  processes: [S24]
  - SWE.1 Software Requirements Analysis
  - SWE.2 Software Architectural Design
  - SWE.3 Software Detailed Design and Unit Construction
  - SWE.4 Software Unit Verification
  - SWE.5 Software Component Verification and Integration Verification
  - SWE.6 Software Verification
- Note that ASPICE 4.0 **renamed** SWE.3 and SWE.5 relative to earlier versions (the PAM's
  own annex still refers to SWE.5 as "Software Integration & Integration Verification" in
  prose). [S24]
- **Bidirectional traceability is a first-class, recurring concept.** In ASPICE 4.0 the
  phrase "Ensure consistency and establish bidirectional traceability" is the name of a base
  practice in multiple processes (e.g. SYS.2.BP5, SYS.3.BP4, SYS.4.BP4, SYS.5.BP4,
  SWE.1.BP5). [S24]
- SWE.1 Software Requirements Analysis has 7 process outcomes; outcomes 5 and 6 are
  "Consistency and bidirectional traceability are established between software requirements
  and system requirements" and "…and system architecture". SWE.1.BP5 requires establishing
  bidirectional traceability in both directions. [S24]
- The rationale given in the PAM (Note 11 to SWE.1.BP5) is explicitly that traceability is
  not an end in itself: "Bidirectional traceability supports consistency, and facilitates
  impact analysis of change requests, and demonstration of verification coverage.
  Traceability alone, e.g., the existence of links, does not necessarily mean that the
  information is consistent with each other." Note 9: "Redundant traceability is not
  intended." [S24]
- Traceability/consistency is materialised as a work product: SWE.1's output information
  items include "13-51 Consistency Evidence" mapped to outcomes 5 and 6. [S24]
- SWE.3's outcome 3 requires bidirectional traceability between software detailed design and
  software architecture, **and between source code and software detailed design** — i.e.
  traceability reaches down to code. [S24]
- SWE.4 and SWE.6 require bidirectional traceability between verification measures and the
  verified artefact, and between verification results and verification measures. [S24]
- SWE.1.BP1 references requirement-quality characteristics from other standards:
  "Characteristics of requirements are defined in standards such as ISO IEEE 29148,
  ISO 26262-8:2018, or the INCOSE Guide for Writing Requirements", and names examples such
  as "verifiability …, unambiguity/comprehensibility, freedom from design and
  implementation, and not contradicting any other requirement". [S24]
- ASPICE's scope statement allows augmenting with other PRMs: "If processes beyond the scope
  of Automotive SPICE are needed, appropriate processes from other process reference models
  such as ISO/IEC 12207 or ISO/IEC/IEEE 15288 may be added". [S24]
- Ecosystem scale (VDA QMC self-reported figures on their ASPICE landing page): 7619
  Automotive SPICE® assessors, 51 countries, 5 available languages. [S23]
- **Version caveat:** VDA QMC's page says "the current version 4.0" and links the PAM v4.0
  PDF [S23], while intacs.info refers in 2026 to alignment with "Automotive SPICE® 4.1"
  [S26]. A version 4.1 document itself could not be retrieved (automotivespice.com returned
  HTTP 500 during this session). Treat "4.1" as attributed to intacs, not as established.

### Related ISO/IEC 330xx facts (verified separately)

- ISO/IEC 33002:2015 "Information technology — Process assessment — Requirements for
  performing process assessment", published 2015-03, ISO/IEC JTC 1/SC 7, 16 pages, stage
  90.93 (confirmed). It "defines the minimum set of requirements for performing an
  assessment that will ensure assessment results are objective, consistent, repeatable, and
  representative of the assessed processes." Its ISO lifecycle page lists
  **ISO/IEC 15504-2:2003 as the withdrawn predecessor**, confirming the 15504 → 330xx
  lineage. [S27]
- ISO/IEC 33004:2015 "Information technology — Process assessment — Requirements for process
  reference, process assessment and maturity models", published 2015-03 (corrected version
  2017-04), ISO/IEC JTC 1/SC 7. It "sets out the requirements for process reference models,
  process assessment models, and maturity models" and explicitly relates them to "the
  activities and tasks defined in ISO/IEC 12207 and ISO/IEC 15288". [S28]

---

## 6. ISO/IEC/IEEE 12207 and ISO/IEC/IEEE 29148

### Verified facts

- **ISO/IEC/IEEE 12207:2017** "Systems and software engineering — Software life cycle
  processes" is **withdrawn** (stage 95.99); it was published 2017-11, 145 pages,
  ISO/IEC JTC 1/SC 7, and has been revised by **ISO/IEC/IEEE 12207:2026**. Its predecessor
  was ISO/IEC 12207:2008. [S29]
- **ISO/IEC/IEEE 12207:2026** is published and carries the same title. It "establishes a
  common framework for software life cycle processes … that can be applied during the
  acquisition of a software system, product, or service and during the supply, development,
  operation, maintenance, and disposal of software products and services", and "also
  provides processes that can be employed for defining, controlling, and improving software
  life cycle processes within an organization or a project". [S30]
- **ISO/IEC/IEEE 29148:2018** "Systems and software engineering — Life cycle processes —
  Requirements engineering", Edition 2, published 2018-11, 92 pages, ISO/IEC JTC 1/SC 7,
  stage 90.92 (to be revised); a DIS successor is under development. It "specifies the
  required processes implemented in the engineering activities that result in requirements
  for systems and software products (including services) throughout the life cycle",
  "provides guidelines for applying the requirements and requirements-related processes
  described in ISO/IEC/IEEE 15288 and ISO/IEC/IEEE 12207", and "specifies the required
  contents of the required information items". [S31]
- ISO 26262-8:2018 cites ISO/IEC/IEEE 29148 and ISO/IEC/IEEE 12207 in its bibliography, and
  ASPICE 4.0 references "ISO IEEE 29148" as a source of requirement quality characteristics.
  [S6][S24]

> **Paywall note.** The clauses of 29148 that enumerate requirement characteristics and
> traceability are not publicly readable; the ISO abstract does not mention traceability
> explicitly. See UNVERIFIED.

---

## 7. CMMI

### Verified facts

- CMMI is currently owned and operated by **ISACA** through the CMMI Institute
  (cmmiinstitute.com); the site states "Our Partners are selected, trained, and licensed by
  ISACA to deliver authentic services" and refers to "ISACA's CMMI models". [S32][S33]
- CMMI stands for "Capability Maturity Model Integration"; ISACA describes it as "a proven
  set of global best practices that drives business performance through building and
  benchmarking key capabilities", "Originally created for the U.S. Department of Defense to
  assess the quality and capability of their software contractors", and states CMMI models
  "have expanded beyond software engineering". [S33]
- The model is organised into "Practice Areas and Practices" across domains (Data,
  Development, Services, People, Safety, Security, Virtual), and includes "CMMI High
  Maturity practices". [S34]
- Historical provenance is independently corroborated by ISO 26262-8:2018's bibliography
  entry "[12] CMMI for Development, CMMI-DEV, Carnegie Mellon University Software
  Engineering Institute". [S6]

---

## UNVERIFIED — DO NOT USE

The following items were targeted but **could not be verified from a fetched primary
source** in this session. Do not state them in the paper without further verification.

1. **ISO 26262-8:2018 clause number for "Confidence in the use of software tools", and the
   concepts Tool Impact (TI), Tool Error Detection (TD) and Tool Confidence Level (TCL),
   including the TCL1/TCL2/TCL3 classification and the qualification-method tables.**
   Tried: ISO OBP free preview of ISO 26262-8:2018 (the table of contents is truncated after
   clause "10.2 General", so clause 11 and beyond are not shown); the ISO catalogue abstract
   (lists "confidence in the use of software tools" as a topic but gives no clause number);
   ISO 26262-1:2018 vocabulary on OBP (contains "3.158 software tool" but **no** term
   "tool confidence level" or "tool impact"). The normative text is paywalled at CHF 225.
2. **ISO 26262-8:2018 clause numbers/contents for requirements management and traceability
   beyond the clause titles.** Only the clause titles "6 Specification and management of
   safety requirements", "7 Configuration management", "10 Documentation management" are
   publicly visible; their requirements are paywalled. The word "traceability" does not
   appear in the publicly visible portion.
3. **ISO 26262-6:2018 clause 11 ("Testing of the embedded software") number.** The scope
   abstract lists "testing of the embedded software" as a topic, but the OBP TOC is
   truncated after "10.2 General", so the clause number could not be confirmed.
4. **DO-178C traceability objectives** (e.g. bidirectional trace data between high-level
   requirements, low-level requirements, source code and executable object code; Table A-2/
   A-7 objectives; DAL A–E definitions). Tried: RTCA online store product pages, which give
   only title, description, committee and issue date. Documents are paid (USD 525 / 258).
   No public normative text was reachable.
5. **DO-330 TQL-1..TQL-5 tool qualification levels and the DO-178C tool criteria 1/2/3.**
   The RTCA product description confirms the document "explains the process and objectives
   for qualifying tools" but does not name the levels. Not verifiable without purchase.
6. **ISO/IEC/IEEE 29148:2018 statements about traceability and the characteristics of a good
   requirement** (unambiguous, verifiable, complete, singular, traceable, etc.). The ISO
   abstract does not mention them. Note: an *indirect, attributed* statement is available —
   ASPICE 4.0 SWE.1.BP1 Note 1/2 says such characteristics "are defined in standards such as
   ISO IEEE 29148, ISO 26262-8:2018, or the INCOSE Guide for Writing Requirements" and gives
   examples [S24]. Use only in that attributed form.
7. **ISO/IEC/IEEE 12207 statements about traceability.** Not present in the public abstracts
   of either the 2017 or the 2026 edition.
8. **Automotive SPICE 4.1 as a released version.** intacs.info refers to "Automotive SPICE®
   4.1" [S26], but VDA QMC's own ASPICE page says "the current version 4.0" and links the
   v4.0 PDF [S23]. automotivespice.com (the official document site) returned HTTP 500 during
   this session, so no 4.1 document could be retrieved. Attribute or omit.
9. **The exact list of the 32 ASPICE processes / 3 categories / 11 groups.** The counts are
   verified from VDA QMC [S23] and the SYS/VAL/SWE groups are verified from the PAM [S24],
   but the complete enumeration of all 11 groups was not extracted in this session.
10. **CMMI current version number (e.g. "V3.0") and its release date.** cmmiinstitute.com's
    "What is CMMI?" and "Model Viewer" pages carry no version number. Not verified.
11. **IEC 62304 software safety classes A/B/C.** Not stated in the IEC webstore description;
    the normative text is paywalled (CHF 1'150).
12. **IEC 61508 SIL 1–4 definitions and the relation table between SIL and ASIL.** IEC
    61508-5's *title* ("Examples of methods for the determination of safety integrity
    levels") is verified [S14], but no SIL definition text was publicly readable.
13. **FAA/EASA regulatory recognition of DO-178C/ED-12C** (e.g. FAA AC 20-115D). Not fetched
    in this session. RTCA's own wording — "Compliance with the objectives of DO-178C is the
    primary means of obtaining approval of software used in civil aviation products" [S17] —
    and EUROCAE's "European reference standard for airborne software certification" [S20]
    are available and should be quoted as attributed statements by RTCA/EUROCAE.

---

## Sources

Non-authoritative sources are flagged. All accessed **2026-08-30**.

| Ref | Title | Publisher / venue | Year | URL | Verbatim quote |
|---|---|---|---|---|---|
| S1 | ISO 26262-1:2018 — Road vehicles — Functional safety — Part 1: Vocabulary (catalogue page) | ISO (official) | 2018 | https://www.iso.org/standard/68383.html | "Road vehicles — Functional safety — Part 1: Vocabulary … Edition 2 2018-12 … Status : Published … Publication date : 2018-12 … Stage : International Standard to be revised [90.92] … Number of pages : 33 … Technical Committee : ISO/TC 22/SC 32 … This document defines the vocabulary of terms used in the ISO 26262 series of standards." |
| S2 | ISO 26262 road vehicles functional safety (publication/package page PUB200262) | ISO (official) | n.d. | https://www.iso.org/publication/PUB200262.html | "ISO 26262-1:2018 - Road vehicles — Functional safety — Part 1: Vocabulary … ISO 26262-2:2018 - … Part 2: Management of functional safety … Part 3: Concept phase … Part 4: Product development at the system level … Part 5: Product development at the hardware level … ISO 26262-6:2018 - Road vehicles — Functional safety — Part 6: Product development at the software level … Part 7: Production, operation, service and decommissioning … ISO 26262-8:2018 - Road vehicles — Functional safety — Part 8: Supporting processes … Part 9: Automotive safety integrity level (ASIL)-oriented and safety-oriented analyses … Part 10: Guidelines on ISO 26262" |
| S3 | ISO 26262-6:2018 — Part 6: Product development at the software level (catalogue page) | ISO (official) | 2018 | https://www.iso.org/standard/68388.html | "This document specifies the requirements for product development at the software level for automotive applications, including the following: — general topics for product development at the software level; — specification of the software safety requirements; — software architectural design; — software unit design and implementation; — software unit verification; — software integration and verification; and — testing of the embedded software. It also specifies requirements associated with the use of configurable software." / "Edition : 2 … Number of pages : 57 … Technical Committee : ISO/TC 22/SC 32" |
| S4 | ISO 26262-8:2018 — Part 8: Supporting processes (catalogue page) | ISO (official) | 2018 | https://www.iso.org/standard/68390.html | "This document specifies the requirements for supporting processes, including the following: — interfaces within distributed developments; — overall management of safety requirements; — configuration management; — change management; — verification; — documentation management; — confidence in the use of software tools; — qualification of software components; — evaluation of hardware elements; — proven in use argument; — interfacing an application that is out of scope of ISO 26262; and — integration of safety-related systems not developed according to ISO 26262." / "Edition : 2 … Number of pages : 60" |
| S5 | ISO 26262-1:2018(en) — Terms and definitions (free preview, Online Browsing Platform) | ISO OBP (official) | 2018 | https://www.iso.org/obp/ui/en/#iso:std:iso:26262:-1:ed-2:v1:en | "3.6 automotive safety integrity level ASIL one of four levels to specify the item's (3.84) or element's (3.41) necessary ISO 26262 requirements and safety measures (3.141) to apply for avoiding an unreasonable risk (3.176), with D representing the most stringent and A the least stringent level. Note 1 to entry: QM (3.117) is not an ASIL." / "3.158 software tool computer program used in the development of an item (3.84) or element (3.41)" |
| S6 | ISO 26262-8:2018(en) — Foreword, Introduction, Scope, Normative references, Bibliography (free preview, OBP) | ISO OBP (official) | 2018 | https://www.iso.org/obp/ui/en/#iso:std:iso:26262:-8:ed-2:v1:en | "The ISO 26262 series of standards is the adaptation of IEC 61508 series of standards to address the sector specific needs of electrical and/or electronic (E/E) systems within road vehicles." / "b) provides an automotive-specific risk-based approach to determine integrity levels [Automotive Safety Integrity Levels (ASILs)]; c) uses ASILs to specify which of the requirements of ISO 26262 are applicable to avoid unreasonable residual risk" / "The ISO 26262 series of standards is based upon a V-model as a reference process model" / "the specific clauses are indicated in the following manner: “m-n”, where “m” represents the number of the particular part and “n” indicates the number of the clause within that part." / Bibliography: "[1] ISO 26262-11:2018, Road vehicles — Functional safety — Part 11: Guideline on application of ISO 26262 to semiconductors [2] ISO 26262-12:2018, Road vehicles — Functional safety — Part 12: Adaptation of ISO 26262 for motorcycles … [8] ISO/IEC/IEEE 29148 … [10] IEC 61508 (all parts) … [11] RTCA DO-178C, Software Considerations in Airborne Systems and Equipment Certification [12] CMMI for Development, CMMI-DEV, Carnegie Mellon University Software Engineering Institute … [17] Automotive SPICE®4 … [19] ISO/IEC/IEEE 12207 … [20] ISO/IEC 33000 (series)" / "Automotive SPICE® is an example of a suitable product available commercially." |
| S7 | (reserved — unused) | — | — | — | — |
| S8 | ISO 26262-6:2018(en) — Table of contents (free preview, OBP) | ISO OBP (official) | 2018 | https://www.iso.org/obp/ui/en/#iso:std:iso:26262:-6:ed-2:v1:en | "5 General topics for the product development at the software level … 6 Specification of software safety requirements … 7 Software architectural design … 8 Software unit design and implementation … 9 Software unit verification … 10 Software integration and verification" |
| S9 | IEC 61508:2010 CMV — Functional safety of electrical/electronic/programmable electronic safety-related systems - Parts 1 to 7 | IEC (official webstore) | 2010 | https://webstore.iec.ch/en/publication/22273 | "Functional safety of electrical/electronic/programmable electronic safety-related systems - Parts 1 to 7 … Technical committee TC 65/SC 65A System aspects … Publication type International Standard Publication date 2010-04-30 Edition 2.0 … ISBN number 9782889109852" |
| S10 | IEC 61508-1:2010 | IEC (official webstore) | 2010 | https://webstore.iec.ch/en/publication/5515 | "Functional safety of electrical/electronic/programmable electronic safety-related systems - Part 1: General requirements" / "IEC 61508-1:2010 covers those aspects to be considered when electrical/electronic/programmable electronic (E/E/PE) systems are used to carry out safety functions. A major objective of this standard is to facilitate the development of product and application sector international standards by the technical committees responsible for the product or application sector." / "It has the status of a basic safety publication according to IEC Guide 104." |
| S11 | IEC 61508-2:2010 | IEC (official webstore) | 2010 | https://webstore.iec.ch/en/publication/5516 | "Functional safety of electrical/electronic/programmable electronic safety-related systems - Part 2: Requirements for electrical/electronic/programmable electronic safety-related systems" |
| S12 | IEC 61508-3:2010 | IEC (official webstore) | 2010 | https://webstore.iec.ch/en/publication/5517 | "Functional safety of electrical/electronic/programmable electronic safety-related systems - Part 3: Software requirements" / "provides specific requirements applicable to support tools used to develop and configure a safety-related system within the scope of IEC 61508-1 and IEC 61508-2" / "provides, in conjunction with IEC 61508-1 and IEC 61508-2, requirements for support tools such as development and design tools, language translators, testing and debugging tools, configuration management tools." |
| S13 | IEC 61508-4:2010 | IEC (official webstore) | 2010 | https://webstore.iec.ch/en/publication/5518 | "Functional safety of electrical/electronic/programmable electronic safety-related systems - Part 4: Definitions and abbreviations" |
| S14 | IEC 61508-5:2010 | IEC (official webstore) | 2010 | https://webstore.iec.ch/en/publication/5519 | "Functional safety of electrical/electronic/programmable electronic safety-related systems - Part 5: Examples of methods for the determination of safety integrity levels" |
| S15 | IEC 61508-6:2010 | IEC (official webstore) | 2010 | https://webstore.iec.ch/en/publication/5520 | "Functional safety of electrical/electronic/programmable electronic safety-related systems - Part 6: Guidelines on the application of IEC 61508-2 and IEC 61508-3" |
| S16 | IEC 61508-7:2010 | IEC (official webstore) | 2010 | https://webstore.iec.ch/en/publication/5521 | "Functional safety of electrical/electronic/programmable electronic safety-related systems - Part 7: Overview of techniques and measures" |
| S17 | DO-178C product page | RTCA (official store) | 2011 | https://my.rtca.org/productdetails?id=a1B36000001IcmqEAC | "Document Title DO-178C - Software Considerations in Airborne Systems and Equipment Certification … This document provides recommendations for the production of software for airborne systems and equipment that performs its intended function with a level of confidence in safety that complies with airworthiness requirements. Compliance with the objectives of DO-178C is the primary means of obtaining approval of software used in civil aviation products. Errata has been prepared against DO-178C. … Committee SC-205 Issue Date 12/13/2011" |
| S18 | DO-330 product page | RTCA (official store) | 2011 | https://my.rtca.org/productdetails?id=a1B36000001IcflEAC | "Document Title DO-330 - Software Tool Qualification Considerations … In the context of this document a tool is a computer program or a functional part thereof, used to help develop, transform, test, analyze, produce or modify another program, its data or its documentation. … This document explains the process and objectives for qualifying tools. … It provides guidance for airborne and ground-based software. It may also be used by other domains, such as automotive, space, systems, electronic hardware, aeronautical databases and safety assessment processes. … Committee SC-205 Issue Date 12/13/2011" |
| S19 | DO-333 product page | RTCA (official store) | 2011 | https://my.rtca.org/productdetails?id=a1B36000001IcffEAC | "Document Title DO-333 - Formal Methods Supplement to DO-178C and DO-278A … This supplement identifies the additions, modifications and substitutions to DO-178C and DO-278A objectives when formal methods are used as part of a software life cycle, and the additional guidance required. … Formal methods are mathematically-based techniques for the specification, development and verification of software aspects of digital systems. … Committee SC-205 Issue Date 12/13/2011" |
| S20 | Search results for "ED-12C" (document listing) | EUROCAE (official) | n.d. | https://www.eurocae.net/?s=ED-12C | "ED-12C | Software considerations in airborne systems and equipment certification … provides the aviation community with guidance for determining, in a consistent manner and with an acceptable level of confidence, that the software aspects of airborne systems and equipment comply with airworthiness requirements." / "ED-12C aviation software standard, developed by EUROCAE, is the European reference standard for airborne software certification and is equivalent to RTCA DO-178C." / "ED-216 | Formal methods supplement to ED-12C and ED-109A … identifies the modifications and additions to ED-12C objectives, activities, explanatory text, and software life cycle data that should be addressed when formal methods are used as part of the software life cycle." / "ED-217 | Object-oriented technology supplement …" / "ED-218 | Model-based development and verification supplement to ED-12C and ED-109A" / "ED-94C | Supporting Information for ED-12C and ED-109A" |
| S21 | Search results for "ED-215" | EUROCAE (official) | n.d. | https://www.eurocae.net/?s=ED-215 | "…is to provide tool qualification guidance. Additionally, clarification material is provided in the form of Frequently Asked Questions (FAQs)." / "ED-215 Corr 1 | Software Tool Qualification Considerations Corrigendum 1" |
| S22 | IEC 62304:2006+AMD1:2015 CSV | IEC (official webstore) | 2015 | https://webstore.iec.ch/en/publication/22794 | "Medical device software - Software life cycle processes … IEC 62304:2006+A1:2015 Defines the life cycle requirements for medical device software. The set of processes, activities, and tasks described in this standard establishes a common framework for medical device software life cycle processes. Applies to the development and maintenance of medical device software when software is itself a medical device or when software is an embedded or integral part of the final medical device. This standard does not cover validation and final release of the medical device, even when the medical device consists entirely of software. This consolidated version consists of the first edition (2006) and its amendment 1 (2015). … Technical committee TC 62/SC 62A … Publication date 2015-06-26 Edition 1.1 … ISBN number 9782832227657 Pages 170" |
| S23 | Automotive SPICE® (landing page) | VDA QMC (official) | n.d. | https://vda-qmc.de/en/automotive-spice/ | "In 2005, Automotive SPICE® was published for the first time by the “Automotive Special Interest Group”. In the meantime, the current version 4.0 of this standard has been established worldwide … The name “SPICE” stands for “Systems/Software Process Improvement and Capability DEtermination”." / "Automotive SPICE® is based on the ISO/IEC 330xx series of standards and conforms to its basic requirements. Automotive SPICE® has defined its own Process Reference Model (PRM) and Process Assessment Model (PAM) based on the latest version of the ISO/IEC 330xx series." / "The capability dimension contains the six process capability levels from level 0 to level 5, according to the definitions from the ISO/IEC 330xx series." / "The process dimension describes the 32 processes of the Automotive SPICE®model, which are divided into 3 categories and 11 groups." / "7619 Automotive SPICE® Assessors … 51 Countries … 5 Available languages" |
| S24 | Automotive SPICE® Process Reference Model / Process Assessment Model, Version 4.0 (PDF, primary normative document, downloaded and read) | VDA QMC / VDA Working Group 13 (official) | 2023 | https://vda-qmc.de/wp-content/uploads/2023/12/Automotive-SPICE-PAM-v40.pdf | "Automotive SPICE® Process Reference Model Process Assessment Model Version 4.0 … Author(s): VDA Working Group 13 … Date: 2023-11-29 Status: Released" / "The Automotive SPICE process reference model and process assessment model are conformant with the ISO/IEC 33004:2015 … An ISO/IEC 33003:2015 compliant Measurement Framework is defined in section 5." / "It was developed in accordance with the requirements of ISO/IEC 33004:2015." / "If processes beyond the scope of Automotive SPICE are needed, appropriate processes from other process reference models such as ISO/IEC 12207 or ISO/IEC/IEEE 15288 may be added" / "Note: The Automotive SPICE measurement framework is an adaption of ISO/IEC 33020:2019." / "a PRM/PAM according to ISO/IEC 33004 (formerly ISO/IEC 15504-2)" / "There are six capability levels as listed in Table 14, incorporating nine process attributes: Level 0: Incomplete process — The process is not implemented or fails to achieve its process purpose. … Level 5: Innovating process" / "PA 1.1 Process performance … PA 2.1 Performance management … PA 2.2 Work product management … PA 3.1 Process definition … PA 3.2 Process deployment … PA 4.1 Quantitative analysis … PA 4.2 Quantitative control … PA 5.1 Process innovation … PA 5.2 Process innovation implementation" / "SWE.1 Software Requirements Analysis / SWE.2 Software Architectural Design / SWE.3 Software Detailed Design and Unit Construction / SWE.4 Software Unit Verification / SWE.5 Software Component Verification and Integration Verification / SWE.6 Software Verification — Table 7 — Primary life cycle processes – SWE process group" / "SWE.1 … 5) Consistency and bidirectional traceability are established between software requirements and system requirements. 6) Consistency and bidirectional traceability are established between software requirements and system architecture." / "SWE.1.BP5: Ensure consistency and establish bidirectional traceability. … Note 9: Redundant traceability is not intended. … Note 11: Bidirectional traceability supports consistency, and facilitates impact analysis of change requests, and demonstration of verification coverage. Traceability alone, e.g., the existence of links, does not necessarily mean that the information is consistent with each other." / "SWE.1.BP1 … Note 1: Characteristics of requirements are defined in standards such as ISO IEEE 29148, ISO 26262-8:2018, or the INCOSE Guide for Writing Requirements. Note 2: Examples for defined characteristics of requirements shared by technical standards are verifiability (i.e., verification criteria being inherent in the requirements text), unambiguity/comprehensibility, freedom from design and implementation, and not contradicting any other requirement)." / "Output Information Items … 13-51 Consistency Evidence" / "SWE.3 … 3) Consistency and bidirectional traceability are established between software detailed design and software architecture; and consistency and bidirectional traceability are established between source code and software detailed design" / "SWE.6 … The purpose of the Software Verification process is to ensure that the integrated software is verified to be consistent with the software requirements." |
| S25 | (reserved — unused) | — | — | — | — |
| S26 | SPICE Center | intacs (International Assessor Certification Scheme; industry body, semi-authoritative — not a standards body) | n.d. | https://intacs.info/spice-center | "SPICE is an abbreviation and stands for \"Software Process Improvement and Capability dEtermination\". … The project was very successful and finally, ISO issued the standard series ISO/IEC 15504-x … In the meantime, most parts of this standard have been replaced by the ISO 330xx series" / "Automotive SPICE® is a domain-specific adaption of the International Standard ISO/IEC 330xx (SPICE). The purpose of that PRM/PAM is the assessment of process capability." / "Aligned with Automotive SPICE® 4.1" / "Automotive SPICE® is a registered trademark of the VDA-QMC." |
| S27 | ISO/IEC 33002:2015 — Requirements for performing process assessment | ISO (official) | 2015 | https://www.iso.org/standard/54176.html | "ISO/IEC 33002:2015 defines the minimum set of requirements for performing an assessment that will ensure assessment results are objective, consistent, repeatable, and representative of the assessed processes." / "Publication date : 2015-03 … Number of pages : 16 … Technical Committee : ISO/IEC JTC 1/SC 7" / "Previously Withdrawn ISO/IEC 15504-2:2003" |
| S28 | ISO/IEC 33004:2015 — Requirements for process reference, process assessment and maturity models | ISO (official) | 2015 | https://www.iso.org/standard/54178.html | "ISO/IEC 33004:2015 sets out the requirements for process reference models, process assessment models, and maturity models." / "b) the relationship between process reference models and prescriptive/normative models of process performance, as constituted by, for example, the activities and tasks defined in ISO/IEC 12207 and ISO/IEC 15288" / "Publication date : 2015-03 Corrected version (en) : 2017-04" |
| S29 | ISO/IEC/IEEE 12207:2017 — Systems and software engineering — Software life cycle processes | ISO (official) | 2017 | https://www.iso.org/standard/63712.html | "Status : Withdrawn … Publication date : 2017-11 … Stage : Withdrawal of International Standard [95.99] … Number of pages : 145 … Technical Committee : ISO/IEC JTC 1/SC 7" / "Revised by Published ISO/IEC/IEEE 12207:2026" |
| S30 | ISO/IEC/IEEE 12207:2026 — Systems and software engineering — Software life cycle processes | ISO (official) | 2026 | https://www.iso.org/standard/90219.html | "This document establishes a common framework for software life cycle processes. Its terminology can be referenced and applied across the software industry. It contains processes, activities and tasks that can be applied during the acquisition of a software system, product, or service and during the supply, development, operation, maintenance, and disposal of software products and services. … This document also provides processes that can be employed for defining, controlling, and improving software life cycle processes within an organization or a project." |
| S31 | ISO/IEC/IEEE 29148:2018 — Systems and software engineering — Life cycle processes — Requirements engineering | ISO (official) | 2018 | https://www.iso.org/standard/72089.html | "— specifies the required processes implemented in the engineering activities that result in requirements for systems and software products (including services) throughout the life cycle; — provides guidelines for applying the requirements and requirements-related processes described in ISO/IEC/IEEE 15288 and ISO/IEC/IEEE 12207; — specifies the required information items produced through the implementation of the requirements processes; — specifies the required contents of the required information items" / "Status : Published Publication date : 2018-11 Stage : International Standard to be revised [90.92] Edition : 2 Number of pages : 92 Technical Committee : ISO/IEC JTC 1/SC 7" |
| S32 | CMMI Institute — Home | ISACA / CMMI Institute (official owner site; marketing content) | n.d. | https://cmmiinstitute.com/ | "ISACA enables organizations to elevate and benchmark performance across a range of critical business capabilities, including product development, service excellence, supplier management, and cybersecurity." / "Our Partners are selected, trained, and licensed by ISACA to deliver authentic services." |
| S33 | What is CMMI? | ISACA / CMMI Institute (official owner site; marketing content) | n.d. | https://cmmiinstitute.com/cmmi/intro | "The Capability Maturity Model Integration (CMMI)® is a proven set of global best practices that drives business performance through building and benchmarking key capabilities. Originally created for the U.S. Department of Defense to assess the quality and capability of their software contractors, CMMI models have expanded beyond software engineering to help any organization in any industry build, improve, and measure their capabilities and improve performance." / "ISACA's CMMI models help organizations understand their current level of capability and performance" |
| S34 | CMMI Model Viewer | ISACA / CMMI Institute (official owner site; marketing content) | n.d. | https://cmmiinstitute.com/products/cmmi/cmmi-model-viewer | "All CMMI domain Practice Areas and Practices (Data, Development, Services, People, Safety, Security, Virtual) … CMMI High Maturity practices, contexts and guidance … Context-specific information (agile, DevSecOps)" |

### Source authority notes

- **Official standards bodies:** S1–S16, S20–S24, S27–S31 (iso.org, iso.org/obp, webstore.iec.ch,
  eurocae.net, vda-qmc.de).
- **Official standards-adjacent publisher store:** S17–S19 (my.rtca.org — RTCA's own store;
  descriptive text authored by RTCA).
- **Industry certification body, semi-authoritative:** S26 (intacs.info). Use only for
  attributed statements about the assessor community and about model alignment.
- **Vendor/marketing pages of the model owner:** S32–S34 (cmmiinstitute.com). Use only for
  attributed statements ("ISACA describes CMMI as …"), never for absolute claims about the
  model's technical content or version.
- **No vendor blogs, consultancy pages, or Wikipedia were used.**
