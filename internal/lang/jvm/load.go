// Package jvm is the JVM frontend: it reads compiled classes and lowers them
// into the language neutral ir.
//
// It is the second frontend, and its job in the design is to disagree with the
// first. Go reads source through go/types, resolves everything the compiler
// resolved, and knows the exact position of every declaration. This reads
// bytecode, resolves rather more — a class file names its supertypes outright,
// including ones in libraries the Go frontend could never follow — and knows
// almost no positions at all. Where the two disagree is where the neutral model
// had assumptions in it.
package jvm

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/worldiety/speclink/internal/lang/jvm/classfile"
)

// ClassRoots are the directories a build leaves compiled classes in.
//
// They are conventions of three different toolchains and none of them is
// discoverable from the source tree, so they are listed rather than guessed.
// A project that puts them elsewhere says so in the configuration.
//
// asm_instrumented_project_classes is deliberately absent. Android rewrites
// classes there — Hilt replaces the superclass of an @AndroidEntryPoint
// activity, Jacoco threads coverage counters through every method — and reading
// those would mean reporting an architecture that nobody wrote. The
// uninstrumented output is the honest one.
var ClassRoots = []string{
	// Maven.
	"target/classes",
	"target/test-classes",
	// Gradle on the JVM.
	"build/classes/java/main",
	"build/classes/java/test",
	"build/classes/kotlin/main",
	"build/classes/kotlin/test",
	// The fixture, and any project that simply compiles to one directory.
	"classes",
}

// androidGlobs match the Android Gradle Plugin's output, whose layout differs
// from every other Gradle build and carries the build variant in the path.
//
// The Kotlin path is not under build/classes at all: the Kotlin plugin sets the
// destination to build/tmp/kotlin-classes/<variant> for Android variants
// specifically, which is why a scanner written against the JVM convention finds
// a project's Java and none of its Kotlin.
var androidGlobs = []string{
	"build/tmp/kotlin-classes/*",
	"build/intermediates/javac/*/classes",
	"build/intermediates/javac/*/*/classes",
}

// Class is a compiled type together with where it came from.
type Class struct {
	*classfile.Class
	// File is the repository relative path of the class file.
	File string
	// Root is the class directory it was found under, which is what makes a
	// class's package path recoverable.
	Root string
}

// Load reads every class under the given roots.
//
// roots are repository relative; empty means the conventional set. Directories
// that do not exist are skipped in silence, because the conventions cover three
// toolchains and no project has more than one of them.
func Load(root string, roots []string) ([]*Class, []error) {
	if len(roots) == 0 {
		roots = discover(root)
	}

	var (
		out  []*Class
		errs []error
		seen = map[string]bool{}
	)
	for _, rel := range roots {
		dir := filepath.Join(root, rel)
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			continue
		}

		walkErr := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".class") {
				return err
			}
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				errs = append(errs, readErr)
				return nil
			}
			c, parseErr := classfile.Read(data)
			if parseErr != nil {
				errs = append(errs, fmt.Errorf("%s: %w", relTo(root, path), parseErr))
				return nil
			}
			// A class compiled into two roots — main and test, or two Android
			// variants — is one type. Taking the first keeps the model a set of
			// types rather than a set of compilations of them.
			if seen[c.Name] {
				return nil
			}
			seen[c.Name] = true
			out = append(out, &Class{Class: c, File: relTo(root, path), Root: rel})
			return nil
		})
		if walkErr != nil {
			errs = append(errs, walkErr)
		}
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, errs
}

// discover returns the class roots that exist in a project.
func discover(root string) []string {
	var out []string
	for _, rel := range ClassRoots {
		if info, err := os.Stat(filepath.Join(root, rel)); err == nil && info.IsDir() {
			out = append(out, rel)
		}
	}
	for _, pattern := range androidGlobs {
		matches, err := filepath.Glob(filepath.Join(root, pattern))
		if err != nil {
			continue
		}
		for _, m := range matches {
			if info, err := os.Stat(m); err == nil && info.IsDir() {
				out = append(out, relTo(root, m))
			}
		}
	}
	sort.Strings(out)
	return out
}

// Package returns the package a class belongs to, empty for the default one.
func (c *Class) Package() string {
	if i := strings.LastIndexByte(c.Name, '.'); i >= 0 {
		return c.Name[:i]
	}
	return ""
}

// Simple returns the class name without its package, with any nesting kept:
// "com.example.Quote$Inner" gives "Quote$Inner".
func (c *Class) Simple() string {
	if i := strings.LastIndexByte(c.Name, '.'); i >= 0 {
		return c.Name[i+1:]
	}
	return c.Name
}

// IsSynthetic reports whether the class was generated rather than written.
func (c *Class) IsSynthetic() bool { return c.Access&classfile.AccSynthetic != 0 }

// SourcePath guesses where the class was written, from its package and the
// SourceFile attribute.
//
// A class file records the name of its source file and not its path, so the
// path has to be reconstructed from the package — which is exact for Java,
// where the two must agree, and only usually right for Kotlin, where a file may
// declare any package it likes. Callers treat the result as a candidate and
// fall back to the class file's own path when nothing is there.
func (c *Class) SourcePath(sourceRoots []string) []string {
	if c.SourceFile == "" {
		return nil
	}
	dir := strings.ReplaceAll(c.Package(), ".", "/")

	out := make([]string, 0, len(sourceRoots))
	for _, r := range sourceRoots {
		out = append(out, filepath.ToSlash(filepath.Join(r, dir, c.SourceFile)))
	}
	return out
}

func relTo(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}
