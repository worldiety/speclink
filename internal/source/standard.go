package source

import (
	"bytes"
	"encoding/json"
	"os"
	"sort"
	"strings"
)

// StandardSuffix names a catalogue of the clauses of an external standard.
//
// A suffix rather than a plain extension, because .json is already the image
// manifest's and would be several other things besides. A file either announces
// itself as a clause catalogue or it is not one.
const StandardSuffix = ".standard.json"

// standardFile is the on disk form of a clause catalogue.
//
// It is speclink's own schema and a deliberately narrow one. Catalogues in the
// wild carry presentation with them — preambles, table headers, ordering hints
// — and every field of that sort would be a thing this tool appears to
// understand and does not. Converting a catalogue once is cheaper than the
// standing confusion.
type standardFile struct {
	ID     string        `json:"id"`
	Title  string        `json:"title"`
	Clause []clauseEntry `json:"clauses"`
}

type clauseEntry struct {
	Ref   string `json:"ref"`
	Title string `json:"title"`
	Text  string `json:"text,omitempty"`

	// Applicable defaults to true, which is why it is a pointer: the zero
	// value of a bool is false, and a clause that silently defaulted to
	// inapplicable would take itself out of the count without anybody saying
	// so. That is the one direction this whole file must not fail in.
	Applicable *bool  `json:"applicable,omitempty"`
	Because    string `json:"because,omitempty"`
}

// segmentStandard reads a clause catalogue into segments.
//
// A clause is a segment for the same reason a heading is: it is the smallest
// thing a requirement can point at, and the smallest thing whose wording can
// change under a requirement derived from it. Everything that follows — the
// coverage of a standard, the drift of a reworded clause, the citation that
// names a clause which no longer exists — is then the existing machinery
// working on a third kind of document rather than a second implementation of
// it.
//
// # Why the text may be missing
//
// The wording of ISO and DIN standards is not free to reproduce, and a
// catalogue that demanded it could not be committed. Reference and title are
// enough to enumerate the obligations and to say which are answered; the
// fingerprint covers whatever is there, so a project that may carry the text
// also gets drift on it, and one that may not still gets everything else.
func segmentStandard(doc, abs string) (string, []Segment, []error) {
	data, err := os.ReadFile(abs)
	if err != nil {
		return "", nil, []error{&SegmentError{
			Doc: doc,
			Msg: "clause catalogue cannot be read: " + err.Error(),
			How: "Check the file permissions.",
		}}
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	// Unknown fields are refused rather than ignored, on the same grounds as
	// speclink.json: a key with no effect is somebody expecting one.
	dec.DisallowUnknownFields()

	var f standardFile
	if err := dec.Decode(&f); err != nil {
		return "", nil, []error{&SegmentError{
			Doc: doc,
			Msg: "clause catalogue is not readable: " + err.Error(),
			Why: "The catalogue is the list of obligations a standard imposes. One that cannot be read leaves every one of them unaccounted for while looking like a source that is covered.",
			How: `Expected {"id":…, "title":…, "clauses":[{"ref":…, "title":…}]}.`,
		}}
	}

	var (
		segs []Segment
		errs []error
	)
	if strings.TrimSpace(f.ID) == "" || strings.TrimSpace(f.Title) == "" {
		errs = append(errs, &SegmentError{
			Doc: doc,
			Msg: "clause catalogue has no id or title",
			Why: "Both name the standard in every diagnostic and in the chapter derived from it.",
			How: `Set "id" to a short form such as C5 and "title" to the full name of the standard.`,
		})
	}

	seen := map[string]bool{}
	for i, c := range f.Clause {
		ref := strings.TrimSpace(c.Ref)
		if ref == "" {
			errs = append(errs, clauseErr(doc, i, "clause has no ref",
				"The reference is what a requirement cites. A clause without one cannot be pointed at and cannot be counted.",
				`Set "ref" to the identifier the standard uses, such as IAM-05.04B.`))
			continue
		}
		if seen[ref] {
			errs = append(errs, clauseErr(doc, i, "clause "+ref+" appears twice",
				"Two clauses under one reference make every citation of it ambiguous, and the count of what is answered wrong in both directions.",
				"Remove the duplicate."))
			continue
		}
		seen[ref] = true

		applicable := c.Applicable == nil || *c.Applicable
		if !applicable && strings.TrimSpace(c.Because) == "" {
			// The whole value of an exemption is the reason. Collected, these
			// are the statement of applicability, which ISO 27001 requires as
			// a document in its own right — and an entry in it that says only
			// "not applicable" is what an auditor asks about first.
			errs = append(errs, clauseErr(doc, i, "clause "+ref+" is excluded without a reason",
				"An exemption is a decision somebody made, and the reason is the whole of what it is worth. Without one the clause is not excluded, it is skipped.",
				`Set "because" to why this clause does not apply here.`))
			continue
		}

		segs = append(segs, Segment{
			Doc:  doc,
			ID:   ref,
			Kind: KindStandard,
			// The reference leads, because that is what a reader looks up and
			// what a citation carries. The title is the human half.
			Title:       strings.TrimSpace(or(c.Title, ref)),
			Fingerprint: fingerprint([]byte(ref + "\x00" + c.Title + "\x00" + c.Text)),
			Informative: !applicable,
			Because:     strings.TrimSpace(c.Because),
			Pos:         Pos{File: doc},
		})
	}

	sort.Slice(segs, func(i, j int) bool { return segs[i].ID < segs[j].ID })
	return strings.TrimSpace(or(f.Title, f.ID)), segs, errs
}

func clauseErr(doc string, i int, msg, why, how string) error {
	return &SegmentError{
		Doc: doc,
		Msg: msg + " (clause " + itoa(i+1) + ")",
		Why: why,
		How: how,
	}
}

func or(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}
