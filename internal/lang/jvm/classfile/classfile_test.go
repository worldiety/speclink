package classfile

import (
	"os"
	"path/filepath"
	"testing"
)

func read(t *testing.T, rel string) *Class {
	t.Helper()

	path := filepath.Join("..", "..", "..", "..", "testdata", "java", "classes", rel)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v — run testdata/java/build.sh", err)
	}
	c, err := Read(data)
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return c
}

func TestReadsTheShapeOfAClass(t *testing.T) {
	c := read(t, "com/example/sales/SubmitQuote.class")

	if c.Name != "com.example.sales.SubmitQuote" {
		t.Errorf("name is %q", c.Name)
	}
	if c.Access&AccInterface == 0 {
		t.Error("an interface was not reported as one")
	}
	if c.SourceFile != "SubmitQuote.java" {
		t.Errorf("source file is %q", c.SourceFile)
	}
	// --release 17 produces major 61. The reader does not depend on it, but a
	// fixture silently rebuilt against a different toolchain would change what
	// every other test here is measuring.
	if c.Major != 61 {
		t.Errorf("class file major version is %d, want 61; was the fixture rebuilt with a different --release?", c.Major)
	}
	if len(c.Methods) != 1 || c.Methods[0].Name != "submit" {
		t.Errorf("methods are %+v", c.Methods)
	}
}

// The whole design rests on this: a requirement is named by a class literal, so
// the reference is checked by the Java compiler rather than by speclink. With a
// string it would be an unverified reference, which is the one thing this tool
// exists to remove.
func TestClassLiteralArgumentResolves(t *testing.T) {
	c := read(t, "com/example/sales/SubmitQuote.class")

	a, ok := annotation(c.Annotations, "speclink.Satisfies")
	if !ok {
		t.Fatalf("no Satisfies annotation among %v", types(c.Annotations))
	}
	v, ok := a.Values["value"]
	if !ok || len(v.Array) != 1 {
		t.Fatalf("value is %+v", v)
	}
	if got := v.Array[0]; got.Kind != 'c' || got.Class != "com.example.requirements.fun.quote.RQuoteSubmit" {
		t.Errorf("class literal read as %+v", got)
	}
}

// Java's default retention is CLASS, so an annotation written without an
// explicit @Retention lands in RuntimeInvisibleAnnotations. A reader taking
// only the visible table would find nothing at all here — and nothing on
// Android either, where Room, Compose and Hilt all sit in the invisible one.
func TestInvisibleAnnotationsAreRead(t *testing.T) {
	c := read(t, "com/example/requirements/dec/RDecNumbering.class")

	a, ok := annotation(c.Annotations, "speclink.Requirement")
	if !ok {
		t.Fatalf("the requirement annotation was not found among %v", types(c.Annotations))
	}
	if a.Visible {
		t.Error("an annotation with default retention was reported as runtime visible")
	}
	if got := a.Values["id"].String(); got != "R-DEC-NUMBERING" {
		t.Errorf("id is %q", got)
	}
	if got := a.Values["kind"]; got.Kind != 'e' || got.EnumConst != "DECISION" {
		t.Errorf("kind read as %+v", got)
	}
}

// Arguments left at their default are absent from the use site: the default
// lives on the annotation type. A reader that expected them present would treat
// every omitted field as an empty string that somebody wrote.
func TestDefaultedArgumentsAreAbsent(t *testing.T) {
	c := read(t, "com/example/requirements/fun/quote/RQuoteApprove.class")
	a, _ := annotation(c.Annotations, "speclink.Requirement")

	if _, present := a.Values["rationale"]; present {
		t.Error("an argument left at its default was reported as given")
	}
	if _, present := a.Values["title"]; !present {
		t.Error("an argument that was given is missing")
	}
}

func TestNestedAnnotationAndArray(t *testing.T) {
	c := read(t, "com/example/requirements/fun/quote/RQuoteSubmit.class")
	a, _ := annotation(c.Annotations, "speclink.Requirement")

	sources, ok := a.Values["sources"]
	if !ok || len(sources.Array) != 1 {
		t.Fatalf("sources read as %+v", sources)
	}
	nested := sources.Array[0].Nested
	if nested == nil || nested.Type != "speclink.Source" {
		t.Fatalf("nested annotation read as %+v", sources.Array[0])
	}
	if got := nested.Values["anchor"].String(); got != "8-abgabe" {
		t.Errorf("anchor is %q", got)
	}
}

func TestEnumClassIsRecognisable(t *testing.T) {
	c := read(t, "speclink/Kind.class")

	if c.Access&AccEnum == 0 {
		t.Error("an enum was not flagged as one")
	}
	// Its constants are static fields of the enum's own type, plus the
	// synthetic $VALUES array nobody wrote.
	constants, synthetic := 0, 0
	for _, f := range c.Fields {
		if f.IsSynthetic() {
			synthetic++
			continue
		}
		constants++
	}
	if constants != 4 {
		t.Errorf("found %d declared constants, want 4", constants)
	}
	if synthetic == 0 {
		t.Error("the generated $VALUES field was not marked synthetic")
	}
}

// Not UTF-8. U+0000 is encoded as C0 80 and characters above the BMP as a
// surrogate pair of three byte sequences, so a naive string(bytes) survives
// ASCII — which is most identifiers — and then corrupts exactly the strings
// that carry meaning.
func TestModifiedUTF8(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   []byte
		want string
	}{
		{"ascii", []byte("R-QUOTE-SUBMIT"), "R-QUOTE-SUBMIT"},
		{"two byte", []byte{0xC3, 0x9C}, "Ü"},
		{"embedded nul", []byte{'a', 0xC0, 0x80, 'b'}, "a\x00b"},
		{"three byte", []byte{0xE2, 0x82, 0xAC}, "€"},
		// U+1F600, which real UTF-8 writes in four bytes and this encoding
		// writes as two three byte surrogates.
		{"surrogate pair", []byte{0xED, 0xA0, 0xBD, 0xED, 0xB8, 0x80}, "\U0001F600"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := decodeModifiedUTF8(tc.in)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}

	// A four byte sequence is ordinary UTF-8 and cannot occur here. Accepting
	// it would mean accepting input from somewhere this reader does not read.
	if _, err := decodeModifiedUTF8([]byte{0xF0, 0x9F, 0x98, 0x80}); err == nil {
		t.Error("real UTF-8 was accepted as modified UTF-8")
	}
}

// A Long or Double occupies two pool slots with the second unusable. Miss it
// and every index after the first one is off by one — which does not fail, it
// reads the wrong names.
func TestLongTakesTwoPoolSlots(t *testing.T) {
	b := &poolBuilder{}
	b.long(1)               // slots 1 and 2
	utf := b.utf8("marker") // slot 3

	pool := (&reader{b: b.bytes()}).readPool()
	if got, ok := pool.UTF8(utf); !ok || got != "marker" {
		t.Fatalf("entry after a long read as %q (index %d), want \"marker\"", got, utf)
	}
}

func TestRejectsWhatIsNotAClassFile(t *testing.T) {
	if _, err := Read([]byte("package com.example;")); err == nil {
		t.Error("Java source was accepted as a class file")
	}
	if _, err := Read(nil); err == nil {
		t.Error("an empty file was accepted")
	}
}

// poolBuilder writes a constant pool by hand, for the cases a compiler will not
// produce on demand.
type poolBuilder struct {
	buf   []byte
	count int
}

func (b *poolBuilder) long(v int64) {
	b.buf = append(b.buf, TagLong,
		byte(v>>56), byte(v>>48), byte(v>>40), byte(v>>32),
		byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
	b.count += 2
}

func (b *poolBuilder) utf8(s string) uint16 {
	b.buf = append(b.buf, TagUtf8, byte(len(s)>>8), byte(len(s)))
	b.buf = append(b.buf, s...)
	b.count++
	return uint16(b.count)
}

func (b *poolBuilder) bytes() []byte {
	n := b.count + 1
	return append([]byte{byte(n >> 8), byte(n)}, b.buf...)
}

func annotation(in []Annotation, typ string) (Annotation, bool) {
	for _, a := range in {
		if a.Type == typ {
			return a, true
		}
	}
	return Annotation{}, false
}

func types(in []Annotation) []string {
	out := make([]string, 0, len(in))
	for _, a := range in {
		out = append(out, a.Type)
	}
	return out
}
