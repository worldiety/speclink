# Konzept: Annotationsgetriebene Spezifikation

**Status:** Diskussionsentwurf — nicht beschlossen, kein Umsetzungsauftrag
**Zweck:** Grundlage für die Diskussion mit Kollegen und LLMs
**Geltungsbereich:** projektübergreifend; werp/ERP dient nur als Belegfall (Anhang A)

---

## 1. Zusammenfassung

Klassische Spezifikations-Pipelines trennen Anforderung, Modell und Implementierung in
eigene Artefakte und verbinden sie durch Transformationen. Jede Trennung erzeugt
Synchronisationsaufwand, jede Transformation erzeugt Verlust- oder Driftpotential.

Der hier beschriebene Ansatz entfernt die **Distanz**, nicht das Modell: Das
Spezifikationsmodell wandert als **annotierte Subsprache in Kommentarform** direkt in
den Implementierungsquelltext. Ein dedizierter **Annotations-Compiler** bindet diese
Annotationen an die Konstrukte der Wirtssprache, löst sie gegen einen strukturierten
**Anforderungsbaum** auf, prüft Architektur- und Vollständigkeitsinvarianten und
erzeugt sämtliche Dokumentation aus dieser einen Quelle.

Der Quelltext wird dabei überwiegend nicht mehr von Hand geschrieben, sondern von
LLMs — der Compiler wirkt als maschinell prüfbare Zielfunktion, gegen die das LLM
iteriert, bis der Lösungsraum die Architekturvorgaben erfüllt.

**Kernaussage in einem Satz:** Ein separates Spezifikationsmodell ist entbehrlich,
sobald der Implementierungsquelltext selbst spezifikationstragend und maschinell
prüfbar ist.

---

## 2. Problemstellung

### 2.1 Die übliche Kette

```
Rohquelle → Anforderungsdokument → Spezifikationsmodell → Codegenerat → Handarbeit
```

Typische Eigenschaften dieser Kette:

| Eigenschaft | Wirkung |
|---|---|
| Jede Stufe ist ein eigenes Artefakt | eigener Lebenszyklus, eigene Reviews, eigene Diffs |
| Manuelle Verdichtungsschritte | Übersetzungsrisiko, meist ungeprüft |
| Generatoren decken nur Strukturcode ab | Fachlogik bleibt Handarbeit, also außerhalb der Kette |
| Generator braucht eigene Konfiguration | wird faktisch zur zweiten Quelle der Wahrheit |
| Modell läuft der Implementierung voraus | unverbrauchtes Inventar ohne Rückkopplung |

### 2.2 Drei wiederkehrende Fehlermuster

**M1 — Mehrfachpflege.** Dieselbe Tatsache (ein Ereignistyp, eine Zuständigkeit, eine
Berechtigung) steht an drei bis vier Stellen. Eine Änderung erfordert alle. Wird eine
vergessen, driftet das System still.

**M2 — Ungeprüfte Außenkante.** Innerhalb des Modells sind Referenzen oft typisiert
und geprüft. Der Übergang zur Anforderungsquelle ist meist ein Freitextverweis
(`"vertraege.md §8.1"`) — ausgerechnet der teuerste manuelle Schritt ist der einzige
ohne Verifikation.

**M3 — Einseitige Abdeckungsmessung.** Geprüft wird üblicherweise „ist jede
modellierte Anforderung implementiert". Nicht geprüft wird „ist jede Aussage der
Quelle modelliert". Eine vergessene Anforderung ist damit systematisch unsichtbar —
sie taucht nirgends auf, also fehlt sie auch keinem Test.

### 2.3 Die eigentliche Kostenstelle

Nicht die *Anzahl* der Transformationen ist teuer. Automatisierte, deterministische
Transformationen sind praktisch kostenlos. Teuer sind:

1. die **manuellen** Verdichtungsschritte, und
2. die **Synchronisationspflicht** zwischen dauerhaft getrennten Artefakten.

Ein Ansatz, der die Artefaktanzahl reduziert, ohne die manuellen Schritte zu
reduzieren, verbessert wenig. Ein Ansatz, der die Artefakte zusammenlegt, greift an
der richtigen Stelle an.

---

## 3. Kernthese

> Das Spezifikationsmodell ist **irreduzibel** — es ist die Abbildung von Anforderung
> auf Implementierung und muss irgendwo existieren.
> Reduzibel ist ausschließlich der **Transport**: die Tatsache, dass dieses Modell in
> einem eigenen Artefakt lebt, das mit dem Code synchron gehalten werden muss.

Daraus folgt die Entwurfsentscheidung: Das Modell wird nicht abgeschafft, sondern
**ko-lokalisiert** — es steht in derselben Datei, unmittelbar am beschriebenen
Konstrukt. Die Distanz wird null, damit auch das Synchronisationsproblem.

Zweite These, für den LLM-Kontext:

> Wenn der Quelltext überwiegend maschinell erzeugt wird, ist er **Artefakt, nicht
> Quelle**. Die menschlich verantwortete Substanz sind dann Anforderungsbaum und
> Annotationen. Der Code ist deren Ausprägung.

---

## 4. Zielarchitektur

```
┌─────────────────────────────────────────────────────────────┐
│  ANFORDERUNGSBAUM                                           │
│  strukturierte Ablage: ID, Metadaten, Text, Anhänge         │
│  fachlich verantwortet, technologieneutral                  │
└───────────────────────────┬─────────────────────────────────┘
                            │ beidseitig aufgelöst und geprüft
                            ▼
┌─────────────────────────────────────────────────────────────┐
│  QUELLTEXT + ANNOTATIONEN         ◄── SINGLE SOURCE OF TRUTH│
│  Implementierung und Spezifikation in derselben Datei       │
└───────────────────────────┬─────────────────────────────────┘
                            │
                ┌───────────┴───────────┐
                ▼                       ▼
┌──────────────────────────┐  ┌──────────────────────────────┐
│ ANNOTATIONS-COMPILER     │  │ BACKENDS (rein ableitend)    │
│ Bindung · Auflösung ·    │  │ Fachdoku · Wissensgraph ·    │
│ Invarianten · Coverage · │  │ Hilfesystem · Assistent ·    │
│ Diagnostik               │  │ Testgerüste                  │
└───────────┬──────────────┘  └──────────────────────────────┘
            │ strukturierte Diagnostik
            ▼
┌─────────────────────────────────────────────────────────────┐
│  LLM-LOOP: mutiert Quelltext + Annotationen,                │
│  bis Wirtssprachen-Compiler, Annotations-Compiler und Tests │
│  fehlerfrei sind                                            │
└─────────────────────────────────────────────────────────────┘
```

Es gibt **keinen Codegenerator** mehr im klassischen Sinn. An seine Stelle tritt ein
*Verifier*. Die erzeugende Rolle übernimmt das LLM (siehe Abschnitt 9.1 zur ehrlichen
Einordnung dieses Tauschs).

---

## 5. Bestandteile

### 5.1 Anforderungsbaum

Strukturierte Ablage der originären Fachanforderungen. Ein Knoten je Anforderung.

**Mindestbestandteile je Knoten**

| Bestandteil | Zweck |
|---|---|
| stabile ID | Referenzziel; **pfadunabhängig**, damit Umsortieren keine Referenz bricht |
| Kurzfassung | Listen, Übersichten, Tabellen |
| Volltext | normative Formulierung |
| Metadaten | Art (funktional / nicht-funktional / Randbedingung / Entscheidung), Disziplin, Status, Herleitung |
| Anhänge | Screenshots, PDF, Prozessdiagramme, Originalquellen |

**Entwurfsanforderungen**

- Beziehungen zwischen Anforderungen bilden einen **gerichteten azyklischen Graphen**
  (Herleitung, Verfeinerung, Ersetzung), nicht nur den Verzeichnisbaum. Der Baum ist
  Ablageordnung, nicht Beziehungsmodell — Metadaten müssen den DAG tragen können.
- Reiner Fließtext genügt nicht. Ohne strukturierten Metadatenteil entsteht der
  ungeprüfte Freitextverweis (M2) nur an anderer Stelle neu.
- Der Baum trägt zusätzlich alles, was **keinen Ort im Code hat**:
  Negativentscheidungen („X wird bewusst nicht gebaut"), kontextübergreifende
  Abläufe, noch nicht implementierte Anforderungen, Begründungen (Rationale).

### 5.2 Annotationssprache

Eine **geschlossene, deklarative** Subsprache, eingebettet in Kommentare der
Wirtssprache.

**Warum Kommentare.** Wirtssprachen ohne Metadaten-/Annotationssystem (z. B. Go)
bieten keinen anderen Ort, der beliebige Konstrukte adressieren kann. Kommentare sind
für die Wirtssprache bedeutungslos, damit rückwirkungsfrei, und beliebig platzierbar.

**Was annotiert wird** (Konstruktklassen, projektspezifisch zu füllen)

- Architekturrollen: welches Konstrukt ist Aggregat, Ereignis, Anwendungsfall,
  Projektion, Ansicht, Rolle, Berechtigung, Katalog
- Beziehungen: erzeugt, reagiert auf, faltet, besitzt, benötigt, erfüllt
- Erfüllungsbeziehung: welche Anforderung deckt dieses Konstrukt ab
- Invarianten und Vorbedingungen, soweit maschinell prüfbar formulierbar
- Fachtexte: Definition, Glossarbegriff, Handlungsanweisung für Hilfe/Assistent

**Harte Entwurfsregeln**

1. **Referenzieren statt duplizieren.** Annotationen verweisen auf Bezeichner der
   Wirtssprache, nicht auf Zeichenketten, die deren Namen wiederholen. Der Compiler
   löst sie über die Typinformation auf. Damit prüft die Wirtssprache mit, was sie
   prüfen kann, und die Subsprache nur den Rest.
2. **Geschlossen und deklarativ.** Keine Ausdrücke, keine Turing-Vollständigkeit,
   kein Freitext-Escape für Strukturaussagen. Jede Ausweichmöglichkeit wird zum
   bevorzugten Versteck eines LLMs, das die Zielfunktion nicht anders erfüllen kann.
3. **Lokalitätsprinzip.** Eine Annotation beschreibt ausschließlich das Konstrukt,
   an dem sie steht. Übergreifende Aussagen gehören in den Anforderungsbaum, nicht in
   eine willkürlich gewählte Codestelle.
4. **Genau ein Ort je Tatsache.** Ist eine Aussage aus Typen oder Struktur ableitbar,
   wird sie nicht annotiert. Annotation ist für das, was der Code nicht selbst sagt.

### 5.3 Annotations-Compiler

Kein Linter im Sinne von Stilprüfung, sondern ein Übersetzer mit eigener Front- und
Backend-Struktur.

| Phase | Aufgabe |
|---|---|
| Lexer/Parser | Grammatik der Subsprache, eigener AST, präzise Fehlerpositionen |
| Bindung | Zuordnung Annotation → Konstrukt der Wirtssprache über deren Typinformation |
| Auflösung | Anforderungsreferenzen gegen den Anforderungsbaum; Querverweise untereinander |
| Semantikprüfung | Architekturinvarianten (siehe 5.4) |
| Abdeckungsanalyse | beidseitig (siehe 5.4) |
| Diagnostik | menschenlesbar **und** strukturiert maschinenlesbar |
| Backends | Dokumentation, Wissensgraph, Hilfesystem, ggf. Testgerüste |

**Reihenfolge im Build.** Die Bindung setzt übersetzbaren Quelltext voraus. Der Ablauf
ist zwingend gestuft: Wirtssprachen-Compiler → Annotations-Compiler → Tests. Bei
gebrochenem Quelltext gibt es kein Annotationsfeedback; der Loop-Runner muss das
wissen und entsprechend priorisieren.

**Diagnostik als Prompt.** Die Fehlerausgabe wird überwiegend von einem LLM
konsumiert. Sie muss deshalb **präskriptiv** sein („ergänze X, weil Y") statt
deskriptiv („X fehlt"). Die Qualität der Meldungen bestimmt die
Konvergenzgeschwindigkeit des Loops und damit unmittelbar die Praxistauglichkeit des
gesamten Ansatzes. Dieser Punkt ist kein Feinschliff, sondern Kernentwurf.

### 5.4 Prüfklassen

**K1 — Strukturkonformität.** Jedes Konstrukt einer erkannten Architekturklasse trägt
die vorgeschriebene Annotation. Jede Annotation ist auflösbar. Keine Waisen.

**K2 — Beziehungsintegrität.** Alle Verweise (auf Anforderungen, auf andere
Konstrukte) zeigen auf Existierendes. Keine Zyklen, wo keine erlaubt sind.

**K3 — Beidseitige Abdeckung.** Direkte Antwort auf M3:
- *vorwärts:* jedes annotierte Konstrukt ist einer Anforderung zugeordnet
- *rückwärts:* jede normative Anforderung ist von mindestens einem Konstrukt
  referenziert — **oder** trägt eine explizite Markierung (geplant, außerhalb des
  Geltungsbereichs, rein informativ)

Die Rückwärtsrichtung ist der eigentliche Gewinn: sie verwandelt „ich hoffe, es ist
alles umgesetzt" in eine Zahl und macht handgepflegte Statustabellen überflüssig.

**K4 — Architekturmuster.** Projektspezifische Regeln („jede lesende Operation ist
berechtigungspflichtig", „Wirkung über Konsistenzgrenzen hinweg nur als Reaktion").

> **Kostenwarnung.** K1–K3 sind bei deklarierter Struktur billig — im Wesentlichen
> Auflösung und Mengenvergleich. K4 kann Datenflussanalyse über den Quelltext
> erfordern und ist um Größenordnungen teurer und fehleranfälliger. **Je mehr
> deklariert statt inferiert wird, desto kleiner und zuverlässiger das Werkzeug.**
> Empfehlung: mit K1–K3 beginnen, K4-Regeln einzeln nachrüsten, wenn ein konkret
> aufgetretener Fehler sie rechtfertigt.

### 5.5 Backends

Alle Ausgaben sind **rein ableitend** und werden nie von Hand nachbearbeitet:
Fachspezifikation, Wissensgraph (maschinenlesbar für Assistenzsysteme),
kontextsensitives Hilfesystem, Rückverfolgbarkeitsmatrix, Lückenbericht.

Nebeneffekt: der Lückenbericht ersetzt handgeführte Statusübersichten. Handgeführte
Statustabellen lügen erfahrungsgemäß ab der dritten Woche.

### 5.6 LLM-Loop

```
Auftrag (Anforderungsknoten)
   → LLM erzeugt/ändert Quelltext + Annotationen
   → Wirtssprachen-Compiler
   → Annotations-Compiler (strukturierte Diagnostik)
   → Tests
   → bei Fehlern: Diagnostik zurück ins LLM, wiederholen
   → bei Erfolg: menschliches Review der *fachlichen* Korrektheit
```

Der Compiler schränkt den Lösungsraum so stark ein, dass das Ergebnis
architekturkonform ist. Er kann **nicht** prüfen, ob es fachlich richtig ist
(Abschnitt 7).

---

## 6. Designprinzipien

| # | Prinzip | Begründung |
|---|---|---|
| P1 | Ko-Lokalisierung | Distanz erzeugt Synchronisationspflicht; Distanz null, Pflicht null |
| P2 | Genau eine Quelle je Tatsache | Mehrfachpflege ist die Hauptfehlerquelle (M1) |
| P3 | Referenz statt Wiederholung | Wirtssprache prüft mit, Fehlerklassen halbieren sich |
| P4 | Geschlossene Sprache | Escape-Hatches werden vom LLM ausgenutzt |
| P5 | Deklaration vor Inferenz | bestimmt Größe und Zuverlässigkeit des Werkzeugs |
| P6 | Beidseitige Abdeckung | einseitige Messung kann Vergessenes prinzipiell nicht finden |
| P7 | Diagnostik ist Schnittstelle | Fehlermeldung ist Prompt, nicht Beiwerk |
| P8 | Ableitung ist Einbahnstraße | kein Backend-Output wird je manuell nachbearbeitet |

---

## 7. Grenzen — was der Ansatz nicht leistet

Bewusst und deutlich, weil hier die realistische Erwartungshaltung entsteht.

- **Fachliche Korrektheit.** Der Compiler prüft, dass eine Anforderung *referenziert*
  ist — nicht, dass der Code sie *erfüllt*. Diese Lücke ist prinzipiell und bleibt.
- **Alles ohne Codeort.** Kontextübergreifende Abläufe, Negativentscheidungen,
  Zukunftsplanung. Gehört in den Anforderungsbaum; die Prüfbarkeit ist dort
  schwächer.
- **Querschnittliche Qualitäten.** Nebenläufigkeit, Performanz, Bedienbarkeit,
  Ästhetik entziehen sich der Annotation weitgehend. Keine 100-%-Abdeckung anstreben.
- **Determinismus.** Wiederholte Erzeugung liefert unterschiedlichen Code.
  Golden-File-artige Reproduzierbarkeitstests entfallen ersatzlos; was der Verifier
  nicht prüft, ist keine Garantie mehr (siehe 9.1).
- **Review-Ersatz.** Der Loop ersetzt Architektur-Review, nicht Fachreview.

---

## 8. Risiken und Gegenmittel

| Risiko | Wirkung | Gegenmittel |
|---|---|---|
| **Überanpassung an die Zielfunktion** — LLM ergänzt Annotationen, bis grün, ohne die Anforderung zu erfüllen | schleichender, schwer entdeckbarer Qualitätsverlust | beidseitige Abdeckung; ausführbare Anforderungen wo möglich; generierte Doku als Reviewartefakt — falsche Zuordnung fällt beim Lesen auf |
| **Falschmeldungen des Compilers** | LLM mutiert Code, um eine falsche Regel zu befriedigen | konservative Regelmenge; K4 nur bei nachgewiesenem Bedarf; jede Regel abschaltbar mit begründeter Ausnahme |
| **Werkzeug wird zum Engpass** | jede Architekturänderung braucht Compilerarbeit | Regeln als Daten/Konfiguration, nicht als hartcodierte Logik |
| **Annotationsverfall** | Annotation beschreibt den Code nicht mehr | historisches Hauptversagen vergleichbarer Ansätze; hier abgeschwächt, weil der Loop bei jedem Lauf prüft und ein LLM die Pflege trägt |
| **Sprachdesign zu früh eingefroren** | teure Nachmigration | Vokabular aus bestehenden Invarianten *ableiten*, nicht erfinden |
| **Bindung an ein LLM-Leistungsniveau** | Ansatz trägt nur mit Modellen ab Güte X | Fallback muss Handarbeit bleiben; Annotationen sollen auch für Menschen erträglich sein |
| **Verifikationsmonokultur** | der Verifier ist die einzige Schranke | klassische Tests bleiben verpflichtend, nicht ersetzt |

---

## 9. Ehrliche Einordnung

### 9.1 Der Generator verschwindet nicht

Formulierungen wie „es gibt keinen Generator mehr" sind unpräzise. Zutreffend ist:

> Ein **deterministischer, enger** Generator (deckt nur Strukturcode ab) wird durch
> einen **stochastischen, allgemeinen** ersetzt (das LLM, deckt auch Fachlogik ab),
> flankiert von einem starken Verifier.

Der Tausch lautet **Determinismus gegen Ausdrucksmächtigkeit**. Er ist gut begründet,
weil klassische Generatoren typischerweise nur den kleineren, mechanischen Teil
abdecken und der große Rest ohnehin außerhalb der Kette liegt. Aber er ist ein
Tausch, kein reiner Gewinn — die eingebüßte Reproduzierbarkeit muss der Verifier
kompensieren.

### 9.2 Aufwandsbilanz

Der Annotations-Compiler ist kein kleines Werkzeug: Grammatik, Parser, Bindung an die
Typinformation der Wirtssprache, Semantikprüfung, Backends, dazu dauerhafte Pflege bei
jeder Architekturänderung.

Die Bilanz ist trotzdem günstig, wenn er ein bestehendes Spezifikationsmodell,
dessen Generator und dessen Konfigurationsduplikate **ersetzt** statt ergänzt. Wird er
zusätzlich gebaut, verschlechtert er die Lage. Die Ersetzung ist damit nicht
Nebenbedingung, sondern Voraussetzung für die Wirtschaftlichkeit.

### 9.3 Was empirisch offen ist

Die Annahme, dass der überwiegende Teil des Quelltextes dauerhaft von LLMs erzeugt
**und gewartet** werden kann, ist eine begründete Wette, kein Befund. Sie entscheidet
über den gesamten Ansatz und ist billig überprüfbar (Abschnitt 12).

---

## 10. Abgrenzung zu bekannten Ansätzen

| Ansatz | Gemeinsamkeit | Unterschied |
|---|---|---|
| Model Driven Architecture | Modell als Quelle | MDA generiert Code aus einem *getrennten* Modell — genau die hier abgeschaffte Distanz |
| Formale Kontraktsprachen in Kommentaren (JML, ACSL, SPARK) | Subsprache in Kommentaren, eigener Übersetzer | dort Beweisbarkeit einzelner Funktionen; hier Architektur- und Rückverfolgbarkeitsebene, kein Beweisanspruch |
| Design by Contract | ausführbare Zusicherungen am Code | beschränkt auf Vor-/Nachbedingungen; keine Anforderungsbindung |
| Architektur-Testbibliotheken (ArchUnit u. ä.) | Architekturregeln maschinell prüfen | prüfen nur Struktur, kennen keine Anforderungen und erzeugen keine Dokumentation |
| Doku aus Kommentaren (JavaDoc, OpenAPI-Annotationen) | Doku aus dem Code | reine Beschreibung, keine Semantikprüfung, keine Abdeckungsmessung |
| Literate Programming | Text und Code in einer Datei | dort Erzählung für Menschen; hier maschinenprüfbare Struktur |
| Behaviour Driven Development | Anforderung maschinennah, ausführbar | dort Verhaltensprüfung in eigenen Dateien; hier Bindung *am* Konstrukt, ohne Ausführungsanspruch |

Neu ist nicht ein einzelner Bestandteil, sondern die Kombination:
**Anforderungsbindung + Architekturprüfung + Dokumentationsableitung + LLM-Loop, alles
ko-lokalisiert.**

---

## 11. Offene Entwurfsfragen

Für die Diskussion — jeweils mit den erkennbaren Alternativen.

**F1 — Annotationsmechanik.** Kommentar-Subsprache vs. typisierte Registrierung in
der Wirtssprache. Kommentare: universell platzierbar, sprachunabhängig, aber eigener
Compiler nötig. Typisierte Registrierung: sofortige Compilerprüfung, aber nur dort
möglich, wo die Wirtssprache es zulässt, und wortreicher. *Hybrid denkbar: Referenzen
typisiert, Freitext annotiert.*

**F2 — Granularität.** Wie fein wird annotiert? Modul, Typ, Funktion, Feld? Feiner =
präzisere Rückverfolgbarkeit, mehr Rauschen und mehr Pflege.

**F3 — Identität der Anforderungen.** Sprechende IDs (lesbar, aber umbenennungsanfällig)
vs. opake IDs (stabil, aber im Code nichtssagend)? Wie werden Ersetzung und
Versionierung von Anforderungen abgebildet?

**F4 — Wirtssprachenunabhängigkeit.** Soll der Compiler nur eine Sprache bedienen
oder generisch sein? Die Bindungsphase ist sprachspezifisch, Auflösung und Backends
nicht. Sauberer Schnitt zwischen beiden entscheidet über die Wiederverwendbarkeit in
anderen Projekten.

**F5 — Ausführbare Annotationen.** Sollen Annotationen Prädikate tragen, aus denen
Tests entstehen (Pflichtfelder, Zustandsübergänge, Berechtigungspflicht)? Das ist der
Hebel gegen Überanpassung — erhöht aber die Sprachkomplexität deutlich und kollidiert
mit P4.

**F6 — Grenze zwischen Anforderungsbaum und Annotation.** Wo genau verläuft sie?
Faustregel im Entwurf: „hat einen Codeort" → Annotation, sonst → Baum. Belastbar?

**F7 — Umgang mit Bestandscode.** Wie wird eine bestehende, nicht annotierte
Codebasis überführt? Schrittweise mit Toleranzmodus, oder Big Bang?

**F8 — Reviewmodell.** Wenn der Code Artefakt ist: wird der Diff des Codes noch
reviewt, oder nur der Diff von Annotationen und Anforderungen? Was heißt das für
Freigabe, Haftung und Auditierbarkeit?

**F9 — Mehrsprachigkeit der Fachtexte.** Annotationen tragen Fachtext für Doku und
Hilfesystem. Wie werden Übersetzungen gehalten, ohne die Ko-Lokalisierung zu brechen?

---

## 12. Validierungsvorschlag

Vor Investition in Grammatik, Parser und Backends ist genau eine Frage zu beantworten:
**trägt der Loop?**

1. **Vokabular ableiten, nicht erfinden.** Bestehende Architekturinvarianten und
   Modellkonzepte des Zielprojekts bilden das Anforderungsprofil der Sprache.
2. **Zielbild von Hand.** Einen abgeschlossenen Ausschnitt vollständig annotieren,
   ohne Compiler. Prüffragen: Ist die Sprache erträglich? Welche Information findet
   keinen Ort?
3. **Rückwärtsprobe.** Lässt sich aus den Annotationen die bisherige Dokumentation in
   vergleichbarer Qualität erzeugen? Die Differenz benennt die Lücke exakt.
4. **Loop-Test — der eigentliche Beweis.** Einen nichttrivialen Implementierungsteil
   löschen und allein aus Annotationen, Anforderungsbaum und Compilerdiagnostik
   rekonstruieren lassen. Messgrößen: Konvergenz ja/nein, Anzahl Runden, fachliche
   Korrektheit des Ergebnisses (nicht nur formales Grün).
5. Erst danach: Format des Anforderungsbaums, Compilerarchitektur, Migration,
   Rückbau der abgelösten Artefakte.

Schritt 4 ist billig und beantwortet die einzige wirklich offene Frage. Alles davor
ist Vorbereitung, alles danach Ausführung.

---

## Anhang A — Belegfall

Beobachtungen aus einer realen Spezifikations-Pipeline (ERP-Projekt, Go, ereignis-
und kontextorientierte Architektur). Sie motivieren das Konzept, sind aber nicht
Gegenstand dieses Dokuments.

**Pipeline im Ist-Zustand**

```
Rohquellen (91 Screenshots, 6 PDF, Prozessdiagramme)
  └─ manuell ─► Anforderungsdokumente        11.386 Zeilen Markdown, 49 Dateien
       └─ manuell ─► Spezifikationsmodell (Go) 27.104 LOC, 37 Kontexte
            ├─ automatisch ─► Fachdokumentation 18.136 Zeilen (+ identische Golden-Kopie)
            └─ automatisch ─► Codegenerat        9.061 LOC
                 └─ manuell ─► Implementierung gesamt 56.293 LOC
```

**Befunde**

| Befund | Zahl | Deutung |
|---|---|---|
| Aufblähung Anforderung → Modell | 11.386 → 27.104 Zeilen (×2,4) | die Kette verdichtet nicht, sie expandiert |
| Anteil generierten Codes | 9.061 / 56.293 = 16 % | der Generator deckt nur den mechanischen Teil ab |
| Generatorabdeckung | 57 von 146 Aggregaten | die Konfigurationstabelle des Generators ist selbst eine Modellkopie (M1) |
| Modell ohne Implementierung | 22 von 37 Kontexten | Modell weit vor der Implementierung, ohne Rückkopplung |
| Außenkante | Anforderungsverweis als Freitext, kein Test | M2 in Reinform |
| Abdeckungsrichtung | nur „Anforderung → implementiert" | M3: vergessene Quellaussagen unauffindbar |

**Entscheidender Befund — spontane Rekonstruktion.** Innerhalb der Implementierung
existiert ein **handgepflegter Wissensgraph** (8 Pakete, 518 LOC), der das
Spezifikationsmodell mit denselben Konzepten nachbildet — Kontext, Aggregat, Ereignis,
Anforderung, Entscheidung, Rolle, Anwendungsfall, Herleitung — inklusive
dupliziertem Anforderungstext. Er ist an zwei Stellen **reicher** als das offizielle
Modell (Berechtigungen, Glossar, Handlungsanweisungen für Assistenzsysteme), weil
diese Information nur nah am Code nutzbar ist.

Damit existieren faktisch **vier** Pflegestellen derselben Tatsachen.

Die Deutung ist der stärkste Einzelbeleg für dieses Konzept: Das Modell wurde
unaufgefordert im Implementierungsquelltext nachgebaut, weil es dort gebraucht wird.
Die Single Source of Truth zieht es empirisch zur Implementierung hin. Der Vorschlag
formalisiert lediglich, was sich ohnehin durchgesetzt hat.

---

## Anhang B — Glossar

| Begriff | Bedeutung in diesem Dokument |
|---|---|
| **Annotation** | maschinenlesbare Aussage in Kommentarform, die ein Codekonstrukt beschreibt oder mit einer Anforderung verbindet |
| **Annotations-Compiler** | Werkzeug, das Annotationen übersetzt, bindet, prüft und daraus Artefakte erzeugt |
| **Anforderungsbaum** | strukturierte Ablage der originären Fachanforderungen mit stabilen IDs |
| **Backend** | ableitendes Ausgabeformat des Compilers; nie manuell nachbearbeitet |
| **Bindung** | Zuordnung einer Annotation zum Konstrukt der Wirtssprache über deren Typinformation |
| **Beidseitige Abdeckung** | gleichzeitige Prüfung beider Richtungen zwischen Anforderung und Code |
| **Ko-Lokalisierung** | Spezifikation steht in derselben Datei wie die Implementierung |
| **Verifier** | prüfende, nicht erzeugende Komponente |
| **Wirtssprache** | Programmiersprache, in deren Kommentaren die Subsprache eingebettet ist |
| **Zielfunktion** | maschinell prüfbares Abbruchkriterium des LLM-Loops |
