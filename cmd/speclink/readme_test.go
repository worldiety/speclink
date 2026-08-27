package main

import (
	"os"
	"strings"
	"testing"
)

// TestReadmeListsEveryRule guards the rule index of the README.
//
// The README is the only document an agent reads, and it maps a finding to a
// fix. A rule that exists in the code but not in the index is therefore a rule
// nobody can act on: the agent sees an ID it cannot look up and has to guess.
//
// This is the same reasoning that makes the diagnostic texts public API. A new
// rule is not finished when it fires; it is finished when it can be answered.
func TestReadmeListsEveryRule(t *testing.T) {
	readme, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	text := string(readme)

	// Every rule ID the tool can emit, taken from the rule constants.
	rules := []string{
		"K1-CONSTRUCT-UNBOUND",
		"K10-REQ-CHANGED",
		"K11-REQ-UNSOURCED",
		"K11-SOURCE-UNANCHORED",
		"K12-SOURCE-UNCOVERED",
		"K13-SOURCE-DRIFT",
		"K14-REQ-UNVERIFIED",
		"K14-VERIFICATION-STALE",
		"K3-REQ-UNCOVERED",
		"K3-ABSTRACT-COVERED",
		"K3-SUPERSEDED-COVERED",
		"K4-NO-GENERIC-CRUD",
		"K5-UC-FILE",
		"K5-UC-CONSTRUCTOR",
		"K5-UC-SIGNATURE",
		"K5-UC-AUTHZ",
		"K5-UC-PERMISSION",
		"K5-UC-PERMISSION-I18N",
		"K5-UC-DEPS",
		"K6-CTX-NO-UI-IMPORT",
		"K6-CTX-UI-PKG",
		"K6-CTX-USECASES",
		"K7-INFRA-DOMAIN-FREE",
		"K8-MAIN-EXISTS",
		"K8-MAIN-LOCATION",
		"K9-DRAFT-REDUNDANT",
		"K9-BASELINE-MISSING",
		"K9-DISCRIMINATOR-FROZEN",
		"K9-FIELD-REMOVED",
		"K9-FIELD-RENAMED",
		"K9-TYPE-REMOVED",
		"K9-DRAFT-FROZEN",
		"K9-FIELD-SHAPE",
		"K9-OPTIONAL-REVOKED",
		"K9-FIELD-ADDED-REQUIRED",
		"K9-DISCRIMINATOR-COLLISION",
	}
	for _, rule := range rules {
		if !strings.Contains(text, rule) {
			t.Errorf("README does not document rule %s", rule)
		}
	}
}

// TestReadmeIsSelfContained guards the one editorial decision that matters for
// an agent: the README must not send it into the German design documents. Those
// are discussion drafts, not a contract, and following them produces code that
// the tool then rejects.
//
// There used to be German design records beside it, and the check was on how
// often the README pointed at them. They have been deleted: they described a
// tool that no longer existed, and a design record that is wrong is worse than
// none, because it is read with the same trust as a right one. What was still
// live in them — the reasoning behind having no severities, and the measured
// blockers of the reference ERP — was moved into the code and into section 12.
//
// The check now guards the absence, so that reintroducing a second document
// takes a deliberate decision rather than a commit.
func TestReadmeIsSelfContained(t *testing.T) {
	readme, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	text := string(readme)

	for _, doc := range []string{"docs/annotations.md", "docs/plan.md", "konzept-annotationscompiler.md"} {
		if strings.Contains(text, doc) {
			t.Errorf("README points at %s, which no longer exists", doc)
		}
	}
}
