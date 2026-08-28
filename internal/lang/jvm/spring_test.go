package jvm

import (
	"strings"
	"testing"

	"github.com/worldiety/speclink/internal/ir"
)

// Inference is not a property of nago.
//
// The suspicion after the Go frontend was that recognisers only worked because
// that one framework's shapes happened to be legible — method sets, type
// aliases, a named func type with a particular first parameter. Spring declares
// its architecture in annotations instead, and annotations are exactly what a
// declaration level reader sees. This is the evidence either way.
func TestInfersSpringRoles(t *testing.T) {
	root, classes := fixture(t)
	r := NewReader(root, classes, nil, "")

	byName := map[string]ir.Construct{}
	for _, c := range r.Infer() {
		byName[c.Name] = c
	}

	for name, want := range map[string]ir.ConstructKind{
		"com.example.sales.QuoteController.submit":  ConstructEndpoint,
		"com.example.sales.QuoteController.approve": ConstructEndpoint,
		"com.example.sales.NumberRegistry.next":     ConstructService,
		"com.example.sales.Quote":                   ConstructEntity,
		"com.example.sales.QuoteRepository":         ConstructRepository,
	} {
		got, ok := byName[name]
		if !ok {
			t.Errorf("%s was not recognised at all", name)
			continue
		}
		if got.Kind != want {
			t.Errorf("%s recognised as %v, want %v", name, got.Kind, want)
		}
		if got.Evidence == "" {
			t.Errorf("%s carries no evidence, so a reader cannot see why speclink believes it", name)
		}
	}
}

// A method without a request mapping is not an endpoint, and a private method
// of a service is not an operation anybody asked for. Recognising them would
// make the rule fire on plumbing, which is how a rule teaches people to switch
// it off.
func TestPrivateAndUnmappedMethodsAreNotConstructs(t *testing.T) {
	root, classes := fixture(t)
	r := NewReader(root, classes, nil, "")

	for _, c := range r.Infer() {
		if strings.HasSuffix(c.Name, ".describe") || strings.HasSuffix(c.Name, ".peek") {
			t.Errorf("%s was recognised as %v", c.Name, c.Kind)
		}
	}
}

// Constructors, static initialisers and the bridges a compiler emits were
// written by nobody and can carry no intent.
func TestGeneratedMembersAreNotConstructs(t *testing.T) {
	root, classes := fixture(t)
	r := NewReader(root, classes, nil, "")

	for _, c := range r.Infer() {
		if strings.Contains(c.Name, "<") || strings.Contains(c.Name, "$") {
			t.Errorf("a generated member was recognised: %s", c.Name)
		}
	}
}

// The three questions the neutral rules ask have to be answered per role, and
// answering them wrongly is invisible until a rule fires somewhere it should
// not. An entity carries no requirement of its own; an endpoint does.
func TestRolesAnswerTheThreeQuestions(t *testing.T) {
	for _, tc := range []struct {
		kind                              ir.ConstructKind
		requirement, domainModel, storage bool
	}{
		{ConstructEndpoint, true, false, false},
		{ConstructService, true, false, false},
		{ConstructEntity, false, true, true},
		{ConstructRepository, false, false, true},
	} {
		t.Run(tc.kind.String(), func(t *testing.T) {
			if tc.kind.NeedsRequirement() != tc.requirement {
				t.Errorf("NeedsRequirement is %v", tc.kind.NeedsRequirement())
			}
			if tc.kind.IsDomainModel() != tc.domainModel {
				t.Errorf("IsDomainModel is %v", tc.kind.IsDomainModel())
			}
			if tc.kind.EmbodiesStorageDecision() != tc.storage {
				t.Errorf("EmbodiesStorageDecision is %v", tc.kind.EmbodiesStorageDecision())
			}
		})
	}
}

// An entity's fields have to arrive, because that is what the field level rules
// work from. The class file carries them; their lines do not exist there and
// come from the source.
func TestEntityCarriesItsFields(t *testing.T) {
	root, classes := fixture(t)
	r := NewReader(root, classes, nil, "")

	for _, c := range r.Infer() {
		if c.Name != "com.example.sales.Quote" {
			continue
		}
		if len(c.Fields) != 3 {
			t.Fatalf("read %d fields, want 3: %+v", len(c.Fields), c.Fields)
		}
		for _, f := range c.Fields {
			if f.Pos.Line == 0 {
				t.Errorf("field %s has no line, so a finding about it cannot be acted on", f.Name)
			}
		}
		return
	}
	t.Fatal("the entity was not recognised")
}

// TestFrameworkContractIsComplete pins that every name a recogniser matches on
// is listed with the role it carries.
//
// The coupling is by string and its failure is silence: a renamed annotation
// does not break a rule, it stops the rule matching anything. In the Go
// frontend the equivalent list had already drifted — one framework package was
// never in it, so a recogniser went unchecked from the day it was written.
func TestFrameworkContractIsComplete(t *testing.T) {
	listed := map[string]bool{}
	for _, s := range frameworkContract {
		listed[s.Name] = true
		if s.Breaks == "" {
			t.Errorf("%s does not say what it takes with it when it stops matching", s.Name)
		}
	}
	for _, m := range mappingAnnotations {
		if !listed[m] {
			t.Errorf("%s is matched on but named by no entry of frameworkContract", m)
		}
	}
	for _, name := range []string{
		springWeb + ".RestController",
		springStereotype + ".Controller",
		springStereotype + ".Service",
		springStereotype + ".Repository",
		jakartaPersist + ".Entity",
		javaxPersist + ".Entity",
	} {
		if !listed[name] {
			t.Errorf("%s is matched on but named by no entry of frameworkContract", name)
		}
	}
}
