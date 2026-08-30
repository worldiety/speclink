package source

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A standard is a source document whose segments are its clauses.
//
// Making it that rather than a mechanism of its own is the whole point: the
// coverage, the drift and the citation checks are then the ones that already
// exist, and an audit chapter is the ordinary machinery read from the other
// end rather than a second implementation written for auditors.

func write(t *testing.T, body string) (root, doc string) {
	t.Helper()
	root = t.TempDir()
	doc = "c5" + StandardSuffix
	if err := os.WriteFile(filepath.Join(root, doc), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return root, doc
}

func TestClausesBecomeSegments(t *testing.T) {
	root, doc := write(t, `{
	  "id": "C5", "title": "BSI C5",
	  "clauses": [
	    {"ref": "IAM-01", "title": "Berechtigung", "text": "Vor jeder Wirkung."},
	    {"ref": "PS-01", "title": "Zutritt", "applicable": false, "because": "kein Betrieb"}
	  ]}`)

	d := Load(root, doc)
	if d.Err != nil {
		t.Fatalf("catalogue did not load: %v", d.Err)
	}
	if d.Kind != KindStandard {
		t.Errorf("kind is %v, want standard", d.Kind)
	}
	if d.Title != "BSI C5" {
		t.Errorf("title is %q", d.Title)
	}
	if len(d.Segments) != 2 {
		t.Fatalf("got %d segments, want 2", len(d.Segments))
	}

	// The reference is what a requirement cites, so it is the segment ID.
	if d.Segments[0].ID != "IAM-01" {
		t.Errorf("clause reference did not become the segment id: %q", d.Segments[0].ID)
	}
	// An excluded clause is informative in exactly the sense the other kinds
	// use: not expected to produce a requirement.
	if !d.Segments[1].Informative {
		t.Error("an inapplicable clause is still counted as an obligation")
	}
	if d.Segments[1].Because != "kein Betrieb" {
		t.Errorf("the reason was lost: %q", d.Segments[1].Because)
	}
}

// The wording of ISO and DIN standards may not be reproduced, so a catalogue
// that demanded it could not be committed. Reference and title are enough to
// enumerate the obligations and say which are answered.
func TestClauseTextMayBeAbsent(t *testing.T) {
	root, doc := write(t, `{"id":"ISO","title":"ISO 27001",
	  "clauses":[{"ref":"A.8.2","title":"Klassifizierung von Information"}]}`)

	d := Load(root, doc)
	if d.Err != nil {
		t.Fatalf("a catalogue without clause text did not load: %v", d.Err)
	}
	if d.Segments[0].Fingerprint == "" {
		t.Error("a clause without text got no fingerprint, so drift on its title would go unseen")
	}
}

func TestCatalogueRefusesWhatItCannotMean(t *testing.T) {
	for _, tc := range []struct{ name, body, want string }{
		{
			// The whole value of an exemption is the reason. Collected, these
			// are the statement of applicability.
			name: "an exclusion without a reason",
			body: `{"id":"C5","title":"C5","clauses":[{"ref":"PS-01","title":"Zutritt","applicable":false}]}`,
			want: "excluded without a reason",
		},
		{
			// A key with no effect is somebody expecting one.
			name: "a field speclink does not understand",
			body: `{"id":"C5","title":"C5","clauses":[{"ref":"A","title":"B","severity":"high"}]}`,
			want: `unknown field "severity"`,
		},
		{
			name: "two clauses under one reference",
			body: `{"id":"C5","title":"C5","clauses":[{"ref":"A","title":"x"},{"ref":"A","title":"y"}]}`,
			want: "appears twice",
		},
		{
			name: "a clause with no reference",
			body: `{"id":"C5","title":"C5","clauses":[{"title":"x"}]}`,
			want: "has no ref",
		},
		{
			name: "a catalogue that names no standard",
			body: `{"clauses":[{"ref":"A","title":"x"}]}`,
			want: "no id or title",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, doc := write(t, tc.body)
			d := Load(root, doc)

			var got []string
			for _, e := range d.Errors() {
				got = append(got, e.Error())
			}
			joined := strings.Join(got, "\n")
			if !strings.Contains(joined, tc.want) {
				t.Errorf("expected %q, got:\n%s", tc.want, joined)
			}
		})
	}
}

// Applicability defaults to true, and the pointer is why.
//
// The zero value of a bool is false, so a plain field would make every clause
// that did not mention it exclude itself — silently taking the whole catalogue
// out of the count. That is the one direction this must not fail in.
func TestApplicabilityDefaultsToApplicable(t *testing.T) {
	root, doc := write(t, `{"id":"C5","title":"C5","clauses":[{"ref":"A","title":"x"}]}`)

	d := Load(root, doc)
	if d.Err != nil {
		t.Fatal(d.Err)
	}
	if d.Segments[0].Informative {
		t.Error("a clause that says nothing about applicability excluded itself")
	}
}

// A catalogue is a .json and so is an image manifest. Announcing itself is what
// makes it one.
func TestOnlyAnAnnouncedCatalogueIsOne(t *testing.T) {
	for _, tc := range []struct {
		path string
		want Kind
		ok   bool
	}{
		{"requirements/_sources/c5.standard.json", KindStandard, true},
		{"requirements/_sources/screen.png", KindImage, true},
		{"requirements/_sources/flow.md", KindMarkdown, true},
		{"requirements/_sources/anything.json", 0, false},
	} {
		got, ok := KindOf(tc.path)
		if ok != tc.ok || (ok && got != tc.want) {
			t.Errorf("KindOf(%q) = %v, %v; want %v, %v", tc.path, got, ok, tc.want, tc.ok)
		}
	}
}
