package main

import (
	"strings"
	"testing"
)

// A specification is read by the person furthest from the code, and a chapter
// that is not there tells them nothing at all. Three different facts used to
// print as the same blank page: nothing was declared, nothing exists, or this
// frontend cannot read them. Only the first two are about the system.

// TestTheDocumentSaysWhatItDoesNotContain is the case that made this worth
// fixing.
//
// The JVM frontend reads no topology, no processes and no routes. A Spring
// application is very little other than routes, and its specification came out
// with no boundary chapter, no courses of business and no surface — reading
// exactly like a system that has none of those things. An auditor cannot tell,
// and the auditor is who this document is for.
func TestTheDocumentSaysWhatItDoesNotContain(t *testing.T) {
	t.Parallel()
	out, _ := runSpeclink(t, "generate", "../../testdata/java", "./...")

	for _, want := range []string{
		"## The boundary\n\n_Not measured: this frontend reads no topology declarations",
		"## What answers from outside\n\n_Not measured: this frontend recognises no routes",
		"## Courses of business\n\n_Not measured: this frontend reads no process declarations",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the document omits a chapter without saying so:\n  wanted %q", want)
		}
	}
}

// TestAnEmptyChapterIsNotAnUnreadableOne keeps the other half honest.
//
// The frameworkless fixture declares no course of business, and that is a fact
// about the project rather than about the tool. It has to read as one.
func TestAnEmptyChapterIsNotAnUnreadableOne(t *testing.T) {
	t.Parallel()
	out, _ := runSpeclink(t, "generate", "../../testdata/bare", "./...")

	if !strings.Contains(out, "## Courses of business\n\n_No course of business is declared") {
		t.Errorf("an undeclared chapter must say it was undeclared:\n%s", chaptersOf(out))
	}
	if strings.Contains(out, "Not measured") {
		t.Errorf("a frontend that can read every chapter reported that it cannot:\n%s", chaptersOf(out))
	}
}

// TestEveryChapterIsAlwaysPresent is what makes two generations comparable.
//
// The document keeps its shape whether or not it has content for a section, so
// a diff between two runs shows what changed rather than what moved.
func TestEveryChapterIsAlwaysPresent(t *testing.T) {
	t.Parallel()
	chapters := []string{
		"## Where it stands",
		"## Gaps",
		"## The material",
		"## Themes",
		"## Standards",
		"## The boundary",
		"## What answers from outside",
		"## Courses of business",
		"## Requirements",
	}
	for _, fixture := range []string{"../../testdata/example", "../../testdata/bare", "../../testdata/java"} {
		out, _ := runSpeclink(t, "generate", fixture, "./...")
		for _, c := range chapters {
			if !strings.Contains(out, c) {
				t.Errorf("%s is missing %q entirely", fixture, c)
			}
		}
	}
}

func chaptersOf(doc string) string {
	var out []string
	for _, line := range strings.Split(doc, "\n") {
		if strings.HasPrefix(line, "## ") || strings.HasPrefix(line, "_") {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}
