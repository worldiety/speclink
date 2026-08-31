package schema

import "testing"

// TestTheShapeGrammarIsReadBack covers every form speclink writes.
//
// The parser reads a grammar this project produces, so anything it cannot read
// is a defect here rather than in a project. That makes the round trip worth
// asserting for each form rather than for a sample.
func TestTheShapeGrammarIsReadBack(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		shape string
		kind  string
		check func(*testing.T, Type)
	}{
		{shape: "string", kind: "string"},
		{shape: "int", kind: "int"},
		{shape: "bool", kind: "bool"},
		{shape: "float64", kind: "number"},
		// Three different reasons for the same silence: an interface decides
		// its content elsewhere, an unread type was never resolved, and a
		// recursive one was cut off. None of them is a structure.
		{shape: "any", kind: "any"},
		{shape: "unknown", kind: "any"},
		{shape: "[]string", kind: "array", check: func(t *testing.T, ty Type) {
			if ty.Elem.Kind != "string" {
				t.Errorf("element is %q, want string", ty.Elem.Kind)
			}
		}},
		{shape: "map[string]int", kind: "map", check: func(t *testing.T, ty Type) {
			if ty.Elem.Kind != "int" {
				t.Errorf("value is %q, want int", ty.Elem.Kind)
			}
		}},
		{shape: "{a:string,b:int}", kind: "object", check: func(t *testing.T, ty Type) {
			if len(ty.Fields) != 2 || ty.Fields[0].Wire != "a" || ty.Fields[1].Wire != "b" {
				t.Errorf("got %#v, want fields a and b in order", ty.Fields)
			}
		}},
		// The one that breaks a naive split on commas.
		{shape: "{a:{b:string,c:int},d:[]bool}", kind: "object", check: func(t *testing.T, ty Type) {
			if len(ty.Fields) != 2 {
				t.Fatalf("got %d fields, want 2: %#v", len(ty.Fields), ty.Fields)
			}
			if ty.Fields[0].Type.Kind != "object" || len(ty.Fields[0].Type.Fields) != 2 {
				t.Errorf("the nested object was not read: %#v", ty.Fields[0])
			}
		}},
		{shape: "{}", kind: "object"},
	} {
		t.Run(tc.shape, func(t *testing.T) {
			t.Parallel()
			got, err := Parse(tc.shape)
			if err != nil {
				t.Fatalf("Parse(%q): %v", tc.shape, err)
			}
			if got.Kind != tc.kind {
				t.Fatalf("got kind %q, want %q", got.Kind, tc.kind)
			}
			if tc.check != nil {
				tc.check(t, got)
			}
		})
	}
}

// TestAnUnreadableShapeIsRefused keeps the parser from inventing structure.
//
// A schema derived from a shape nobody could read would be a confident
// description of nothing, which is worse than no file at all.
func TestAnUnreadableShapeIsRefused(t *testing.T) {
	t.Parallel()

	for _, shape := range []string{"chan int", "{a:string", "{a}", "string extra"} {
		if _, err := Parse(shape); err == nil {
			t.Errorf("Parse(%q) succeeded; an unreadable shape must be refused", shape)
		}
	}
}

// TestNestedObjectsClaimNothingAboutPresence is the honest half.
//
// speclink records whether a top level field may be omitted and never asks the
// question below that. A required list there would be invented, and a schema
// confidently wrong about presence turns into a parser that dereferences an
// absent value.
func TestNestedObjectsClaimNothingAboutPresence(t *testing.T) {
	t.Parallel()

	doc, err := Of(Shape{
		Type:     "example.com/x.Outer",
		Shape:    "{a:string,b:{c:string}}",
		Optional: map[string]bool{"a": true},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	req, _ := doc.Body["required"].([]string)
	if len(req) != 1 || req[0] != "b" {
		t.Errorf("top level required is %v, want only b — a is optional", req)
	}

	props := doc.Body["properties"].(map[string]any)
	inner := props["b"].(map[string]any)
	if _, claimed := inner["required"]; claimed {
		t.Error("the nested object claims a required list, which was never recorded")
	}
	if _, said := inner["$comment"]; !said {
		t.Error("the nested object does not say why it makes no claim")
	}
}
