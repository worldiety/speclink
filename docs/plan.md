# Umsetzungsplan

**Status:** beschlossen, in Arbeit
**Bezug:** `konzept-annotationscompiler.md` (Prinzipien, Prüfklassen K1–K4, Fehlermuster M1–M3) · `docs/annotations.md` (Sprachentwurf)

> **Namenskonvention.** Stufen heißen `S0`–`S5`. Die Kürzel `M1`–`M3` sind im Konzept für die Fehlermuster vergeben (Mehrfachpflege, ungeprüfte Außenkante, einseitige Abdeckung) und werden hier nicht für Meilensteine verwendet.

---

## 1. Ziel

Ein Werkzeug, das Anforderungen an Codekonstrukte bindet, beide Richtungen der Abdeckung misst, Architekturinvarianten prüft und daraus sämtliche Dokumentation ableitet — mit dem Quelltext als einziger Quelle.

**Der Nutzen entsteht erst beim Rückbau.** Solange speclink zusätzlich zu Spezifikationsmodell, Generator und Wissensgraph existiert, verschlechtert es die Lage (Konzept §9.2). Die Ablösung ist Voraussetzung, nicht Nebenbedingung. Deshalb ist S4 kein Anhang, sondern der Zweck.

---

## 2. Beschlossene Entscheidungen

| Punkt | Entscheidung | Fundstelle |
|---|---|---|
| Trägerform | Nebendatei `<quelle>.annotation.go`, im **normalen Build** | annotations §2 |
| Sprache | Go-Teilmenge, definiert als `go/ast`-Knoten-Whitelist | annotations §3 |
| Prüfung | der Go-Compiler prüft Arität, Typen, Feldnamen, Referenzen | annotations §7.2 |
| Deklaration | `var X = spec.Requirement{…}` in `<ID>.spec.go` | annotations §2.2 |
| Aussage | `var _ = spec.For[T](…)` | annotations §5.3 |
| Auswertung | keine; zwei Pässe, Reihenfolgefreiheit als Property-Test | annotations §4.1 |
| Anforderungsbaum | `dec/`, `nfr/`, `con/`, `fun/<domäne>/`; ID im Dateinamen | annotations §10 |
| Inferenz | nago-Recognizer, hartkodiert; Annotation inferierbarer Fakten ist ein Fehler | annotations §1.2 |
| Generisches CRUD | verboten (`K4-NO-GENERIC-CRUD`), nicht nachgebaut | annotations §1.6, §8.2 |
| Schweregrade | **keine.** Steuerung über den Geltungsbereich | annotations §1.8 |
| Laufzeitregister | immer aktiv; dient als Kreuzprobe, nie als Wahrheit | annotations §4.2, §7.4 |
| Ausführung | monolithisch, in-process | dieses Dokument §3 |
| Abgrenzung | speclink kennt das Framework, nie das Projekt | annotations §1.7 |

---

## 3. Paketschnitt

```
cmd/speclink               verify · generate · selfreport
internal/ir                sprachneutrales Zwischenmodell
internal/reqtree           Anforderungsbaum laden, DAG prüfen
internal/lang/golang       Bindung über go/packages + go/types
  ├── infer                nago-Recognizer (hartkodiert)
  └── read                 Bindungs- und Aussageterme lesen
internal/check             K1–K3, K4-Teilmenge; je Regel eine stabile ID
internal/diag              Text und JSON aus einer Quelle
internal/backend           Fachdoku · Matrix · Lückenbericht · JSON-LD
spec/                      der öffentliche Direktivenkatalog (importiert vom Zielprojekt)
```

Frontend und Kern sind durch die **Paketgrenze und die IR** getrennt, nicht durch eine Prozessgrenze. Es gibt keinen Plugin-Lader. Ein zweites Sprach-Frontend wäre ein weiteres Paket, das dasselbe Go-Interface erfüllt; braucht es externe Werkzeuge, kapselt es sie selbst.

### 3.1 Wozu `internal/ir` — und wozu nicht

Die IR ist **keine Schnittstelle nach außen**. Sie hat kein Serialisierungsformat, keine Version und braucht keine Round-Trip-Tests. In einem monolithischen Binary gibt es dafür keinen Anlass: eine Version, ein Prozess, kein Übergang.

Ihr einziger Zweck ist **Typisolation**:

> `go/types.Type`, `token.Pos` und `ast.Node` dürfen `internal/check`, `internal/diag` und `internal/backend` nie erreichen.

Ohne diese Grenze hingen Prüfregeln und Backends unmittelbar an der Go-Typinformation, und ein zweites Sprach-Frontend wäre nicht nachrüstbar, ohne sie neu zu schreiben. Die Grenze kostet fast nichts — sie bestimmt nur, wo die Datentypen liegen.

Konkrete Folge: Positionen sind in der IR `{Datei, Zeile, Spalte}`, nicht `token.Pos` — letzteres ist ohne `FileSet` bedeutungslos und damit sprachgebunden.

### 3.2 Wo Serialisierung tatsächlich stattfindet

Drei Übergänge, keiner davon die IR:

| Übergang | Konsument | Stabilität |
|---|---|---|
| `internal/diag` → JSON | der LLM-Loop | **stabiles, versioniertes Schema** — es ist die Schnittstelle, an der P7 hängt |
| `spec.DumpJSON` → speclink | die Kreuzprobe (§ S3) | **versionsrelevant, siehe unten** |
| Backend → JSON-LD | Assistent und Hilfesystem zur Laufzeit | stabiles Artefaktformat |

**Der einzige Ort mit echtem Versionsversatz** ist der zweite: das Zielprojekt pinnt `speclink/spec` in seiner `go.mod`, während der Entwickler ein beliebiges `speclink`-Binary ausführt. Beide Versionen können auseinanderlaufen. Das Dump-Format braucht deshalb eine Versionsangabe und eine klare Fehlermeldung bei Nichtübereinstimmung — nicht die IR.

---

## 4. Stufen

### S0 — Sprachentwurf ✅ abgeschlossen

`docs/annotations.md`. Zwei Annahmen wurden dabei nicht behauptet, sondern verifiziert:

| verifiziert | Ergebnis |
|---|---|
| Annotationsdatei übersetzt, `go vet` und `gofmt` greifen | ✅ |
| fünf Fehlerklassen erzeugen echte Compilerfehler mit Position | ✅ annotations §7.3 |
| gelöschte Anforderung bricht den Produktivbuild | ✅ |
| statische Extraktion aller vier Bindungsarten über `go/packages` | ✅ Frontendkern < 100 LOC |
| Laufzeitregister liefert exakte Positionen | ✅ annotations §7.4 |
| Init-Reihenfolge ≠ Quelltextreihenfolge (deterministisch, aber verdreht) | ✅ erzwingt Mengenvergleich |

### S1 — IR, Anforderungsbaum, Go-Frontend ✅ abgeschlossen

**Inhalt**
- `internal/ir` — Zwischenmodell als Go-Datentypen (§3.1)
- `internal/reqtree` — Laden, ID-Eindeutigkeit, DAG-Zyklenfreiheit, Konsistenz von Pfad/Präfix/`Kind`
- `internal/lang/golang/read` — Bindungs- und Aussageterme aus dem getypten AST
- `internal/lang/golang/infer` — nago-Recognizer
- `spec/` — der Direktivenkatalog als übersetzbares Go

**Erkannte Konstrukte.** Use Case (benannter Functype mit `auth.Subject` als erstem Parameter), Query (dito, aber mit Datenrückgabe), Command (`Decide`), Event (`Evolve` + `Discriminator`), Aggregat (`Identity`), Permission (`permission.Declare[UC]`). Alles über aufgelöste Typen, nie über die Schreibweise eines Bezeichners — ein Alias-Import oder ein verdeckter Name kann die Erkennung nicht täuschen.

**Abnahmekriterium.** Die aus `werp/` inferierten Aggregate, Events und Use Cases werden gegen das normative `spec/`-Modell abgeglichen. Erwartete Größenordnung: 146 Aggregate, 417 Events, 130 Use Cases. Jede Abweichung ist entweder ein Recognizer-Defekt oder echter Drift — beides wertvoll. Das ist der Grund für den schrittweisen Weg: `spec/` ist eine unabhängige Referenz, gegen die sich die gesamte Inferenzschicht validieren lässt. *Noch offen, siehe §6 Punkt 7.*

**Bereits belegter Drift, der dabei fallen muss:** `R-QUOTE-SUBMIT` existiert nur in `werp/`, nicht im normativen Modell, wird aber im Wissensgraphen als erfüllt ausgewiesen.

**Property-Test.** Permutation der Eingabereihenfolge ⇒ bitgleiches IR und bitgleiche Diagnostik. Zwanzig Permutationen, Graph *und* Diagnostik verglichen.

**Befund: Phase V4 ist leer — und das ist ein gutes Zeichen.** Die Redundanzprüfung („annotierte Tatsache ist inferierbar") hat nichts zu prüfen, weil der Direktivenkatalog so entworfen wurde, dass alles Inferierbare gar nicht erst eine Direktive hat. Rolle, Permission, Aggregat, Kontext und Eventfluss kommen ausschließlich aus den Recognizern. V4 wird erst dann nicht-leer, wenn der Katalog wächst; bis dahin ist die Vorrangregel durch Abwesenheit erfüllt statt durch Prüfung.

### S2 — Prüfungen und Diagnostik (teilweise umgesetzt)

**Umgesetzt: die Architektur-Linter.** Fünf Regelgruppen über der Projektstruktur, gemessen am realen `werp`:

| Regel | Befunde in werp |
|---|---:|
| `K5-UC-FILE` Use Case nicht in `uc_<name>.go` | 238 |
| `K5-UC-CONSTRUCTOR` Konstruktor fehlt oder falsch platziert | 238 |
| `K5-UC-PERMISSION` keine eigene Permission | 16 |
| `K5-UC-PERMISSION-I18N` hartkodierte Texte | 10 |
| `K6-CTX-USECASES` Use Case fehlt im Bündel | 10 |
| `K6-CTX-NO-UI-IMPORT` Domäne importiert Präsentation | 9 |
| `K6-CTX-UI-PKG` UI-Paket falsch benannt | 1 |
| `K7`, `K8` | 0 |

Die 238 Konstruktor-Befunde haben eine gemeinsame Ursache: werp nennt den Typ `DraftQuoteUC`, den Konstruktor aber `NewDraftQuote`. Die Regel verlangt `New` + exakter Typname. Entweder fällt das `UC`-Suffix, oder die Regel muss es tolerieren — offener Punkt 9.

**Projektlayout als einzige Konfiguration.** `speclink.json` im Wurzelverzeichnis, nur bei Abweichung von der Konvention:

```json
{"contextRoot": ".", "cmdRoot": "cmd", "infraRoots": ["pkg", "kernel"]}
```

Das ist die bewusste Ausnahme zu §1.7: Regeln und Erkennung bleiben hartkodiert, aber ein Verzeichnislayout ist Projektwissen, das kein Framework-Verständnis erschließt.

### S2 — Prüfungen und Diagnostik

**Inhalt**
- Portierung der 33 im Zielprojekt bereits maschinell erzwungenen Invarianten als Regeln mit stabilen IDs
- Whitelist-Durchsetzung über `*.annotation.go` und `*.spec.go` — sicherheitsrelevant, nicht formal (annotations §3)
- K4 nur mit belegtem Bedarf: Query-Signatur `func(auth.Subject) (…, error)` (realer projektweiter Bug), `subject.Audit` als erste Anweisung in `Decide`, Produces/Emits/Reacts
- `K4-NO-GENERIC-CRUD` ist bereits umgesetzt (S1): erkennt `cfgent.Enable`, `ent.DeclarePermissions`, `ent.NewUseCases` und den Import der generischen CRUD-UI
- `internal/diag` — Text und JSON aus einer Quelle, präskriptiv, Meldungstexte per Golden-Test festgehalten

**Nicht enthalten:** Schweregrade. Der Geltungsbereich wird hier als Konfiguration eingeführt (annotations §1.8, offene Form).

### S3 — Backends und Kreuzprobe

**Inhalt**
- Fachdokumentation, Rückverfolgbarkeitsmatrix, Lückenbericht, Wissensgraph als JSON-LD
- `speclink selfreport` und der Mengenvergleich statisch ↔ Laufzeit
- `speclink verify --check-generated` für CI: erzeugen und diffen

**Rückwärtsprobe mit hartem Maßstab.** Ziel ist `spec/out/spec.adoc` — 18.136 Zeilen — in vergleichbarer Qualität. Die Differenz benennt die Lücke exakt (Konzept §12.3).

**Vorab zu klären**
1. Ob `go run -overlay` ein nur virtuell existierendes Paketverzeichnis akzeptiert. Rückfall: temporäre, gitignorierte Datei im Zielmodul.
2. Ob der Blank-Import aller Pakete im ERP gefahrlos ist — laufen `init()`-Seiteneffekte, die Verbindungen öffnen oder Dateien anlegen?

### S4 — Rückbau im Zielprojekt

Hier entsteht die Wirtschaftlichkeit. Reihenfolge nach aufsteigendem Risiko:

| Reihenfolge | Artefakt | Umfang |
|---|---|---|
| 1 | `werp/*/requirements.go` | 9 Dateien, wörtlich duplizierte Anforderungstexte |
| 2 | `werp/*/knowledge.go` | 10 Pakete, 863 LOC Selbst-Eintragung → wird inferiert |
| 3 | `werp/knowledge/` | 373 LOC Infrastruktur → Backend |
| 4 | `spec/genwerp` Konfigurationstabelle | 57 Einträge, größtenteils Namenskonvention |
| 5 | `spec/contexts/` | 279 Dateien, 27.104 LOC |

Nach Schritt 5 verliert S1 seine unabhängige Referenz. Deshalb erst hier — und erst, wenn die Backends (S3) die Ablösung tragen.

### S5 — Loop-Test

Der eigentliche Beweis (Konzept §12.4). Einen nichttrivialen Implementierungsteil löschen und allein aus Anforderungsbaum, Annotationen und `speclink verify --format json` rekonstruieren lassen.

Messgrößen: Konvergenz ja/nein, Anzahl Runden, **fachliche** Korrektheit — nicht nur formales Grün.

> Fällt er negativ aus, ist nicht der Compiler zu reparieren, sondern die Diagnostikqualität. Sie ist die Stellschraube, die über die Konvergenzgeschwindigkeit und damit über den gesamten Ansatz entscheidet (Konzept §5.3).

---

## 5. Paralleler Strang: Migration im Zielprojekt

**Nicht Teil von speclink** (annotations §1.7). Einmalig, projektspezifisch, im ERP-Repository zu erledigen — ein Skript, das nach dem Lauf verschwindet.

Gemessener Umfang der Quellverweis-Überführung:

| Form | Anzahl | automatisch |
|---|---:|---|
| voller Pfad `anforderungen/…` | 376 | 208 von 209 distinkten Pfaden auflösbar |
| davon mit `§`-Anker | 302 | 104 von 131 distinkten über die Überschriftenregel |
| nackter Dateiname (`vertraege.md §7`) | 42 | über eindeutigen Basisnamen |
| Norm/Gesetz | 27 | → `Source.Extern`, keine Prüfung nötig |
| Verzeichnisverweis, Freitext | 31 | **manuell** |
| leer | 77 | **manuell** |

**Restvorrat Handarbeit: rund 135 Positionen.** Der Pfadteil ist nahezu geschenkt.

Abhängigkeit: Der Strang kann parallel zu S1 laufen, muss aber vor S3 abgeschlossen sein, weil die Backends den Anforderungsbaum brauchen.

---

## 6. Offene Punkte

| # | Punkt | Wann zu entscheiden |
|---|---|---|
| 1 | Form der Geltungsbereichs-Konfiguration: Positivliste, Negativliste oder Marker im Paket | S2 |
| 2 | Modulschnitt: liegt der Anforderungsbaum im selben Go-Modul wie das Zielprojekt? | vor S1, betrifft Versionierung und Release |
| 3 | F9 Mehrsprachigkeit der Fachtexte (`Help`, `Term`) | nach S3 |
| 4 | F8 Reviewmodell: wird der Code-Diff noch reviewt oder nur der von Annotationen und Anforderungen? | vor S4 |
| 5 | Verdrahtungswahrheit als eigener Befehl (braucht Eingriff im Zielprojekt) | nach S3 |
| 6 | Bildregionen als prüfbarer Verweis — derzeit bewusste Lücke | offen, evtl. nie |
| 7 | Abnahme der Inferenzschicht gegen das reale `spec/`-Modell des ERP (146 Aggregate, 417 Events, 130 Use Cases) | vor S2 |
| 8 | Weitere Recognizer: Eventfluss (`bus.Publish`, `events.SubscribeFor`), Routen (`cfg.RootView`, `admin.Card`), ReBAC-Namespaces | S2, nach Bedarf |
| 9 | `New<Typname>` gegen werps `UC`-Suffix: 238 Befunde aus einer einzigen Namenskonvention | vor der Einführung im ERP |

---

## 7. Bewusst nicht im Umfang

| | Begründung |
|---|---|
| Zweites Sprach-Frontend | die Fähigkeit wird gebaut (IR-Grenze), das Frontend nicht |
| Prädikatsprache / ausführbare Annotationen (F5) | kollidiert mit P4; nachrüstbar ohne Migration |
| K4 über Datenflussanalyse | Kostenwarnung Konzept §5.4 |
| Schweregrade, Toleranzmodus | annotations §1.8 |
| Migrationswerkzeuge für Fremdmodelle | annotations §1.7 |
