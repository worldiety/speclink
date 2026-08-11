# Annotationssprache — Entwurf

**Status:** Entwurf, beschlussreif
**Zweck:** Festlegung von Trägerform, Grammatik, Direktivenkatalog, Bindung, Prüfung und Diagnostik
**Bezug:** `konzept-annotationscompiler.md` (Prinzipien P1–P8, Prüfklassen K1–K4, Fehlermuster M1–M3)

---

## 1. Zweck und Abgrenzung

### 1.1 Drei Orte, drei Rollen

| Ort | Trägt | Beispiel |
|---|---|---|
| **Anforderungsbaum** (`<ID>.spec.go`) | Deklarationen: Anforderungen, Entscheidungen, Begriffe, deren Text, Herleitung und Quelle | `var R_QUOTE_SUBMIT = spec.Requirement{…}` |
| **Annotationsdatei** (`<quelle>.annotation.go`) | Aussagen über die Konstrukte der Nachbardatei | `spec.For[SubmitQuoteUC](spec.Satisfies(…))` |
| **Recognizer** (hartkodiert in speclink) | alles, was aus der Framework-Nutzung ableitbar ist | Rolle, Permission, Aggregat, Eventfluss |

### 1.2 Die Vorrangregel

> **Ist eine Tatsache inferierbar, ist ihre Annotation ein Fehler — kein Duplikat.**

Unmittelbare Umsetzung von P2 („genau eine Quelle je Tatsache") und zugleich das wirksamste Mittel gegen Annotationsverfall: Was nicht annotiert ist, kann nicht veralten. Der Compiler meldet Redundanz präskriptiv (Phase V4, §8).

### 1.3 Was inferiert wird (nago)

Nicht annotieren — speclink erkennt es:

| Marker im Zielcode | inferierte Tatsache |
|---|---|
| `permission.Declare[DraftQuoteUC]("sales.quote.draft", …)` | Permission ↔ Use Case, ID, Klartextname |
| benannter Functype mit `auth.Subject` als erstem Parameter | Use-Case-Eigenschaft |
| `subject.Audit(PermX)` als erste Anweisung in `Decide` | Implementierung → Permission |
| `evs.Cmd` / `evs.Evt` (Interface-Erfüllung) | Command bzw. Domain-Event |
| Rückgabe von `Decide` | produzierte Events |
| `data.Repository[E, ID]`, `data.Aggregate[ID]` | Aggregatwurzel und ID-Typ |
| `bus.Publish(T{})`, `events.SubscribeFor[T]` | Eventfluss |
| `cfg.RootView(path, factory)`, `admin.Card{Target, Role, Permission}` | Routen, Route ↔ Rolle ↔ Permission |
| `const Namespace rebac.Namespace = "…"` | Autorisierungs-Ressourcenmodell |
| `var _ DraftQuoteUC = NewDraftQuote(nil)` | Konstruktor ↔ Use-Case-Typ |

### 1.4 Was übrig bleibt

Nur was weder aus dem Code folgt noch in den Anforderungsbaum gehört:

1. **Die Anforderungsbindung** (`Satisfies`) — prinzipiell nicht inferierbar. Der Kern des Werkzeugs.
2. **Fachtext für Endnutzer** (`Help`, `Term`) — Handlungsanweisung und Glossar.
3. **Nicht-triviale Semantik** (`Transition`, `External`) — Aussagen, die der Code nicht selbst trägt.
4. **Begründete Ausnahmen** (`Waive`).

### 1.5 Die Trennlinie zum Anforderungsbaum (F6)

Faustregel des Konzepts: „hat einen Codeort → Annotation, sonst → Baum". Präzisiert:

- Eine **Deklaration** (was gilt) gehört immer in den Baum, auch wenn sie genau ein Konstrukt betrifft. Sie ist fachlich verantwortet und überlebt die Implementierung.
- Eine **Aussage** (dieses Konstrukt tut X) gehört immer an das Konstrukt.

Konsequenz: `spec.Requirement{…}` ist in Annotationsdateien **nicht zulässig** (§6.4). Damit kann eine Anforderung nicht an einer willkürlichen Codestelle entstehen — das Lokalitätsprinzip wird maschinell erzwungen statt appelliert.

### 1.6 P9 — Analysierbarkeit vor Kürze

Ein Prinzip, das aus der zweiten These des Konzepts folgt, dort aber nicht steht:

> **Wenn der Quelltext überwiegend maschinell erzeugt wird, ist er Artefakt, nicht Quelle.**

Generische Fabriken existieren, um Menschen Tipparbeit zu sparen. Schreibt ein LLM den Code, ist dieser Nutzen wertlos — die Kosten bleiben:

1. Sie erzeugen Fakten **zur Laufzeit**, die eine statische Analyse nur sieht, wenn speclink Framework-Interna nachbaut. Ändert das Framework seine Namensregel, lügt speclink still.
2. Sie liefern Module, die **nur ganz oder gar nicht** anpassbar sind. Eine handgeschriebene Use-Case-Funktion ist einzeln änderbar.

> **P9 — Analysierbarkeit vor Kürze.** Konstrukte, die Spezifikationsfakten zur Laufzeit erzeugen statt sie zu deklarieren, sind verboten. Der Codeumfang, den sie sparen, hat keinen Wert mehr; die Undurchsichtigkeit, die sie erzeugen, hat ihn sehr wohl.

Betroffen sind die generischen CRUD-Konstrukte von nago: `cfgent.Enable`, `application/ent` (`ent.NewUseCases`, `ent.DeclarePermissions`) und die generische CRUD-UI `uient`. Durchgesetzt als Regel `K4-NO-GENERIC-CRUD` (§8.2).

**Das Verbot ist empirisch kostenlos.** Gemessen im ERP:

| | Vorkommen |
|---|---:|
| `cfgent.Enable` | **0** |
| `application/ent` | **0** |
| `uient.` | **0** |
| handgeschriebene `permission.Declare[UC]` | **160** |

Es verbietet nichts, was benutzt wird — es schließt eine Tür, bevor jemand hindurchgeht.

### 1.7 Abgrenzung: was speclink nicht ist

speclink ist ein **generischer, dauerhafter Verifier**. Die Trennlinie verläuft nicht zwischen „generisch" und „spezifisch", sondern präziser:

> **speclink kennt das Framework, nie das Projekt.**

Die nago-Recognizer (§1.3) und die nago-bezogenen Regeln (§8.2) sind hartkodiert — bewusst, denn ein Recognizer *muss* sein Framework kennen. Ein Framework wird von vielen Projekten geteilt; das Wissen amortisiert sich. Projektwissen amortisiert sich nie.

Nicht Bestandteil von speclink:

| | warum nicht |
|---|---|
| **Migration eines Bestandsmodells** — etwa das Überführen der 481 `spec.RegisterRequirement`-Aufrufe des ERP in den Anforderungsbaum | Der Importer müsste eine **fremde** DSL interpretieren (`spec.git/spec.Requirement`, nicht `speclink/spec.Requirement`), ein projekteigenes Dateilayout kennen und eine projekteigene Freitextkonvention (`"vertraege.md §7"`) umschreiben. Alles drei ist einmalig und wäre danach totes Gewicht |
| **Normalisierung von Alt-Quellverweisen** | speclink *definiert* das Anker-Format (§5.2) und *prüft* es (V5). Die Überführung eines Altformats in dieses Format ist Projektarbeit |
| **Statustabellen, Reports zum Projektfortschritt** | fällt als Backend ab (Lückenbericht), ist aber kein eigenes Werkzeug |

Was speclink zur Einführung beiträgt, ist **kein Werkzeug, sondern eine Stellschraube**: der Geltungsbereich (§1.8). Ein Paket wird geprüft oder nicht — es gibt nichts dazwischen.

> Die Versuchung, projektspezifische Einmalwerkzeuge in den Verifier zu ziehen, weil „die Lesemechanik doch dieselbe ist", ist real. Sie ist es nicht: gelesen wird zwar in beiden Fällen Go, aber die *interpretierten Typen* sind verschieden. Migration gehört ins Projekt, nicht ins Werkzeug.

### 1.8 Kein Schweregrad — Geltungsbereich

speclink kennt **keine Warnungen**. Ein Befund ist ein Fehler, der Lauf schlägt fehl. Es gibt kein `--severity warn`, keine Schweregrade, keine „vorerst tolerierten" Verstöße.

Begründung, in aufsteigender Stärke:

1. **Der Go-Compiler macht es genauso.** Entweder übersetzt es, oder nicht. Es gibt keinen Grund, hier weicher zu sein.
2. **Es wäre in sich unstimmig.** Weil die Annotationen im normalen Build liegen, ist V2 (§8) ohnehin binär — an einem Übersetzungsfehler kommt man nicht mit einer Warnung vorbei. V1 und V3–V6 abschwächbar zu machen, während V2 tödlich bleibt, ergibt kein kohärentes Verhalten.
3. **Warnungen werden ignoriert.** Ein Toleranzmodus, der zur Migration gedacht ist, wird zur Dauerausrede. Das ist keine Vermutung, sondern das Regelverhalten jeder Codebasis mit Warnungsrückstand.
4. **Der Adressat ist ein LLM.** Es iteriert, bis grün — oder kommentiert aus, bis es weiterkommt. Auskommentierter Code ist sichtbar unfertig; eine unterdrückte Warnung ist unsichtbar unfertig.

#### Wie dann schrittweise eingeführt wird

Über den **Geltungsbereich**, nicht über den Schweregrad. Beide verfügbaren Mittel sind binär:

| Mittel | Wirkung | Rechenschaft |
|---|---|---|
| **Paket außerhalb des Geltungsbereichs** | wird gar nicht geprüft | steht explizit in der Konfiguration |
| **`spec.Waive(rule, reason)`** | eine Regel für ein Konstrukt ausgesetzt | Begründung ist Pflicht, erscheint im Lückenbericht |

Innerhalb des Geltungsbereichs gilt Vollständigkeit; außerhalb wird nichts behauptet. Eine Codebasis wird also paketweise überführt, nicht regelweise aufgeweicht. Der Unterschied ist entscheidend: „dieses Paket ist noch nicht dran" ist eine wahre Aussage, „diese Regel gilt hier nur halb" ist keine.

---

## 2. Trägerform

### 2.1 Annotationsdatei

Neben jeder annotierten Quelldatei liegt eine gleichnamige Datei mit der Endung `.annotation.go`, im **selben Paket** und im **normalen Build**.

```
werp/sales/
  commands.go
  commands.annotation.go
  usecases.go
  usecases.annotation.go
  types.go
```

```go
package sales

import (
	"github.com/worldiety/speclink/spec"
	"gitlab.worldiety.net/worldiety-erp/spec.git/anforderungen/fun/quote"
)

var _ = spec.For[SubmitQuoteUC](
	spec.Satisfies(quote.R_QUOTE_SUBMIT),
	spec.Transition[QuoteSubmitted]("submitted"),
	spec.Help(`Gib das freigegebene Angebot ab. Das System zieht die nächste
Angebotsnummer aus dem Nummernkreis.`),
)
```

Regeln:

- Eine Annotationsdatei je Quelldatei, `<basis>.annotation.go` zu `<basis>.go`.
- Sie annotiert ausschließlich Konstrukte ihrer Nachbardatei. Eine Annotationsdatei ohne Nachbardatei ist ein Fehler — ein verwaister Rest nach Umbenennung oder Löschung.
- Sie enthält **nur** Bindungsterme, Importe und die Paketklausel. Nichts sonst (§3).

### 2.2 Anforderungsbaum

Reguläre Go-Dateien mit der Endung `.spec.go`, abgelegt unter `anforderungen/` (§10).

```go
package quote

import "github.com/worldiety/speclink/spec"

var R_QUOTE_SUBMIT = spec.Requirement{
	ID:         "R-QUOTE-SUBMIT",
	Kind:       spec.Functional,
	Discipline: spec.Fachlich,
	Status:     spec.Normative,
	Title:      "Angebotsnummer bei Abgabe",
	Text: `Bei Abgabe eines freigegebenen Angebots MUSS eine fortlaufende, dublettenfreie
Angebotsnummer aus dem zentralen Nummernkreis gezogen werden.`,
	Sources: []spec.Source{
		{Doc: "anforderungen/_quellen/worldiety/angebote/angebotsflow.md", Anchor: "8-abgabe"},
	},
}
```

Beide Träger nutzen dasselbe `spec`-Paket, dieselbe Knoten-Whitelist und dieselben Prüfphasen.

### 2.3 Warum diese Form

| | |
|---|---|
| **Kein Parser, keine Extraktion.** | Es ist Go. `go/packages` lädt es, `go/types` prüft es, die Positionen stimmen. |
| **Kein eigenes Schema.** | Der Direktivenkatalog ist eine Go-Typdeklaration; `go doc` dokumentiert ihn. |
| **Der Go-Compiler prüft mit.** | Arität, Argumenttypen, Feldnamen, Aufzählungswerte, Bezeichnerreferenzen (§7.2). P3 ist nicht abgeschwächt, sondern eingelöst. |
| **IDE-Werkzeuge greifen.** | Umbenennen erfasst die Annotation, Sprungnavigation und Autovervollständigung funktionieren, `gofmt` formatiert, `go vet` prüft. Bei einer Kommentar-Subsprache prinzipiell unmöglich. |
| **Drift bricht den Build.** | Ein gelöschter Typ, eine entfernte Anforderung: Übersetzungsfehler, nicht Prüfbefund (§2.4). |
| **Mechanische Migration.** | Das bestehende `spec/`-Modell ist bereits dieselbe Form ohne `Register`-Wrapper. |

### 2.4 Die Kopplung, bewusst gewählt

Annotationsdateien sind Teil des normalen Builds — kein `//go:build`-Ausschluss. Daraus folgt:

1. **Eine gebrochene Annotation bricht den Produktivbuild.** Die schärfste denkbare Fassung von K1/K2: Strukturdrift ist kein Befund, sondern ein Übersetzungsfehler. Auslieferung ohne konsistente Spezifikation wird unmöglich.
2. **Die Anforderungstexte liegen im Binary** — grob 250–400 KB. Das ist gewollt: `werp/knowledge` ist heute aus genau diesem Grund einkompiliert, weil `werp/assistant` den Wissensgraphen zur Laufzeit liest. Die Möglichkeit bleibt erhalten.
3. **Das Zielprojekt hängt zur Übersetzungszeit am Anforderungsbaum.** Ohne ihn baut es nicht. Das ist die Außenkante, die heute Freitext ist (§10.4), als echte Abhängigkeit.

Die Alternative (`//go:build speclink`) wäre rückwirkungsfrei, verlöre aber alle drei Eigenschaften und brächte einen still verrottenden, von IDE und `go vet` unbeachteten Parallelbuild.

---

## 3. Grammatik: eine Knoten-Whitelist, nicht „gültiges Go"

Die Sprache ist **nicht** Go. Sie ist eine explizit aufgezählte Teilmenge der `go/ast`-Knotentypen.

> **Da Annotationsdateien echte, normal übersetzte Go-Dateien sind, gibt es keinerlei syntaktische Schranke.** Die Whitelist ist die einzige Barriere dagegen, dass aus `*.annotation.go` eine zweite Codebasis wird — mit Hilfsfunktionen, Konstanten und irgendwann Logik. Ihre Durchsetzung ist deshalb nicht formal, sondern tragend.

### 3.1 Erlaubt in `*.annotation.go`

| Knoten | Einschränkung |
|---|---|
| `*ast.File` mit Paketklausel | Paket muss dem der Nachbardatei entsprechen |
| `*ast.GenDecl` mit `tok == IMPORT` | beliebig |
| `*ast.GenDecl` mit `tok == VAR` | ausschließlich `var _ = <Bindungsterm>` |
| `*ast.CallExpr` | Callee muss in `speclink/spec` auflösen |
| `*ast.IndexExpr`, `*ast.IndexListExpr` | nur als Callee — generische Instanziierung (`spec.For[T]`) |
| `*ast.SelectorExpr`, `*ast.Ident` | |
| `*ast.BasicLit` | `STRING`, `INT` |
| `*ast.UnaryExpr` | nur `&`, nur als Argument von `ForVar` |

### 3.2 Zusätzlich erlaubt in `*.spec.go`

| Knoten | Einschränkung |
|---|---|
| `*ast.GenDecl` mit `tok == VAR` | `var <Name> = spec.Requirement{…}` bzw. `spec.Glossary{…}` |
| `*ast.CompositeLit` | nur mit `Key: Value`; positionale Felder verboten |
| `*ast.KeyValueExpr` | Schlüssel ist Bezeichner |
| `*ast.ArrayType` + Slice-Literal | nur als Feldwert (`Sources: []spec.Source{…}`) |

### 3.3 Verboten in beiden

`FuncDecl`, `TypeSpec`, `ConstDecl`, `FuncLit`, `IfStmt`, `ForStmt`, `RangeStmt`, `SwitchStmt`, `TypeSwitchStmt`, `SelectStmt`, `ReturnStmt`, `BranchStmt`, `DeferStmt`, `GoStmt`, `LabeledStmt`, `BlockStmt`, `BinaryExpr`, `SliceExpr`, `TypeAssertExpr`, `StarExpr`, `ChanType`, `MapType`, `InterfaceType`, `StructType`, `IncDecStmt`, `SendStmt`, `AssignStmt`.

Jeder verbotene Knoten erzeugt eine eigene, präskriptive Diagnose — niemals ein generisches „nicht erlaubt" und niemals stilles Ignorieren.

### 3.4 Totalitätsgarantie

Ohne `ForStmt`, `RangeStmt` und Funktionsdefinitionen ist die Sprache **total**: keine Schleifen, keine Rekursion, jede Auswertung terminiert konstruktiv.

> Diese Garantie ist der Preis jeder künftigen Erweiterung. Wer `ForStmt` aufnimmt, gibt sie auf und braucht Rekursionsgrenzen und Fixpunkterkennung. Erweiterungen der Whitelist sind bewusste, zu begründende Entscheidungen.

---

## 4. Deklaration und Aussage

| | Form | Ort | Semantik |
|---|---|---|---|
| **Deklaration** | `var X = spec.Requirement{…}` | Anforderungsbaum | ein Term |
| **Aussage** | `var _ = spec.For[T](…)` | Annotationsdatei | ein Term |

### 4.1 speclink wertet nie aus

> **Harte Regel:** Keine Annotation verändert die Bedeutung einer anderen.

speclink liest den **getypten AST**, es führt nichts aus. Zwei Begriffe, die auseinandergehalten gehören:

| | speclink |
|---|---|
| **Typauflösung** — Bezeichner auf Deklarationen abbilden, Typparameter instanziieren, paketübergreifend nachschlagen | **ja, vollständig** |
| **Auswertung** — die Terme aufrufen, Werte berechnen, Registries füllen | **nein** |

Ohne Typauflösung ginge nichts: `QuoteSubmitted` im rohen AST ist nur ein Bezeichner. Erst `go/types` sagt, welcher Typ welchen Pakets gemeint ist. Geladen wird mit `go/packages` im Modus `NeedTypes｜NeedTypesInfo｜NeedDeps`; `types.Info.Instances` liefert die Typargumente generischer Aufrufe, `types.Info.Uses` die Deklaration jeder Referenz samt Position (verifiziert, §7.3).

Aus dem Verzicht auf **Auswertung** folgt:

- Terme werden in einem Durchlauf eingesammelt, nicht expandiert.
- Kein Fixpunkt, keine Rekursionsgrenze, keine Hygiene, keine Sichtbarkeitsregeln zwischen Annotationen.
- **Zwei Pässe:** Pass 1 sammelt alle Deklarationen über alle Pakete. Pass 2 löst alle Aussagen dagegen auf. Vorwärtsreferenzen sind zulässig, Reihenfolge irrelevant.

**Testbare Eigenschaft:** Permutation der Eingabereihenfolge (Pakete, Dateien, Deklarationen) muss zu bitgleichem IR und bitgleicher Diagnostik führen. Das liefert nebenbei den Determinismus, dessen Verlust Kap. 7 des Konzepts für LLM-erzeugten Code beklagt.

### 4.2 Der Laufzeitpfad ist ein anderer

Weil die Terme im normalen Build liegen, werden sie zur Init-Zeit *ausgewertet* — Go-Semantik, deterministisch, aber reihenfolgeabhängig (Paketabhängigkeit, dann Dateiname, dann Deklarationsreihenfolge).

Die Laufzeitregistrierung ist **immer aktiv** (Entscheidung, §7.4). Sie hält die Option offen, den Wissensgraphen zur Laufzeit zu nutzen — wie `werp/knowledge` es heute für den Assistenten tut — und sie verhindert, dass der Compiler `var _ = For(…)` als wirkungslos eliminiert.

Sie hat **keinerlei Einfluss auf speclink**, das rein statisch arbeitet. Die beiden Modelle sind sauber getrennt und dürfen nicht vermischt werden:

| | speclink | Laufzeitregister |
|---|---|---|
| liest | AST + Typinformation | ausgeführte Terme |
| Reihenfolge | irrelevant, testbar erzwungen | Go-Init-Ordnung |
| Zweck | Prüfung, Backends | Hilfesystem, Assistent, **Kreuzprobe** (§7.4) |
| Verbindlichkeit | **maßgeblich** | nie Korrektheitsgrundlage (§13.1) |

### 4.3 Verworfene Alternative

Kommentar-Subsprache (`// #[ … // ]`). Sie erzwänge Blockextraktion, den Bau einer synthetischen Datei, eine Positionsabbildung und einen eigenen `types.Config.Check`-Lauf — und verlöre IDE-Refactoring, `gofmt` und `go vet`. Ihr einziger Vorteil wäre positionelle Bindung (§6.5).

---

## 5. Das Paket `speclink/spec`

Der Direktivenkatalog ist hartkodiert — als Go-Typdeklarationen. Damit prüft der Go-Compiler Arität, Argumenttypen, Feldnamen und Aufzählungswerte.

### 5.1 Identitäten und Aufzählungen

```go
package spec

type RequirementID string
type RuleID string
type TermID string
type State string

// Kind entspricht spec.RequirementKind des Bestandsmodells und bestimmt die
// erste Verzeichnisebene des Anforderungsbaums (§10.2).
type Kind int

const (
	Functional Kind = iota + 1
	NonFunctional
	Constraint
	Decision
)

type Discipline int

const (
	Fachlich Discipline = iota + 1
	Technisch
	Gemischt
)

// Status steuert die Rückwärtsrichtung der Abdeckungsprüfung (K3).
// Nur Normative geht in die Pflichtabdeckung ein; die übrigen sind die
// explizit markierten, begründungspflichtigen Ausnahmen.
type Status int

const (
	Normative Status = iota + 1
	Abstract         // reiner Herleitungsknoten, wird NICHT abgedeckt
	Planned
	OutOfScope
	Informative
	Superseded
)

// Role klassifiziert begleitendes Material. Spiegelt bewusst die Aufzählung
// aus R-QUOTE-ASSETS der Fachdomäne.
type Role int

const (
	Mockup Role = iota + 1
	Scribble
	Diagramm
	Akzeptanzkriterien
	Protokoll
	Dokument
)
```

### 5.2 Deklarationsterme

```go
// Source benennt die Herkunft der Anforderung. Doc ist ein repo-relativer Pfad
// nach anforderungen/_quellen/ und MUSS existieren. Extern trägt Gesetze und
// Normen ohne Dokument im Repository.
// Genau eines von {Doc, Extern} ist zu setzen.
//
// Anchor ist der Slug einer Überschrift des Zieldokuments: kleingeschrieben,
// Leerzeichen zu "-", Satzzeichen entfernt — die übliche Markdown-Ankerbildung.
// "## 8.1 Angebot (Kopf)" ergibt "81-angebot-kopf". speclink prüft, dass eine
// Überschrift mit diesem Slug existiert (V5). Bei Nicht-Textdokumenten bleibt
// Anchor leer.
//
// Note beschreibt bei Bildern die gemeinte Stelle. Es ist nicht prüfbar und die
// bewusst in Kauf genommene Restlücke von M2 (§13.1).
type Source struct {
	Doc    string
	Anchor string
	Extern string
	Note   string
}

// Attachment ist begleitendes Material — im Unterschied zu Source nicht die
// Herkunft, sondern Zubehör. Path ist relativ zum Anhangsordner der Anforderung
// oder repo-relativ bei geteiltem Material (§10.5).
type Attachment struct {
	Path string
	Role Role
	Note string
}

type Requirement struct {
	ID          RequirementID
	Kind        Kind
	Discipline  Discipline
	Status      Status
	Title       string
	Text        string // normativ, kurz. Für Listen, Matrizen, Diagnosemeldungen.
	Detail      string // optionale Markdown-Datei im Anhangsordner für Langform
	Rationale   string // Pflicht bei Kind == Decision
	DerivedFrom []Requirement
	Supersedes  []Requirement
	Sources     []Source
	Attachments []Attachment
}

type Glossary struct {
	ID         TermID
	Title      string
	Definition string
}
```

### 5.3 Bindungsterme

Die **einzigen fünf Funktionen der Sprache mit Seiteneffekt.** Sie tragen den Term in das Laufzeitregister ein (§4.2) und erfassen dabei ihre Quellposition.

```go
// Binding ist der opake Rückgabetyp aller Bindungen. Er trägt keine Information
// und existiert, damit Bindungen als `var _ = …` auf Paketebene stehen können.
type Binding struct{}

// For bindet an einen benannten Typ — Use-Case-Functype, Event-Struct, Aggregat.
// Der Regelfall; das Ziel ist vom Go-Compiler geprüft.
func For[T any](as ...Assertion) Binding

// ForFunc bindet an eine Funktion. Das Argument ist der Funktionswert selbst.
func ForFunc(fn any, as ...Assertion) Binding

// ForVar bindet an eine Variable oder Konstante. Das Argument ist ihre Adresse.
func ForVar(ptr any, as ...Assertion) Binding

// ForField bindet an ein Struct-Feld. Der Feldname ist ein String und damit
// NICHT vom Go-Compiler geprüft; speclink prüft ihn gegen den Typ (§6.5).
func ForField[T any](field string, as ...Assertion) Binding

// ForPackage bindet an das Paket der Nachbardatei.
func ForPackage(as ...Assertion) Binding
```

Jede dieser Funktionen erfasst `runtime.Caller(1)`. Das ist nicht Beiwerk, sondern trägt zwei Dinge:

1. **Die Kreuzprobe wird positionsbasiert** (§7.4) — statisches und Laufzeitmodell werden über `(datei, zeile, …)` verglichen.
2. **Es umgeht einen sonst harten Verlust.** `ForVar(&PermSubmitQuote)` liefert zur Laufzeit nur einen Zeiger; der Variablenname ist weg. Über die Position ist der Eintrag trotzdem eindeutig zuzuordnen.

Präzedenz im Zielumfeld: nagos `permission.register(permission, skip int)` existiert aus genau diesem Grund.

### 5.4 Aussageterme

Alle Aussageterme sind **rein** — sie berechnen eine Nutzlast und geben sie zurück, ohne Seiteneffekt.

Das ist keine Konvention, sondern von Go erzwungen: Argumente werden **vor** dem Aufruf ausgewertet.

```go
var _ = spec.For[SubmitQuoteUC](
	spec.Satisfies(quote.R_QUOTE_SUBMIT),          // läuft zuerst
	spec.Transition[QuoteSubmitted]("submitted"),  // dann
)                                                   // For läuft zuletzt
```

`Satisfies` kann sich also gar nicht selbst registrieren — es kennt sein Bindungsziel nicht. Es *muss* zurückgeben, und `For` sammelt ein. Deshalb trägt `Assertion` eine unexportierte Nutzlast statt leer zu sein.

```go
// Assertion trägt ihre Nutzlast in unexportierten Feldern und ist für den
// Benutzer opak. Aussageterme sind rein; die Seiteneffektfläche der gesamten
// Sprache liegt bei den fünf Bindungsfunktionen aus §5.3.
type Assertion struct {
	kind assertionKind
	reqs []Requirement
	typ  reflect.Type
	text string
	rule RuleID
}

// Satisfies bindet das Konstrukt an eine oder mehrere Anforderungen.
// Die zentrale Aussage der Sprache; nicht inferierbar.
func Satisfies(reqs ...Requirement) Assertion

// Transition erklärt, welchen groben Lebenszyklus-Zustand das Aggregat nach dem
// Event T einnimmt. Der Zustand ist bewusst kein Eventfeld; er folgt aus dem Typ.
func Transition[T any](to State) Assertion

// External markiert ein Event als von außen kommend (ACL/Föderation) und nimmt es
// von der Journey-Erreichbarkeitsprüfung aus.
func External() Assertion

// Help trägt die Handlungsanweisung für Endnutzer, Hilfesystem und Assistent.
func Help(text string) Assertion

// Term verankert einen Glossarbegriff am definierenden Konstrukt.
func Term(g Glossary) Assertion

// Rationale begründet eine Entscheidung am Konstrukt, das sie umsetzt.
func Rationale(text string) Assertion

// Waive setzt eine Regel für dieses Konstrukt aus. reason ist Pflicht und wird
// in den Lückenbericht übernommen.
func Waive(rule RuleID, reason string) Assertion
```

### 5.5 Mehrzeiliger Fachtext

Ein Sonderfall ist nicht nötig — Go-Rohstrings tragen Zeilenumbrüche:

```go
spec.Help(`Gib das freigegebene Angebot ab. Das System zieht die nächste
Angebotsnummer aus dem Nummernkreis. Die Freigabe muss vorher erteilt sein.`)
```

### 5.6 Typreferenzen über Typparameter

Go kann Typen nicht als Werte übergeben. Direktiven, die auf einen Typ zeigen, nutzen deshalb einen Typparameter — dieselbe Idiomatik wie `permission.Declare[DraftQuoteUC]` und `spec.RegisterEvent[T]` im Bestand. Der Vorteil ist entscheidend: **die Referenz wird vom Go-Typechecker aufgelöst**, ein Tippfehler ist ein Compilerfehler.

---

## 6. Bindung

### 6.1 Explizit statt positionell

Eine Nebendatei hat keine Position relativ zum Konstrukt. Das Ziel wird deshalb benannt. Für benannte Typen ist das *stärker* als positionelle Bindung, weil ein nicht existierendes Ziel gar nicht übersetzt.

```go
var _ = spec.For[QuoteSubmitted](
	spec.Transition[QuoteSubmitted]("submitted"),
)

var _ = spec.ForFunc(NewSubmitQuote,
	spec.Satisfies(quote.R_NUMBERING_ALLOCATION),
)

var _ = spec.ForVar(&PermSubmitQuote,
	spec.Rationale(`Die Abgabe ist berechtigungspflichtig, weil sie eine Nummer
aus dem lückenlosen Kreis zieht und damit nicht folgenlos wiederholbar ist.`),
)
```

### 6.2 Zulässige Ziele je Aussage

| Aussage | `For[T]` | `ForFunc` | `ForVar` | `ForField[T]` | `ForPackage` |
|---|:-:|:-:|:-:|:-:|:-:|
| `Satisfies` | ✓ | ✓ | ✓ | ✓ | ✓ |
| `Transition[T]` | ✓ | ✓ | ✓ | | |
| `External` | ✓ | | | | |
| `Help` | ✓ | ✓ | ✓ | | |
| `Term` | ✓ | | | | ✓ |
| `Rationale` | ✓ | ✓ | ✓ | | ✓ |
| `Waive` | ✓ | ✓ | ✓ | ✓ | ✓ |

### 6.3 Wiederholbarkeit

`Satisfies`, `Transition`, `Waive`, `Term` sind wiederholbar. `External`, `Help`, `Rationale` sind je Ziel genau einmal zulässig. Mehrere `For`-Terme für dasselbe Ziel sind zulässig und werden vereinigt.

### 6.4 Nicht zulässig in Annotationsdateien

`spec.Requirement{…}` und `spec.Glossary{…}` sind Deklarationen und ausschließlich in `.spec.go` erlaubt (§1.5):

```
werp/sales/commands.annotation.go:14:11: [SPEC-V3-011] spec.Requirement darf nicht im Quelltext deklariert werden.
    Eine Anforderung ist fachlich verantwortet und überlebt die Implementierung.
    Lege sie unter anforderungen/fun/<domäne>/ als <ID>.spec.go an und verweise
    hier mit spec.Satisfies(…) darauf.
```

### 6.5 Die Feldebene als bekannte Einbuße

`ForField[T](name string, …)` trägt den Feldnamen als String. Der Go-Compiler prüft ihn nicht — als einzige Referenz der ganzen Sprache. speclink prüft ihn gegen die Typinformation von `T` und meldet ihn präzise, aber die Prüfung wandert aus der Wirtssprache zurück ins Werkzeug.

Das ist der Preis der Nebendatei; eine Kommentarform könnte hier positionell binden. Er ist vertretbar, weil Feldmetadaten im Bestand ohnehin über Struct-Tags getragen werden (`doc:`, `example:`, `optional:`, `catalog:`) und diese bleiben — `ForField` ist nur für die Anforderungsbindung nötig.

---

## 7. Mechanik

### 7.1 Ablauf

1. **Laden.** `go/packages` mit Syntax und Typen über Zielmodul und Anforderungsbaum. Der Go-Compiler hat bereits alles geprüft, was §7.2 auflistet.
2. **Vorprüfung.** Knoten-Whitelist über allen `*.annotation.go` und `*.spec.go` (§3).
3. **Sammeln.** Pass 1: Deklarationen. Pass 2: Bindungen und Aussagen, über den getypten AST.
4. **Prüfen und ableiten.** Auflösung, Redundanz, Semantik, Backends.

Mehr ist nicht nötig. Es gibt keine Extraktion, keine synthetische Datei, keine Positionsabbildung und keinen zusätzlichen Typprüflauf.

### 7.2 Was der Go-Compiler dabei prüft

| geprüft | Beispielfehler |
|---|---|
| Existenz der Direktive | `undefined: spec.Satsifies` |
| Arität und Argumenttypen | `cannot use R_QUOTE_DRAFT (variable of type spec.Requirement) as spec.RuleID value` |
| Feldnamen im Literal | `unknown field Diszipline in struct literal of type spec.Requirement` |
| Aufzählungswerte | `undefined: spec.Technikal` |
| **Referenzen auf Go-Typen** | `undefined: QuoteSubmittd` |
| **Existenz der Anforderung** | `undefined: quote.R_QUOTE_SUBMIT` |
| Sichtbarkeit und Importe | `undefined: sales.internalThing` |

Die Wirtssprache prüft also tatsächlich mit. P3 ist vollständig eingelöst, eine eigene Schemaprüfung entfällt ersatzlos.

### 7.3 Verifiziert

Der Mechanismus ist an einem Durchstich geprüft, nicht nur behauptet. Aufbau: ein `spec`-Paket mit den Termen aus §5, ein Anforderungspaket, ein `sales`-Paket mit `commands.go` und `commands.annotation.go`.

**Übersetzung.** `go build`, `go vet` und `gofmt` laufen sauber über die Annotationsdatei.

**Fehlerklassen** — alle erzeugen echte Compilerfehler mit exakter Position:

| Verletzung | Meldung des Go-Compilers |
|---|---|
| Tippfehler im Anforderungsbezeichner | `commands.annotation.go:9:23: undefined: quote.R_QUOTE_SUBMITT` |
| Tippfehler im Go-Typ | `commands.annotation.go:10:18: undefined: QuoteSubmittd` |
| Bindungsziel existiert nicht | `commands.annotation.go:8:18: undefined: SubmitQuoteUseCase` |
| unbekanntes Feld im Literal | `R-QUOTE-SUBMIT.spec.go:7:2: unknown field Kindd in struct literal of type spec.Requirement` |
| **Anforderung gelöscht** | Übersetzung des Produktivpakets bricht — §2.4 (1) bestätigt |

**Statische Extraktion.** Ein Leser über `go/packages` (Modus `NeedSyntax|NeedTypes|NeedTypesInfo|NeedDeps`) gewinnt alle Fakten aus dem getypten AST:

```
Datei: sales/commands.annotation.go
  Bindung For     -> Ziel spike/sales.SubmitQuoteUC
    satisfies quote.R_QUOTE_SUBMIT  (deklariert: anforderungen/fun/quote/R-QUOTE-SUBMIT.spec.go:5:5)
    transition on spike/sales.QuoteSubmitted -> "submitted"
  Bindung ForFunc -> Ziel NewSubmitQuote
  Bindung ForVar  -> Ziel &PermSubmitQuote
  Bindung ForField-> Ziel spike/sales.SubmitQuoteCmd
```

Typparameter werden über `types.Info.Instances` aufgelöst, Anforderungsreferenzen über `types.Info.Uses` samt Deklarationsposition im fremden Paket. Der Kern des Frontends ist damit kleiner als 100 Zeilen.

### 7.4 Kreuzprobe: `speclink selfreport`

Statische Analyse ist **maßgeblich**. Sie hat aber blinde Flecken, und weil das Laufzeitregister ohnehin vorhanden ist (§4.2), lässt sich eine zweite, unabhängige Sicht auf dasselbe Modell fast umsonst gewinnen.

#### Arbeitsteilung

| | statisch | `selfreport` |
|---|---|---|
| maßgeblich für | Diagnostik, K1–K4, Backends | nichts — reine Kontrolle |
| braucht | übersetzbare Pakete | lauffähige, verlinkte Pakete |
| liefert | Positionen, Anweisungsreihenfolge | Vollständigkeit über alle Build-Sichten |

Warum statisch die Grundlage bleibt: Diagnostik braucht Positionen; K4-Regeln wie „`subject.Audit` ist die erste Anweisung in `Decide`" sind zur Laufzeit unsichtbar; und im LLM-Loop gibt es oft übersetzbare Pakete, aber kein startbares Programm.

#### Was die Kreuzprobe erfasst

Nach dem CRUD-Verbot (§1.6) bleibt übrig:

| Divergenzquelle | erfasst |
|---|---|
| Build-Tag-abhängige Dateien — statisch wird *eine* Konfiguration geladen | ✓ |
| dynamische Registrierung im Zielcode (Schleife über Typen) | ✓ |
| Recognizer-Defekt: der statische Leser übersieht ein Konstrukt | ✓ |
| unerreichtes Paket, toter Code | ✓ |
| **Verdrahtung** — welche Module sind in der gebauten App aktiviert | ✗ (siehe unten) |

> **Ehrliche Ertragsschätzung: gering.** Der Wert liegt in dauerhafter Absicherung, nicht in einmaliger Entdeckung. Das soll die Dimensionierung bestimmen — es lohnt kein großes Werkzeug.

#### Aufbau

```
1. packages.Load("./...")     alle Pakete des Zielmoduls, `package main` aussortieren
2. main.go im Speicher erzeugen:
       package main
       import (
           _ "modul/werp/sales"
           _ "modul/werp/comment"
           …                      ein Blank-Import je Paket
           "github.com/worldiety/speclink/spec"
       )
       func main() { spec.DumpJSON(os.Stdout) }
3. Datei nach $TMPDIR schreiben, overlay.json erzeugen, das sie virtuell
   nach <modul>/.speclink/selfreport/main.go abbildet
4. go run -overlay=<tmp>/overlay.json ./.speclink/selfreport
5. stdout als JSON einlesen und vergleichen
```

Schritt 3 nutzt `go build -overlay` — damit wird **nichts in das Zielrepository geschrieben**. Es ist dasselbe Muster, das `spec/gen` heute von Hand anwendet: alle 37 Kontexte blank-importieren und die gefüllten Registries auslesen.

> **Unverifizierte Annahme.** Ob `-overlay` ein Paketverzeichnis akzeptiert, das nur virtuell existiert, ist nicht geprüft. Rückfall: eine temporäre, gitignorierte Datei im Zielmodul, die nach dem Lauf gelöscht wird. **Zu klären, bevor die Kreuzprobe gebaut wird** — analog zum Typcheck-Spike, der in §7.3 erledigt ist.

> **Versionsversatz.** Das Zielprojekt pinnt `speclink/spec` in seiner `go.mod`; der Entwickler führt ein beliebiges `speclink`-Binary aus. Beide Versionen können auseinanderlaufen. Das Dump-Format trägt deshalb eine Versionsangabe, und speclink bricht bei Nichtübereinstimmung mit einer klaren Meldung ab, statt einen unvollständigen Vergleich zu melden. Es ist der **einzige** Ort im Werkzeug mit echtem Versionsversatz.

#### Der Vergleich

Beide Seiten liefern Tupel gleicher Form: `(datei, zeile, zielart, ziel, aussageart, argumente)`.

| Differenz | Bedeutung |
|---|---|
| nur zur Laufzeit vorhanden | der statische Leser hat etwas übersehen — Recognizer-Defekt, andere Build-Sicht, dynamische Registrierung |
| nur statisch vorhanden | die Registrierung lief nicht — unerreichbares Paket, per Tag ausgeschlossen, toter Code |

Ausgabe im normalen Diagnostikformat mit Position. Kein eigener Report-Formalismus.

#### Verifiziert

Der Laufzeitpfad ist am selben Durchstich wie §7.3 geprüft. `spec` mit Registry, `runtime.Caller` und `DumpJSON`, ein `selfreport`-Programm mit Blank-Import.

**Positionen stimmen.** Alle vier Bindungsarten melden exakt die Zeile ihrer Annotationsdatei — dieselben Positionen, die der statische Leser liefert:

```json
{"File": ".../commands.annotation.go", "Line": "8",  "TargetKind": "type",  "Target": "sales.SubmitQuoteUC",
 "Assertions": ["satisfies:R-QUOTE-SUBMIT", "transition:QuoteSubmitted->submitted", "help"]}
{"File": ".../commands.annotation.go", "Line": "15", "TargetKind": "func",  "Target": "spike/sales.NewSubmitQuote"}
{"File": ".../commands.annotation.go", "Line": "17", "TargetKind": "field", "Target": "sales.SubmitQuoteCmd.Title"}
{"File": ".../commands.annotation.go", "Line": "16", "TargetKind": "var",   "Target": "*string"}
```

**`ForVar` verliert den Namen — bestätigt.** Das Ziel meldet sich als `*string`, nicht als `PermSubmitQuote`. Der Vergleich über die Position (§5.3) ist damit nicht Bequemlichkeit, sondern notwendig.

**Die Init-Reihenfolge ist nicht die Quelltextreihenfolge.** Die Einträge erscheinen als `8, 15, 17, 16` — über drei Läufe identisch, also deterministisch, aber textlich verdreht. Go initialisiert Paketvariablen nach Abhängigkeit, nicht nach Zeile.

> Daraus folgt eine harte Vorgabe für den Vergleich: **er muss auf Mengen arbeiten, geschlüsselt über die Position — nie auf Folgen.** Ein sequenzieller Vergleich würde reihenweise Falschmeldungen erzeugen.

#### Abgrenzung: Verdrahtungswahrheit

`package main` lässt sich nicht importieren. Die Verdrahtung liegt aber in `werp/cmd/erp/main.go` (406 Zeilen `Configure`-Aufrufe) und bleibt damit außen vor.

Sie zu erfassen bräuchte einen Eingriff im Zielprojekt — ein Flag, das nach dem Aufbau des `Configurator` das Modell ausgibt. Das beantwortet aber eine **andere Frage** („ist die zusammengebaute App konsistent?") als `selfreport` („ist das statische Modell vollständig?") und gehört deshalb in einen eigenen, späteren Befehl.

---

## 8. Validierungsphasen

Streng geschichtet, jede mit eigenem Diagnosecode-Präfix.

| Phase | Code | Prüft | Beispiel |
|---|---|---|---|
| **V1** | `SPEC-V1` | Knoten-Whitelist | `func` in einer Annotationsdatei; `if`; positionales Struct-Feld |
| **V2** | — | **Go-Übersetzung** | siehe §7.2; die Fehler kommen vom Go-Compiler |
| **V3** | `SPEC-V3` | Bindung | Aussage an unzulässigem Zieltyp; nicht wiederholbare Aussage doppelt; Deklaration in Annotationsdatei; unbekannter Feldname bei `ForField`; verwaiste Annotationsdatei |
| **V4** | `SPEC-V4` | Redundanz gegen Inferenz | annotierte Tatsache ist ableitbar |
| **V5** | `SPEC-V5` | Auflösung | `Source.Doc` existiert nicht; `Anchor` nicht auflösbar; `Detail`-Datei fehlt; unbekannte `RuleID`; Zyklus in `DerivedFrom`; Pfad, ID-Präfix und `Kind` widersprechen sich |
| **V6** | `SPEC-V6` | Semantik K1–K4 | Abdeckung, Beziehungsintegrität, Architekturmuster |

V2 ist keine speclink-Phase, sondern eine Build-Stufe: `go build ./...` läuft ohnehin vor speclink. Schlägt sie fehl, gibt es kein Annotationsfeedback — der Loop-Runner muss das wissen und entsprechend priorisieren (Konzept §5.3).

### 8.1 Regelidentität

Jede V6-Regel trägt eine stabile `RuleID` (`K3-REQ-UNCOVERED`, `K4-QUERY-SUBJECT`). Regeln sind Go-Code, aber einzeln abschaltbar und per `spec.Waive(rule, reason)` mit Begründungspflicht aussetzbar — das Muster der bestehenden `readSideUnfoldedAllowlist` im Zielprojekt.

### 8.2 `K4-NO-GENERIC-CRUD`

Die einzige Regel, die nicht Konsistenz prüft, sondern **Konstrukte verbietet**. Sie setzt P9 durch (§1.6).

**Verboten im Zielcode:**

| Konstrukt | erzeugt zur Laufzeit |
|---|---|
| `cfgent.Enable[T, ID](cfg, prefix, entityName, opts)` | 6 Permissions nach Namensregel, Repository, 3 Routen |
| `ent.DeclarePermissions[T, ID](prefix, entityName)` | 6 Permission-IDs aus einem Präfix |
| `ent.NewUseCases[T, ID](perms, repo, opts)` | 6 Use Cases |
| `uient.Pages` und die generische CRUD-UI | Ansichten ohne deklarierte Bindung |

Rein syntaktisch prüfbar: es genügt, die Aufrufstelle über `types.Info.Uses` zu erkennen. Kosten vernachlässigbar.

```
werp/person/cfg.go:24:12: [SPEC-V6-K4-NO-GENERIC-CRUD] Generische CRUD-Fabriken sind nicht zulässig.
    cfgent.Enable erzeugt sechs Permissions, ein Repository und drei Routen erst zur
    Laufzeit aus dem Präfix "my.person". Diese Fakten sind statisch nicht sichtbar,
    und das Modul ist nur als Ganzes anpassbar (P9: Analysierbarkeit vor Kürze).
    Schreibe die Use Cases einzeln aus: je einen benannten Funktionstyp mit
    auth.Subject als erstem Parameter, je ein permission.Declare[UC](…) und je ein
    subject.Audit(…) als erste Anweisung. Vorbild: werp/sales/commands.go.
```

Die Regel ist wie jede andere per `spec.Waive` mit Begründung aussetzbar — der Notausstieg bleibt, aber er hinterlässt eine Spur im Lückenbericht.

---

## 9. Diagnostik

Die Fehlerausgabe wird überwiegend von einem LLM konsumiert (P7). Sie ist deshalb **präskriptiv**, nicht deskriptiv.

### 9.1 Aufbau

```
<datei>:<zeile>:<spalte>: [<code>] <was ist falsch>
    <warum es falsch ist — Regel oder Prinzip>
    <was konkret zu tun ist>
```

Vorbild ist die bestehende Meldung aus `TestProducesIsOwnAggregate` im Zielprojekt: *„Fremd-Aggregat gehört in Emits, nicht Produces."*

### 9.2 Whitelist-Verstoß

Die erwartbar häufigste Fehlerklasse, weil die Datei aussieht wie gewöhnliches Go:

```
werp/sales/commands.annotation.go:22:1: [SPEC-V1-001] Funktionsdefinitionen sind in Annotationsdateien nicht zulässig.
    *.annotation.go trägt ausschließlich Bindungsterme. Die Whitelist ist die einzige
    Schranke dagegen, dass hier eine zweite Codebasis entsteht (P4).
    Verschiebe die Funktion nach commands.go oder entferne sie.
```

### 9.3 Redundanz (V4)

```
werp/sales/commands.annotation.go:9:2: [SPEC-V4-002] Die Rolle ist bereits ableitbar und darf nicht annotiert werden.
    permission.Declare[ApproveQuoteUC] in commands.go:29 bindet die Permission an den
    Use-Case-Typ; die Rollenzuordnung folgt daraus über das IAM
    (P2: genau eine Quelle je Tatsache).
    Entferne spec.Role(…) aus dem Bindungsterm.
```

### 9.4 Ausgabeformen

- **Text** — `datei:zeile:spalte: [code] …`, copy-paste-fähig, für Menschen und Editoren.
- **JSON** — stabiles, versioniertes Schema mit Code, Position, Konstrukt, Anforderungskontext und Handlungsvorschlag als getrennte Felder. Für den LLM-Loop.

Beide entstehen aus derselben Quelle. Meldungstexte werden wie eine API-Oberfläche behandelt und mit Golden-Tests festgehalten.

---

## 10. Anforderungsbaum

### 10.1 Ablage

```
anforderungen/
  dec/                                    Kind = Decision          → Paket dec
    R-DEC-EVENTSOURCING.spec.go
    R-DEC-NUMBERING-REGISTRY.spec.go
  nfr/                                    Kind = NonFunctional     → Paket nfr
  con/                                    Kind = Constraint        → Paket con
  fun/                                    Kind = Functional
    quote/                                                          → Paket quote
      doc.go                              Gruppenbeschreibung
      R-QUOTE-SUBMIT.spec.go
      R-QUOTE-SUBMIT/                     Anhänge; kein .go → Go ignoriert es
        akzeptanzkriterien.md
        abnahmeprotokoll.pdf
      R-QUOTE-EDITOR.spec.go
      R-QUOTE-EDITOR/
        mockup-angebotseditor.png         geteiltes Entwurfsmaterial (§10.5)
      _material/                          Rückfall für geteiltes Material
        prozess-angebotsfreigabe.svg
    shop/
  _quellen/                               `_` → vom Go-Werkzeug ignoriert
    worldiety/vertraege/vertraege.md      Rohquellen, unverändert
    monday/potentials/potentials_form.png
```

### 10.2 Erste Achse: `Kind`

`spec/requirement.go` des Bestands nennt es bereits so: *„RequirementKind klassifiziert, WAS eine Anforderung ist (erste Achse)."* Die Konvention existiert dort vollständig und ausnahmslos, nur unaufgeschrieben und ungeprüft:

| Präfix | Anzahl | `Kind` | Ausnahmen |
|---|---:|---|---|
| `R-DEC-` | 125 | `Decision` | 0 |
| `R-NFR-` | 10 | `NonFunctional` | 0 |
| `R-CON-` | 7 | `Constraint` | 0 |
| Domänenpräfix | 339 | `Functional` | 0 |

Für die querschnittlichen Arten ist der Typ die natürliche Gruppierung, weil sie definitionsgemäß keine fachliche Heimat haben — alle 142 stehen im Bestand in *einer* Datei. Für `Functional` (70 %) trägt er nicht; dort ist die Domäne die zweite Achse. `fun/` selbst ist reine Gliederung ohne eigenes Paket: der Import bleibt `quote`, nicht `fun/quote`.

speclink prüft die Übereinstimmung von Verzeichnis, ID-Präfix und `Kind`-Feld. Wie bei ID und Dateiname gilt: **Redundanz ist unschädlich, sobald sie geprüft wird** — ungeprüfte Redundanz ist M1.

Ändert sich der Typ einer Anforderung, ändern sich Pfad und ID. Das ist eine Identitätsänderung und erzwingt `Supersedes`. Korrekt so: eine funktionale Anforderung, die sich als Entscheidung entpuppt, *ist* ein anderer Knoten.

### 10.3 Identität (F3)

- Die ID steht im Term. Dateiname und Anhangsordner tragen sie ebenfalls; speclink gleicht ab.
- Schema `R-<PRÄFIX>-<NAME>` — das im Bestand bereits verwendete Muster wird zur Norm erhoben.
- Umbenennung erfolgt nie stillschweigend, sondern über einen neuen Knoten mit `Supersedes`; der alte erhält `Status: spec.Superseded`.
- `DerivedFrom` und `Supersedes` spannen einen gerichteten azyklischen Graphen auf. Das Verzeichnis ist Ablageordnung, der DAG das Beziehungsmodell (Konzept §5.1).
- Referenziert wird über den **Go-Bezeichner** (`quote.R_QUOTE_SUBMIT`), nicht über Pfad oder ID-String. Umsortieren bricht deshalb keine Referenz.

### 10.4 Die Außenkante wird prüfbar

`Source.Doc` muss auf eine existierende Datei unter `_quellen/` zeigen, `Source.Anchor` auf eine vorhandene Überschrift des Zieldokuments (Slug-Regel, §5.2). Genau eines von `Doc` und `Extern` ist gesetzt. Damit endet M2.

Warum das nötig ist, zeigt der Ausgangszustand im ERP: 553 `Source:`-Belegungen, davon 77 leer, in vier koexistierenden Formatvarianten (`"anforderungen/x/y.md §2.3"`, `"vertraege.md §7"`, `"kandidaten §4.8"`, `"HGB §§ 383 ff."`), teils auf Verzeichnisse statt Dateien zeigend. **Kein einziger Test prüft sie** — die teuerste manuelle Übersetzung im ganzen Prozess ist die einzige ohne Verifikation.

Die Überführung eines solchen Altbestands in das hier definierte Format ist **Projektarbeit, nicht Aufgabe von speclink** (§1.7). speclink definiert das Format und prüft es; wer von Freitext darauf migriert, schreibt dafür ein Einmalskript in seinem eigenen Repository.

### 10.5 Materialien

Der **Ablageort folgt der Reichweite** (Duplikatvermeidung), die **Ablagezone der Herkunft** (Veränderlichkeit):

| | deckt genau eine Anforderung | deckt mehrere |
|---|---|---|
| **vorgefunden** (Altsystem, Kunde) | Anhangsordner | `_quellen/` |
| **erstellt** (Team, UX) | Anhangsordner | Elternknoten, sonst `_material/` |

- **Beleg** — Screenshot des Altsystems, Kunden-PDF, Prozessdiagramm. Existierte vor der Anforderung, deckt meist mehrere ab, wird nie geändert. Liegt in `_quellen/` in Herkunftsstruktur und wird über `Sources` referenziert. Ein Screenshot liegt deshalb **nie** in einem Anforderungsordner.
- **Eigenmaterial** — Akzeptanzkriterien, Abnahmeprotokoll. Entstand wegen dieser einen Anforderung. Liegt in ihrem Anhangsordner.
- **Entwurfsmaterial** — Mockup, Scribble, neu gezeichnetes Ablaufdiagramm. Erstellt, entwickelt sich, deckt meist mehrere Anforderungen ab.

Für geteiltes Entwurfsmaterial gilt: **bevorzugt trägt es der gemeinsame Elternknoten.** Ein Mockup eines ganzen Bildschirms entspricht fast immer einer gröberen Anforderung, von der die feinen per `DerivedFrom` abgeleitet sind — dann ist es deren Eigenmaterial und die Kinder erreichen es über den DAG. `_material/` auf Gruppenebene ist der Rückfall, wenn es keinen solchen Elternknoten gibt und einer nur zum Hosten einer Datei erfunden würde. Künstliche Knoten verfälschen die Abdeckungszahlen.

### 10.6 Zwei Textebenen

`Text` trägt den normativen Satz — kurz, für Listen, Matrizen und Diagnosemeldungen. `Detail` verweist auf eine Markdown-Datei im Anhangsordner für Tabellen, Akzeptanzkriterien und Prozessbeschreibungen. Grund: die Rohquellen enthalten reichlich Langform, die als Go-String unerträglich wäre — `vertraege.md §8.1` allein ist eine zwölfzeilige Tabelle.

### 10.7 Rückwärtsabdeckung (K3)

Jede Anforderung mit `Status: spec.Normative` muss von mindestens einem Konstrukt referenziert sein. `Abstract`, `Planned`, `OutOfScope`, `Informative` und `Superseded` sind die explizit markierten Ausnahmen. Damit wird aus „ich hoffe, es ist alles umgesetzt" eine Zahl, und handgeführte Statustabellen werden überflüssig.

---

## 11. Zielbild: `werp/sales`

### 11.1 Ausgangslage — vier Pflegestellen

Dieselbe Tatsache „bei Abgabe wird eine Angebotsnummer gezogen" steht heute an vier Orten:

1. `anforderungen/worldiety/angebote/angebotsflow.md §8` — Fließtext, keine ID
2. `spec/contexts/sales/requirements.go` — `R_QUOTE_VERSION`, `Source: "vertraege.md §10"`
3. `spec/genwerp/main.go` — Eintrag `sales.quote` in der 57-zeiligen Konfigurationstabelle
4. `werp/sales/requirements.go` + `werp/sales/knowledge.go` — `ReqSubmit`, Volltext wörtlich dupliziert

### 11.2 Belegter Drift

Die ID **`R-QUOTE-SUBMIT` existiert ausschließlich in `werp/`** — im normativen `spec/`-Modell gibt es sie nicht. Sie wurde in der Implementierung erfunden, ist in `werp/knowledge/out/knowledge.jsonld` und `KNOWLEDGE.md` exportiert und wird dort als erfüllt ausgewiesen. Kein Test bemerkt das, weil die beiden ID-Räume nie verglichen werden.

Genau dieser Abgleich — inferierte Fakten aus `werp/` gegen das normative `spec/`-Modell — ist das Abnahmekriterium des Go-Frontends.

### 11.3 Nachher

`anforderungen/fun/quote/R-QUOTE-SUBMIT.spec.go`:

```go
package quote

import (
	"github.com/worldiety/speclink/spec"
	"gitlab.worldiety.net/worldiety-erp/spec.git/anforderungen/dec"
)

var R_QUOTE_SUBMIT = spec.Requirement{
	ID:         "R-QUOTE-SUBMIT",
	Kind:       spec.Functional,
	Discipline: spec.Fachlich,
	Status:     spec.Normative,
	Title:      "Angebotsnummer bei Abgabe",
	Text: `Bei der Abgabe eines freigegebenen Angebots MUSS eine fortlaufende,
dublettenfreie Angebotsnummer aus dem zentralen Nummernkreis gezogen werden;
die Nummer wird im QuoteSubmitted-Fakt geführt.`,
	DerivedFrom: []spec.Requirement{dec.R_DEC_NUMBERING_REGISTRY},
	Sources: []spec.Source{
		{Doc: "anforderungen/_quellen/worldiety/angebote/angebotsflow.md", Anchor: "8-abgabe"},
	},
}
```

`werp/sales/commands.annotation.go`:

```go
package sales

import (
	"github.com/worldiety/speclink/spec"
	"gitlab.worldiety.net/worldiety-erp/spec.git/anforderungen/fun/quote"
)

var _ = spec.For[SubmitQuoteUC](
	spec.Satisfies(quote.R_QUOTE_SUBMIT),
	spec.Transition[QuoteSubmitted]("submitted"),
	spec.Help(`Gib das freigegebene Angebot ab. Das System zieht die nächste
Angebotsnummer aus dem Nummernkreis.`),
)
```

`werp/sales/commands.go` bleibt **unverändert**.

### 11.4 Was entfällt — und warum

Aus dem realen Code inferiert, deshalb **verboten** zu annotieren:

| Tatsache | Quelle der Inferenz (real, `werp/sales`) |
|---|---|
| ist ein Use Case | benannter Functype, `auth.Subject` als erster Parameter — `commands.go:19` |
| Permission `sales.quote.submit` | `permission.Declare[SubmitQuoteUC](…)` — `commands.go:30` |
| Berechtigung wird durchgesetzt | `s.Audit(PermSubmitQuote)` als erste Anweisung — `usecases.go:56` |
| Kontext `sales` | Paketzugehörigkeit |
| Aggregat `sales.quote` | `*QuoteAggregate` im `Decide`-Empfänger — `commands.go:115` |
| produziert `QuoteSubmitted` | Rückgabe von `Decide` — `commands.go:122` |
| Konstruktor gehört zum Use-Case-Typ | `var _ SubmitQuoteUC = NewSubmitQuote(nil, nil)` — `usecases.go:112` |
| Abhängigkeit zu `numbering` | Parameter `allocate numbering.Allocate` — `usecases.go:54` |

Bilanz für diesen einen Use Case: aus einem Eintrag in `spec/contexts/sales/usecases.go`, drei Konstanten in `werp/sales/requirements.go`, einem `RegisterUseCase`-Aufruf in `werp/sales/knowledge.go` und einer Zeile in der `genwerp`-Tabelle werden **ein Bindungsterm** plus ein Knoten im Anforderungsbaum.

---

## 12. Bewusst ausgeschlossen

| | Begründung |
|---|---|
| **Kommentar-Subsprache** | erzwänge Extraktion, synthetische Datei, Positionsabbildung, eigenen Typcheck-Lauf; verlöre IDE-Refactoring, `gofmt`, `go vet` (§4.3) |
| **Prozedurale Makros / Codeerzeugung** | dreht die Richtung um (Code→Code statt Code→Fakten), bringt Expansionsreihenfolge, Fixpunkt und Hygiene, und holt den Codegenerator zurück, den Kap. 4 abgeschafft hat |
| **Prädikatsprache** | wäre der Ort für F5 „ausführbare Annotationen"; erhöht die Sprachkomplexität deutlich und kollidiert mit P4. Nachrüstbar ohne Migration |
| **Schleifen, Rekursion, Ausdrücke** | opfern die Totalitätsgarantie (§3.4) |
| **Build-Tag-Ausschluss** | wäre rückwirkungsfrei, verlöre aber Build-Kopplung, Laufzeitverfügbarkeit und IDE-Unterstützung (§2.4) |
| **Generische CRUD-Fabriken im Zielcode** | erzeugen Fakten zur Laufzeit und Module, die nur als Ganzes anpassbar sind. Verboten statt nachgebaut (P9, §1.6; Regel §8.2) |
| **Nachbau von Framework-Namensregeln** | speclink würde `<prefix>.create` & Co. modellieren müssen; ändert nago die Regel, lügt speclink still. Durch das CRUD-Verbot gegenstandslos |
| **Mehrsprachigkeit der Fachtexte** (F9) | ungelöst, siehe §13 |
| **K4 über Datenflussanalyse** | Kostenwarnung Konzept §5.4: nur bei nachgewiesenem Bedarf, einzeln nachrüsten |

---

## 13. Risiken und offene Punkte

### 13.1 Risiken

| Risiko | Gegenmittel |
|---|---|
| **Annotationsdateien werden zur zweiten Codebasis.** Es sind echte, normal übersetzte Go-Dateien ohne syntaktische Schranke. | Whitelist-Durchsetzung (§3) mit eigenen Regel-IDs; sie ist tragend, nicht formal |
| **Produktivbuild hängt an der Spezifikation.** Eine gebrochene Annotation blockiert die Auslieferung. | bewusst gewählt (§2.4); Notausstieg ist `spec.Waive` mit Begründungspflicht |
| **Die Verweisstelle in Bildern ist nicht prüfbar.** `Source.Note` ist Freitext — die verbliebene Restlücke von M2. | als bewusste Lücke geführt; Bildregionen oder Callout-Dateien wären möglich, kosten aber spürbaren Pflegeaufwand bei 91 Bildern |
| **Feldnamen bei `ForField` sind Strings.** Einzige nicht compilergeprüfte Referenz der Sprache. | Prüfung durch speclink (§6.5); Feldmetadaten bleiben ohnehin Struct-Tags |
| **speclink erzwingt einen Codierstil, nicht nur Konsistenz.** `K4-NO-GENERIC-CRUD` (§8.2) verbietet Framework-Konstrukte, statt Widersprüche zu melden. Das ist eine Ausweitung des Mandats vom Verifier zum Stilwächter. | bewusst und begründet (P9, §1.6); empirisch kostenlos (0 Nutzungen im Bestand); `spec.Waive` bleibt als begründungspflichtiger Notausstieg |
| **`selfreport` führt alle `init()` aus.** Der Blank-Import jedes Pakets löst sämtliche Paketinitialisierung aus. Ein Paket, das dabei eine Verbindung öffnet oder Dateien anlegt, wäre ein Problem. | im ERP vermutlich harmlos (im Wesentlichen `permission.Declare`), aber **vor dem Bau der Kreuzprobe zu prüfen** |
| **Die Laufzeitregistry wird zur zweiten Wahrheit.** Wenn sie je Korrektheitsgrundlage statt Kreuzprobe wird, haben wir zwei Modelle statt einem — genau M1. | statisch ist maßgeblich (§4.2, §7.4); die Registry darf ausschließlich vergleichen, nie entscheiden |

### 13.2 Offene Punkte

1. **F9 — Mehrsprachigkeit.** `Help` und `Term` tragen Fachtext für Doku und Hilfesystem. Wie werden Übersetzungen gehalten, ohne die Ko-Lokalisierung zu brechen? nago bringt `i18n.Bundle` und Label-Schlüssel (`label:"nago.common.label.name"`) mit — möglicher Anknüpfungspunkt.
2. **Modulschnitt.** Der Anforderungsbaum wird zu einem Go-Paketbaum, den `werp` importiert. Ob er im selben Modul liegt oder als eigenes, ist eine Projektentscheidung mit Auswirkung auf Versionierung und Release.
3. **Bestandscode (F7).** Überführung paketweise über den Geltungsbereich (§1.8). Offen ist die Form der Konfiguration — Positivliste, Negativliste oder ein Marker im Paket.
4. **Reviewmodell (F8).** Wenn der Code Artefakt ist: wird sein Diff noch reviewt oder nur der von Annotationen und Anforderungen? Beeinflusst die Ausgestaltung der Backends.
5. **Verdrahtungswahrheit.** `selfreport` erfasst die Modellvollständigkeit, nicht die Frage, welche Module in der gebauten App aktiviert sind (§7.4). Ob dafür ein eigener Befehl mit Eingriff im Zielprojekt entsteht, ist offen.

---

## Anhang — Prinzipienbezug

| Prinzip | Umsetzung in diesem Entwurf |
|---|---|
| P1 Ko-Lokalisierung | Nebendatei im selben Paket. Kein eigener Lebenszyklus, kein eigener Review, kein eigener Commit, Drift bricht den Build — vier von fünf Kriterien des Konzepts erfüllt. Nicht erfüllt: dieselbe Datei (§2.4) |
| P2 eine Quelle je Tatsache | Vorrangregel Inferenz vor Annotation, V4 als Fehler (§1.2, §8) |
| P3 Referenz statt Wiederholung | Typparameter und Go-Bezeichner, vom Go-Typechecker aufgelöst — einzige Ausnahme `ForField` (§5.6, §6.5, §7.2) |
| P4 geschlossene Sprache | Knoten-Whitelist statt „gültiges Go", Totalitätsgarantie (§3) |
| P5 Deklaration vor Inferenz | Recognizer nur für belegte Framework-Marker; K4 nur bei nachgewiesenem Bedarf (§1.3, §12) |
| P6 beidseitige Abdeckung | `Status` steuert die Rückwärtsrichtung (§10.7) |
| P7 Diagnostik ist Schnittstelle | präskriptiver Meldungsaufbau, Text und JSON aus einer Quelle (§9) |
| P8 Ableitung ist Einbahnstraße | Backends erzeugen nur; kein Ausgabeartefakt wird nachbearbeitet |
| **P9 Analysierbarkeit vor Kürze** *(neu, folgt aus der zweiten These des Konzepts)* | Konstrukte, die Spezifikationsfakten zur Laufzeit erzeugen statt sie zu deklarieren, sind verboten. Durchgesetzt als `K4-NO-GENERIC-CRUD` (§1.6, §8.2) |
