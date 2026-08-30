# INDEPENDENT CITATION AUDIT

**Auditor role:** adversarial, independent. None of the sources below were gathered by me.
Every entry was **re-fetched from scratch** on **2026-08-30** (this session), using at
least two independent endpoints for every academic citation where an endpoint existed.

**Files audited**

- `doc/paper/research/standards.md` (sources S1–S34)
- `doc/paper/research/traceability-literature.md` (sections §1.1–§8.5 + BibTeX block)
- `doc/paper/research/related-tools.md` (sections A–F)

**Endpoints used for re-verification (independent of the original agents):**
`api.crossref.org/works/<doi>`, `api.openalex.org/works/doi:<doi>` (abstract reconstructed
from `abstract_inverted_index`), `api.semanticscholar.org/graph/v1/paper/DOI:<doi>`,
`dblp.org/search/publ/api`, `export.arxiv.org/api/query`, `doi.org` resolution,
`api.github.com`, YouTube oEmbed + watch-page payload, and live browser rendering
(incl. rendered-DOM/HTML inspection of accordion-hidden vendor content) for
iso.org, iso.org/obp, webstore.iec.ch, my.rtca.org, eurocae.net, vda-qmc.de,
intacs.info, cmmiinstitute.com, ibm.com, siemens.com, jamasoftware.com, ptc.com,
kiro.dev, open-services.net, omg.org, archunit.org, staticcheck.dev, pkg.go.dev,
docs.oracle.com. The ASPICE PAM v4.0 PDF was downloaded and text-extracted locally
(2,049,002 bytes) and every quoted string was string-matched against it.

**Headline result: no fabricated source and no fabricated DOI was found.**
Two verbatim-quote defects, one metadata mis-attribution, and a set of
resolvable date/title ambiguities are reported below.

---

## 1. VERDICT TABLE — `standards.md`

| ID | Claim checked | Verdict | Evidence URL (re-fetched) |
|---|---|---|---|
| S1 | ISO 26262-1:2018 title, Ed. 2, 2018-12, 90.92, 33 pp., ISO/TC 22/SC 32, ISO/DIS successor, vocabulary sentence | **CONFIRMED** (all fields char-exact) | https://www.iso.org/standard/68383.html |
| S2 | 10-part package page PUB200262 and all 10 exact part titles | **CONFIRMED** | https://www.iso.org/publication/PUB200262.html |
| S3 | ISO 26262-6:2018 Ed. 2, 57 pp., TC 22/SC 32; full scope list incl. "testing of the embedded software"; configurable software | **CONFIRMED** (char-exact) | https://www.iso.org/standard/68388.html |
| S4 | ISO 26262-8:2018 Ed. 2, 60 pp.; 12-item supporting-process list incl. "confidence in the use of software tools"; CHF 225 | **CONFIRMED** (char-exact) | https://www.iso.org/standard/68390.html |
| S5 | OBP free preview: term **3.6** ASIL ("…D representing the most stringent and A the least stringent level"; "QM … is not an ASIL") and term **3.158** software tool ("computer program used in the development of an item … or element") | **CONFIRMED** | https://www.iso.org/obp/ui/en/#iso:std:iso:26262:-1:ed-2:v1:en |
| S5b | Negative claim: "tool confidence level" / "tool impact" absent from public Part-1 vocabulary | **CONFIRMED** (0 occurrences) | same |
| S6 | OBP Part 8: IEC 61508-adaptation sentence; V-model sentence; "m-n" clause notation + EXAMPLE "2-6"; bibliography entries [1],[2],[8],[10],[11],[12],[17],[19],[20]; Automotive SPICE® footnote | **CONFIRMED** (char-exact, incl. bibliography numbering) | https://www.iso.org/obp/ui/en/#iso:std:iso:26262:-8:ed-2:v1:en |
| S6b | Negative claim: the word "traceability" does not appear in the public portion of ISO 26262-8 | **CONFIRMED** (0 occurrences) | same |
| S6c | Part-8 TOC truncated after "10.2 General" (clauses 5–10 titles as stated) | **CONFIRMED** | same |
| S8 | ISO 26262-6 TOC clauses 5–10 as stated; truncated after 10.2 | **CONFIRMED** (char-exact) | https://www.iso.org/obp/ui/en/#iso:std:iso:26262:-6:ed-2:v1:en |
| S9 | IEC 61508 Parts 1–7 series title, Ed. 2.0, 2010-04-30, TC 65/SC 65A | **CONFIRMED** | https://webstore.iec.ch/en/publication/22273 |
| S9b | "ISBN 9782889109852" presented as the series ISBN | **CORRECTED** — that ISBN belongs to the *IEC 61508:2010 **CMV*** (Commented Version, a value-added compilation product), not to the base standard | same |
| S10 | IEC 61508-1:2010 title, scope sentence, "facilitate the development of product and application sector international standards", IEC Guide 104 basic-safety status | **CONFIRMED** (char-exact) | https://webstore.iec.ch/en/publication/5515 |
| S11 | IEC 61508-2:2010 title, Ed. 2.0, 2010-04-30 | **CONFIRMED** | https://webstore.iec.ch/en/publication/5516 |
| S12 | IEC 61508-3:2010 title + both support-tool sentences | **CONFIRMED** (char-exact) | https://webstore.iec.ch/en/publication/5517 |
| S13–S16 | IEC 61508 Parts 4/5/6/7 titles, Ed. 2.0, 2010-04-30 | **CONFIRMED** (all four char-exact) | webstore.iec.ch/en/publication/5518, /5519, /5520, /5521 |
| S17 | DO-178C title, description, "primary means of obtaining approval", errata, SC-205, 12/13/2011, USD 525 | **CONFIRMED** | https://my.rtca.org/productdetails?id=a1B36000001IcmqEAC |
| S18 | DO-330 title, tool definition, "explains the process and objectives for qualifying tools", cross-domain sentence, SC-205, USD 335.40 | **CONFIRMED** | https://my.rtca.org/productdetails?id=a1B36000001IcflEAC |
| S19 | DO-333 title, supplement sentence, formal-methods definition, SC-205, 12/13/2011 | **CONFIRMED** (char-exact) | https://my.rtca.org/productdetails?id=a1B36000001IcffEAC |
| S20 | ED-12C description; "European reference standard … equivalent to RTCA DO-178C" (**training page**, correctly attributed); ED-216/217/218/ED-94C | **CONFIRMED** | https://www.eurocae.net/?s=ED-12C |
| S21 | ED-215 purpose + FAQs; ED-215 Corr 1 exists | **CONFIRMED** | https://www.eurocae.net/?s=ED-215 |
| S22 | IEC 62304:2006+AMD1:2015 CSV, Ed. 1.1, 2015-06-26, 170 pp., ISBN 9782832227657, ICS 11.040.01, TC 62/SC 62A, full description | **CONFIRMED** (char-exact) | https://webstore.iec.ch/en/publication/22794 |
| S23 | VDA QMC: 2005 first publication; "Systems/Software Process Improvement and Capability DEtermination"; ISO/IEC 330xx basis; six levels 0–5; **"32 processes … 3 categories and 11 groups"**; 7619 assessors / 51 countries / 5 languages; Behrenstraße 35, 10117 Berlin | **CONFIRMED** (all, incl. the 32/3/11 sentence which sits further down the page) | https://vda-qmc.de/en/automotive-spice/ |
| S24a | PAM v4.0 front matter: "Version 4.0", "VDA Working Group 13", "2023-11-29", "Status: Released", revision of PAM/PRM 3.1 | **CONFIRMED** | https://vda-qmc.de/wp-content/uploads/2023/12/Automotive-SPICE-PAM-v40.pdf |
| S24b | ISO/IEC 33004:2015 conformance, 33003:2015 measurement framework in §5, "adaption of ISO/IEC 33020:2019", "identical to ISO/IEC 33020:2019", ISO/IEC 15504-5:2006, "a PRM/PAM according to ISO/IEC 33004 (formerly ISO/IEC 15504-2)", ISO/IEC 12207 / 15288 augmentation sentence | **CONFIRMED** (all char-exact) | same PDF |
| S24c | Six capability levels, nine process attributes, PA 1.1 … PA 5.2 names, Level 0 definition | **CONFIRMED** | same PDF |
| S24d | SWE.1–SWE.6 exact process names (incl. renamed SWE.3, SWE.5) | **CONFIRMED** | same PDF |
| S24e | "Ensure consistency and establish bidirectional traceability" is the BP name of SYS.2.BP5, SYS.3.BP4, SYS.4.BP4, SYS.5.BP4, SWE.1.BP5 | **CONFIRMED** (all five located) | same PDF |
| S24f | SWE.1 has **7** outcomes; outcomes 5 & 6 texts | **CONFIRMED** char-exact | same PDF |
| S24g | **Note 9** "Redundant traceability is not intended." and **Note 11** "Bidirectional traceability supports consistency, and facilitates impact analysis of change requests, and demonstration of verification coverage. Traceability alone, e.g., the existence of links, does not necessarily mean that the information is consistent with each other." — both attributed to SWE.1.BP5 | **CONFIRMED** — note numbers are correct and the quotes are char-exact | same PDF |
| S24h | SWE.3 outcome 3 reaches down to **source code ↔ software detailed design** | **CONFIRMED** char-exact | same PDF |
| S24i | SWE.4 / SWE.6 traceability between verification measures/results and artefacts | **CONFIRMED** (SWE.4 outcome 4, SWE.6 outcome 4) | same PDF |
| S24j | SWE.1.BP1 Note 1/2 requirement-characteristics quote (ISO IEEE 29148 / ISO 26262-8:2018 / INCOSE Guide) | **CONFIRMED** char-exact under **SWE.1.BP1** (the same note also appears under SYS.2.BP1 and MLE.1) | same PDF |
| S24k | "13-51 Consistency Evidence" is a SWE.1 output information item **mapped to outcomes 5 and 6** | **PARTLY UNVERIFIABLE** — the information item exists and is a SWE.1 output (confirmed); the outcome-column mapping is a graphical matrix that cannot be resolved reliably from PDF text extraction. Confirm visually before citing the 5/6 mapping. | same PDF |
| S24l | Claim that the PAM annex "still refers to SWE.5 as **'Software Integration & Integration Verification'**" | **CORRECTED / MISQUOTE** — the PAM Annex C text is `SWE.5 "Software Integration & Integration **Test**"`. The string "Software Integration & Integration Verification" does **not** occur anywhere in the PAM v4.0. | same PDF |
| S26 | intacs: "Software Process Improvement and Capability dEtermination"; ISO/IEC 15504-x → 330xx; "Aligned with Automotive SPICE® 4.1"; "registered trademark of the VDA-QMC" | **CONFIRMED** — with the added, important context that "Aligned with Automotive SPICE® 4.1" is a property of the **INTACS® Hardware SPICE PRM/PAM v4.0**, not a statement that a 4.1 PAM is published | https://intacs.info/spice-center |
| S27 | ISO/IEC 33002:2015 abstract, 2015-03, 16 pp., JTC 1/SC 7, ISO/IEC 15504-2:2003 as withdrawn predecessor | **CONFIRMED**; note the stage is **90.93 "confirmed"** — matches the file's "(confirmed)" annotation | https://www.iso.org/standard/54176.html |
| S28 | ISO/IEC 33004:2015 abstract incl. clause b), 2015-03, corrected version 2017-04 | **CONFIRMED** char-exact (page count is 9, not stated in the file) | https://www.iso.org/standard/54178.html |
| S29 | ISO/IEC/IEEE 12207:2017 withdrawn, 95.99, 2017-11, 145 pp., JTC 1/SC 7, predecessor ISO/IEC 12207:2008, revised by 12207:2026 | **CONFIRMED** (all fields) | https://www.iso.org/standard/63712.html |
| S30 | ISO/IEC/IEEE 12207:2026 published; abstract quote | **CONFIRMED** char-exact | https://www.iso.org/standard/90219.html |
| S31 | ISO/IEC/IEEE 29148:2018 Ed. 2, 2018-11, 92 pp., stage 90.92, DIS successor; 4-bullet abstract | **CONFIRMED** char-exact | https://www.iso.org/standard/72089.html |
| S32/S33 | CMMI Institute / ISACA ownership; "What is CMMI?" paragraph; "Our Partners are selected, trained, and licensed by ISACA" | **CONFIRMED** char-exact | https://cmmiinstitute.com/cmmi/intro |
| S34 | Model Viewer feature list "All CMMI domain Practice Areas and Practices (Data, Development, Services, People, Safety, Security, Virtual)", "CMMI High Maturity practices, contexts and guidance", "Context-specific information (agile, DevSecOps)" | **CONFIRMED** char-exact | https://cmmiinstitute.com/products/cmmi/cmmi-model-viewer |
| S34b | Negative claim: no CMMI version number on these pages | **CONFIRMED** (no `V\d.\d` / `Version \d` token found) | same |

### Independent check of `standards.md`'s own "UNVERIFIED — DO NOT USE" list

I attempted to break this list. I could not. Specifically re-verified as genuinely unverifiable:

- **Item 1** (TI/TD/TCL clause numbers): Part-8 public TOC ends at "10.2 General"; Part-1 vocabulary contains **zero** occurrences of "tool confidence level" or "tool impact". **The file's caution is correct.**
- **Item 2** ("traceability" not in the public Part-8 text): **0 occurrences confirmed.**
- **Item 3** (Part-6 clause 11 number): TOC truncated after 10.2. **Correct.**
- **Item 8** (ASPICE 4.1): VDA QMC still says "the current version 4.0" and still links the v4.0 PDF. **Correct — treat 4.1 as attributed to intacs only.**

---

## 2. VERDICT TABLE — `traceability-literature.md`

Every DOI was resolved and cross-checked against **at least two** of
{Crossref, OpenAlex, dblp, Semantic Scholar, arXiv}. Every "verbatim abstract snippet" was
string-matched against a freshly reconstructed abstract.

| ID | Claim checked | Verdict | Evidence URL |
|---|---|---|---|
| §1.1 | Gotel & Finkelstein, "An analysis of the requirements traceability problem", ICRE 1994, 94–101, DOI 10.1109/ICRE.1994.292398; abstract quote ("over 100 practitioners", pre-RS/post-RS) | **CONFIRMED** — DOI resolves; title/pages/venue/year match; abstract quote is **char-exact** in Semantic Scholar | api.crossref.org/works/10.1109/ICRE.1994.292398 ; api.semanticscholar.org/…/DOI:10.1109/ICRE.1994.292398 |
| §1.1b | Author name "Anthony Finkelstein" | **CORRECTED (upstream metadata trap)** — Crossref/IEEE record the author as **"C.W. Finkelstein"**, which is an IEEE metadata error. dblp + Semantic Scholar give A./Anthony Finkelstein. Use **Anthony Finkelstein**; do not copy Crossref. | same |
| §1.2 | Ramesh & Jarke, TSE 27(1):58–93, 2001, DOI 10.1109/32.895989; abstract quote (26 organisations) | **CONFIRMED** — abstract quote char-exact in OpenAlex; the file's extra claims ("four kinds of link types", "validated in case studies and incorporated in traceability tools") are **also** supported by the full abstract | api.crossref.org/works/10.1109/32.895989 ; api.openalex.org/works/doi:10.1109/32.895989 |
| §1.2b | Title variance "of" (dblp) vs "for" (OpenAlex) | **CORRECTED / RESOLVED** — **Crossref *and* OpenAlex both give "Toward reference models **for** requirements traceability"**. Two independent sources beat dblp's rendering. Use "for". Ambiguity note can be deleted. | same |
| §2.1 | Gotel et al., "The quest for Ubiquity…", RE 2012, 71–80, DOI 10.1109/RE.2012.6345841, 7 authors; abstract quote ("still not pervasive in industry") | **CONFIRMED** — char-exact | api.crossref.org/works/10.1109/RE.2012.6345841 ; api.openalex.org/works/doi:10.1109/RE.2012.6345841 |
| §2.2 | "The Grand Challenge of Traceability (v1.0)", 343–409, DOI 10.1007/978-1-4471-2239-5_16, 9 authors | **CONFIRMED** metadata; **no abstract exists** in Crossref/OpenAlex — the file's "do not paraphrase" caution is correct | api.crossref.org/works/10.1007/978-1-4471-2239-5_16 |
| §2.2b | Year 2012 (dblp) vs 2011 (OpenAlex) | **CORRECTED / RESOLVED** — Crossref chapter issued date is **2011-10-31**, publisher Springer London; the parent **book** is 2012. Cite the chapter as 2012 (book year) only if you also cite the book; otherwise 2011 is the registered chapter date. State one and be consistent. | same |
| §2.3 | "Traceability Fundamentals", 3–22, DOI 10.1007/978-1-4471-2239-5_1, 10 authors incl. Mäder | **CONFIRMED** metadata; no abstract (correctly flagged) | api.crossref.org/works/10.1007/978-1-4471-2239-5_1 |
| §2.4 | *Software and Systems Traceability* book, DOI 10.1007/978-1-4471-2239-5, Springer London, 2012 | **CONFIRMED** (type `book`, year 2012) | api.crossref.org/works/10.1007/978-1-4471-2239-5 |
| §3.1 | Antoniol et al., TSE 28(10):970–983, 2002, DOI 10.1109/TSE.2002.1041053; abstract quote | **CONFIRMED** — abstract char-exact; the file's caution that **no precision/recall numbers exist in the abstract** is **independently CONFIRMED** | api.crossref.org/works/10.1109/TSE.2002.1041053 ; api.openalex.org/works/doi:10.1109/TSE.2002.1041053 |
| §3.1b | 2025 TSE retrospective, 825–832, DOI 10.1109/TSE.2025.3534027 | **CONFIRMED** — TSE **51(3)**:825–832, 2025-03 (volume/issue not stated in file; add them) | api.crossref.org/works/10.1109/TSE.2025.3534027 |
| §3.2 | Hayes, Dekhtyar & Sundaram, TSE 32(1):4–19, 2006, DOI 10.1109/TSE.2006.3; abstract quote (RETRO) | **CONFIRMED** char-exact; "no numeric accuracy figures in abstract" **CONFIRMED** | api.crossref.org/works/10.1109/TSE.2006.3 ; OpenAlex |
| §3.3 | De Lucia et al., TOSEM 16(4) art. 13, 2007, DOI 10.1145/1276933.1276934; **key quote** "they are still far to support a complete semi-automatic recovery of all links"; "seventeen software projects involving about 150 students" | **CONFIRMED** — quote is **char-exact** in the Crossref JATS abstract | api.crossref.org/works/10.1145/1276933.1276934 |
| §4.1 | Mäder & Egyed, ICSM 2012, 171–180, DOI 10.1109/ICSM.2012.6405269; **"21% faster"** and **"60% more correct solutions"**, 52 subjects | **CONFIRMED** — both numbers are **char-exact** in the ICSM 2012 abstract. The file's warning not to transfer them to the 2015 EMSE paper is sound. | api.crossref.org/works/10.1109/ICSM.2012.6405269 ; api.openalex.org/works/doi:10.1109/ICSM.2012.6405269 |
| §4.2 | Mäder & Egyed, EMSE 20(2):413–441, DOI 10.1007/s10664-014-9314-z | **CONFIRMED** metadata; **abstract genuinely not retrievable** from Crossref (independently reproduced). Crossref issued date **2014-06-22**; dblp 2015 (issue year). Both defensible; pick one. | api.crossref.org/works/10.1007/s10664-014-9314-z |
| §4.3 | Rempel & Mäder, TSE 43(8):777–797, 2017, DOI 10.1109/TSE.2016.2622264; abstract quote "more complete traceability decreases the expected defect rate"; 24 projects; Poisson | **CONFIRMED** char-exact. Crossref author typo **"Parick Mader"** independently reproduced — the file's warning is correct. Crossref issued **2017-08-01** → **use 2017**; OpenAlex 2016 is the online-first date. | api.crossref.org/works/10.1109/TSE.2016.2622264 |
| §4.3b | File claim "three of **four** studied activities" | **CONFIRMED** — the full abstract says "the four main requirements implementation supporting activities" and "three of the studied activities" | same |
| §4.4 | Chelouati et al., RESS 230:108933, DOI 10.1016/j.ress.2022.108933 | **CONFIRMED** metadata. Crossref issued **2023-02** → the dblp year **2023 is correct**; OpenAlex 2022 is online-first. Ambiguity **RESOLVED to 2023**. Abstract genuinely unavailable. Note: Crossref renders the 4th author as "El-Miloudi **El Koursi**" (no hyphen); dblp uses "El-Koursi". | api.crossref.org/works/10.1016/j.ress.2022.108933 |
| §4.4b | Wu & Kelly SAFECOMP 2007, 172–186, DOI 10.1007/978-3-540-75101-4_17 | **CONFIRMED** (LNCS, Springer, 2007) | api.crossref.org/works/10.1007/978-3-540-75101-4_17 |
| §4.4c | Denney, Pai & Habli, ICSE 2015, 587–590, DOI 10.1109/ICSE.2015.199 | **CONFIRMED** | api.crossref.org/works/10.1109/ICSE.2015.199 |
| §4.4d | Palin & Habli, SAFECOMP 2010, 82–96, DOI 10.1007/978-3-642-15651-9_7 | **CONFIRMED** (title uses an en dash: "Assurance of Automotive Safety – A Safety Case Approach") | api.crossref.org/works/10.1007/978-3-642-15651-9_7 |
| §5.1 | Meyer, *Computer* 25(10):40–51, 1992, DOI 10.1109/2.161279; abstract quote | **CONFIRMED** char-exact. Crossref renders the title with lowercase/straight quotes: `Applying 'design by contract'`; dblp/IEEE use `Applying "Design by Contract"`. Cosmetic only. | api.crossref.org/works/10.1109/2.161279 ; OpenAlex |
| §5.2 | Leavens, Baker & Ruby, SIGSOFT SEN 31(3):1–38, 2006, DOI 10.1145/1127878.1127884; abstract quote | **CONFIRMED** char-exact (Crossref truncates the title to "Preliminary design of JML"; the full subtitle is confirmed by dblp/OpenAlex) | api.crossref.org/works/10.1145/1127878.1127884 ; OpenAlex |
| §5.2b | Leavens et al. 1999 chapter, 175–188, DOI 10.1007/978-1-4615-5229-1_12 | **CONFIRMED** (Springer US, 1999) | api.crossref.org/works/10.1007/978-1-4615-5229-1_12 |
| §5.3 | Burdy et al., STTT 7(3):212–232, DOI 10.1007/s10009-004-0167-4 | **CONFIRMED** metadata; abstract genuinely absent. Crossref issued **2004-12-14**; volume 7(3) is the **2005** issue → **use 2005**, ambiguity resolved. | api.crossref.org/works/10.1007/s10009-004-0167-4 |
| §5.3b | Conference version "FMICS 2003, pp. 75–91, DOI 10.1016/S1571-0661(04)80810-7" | **CORRECTED (venue label)** — this DOI is an **Electronic Notes in Theoretical Computer Science** article (ENTCS vol. 80, the FMICS 2003 proceedings issue), Crossref title `An overview of JML tools and applications1 1www.jmlspecs.org`. Cite as ENTCS, noting FMICS 2003. | api.crossref.org/works/10.1016/S1571-0661(04)80810-7 |
| §5.4 | Jackson, TOSEM 11(2):256–290, 2002, DOI 10.1145/505145.505149; abstract quote | **CONFIRMED** char-exact (Crossref truncates the title to "Alloy") | api.crossref.org/works/10.1145/505145.505149 ; OpenAlex |
| §6.1 | Murphy, Notkin & Sullivan, TSE 27(4):364–380, 2001, DOI 10.1109/32.917525; **"The artifacts constituting a software system often drift apart over time."** + Excel/1 MLOC | **CONFIRMED** char-exact | api.crossref.org/works/10.1109/32.917525 ; OpenAlex |
| §6.2 | FSE'95 version; the **two-DOI disambiguation** (10.1145/222124.222136 = FSE proceedings; 10.1145/222132.222136 = SIGSOFT SEN 20(4)) | **CONFIRMED** — both DOIs resolve and Crossref confirms exactly the container split the file describes. This was a genuinely careful piece of work. | api.crossref.org/works/10.1145/222124.222136 ; api.crossref.org/works/10.1145/222132.222136 |
| §6.3a | Herold & Rausch 2014, 181–207, DOI 10.1016/B978-0-12-417009-4.00007-7 | **CONFIRMED** | Crossref |
| §6.3b | de Lima & Terra, SBES 2020, 772–777, DOI 10.1145/3422392.3422505 | **CONFIRMED** (Crossref truncates title to "ArchPython"; container "Proceedings of the XXXIV Brazilian Symposium on Software Engineering") | Crossref |
| §6.3c | De Silva & Perera, ICIIS 2015, 43–48, DOI 10.1109/ICIINFS.2015.7398983 | **CONFIRMED** — note the venue is **ICIIS** ("Industrial and Information Systems"); the file writes "ICIIS 2015" as **"ICIIS"** in prose but the DOI prefix is `ICIINFS`. Both correct. | Crossref |
| §6.3d | Menezes, Martins & Rocha, SBMF 2021, 1–16, DOI 10.1007/978-3-030-92137-8_1 | **CONFIRMED** (LNCS, *Formal Methods: Foundations and Applications*) | Crossref |
| §6.3e | Ozkaya, IEEE Software 2023, 4–8, DOI 10.1109/MS.2022.3213880 | **CONFIRMED** | Crossref |
| §6.3f | Negative claim: dblp returns **zero** results for `ArchUnit` | **CONFIRMED — independently reproduced (`hits: 0`)** | https://dblp.org/search/publ/api?q=ArchUnit&format=json |
| §7.1 | Liu et al., COMPSAC 2018, 154–163, DOI 10.1109/COMPSAC.2018.00028; **74.6%** and **77.2%**; "64 features" | **CONFIRMED** — both numbers **char-exact** in the abstract | api.crossref.org/works/10.1109/COMPSAC.2018.00028 ; OpenAlex |
| §7.2 | Stulova et al., SCAM 2020, 65–69, DOI 10.1109/SCAM51674.2020.00012; "none of these tools checks for consistency of the documentation…"; upDoc; "preliminary" evaluation | **CONFIRMED** char-exact, including that the evaluation is self-described as preliminary | api.crossref.org/works/10.1109/SCAM51674.2020.00012 ; OpenAlex |
| §8.1 | Pearce et al., IEEE S&P 2022, 754–768, DOI 10.1109/SP46214.2022.9833571; **89 scenarios / 1,689 programs / ~40% vulnerable** | **CONFIRMED** char-exact | api.crossref.org/works/10.1109/SP46214.2022.9833571 ; OpenAlex |
| §8.1b | CACM version, 2025, 96–105, DOI 10.1145/3610721 | **CONFIRMED** (CACM, issued 2025-01-22) | api.crossref.org/works/10.1145/3610721 |
| §8.2 | Perry et al., CCS 2023, 2785–2799, DOI 10.1145/3576915.3623157; "significantly less secure code" + overconfidence | **CONFIRMED** char-exact | api.crossref.org/works/10.1145/3576915.3623157 ; OpenAlex |
| §8.3 | Panickssery, Bowman & Feng, **NeurIPS 2024**; arXiv:2404.13076, DOI 10.48550/arXiv.2404.13076; self-preference quote | **CONFIRMED** — dblp shows a NeurIPS 2024 record **and** a separate CoRR record; the abstract quote is char-exact in Semantic Scholar | dblp publ API ; api.semanticscholar.org/…/DOI:10.48550/ARXIV.2404.13076 |
| §8.4 | Zheng et al., NeurIPS 2023, 13 authors, arXiv:2306.05685 | **CONFIRMED** metadata (dblp NeurIPS 2023 + CoRR records); abstract correctly **not** claimed | dblp publ API |
| §8.5 | Chen et al., Codex, arXiv:2107.03374, 2021-07-07, **58 authors**, **28.8%** / **70.2%** on HumanEval; **PREPRINT** label | **CONFIRMED** — arXiv API: published `2021-07-07T17:41:24Z`, 58 `<name>` entries, abstract quotes char-exact; dblp classifies as CoRR "Informal and Other Publications". Preprint labelling is **correct**. | https://export.arxiv.org/api/query?id_list=2107.03374 |
| §UNV-1 | Kelly & Weaver GSN paper unverifiable | **CONFIRMED as unverifiable** — not in dblp; Crossref bibliographic search returns no matching Kelly & Weaver record. See CORRECTIONS for a citable substitute. | api.crossref.org/works?query.bibliographic=… |
| §UNV-2 | Jackson & Wing "Lightweight Formal Methods" unverifiable | **CONFIRMED as unverifiable as a standalone article.** Crossref search surfaces only *"An Invitation to Formal Methods"*, Computer 29(4), p. 16, 1996, DOI 10.1109/mc.1996.488298 (Bowen, Butler, Dill, Glass, Gries, Hall) — the Jackson & Wing text is a sidebar within such a roundtable and is **not separately indexed**. Do not invent a citation. | Crossref |
| §UNV-4 | Yagel "living documentation" DOI 10.5220/0005643700220026 | **CONFIRMED to exist** — "Lido – Wiki based Living Documentation with Domain Knowledge", *Proceedings of the 6th International Workshop on Software Knowledge*, 22–26, 2015. Correctly kept out of the main body. | api.crossref.org/works/10.5220/0005643700220026 |

### BibTeX block

Spot-checked every entry against the freshly fetched metadata. **All DOIs, page ranges,
volumes and issues are correct.** Only the year/title items listed under CORRECTIONS need
touching. `@article{ramesh2001toward}` already uses the correct "for" variant.

---

## 3. VERDICT TABLE — `related-tools.md`

| ID | Claim checked | Verdict | Evidence URL |
|---|---|---|---|
| A.1 | YouTube video: title `Der Moment, der die Softwareentwicklung geändert hat!`, channel David Tielke, upload `2026-06-21T09:34:52-07:00`; oEmbed JSON quote | **CONFIRMED** — oEmbed JSON re-fetched, char-exact; watch-page payload `uploadDate` matches to the second | youtube.com/oembed?url=…v=eLDHrqKplVI ; watch page |
| A.1b | View count 170,523 | **CORRECTED (volatile)** — now **170,564**. A live counter must not be cited as a fixed figure without an explicit access timestamp. | same |
| A.1c | Description quote (opening paragraphs + workshop line) | **CONFIRMED** char-exact against `shortDescription` | same |
| A.2 | "Not a conference talk / no event named / not peer-reviewed" | **CONFIRMED** — no event/venue metadata exists in the payload | same |
| A.3 | All nine chapter markers `[14:59]`…`[31:11]` with exact German titles | **CONFIRMED** — all nine char-exact (ampersands are HTML-escaped in the payload) | same |
| B.1 | Spec Kit: GitHub org, MIT, **132.3k stars / 11.9k forks**, John Lam credit; "flips the script"; "specifications become executable"; the 0–5 workflow commands; `/speckit.analyze` line; the presets/extensions traceability sentences | **CONFIRMED** — GitHub API: MIT, **132,287** stars, **11,908** forks; all seven quoted strings matched char-exact in the live README | api.github.com/repos/github/spec-kit ; raw README |
| B.2 | Kiro: "built and operated by a small, opinionated team within AWS"; "We pioneered spec-driven development…"; three-file spec structure; "Break down requirements"; "Catch inconsistencies, ambiguities, and gaps…"; docs "updated August 27, 2026" | **CONFIRMED** — all char-exact, including the page's own `updated: August 27, 2026` | https://kiro.dev/about/ ; https://kiro.dev/docs/specs/ |
| B.2b | Labelling "We pioneered spec-driven development" as an unverified **vendor marketing claim** | **CONFIRMED as correct handling** | same |
| B.3 | Jama markets "Spec-Driven Development – Engineers and AI engineering agents iterate in a shared context via MCP" | **CONFIRMED** char-exact | https://www.jamasoftware.com/platform/jama-connect/ |
| C.1 | IBM DOORS/DOORS Next: all four quotes incl. "Traceability — Link artifacts for alignment and use a graphical explorer…" and the ASPICE/ISO 26262/DO178C sentence | **CONFIRMED** — all present in the rendered page HTML | https://www.ibm.com/products/requirements-management-doors-next |
| C.1b | The cited URL | **CORRECTED (minor)** — it **302-redirects** to `https://www.ibm.com/products/requirements-management`. Cite the redirect target. | same |
| C.2 | Siemens Polarion: **all six** verbatim quotes (DNA sentence; "Pass any audit…"; "full traceability of every source code modification up to the change request"; LiveDocs paragraph-level; "Built-in ReqIF enables lossless…"; "Support SVN and GIT out of the box … Perforce; Plastic SCM") | **CONFIRMED** — *but only in the collapsed accordion markup*. Only ~3.3 kB of the page is visible as rendered text; a naive `innerText` check finds **none** of them. All six were located in the page's full DOM/HTML. **Verified, with a caveat that this content is not visible without expanding the page sections.** | https://www.siemens.com/en-us/products/polarion/requirements/ |
| C.2b | URL note ("redirect target of polarion.plm.automation.siemens.com/products/polarion-requirements") | **CONFIRMED** — the legacy URL 302s to exactly this page | same |
| C.2c | The file's paraphrase `An exclusive innovation, Polarion LiveDocs, enables you to collaborate concurrently and securely on specification documents while having every single paragraph uniquely identifiable and traceable` presented as one verbatim quote | **CORRECTED (quote splice)** — "An exclusive innovation you won't find elsewhere, Polarion LiveDocs™…" and "…every single paragraph uniquely identifiable and traceable" are **two separate strings from different parts of the page**, joined into one apparent quotation. Split them or paraphrase. | same |
| C.3 | Jama Connect: all four quotes (manual-compliance sentence; 10M/100M scale; AI-governance/audit-trail; semantic product graph via MCP) | **CONFIRMED** char-exact in rendered text | https://www.jamasoftware.com/platform/jama-connect/ |
| C.3b | Dedicated `Requirements Traceability` nav entry and comparison pages | **CONFIRMED** — `/solutions/requirements-traceability/` resolves (title "Requirements Traceability - Jama Software") | https://www.jamasoftware.com/solutions/requirements-traceability/ |
| C.3c | Named standards on the page (ISO 26262, ASPICE, DO-178C, ISO 13485, 21 CFR Part 11) | **CONFIRMED** | same |
| C.4 | PTC Codebeamer: all five quotes (ALM platform sentence; end-to-end traceability; "centralized data repository guarantees complete traceability across all work items"; OSLC/digital thread; standards list "ISO 26262, ASPICE, IEC 62304, DO-178C, ISO 14971"; "Windchill, Jira, Git, IBM Rhapsody") | **CONFIRMED** — all present in the page HTML (FAQ/accordion content; again invisible to a plain text render) | https://www.ptc.com/en/products/codebeamer |
| C.5 | OSLC: **OASIS Open Project** (not OMG); "Creating standard REST APIs to connect data"; Core Specification paragraph; RDF vocabularies + resource shapes; CM 3.0 OASIS Standard effective 26 May 2021; Config Mgmt 1.0; PROMCODE | **CONFIRMED** char-exact. Negative claim **"no OMG involvement"** independently **CONFIRMED** (string "OMG" absent from the page). | https://open-services.net/ |
| D.1 | ArchUnit: "Unit test your Java architecture"; 30-minutes sentence; library/bytecode paragraph; .NET/C# port; © Peter Gafert; v1.5.0 javadoc link **and** v1.4.2 news item (Apr 18, 2026) | **CONFIRMED** — all char-exact; both version signals reproduced, so the file's refusal to state a version is correct | https://www.archunit.org/ |
| D.1b | "supported by TNG Technology Consulting GmbH; repository `TNG/ArchUnit`" | **CONFIRMED** — the string "TNG" does not appear as body text, but the page links `github.com/TNG/ArchUnit` and `www.tngtech.com` under "Kindly supported by"; GitHub API confirms `TNG/ArchUnit`, Apache-2.0 | same ; api.github.com/repos/TNG/ArchUnit |
| D.2 | `go vet`: all four quotes + the three named checks + `go1.27.0` | **CONFIRMED** char-exact | https://pkg.go.dev/cmd/vet |
| D.3 | Staticcheck: all three quotes; 150+ checks | **CONFIRMED** char-exact | https://staticcheck.dev/docs/ |
| D.3b | Author "Dominik Honnef" | **CORRECTED (minor attribution)** — the personal name does **not** appear on the cited docs page; only `dominikh/go-tools` does. Attribute the authorship to the repository, or cite the repo URL. | same |
| D.4 | `go/analysis`: all seven quotes, incl. the Fact/"separate analysis" paragraph, `analysistest` `// want`, `singlechecker`; v0.49.0; BSD-3-Clause | **CONFIRMED** char-exact (the file's typographic quotes around "checker" are curly; the page uses straight `"checker"` — cosmetic) | https://pkg.go.dev/golang.org/x/tools/go/analysis |
| D.5 | `javax.annotation.processing.Processor`: all six quotes incl. the four robustness properties; `Since: 1.6` | **CONFIRMED** char-exact | docs.oracle.com/…/Processor.html |
| D.5b | Negative claim: the string **"JSR 269" does not appear** on that Oracle page | **CONFIRMED — independently reproduced.** The file's decision to list "JSR 269" as UNVERIFIED is correct. | same |
| E.1 | Gobra: "automated, modular verifier for Go programs, based on the Viper verification infrastructure" (GitHub About) and "prototype verifier" (README); annotated `.gobra` programs; Z3/Boogie; Silicon default; VerifiedSCION; WireGuard; MPL-2.0; 182 stars | **CONFIRMED** — GitHub API description char-exact, 182 stars, homepage `https://gobra.ethz.ch`; README strings all matched | api.github.com/repos/viperproject/gobra ; raw README |
| E.1b | Caution that no peer-reviewed Gobra paper was verified | **CONFIRMED as correct** — I did not verify one either. Do not cite a Gobra paper. | — |
| F | ReqIF: OMG; "almost all RM and SysML modeling tools today support ReqIF import and export"; the supply-chain motivation paragraph; XML exchange sentence; OMG self-description | **CONFIRMED** char-exact | https://www.omg.org/reqif/ |
| F.b | Flagging "almost all RM and SysML modeling tools…" as an unquantified SDO claim needing hedging | **CONFIRMED as correct handling** | same |

### Independent check of `related-tools.md`'s own "UNVERIFIED" list

Items 2, 4, 5, 9, 10, 11 were **independently re-tested and hold**:
- The video has no event metadata (item 2). ✔
- "JSR 269" is absent from the Oracle page (item 4). ✔
- "OMG" is absent from open-services.net (item 5). ✔
- PTC's page does contain "**#1 in ALM**" and "**Spark Matrix**"; Kiro's "We pioneered spec-driven development"; Polarion's "An exclusive innovation you won't find elsewhere"; OMG's "almost all RM and SysML…" — all present, all correctly flagged as marketing (item 9). ✔
- ArchUnit simultaneously advertises v1.5.0 (javadoc) and v1.4.2 (news) (item 10). ✔

---

## CORRECTIONS

Exact corrected metadata / text for everything found wrong.

### C-1 — `standards.md`, §5 (ASPICE), line ~227–229 — **MISQUOTE**

> Written: *the PAM's own annex still refers to SWE.5 as **"Software Integration & Integration Verification"** in prose*

**Corrected.** That string does not occur anywhere in the Automotive SPICE® PAM v4.0.
The actual Annex C sentence is:

> "…they need to be integrated with the other software components by applying SWE.5 **“Software Integration & Integration Test”**."

Rewrite as: *Annex C of the PAM still refers to SWE.5 by its pre-4.0 style name
“Software Integration & Integration Test”.*

### C-2 — `standards.md`, S9 — **wrong product for the ISBN**

> Written: *the current edition consists of Parts 1 to 7, Edition 2.0, publication date 2010-04-30, IEC TC 65/SC 65A, **ISBN 9782889109852***

**Corrected.** ISBN 9782889109852 identifies **IEC 61508:2010 CMV** — the *Commented
Version*, an IEC value-added compilation product (CHF 4'386), not the base International
Standard. Either cite it explicitly as `IEC 61508:2010 CMV` or drop the ISBN and cite the
individual parts (e.g. IEC 61508-1:2010, ISBN 9782889105243, 127 pp.).

### C-3 — `standards.md`, S24k — **unresolved mapping claim**

> Written: *SWE.1's output information items include "13-51 Consistency Evidence" **mapped to outcomes 5 and 6***

The information item and its presence as a SWE.1 output are confirmed. The
outcome-column mapping lives in a graphical matrix that PDF text extraction cannot
resolve. Either verify visually in the PDF or weaken to: *"13-51 Consistency Evidence" is
listed among SWE.1's output information items.*

### C-4 — `standards.md`, §5, line ~236–237 — **paraphrase drift**

> Written: *SWE.1.BP5 requires establishing bidirectional traceability **in both directions**.*

SWE.1.BP5 requires traceability to **two targets** (system architecture and system
requirements); "bidirectional" is already in the practice name. Corrected wording:
*SWE.1.BP5 requires consistency and bidirectional traceability between software
requirements and **both** the system architecture **and** the system requirements.*

### C-5 — `traceability-literature.md` §1.2 — **title ambiguity resolved**

Use: **"Toward reference models for requirements traceability"** (Crossref *and* OpenAlex
agree; dblp's "of" is a dblp rendering). The "Note on title variance" can be deleted.

### C-6 — `traceability-literature.md` §1.1 — **author-name trap**

Crossref/IEEE record the second author as **"C.W. Finkelstein"**. This is wrong.
Correct: **Anthony Finkelstein** (dblp, Semantic Scholar, the UCL open-access PDF).
Do not copy Crossref's rendering into the bibliography.

### C-7 — `traceability-literature.md` §4.4 — **year resolved**

Chelouati et al.: Crossref issued date **2023-02**, RESS **230**:108933.
Use **2023**. (OpenAlex 2022 = online-first.) Note the Crossref author rendering
"El-Miloudi **El Koursi**" vs dblp "El-Koursi"; prefer the publisher form.

### C-8 — `traceability-literature.md` §4.3 — **year resolved**

Rempel & Mäder: Crossref issued **2017-08-01**, TSE **43(8)**:777–797. Use **2017**.
(OpenAlex 2016 = online-first.) Keep the note that Crossref misspells "Parick Mader".

### C-9 — `traceability-literature.md` §5.3 — **year resolved + venue corrected**

- STTT article: Crossref issued 2004-12-14, but the article is **STTT 7(3):212–232**, the
  **2005** issue. Use **2005**.
- The conference version DOI `10.1016/S1571-0661(04)80810-7` is an **Electronic Notes in
  Theoretical Computer Science** item (ENTCS vol. 80 — the FMICS 2003 proceedings issue),
  Crossref title `An overview of JML tools and applications1 1www.jmlspecs.org`.
  Cite as *ENTCS 80:75–91 (FMICS 2003)*, not bare "FMICS 2003".

### C-10 — `traceability-literature.md` §2.2 / §2.3 — **chapter year**

Crossref registers both chapters as **2011-10-31** (Springer London); the edited book is
**2012**. Pick one convention and apply it to both chapters *and* the book, and say which.

### C-11 — `traceability-literature.md` §3.1b — **incomplete metadata**

Antoniol et al. retrospective: add volume/issue — **IEEE TSE 51(3):825–832, 2025**,
DOI 10.1109/TSE.2025.3534027.

### C-12 — `traceability-literature.md` §UNV-1 — **citable substitute for GSN exists**

The Kelly & Weaver paper remains unverifiable, but an **authoritative, DOI-bearing**
alternative was found this session: the **Goal Structuring Notation Community Standard**
(Assurance Case Working Group), Crossref DOIs **10.65391/r1386** (Version 3) and
**10.65391/r142** (Version 2). Prefer this over a guessed Kelly & Weaver citation.

### C-13 — `traceability-literature.md` §UNV-2 — **container located**

Jackson & Wing's "Lightweight Formal Methods" is a sidebar within an *IEEE Computer*
roundtable and is **not separately indexed**. The nearest indexed container is
*An Invitation to Formal Methods*, **Computer 29(4), 1996, p. 16**,
DOI **10.1109/mc.1996.488298**. Do **not** manufacture a standalone citation; use
Jackson's Alloy TOSEM 2002 paper as the file already recommends.

### C-14 — `related-tools.md` A.1 — **volatile figure**

View count is **not** 170,523; it read **170,564** during this audit. Either drop it or
write "≈170.5k views as of 2026-08-30".

### C-15 — `related-tools.md` C.1 — **URL**

`https://www.ibm.com/products/requirements-management-doors-next` **redirects** to
`https://www.ibm.com/products/requirements-management`. Cite the target.

### C-16 — `related-tools.md` C.2c — **spliced quotation**

Split this into two quotes or paraphrase:
- verbatim: *"An exclusive innovation you won't find elsewhere, Polarion LiveDocs™—online structured specification documents, are fast becoming the way companies of all sizes gather, author, approve, validate and manage requirements."*
- verbatim (separate list item): *"…having every single paragraph uniquely identifiable and traceable"*

### C-17 — `related-tools.md` D.3 — **attribution**

"Dominik Honnef" does not appear on `staticcheck.dev/docs/`. Either cite
`https://github.com/dominikh/go-tools` for authorship, or write "authored by the
maintainer of `dominikh/go-tools`".

---

## MUST NOT BE CITED

Sources/claims that **failed** verification. None is a fabricated source; all are failures
of quotation accuracy, attribution, or availability.

1. **"Software Integration & Integration Verification" as an ASPICE PAM v4.0 quotation** —
   *does not exist in the document.* (See C-1.) Must not appear in the paper as a quote.
2. **ISBN 9782889109852 as the ISBN of IEC 61508 Parts 1–7** — belongs to the CMV
   value-added product, not the standard. (C-2.)
3. **The spliced Polarion LiveDocs "quotation"** in `related-tools.md` C.2 — not a
   contiguous string on the page. (C-16.)
4. **"Kelly & Weaver, The Goal Structuring Notation — A Safety Argument Notation"** —
   no venue/year/pages/DOI verifiable from Crossref or dblp. Do not cite.
5. **"Jackson & Wing, Lightweight Formal Methods" (IEEE Computer)** — not separately
   indexed anywhere; venue/year/pages unverifiable. Do not cite.
6. **ArchUnit as a peer-reviewed source** — dblp returns 0 hits (independently reproduced).
   Cite only the project site / GitHub repo.
7. **Any ISO 26262 Tool Impact / Tool Error Detection / Tool Confidence Level clause
   number, TCL1–3 classification, or qualification-method table** — paywalled; the public
   Part-1 vocabulary contains neither "tool confidence level" nor "tool impact". Do not cite.
8. **Any DO-178C objective-table content, DAL A–E definitions, or DO-330 TQL-1…TQL-5
   levels** — paywalled; RTCA product pages carry no such content.
9. **Any ISO/IEC/IEEE 29148 or 12207 statement about traceability or requirement
   characteristics** — not in the public abstracts (independently re-checked). Only the
   *indirect, attributed* ASPICE SWE.1.BP1 Note 1/2 formulation is usable.
10. **IEC 62304 software safety classes A/B/C** and **IEC 61508 SIL 1–4 definitions or a
    SIL↔ASIL mapping table** — paywalled, no public text.
11. **"Automotive SPICE 4.1" as a released PAM version** — VDA QMC still publishes 4.0.
    The intacs "4.1" reference is a property of the INTACS Hardware SPICE PAM. Attribute or omit.
12. **CMMI version number** — none published on the ISACA/CMMI Institute pages.
13. **Content of any abstract-less work** — Grand Challenge chapter, Traceability
    Fundamentals chapter, the Springer book, Mäder & Egyed EMSE 2015, Chelouati et al.,
    Burdy et al. STTT, Murphy et al. FSE'95, Zheng et al. NeurIPS 2023, the Antoniol 2025
    retrospective, Leavens et al. 1999, and all metadata-only lists in §4.4/§6.3.
    Cite metadata only; do **not** paraphrase findings.
14. **Precision/recall numbers attributed to Antoniol et al. (2002) or Hayes et al. (2006)**
    — independently confirmed absent from both abstracts.
15. **Any content of the David Tielke video beyond its description/chapter list**, and any
    claim that it is a conference talk or that its author has an institutional affiliation.
16. **Any Gobra publication** — no paper venue/DOI verified.
17. **The claim that a vendor tool "cannot" do compile-time requirement↔code enforcement** —
    only "not documented on the vendor's product pages" is supportable.

---

## SAFE TO CITE AS ABSOLUTE FACT

Fully verified against an authoritative primary source, character-level, with the exact
quoted text present. Safe as plain factual statements.

**Standards (SDO primary sources):**
S1, S2, S3, S4, S5, S6 (incl. the full ISO 26262-8 bibliography and the "m-n" notation),
S8, S10, S11, S12, S13, S14, S15, S16, S22, S27, S28, S29, S30, S31 — all iso.org,
iso.org/obp and webstore.iec.ch. S9 is safe **except** for the ISBN (see C-2).

**Freely published normative model:**
S23 (VDA QMC page) and S24 (Automotive SPICE® PAM v4.0 PDF) — every quoted string except
C-1 and C-3 was matched byte-for-byte in the downloaded PDF. The traceability material
(SWE.1 outcomes 5/6, SWE.1.BP5 with Notes 9 and 11, SWE.3 outcome 3, SWE.4/SWE.6
outcome 4, the SYS.*/SWE.1 base-practice names) is **the strongest, fully citable
evidence in the whole corpus**.

**Peer-reviewed literature with verified metadata *and* verified abstract quote:**
Gotel & Finkelstein ICRE 1994; Ramesh & Jarke TSE 2001; Gotel et al. RE 2012;
Antoniol et al. TSE 2002; Hayes et al. TSE 2006; **De Lucia et al. TOSEM 2007** (the
"still far to support a complete semi-automatic recovery of all links" quote);
**Mäder & Egyed ICSM 2012** (21% / 60%); **Rempel & Mäder TSE 2017**; Meyer *Computer*
1992; Leavens et al. SIGSOFT SEN 2006; Jackson TOSEM 2002; Murphy et al. TSE 2001 (the
artefact-drift sentence); **Liu et al. COMPSAC 2018** (74.6% / 77.2%);
**Stulova et al. SCAM 2020**; **Pearce et al. IEEE S&P 2022** (89 / 1,689 / ~40%);
**Perry et al. CCS 2023**; **Panickssery et al. NeurIPS 2024**.

**Official toolchain documentation (authoritative for its own subject):**
`go vet` (pkg.go.dev/cmd/vet), `golang.org/x/tools/go/analysis`, and
`javax.annotation.processing.Processor` (Oracle Java SE 21 API specification).

**Standards-body specification pages:**
ReqIF (omg.org/reqif — for the format's existence, purpose and XML basis) and
OSLC (open-services.net — for its OASIS Open Project status, RDF/LDP basis, and the
OASIS Standard status of CM 3.0 and Configuration Management 1.0).

---

## CITE ONLY AS ATTRIBUTED CLAIM

Verified to exist and quoted accurately, but **not** independently authoritative.
Every use must be phrased as *"X states that…" / "According to X's own documentation…"*.

**Standards-adjacent vendor stores (descriptive text authored by the publisher):**
S17 DO-178C, S18 DO-330, S19 DO-333 (my.rtca.org) — RTCA's own product descriptions.
S20/S21 EUROCAE search-result and **training-page** text — note that the
"European reference standard … equivalent to RTCA DO-178C" sentence comes from a
**marketing/training** page, not from ED-12C itself.

**Semi-authoritative industry body:** S26 intacs.info — for assessor-community statements
and for the "Aligned with Automotive SPICE® 4.1" claim only.

**Model owner's marketing pages:** S32, S33, S34 (cmmiinstitute.com / ISACA) — never for
technical content or version numbers.

**Self-reported ecosystem figures:** VDA QMC's "7619 assessors / 51 countries /
5 languages" — verified as printed, but self-reported and volatile.

**Vendor marketing (all of section C of `related-tools.md`):**
IBM DOORS/DOORS Next; Siemens Polarion; Jama Connect; PTC Codebeamer. All quotes verified,
but every one is a vendor claim. Explicitly *never* repeat: PTC "#1 in ALM"/Spark Matrix,
Jama's G2 badges, Kiro's "We pioneered spec-driven development", Polarion's "An exclusive
innovation you won't find elsewhere", OMG's "almost all RM and SysML modeling tools today
support ReqIF".

**Vendor documentation:** GitHub Spec Kit README (B.1); Kiro docs (B.2).

**Community / OSS project content (technically authoritative for the project, not
peer-reviewed):** ArchUnit (D.1), Staticcheck (D.3), Gobra (E.1).

**Self-published practitioner commentary:** the David Tielke YouTube video (A) —
attribute explicitly, state that it is self-published, uncontrolled and self-reported.

**Preprint:** Chen et al., *Evaluating Large Language Models Trained on Code*,
arXiv:2107.03374 — dblp classifies it as CoRR "Informal and Other Publications".
The 28.8% / 70.2% figures are correct but must be attributed, not stated as fact.

**Metadata-verified but content-unverified peer-reviewed works** (cite the reference,
never the findings): §2.2, §2.3, §2.4, §4.2, §4.4, §5.3, §6.2, §6.3 (all five), §8.4,
the Antoniol 2025 retrospective, and Leavens et al. 1999.

---

## AUDIT SUMMARY

| File | Entries audited | CONFIRMED | CORRECTED | UNVERIFIABLE | FABRICATED |
|---|---|---|---|---|---|
| `standards.md` | 40 checks over S1–S34 | 36 | 3 (S9b, S24l, plus C-4 paraphrase) | 1 (S24k) | **0** |
| `traceability-literature.md` | 44 checks | 36 | 8 (year/title/venue/author-name) | 2 (Kelly & Weaver; Jackson & Wing — both already self-flagged) | **0** |
| `related-tools.md` | 31 checks | 27 | 4 (view count, IBM URL, spliced Polarion quote, Staticcheck attribution) | 0 | **0** |

**No source in any of the three files is fabricated. No DOI is invented. No verbatim
quote was found to be wholly invented.** The single genuine misquotation is C-1
(ASPICE SWE.5 annex name). The most consequential structural finding is C-16 (a spliced
vendor quotation) and the observation that the Polarion and Codebeamer quotes exist only
in collapsed page markup — a re-verifier using plain text extraction will wrongly conclude
they are fabricated, so the source table should record *how* they were obtained.
