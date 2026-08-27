package source

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Walk enumerates every source document below the given roots.
//
// Enumeration is the whole point. Checking only the documents somebody happened
// to cite is circular: a document nobody read would contribute nothing, produce
// no finding, and be indistinguishable from one that is fully covered. That is
// the shape of the worst defect in the chain — a section of a specification
// that never became a requirement, in a tree that is internally consistent and
// reports full coverage.
//
// This is the same move that makes the code side work. Constructs are
// enumerated from the type information rather than expected to announce
// themselves, which is why a use case cannot be forgotten. Sources are
// enumerated from the file system for exactly the same reason.
//
// Region manifests are skipped: they describe a document, they are not one.
func Walk(root string, roots []string) ([]string, []error) {
	var (
		docs []string
		errs []error
		seen = map[string]bool{}
	)

	for _, r := range roots {
		abs := filepath.Join(root, r)
		info, err := os.Stat(abs)
		if err != nil || !info.IsDir() {
			// A configured root that is not there is a configuration defect,
			// not an empty set. Treating it as empty would silently switch the
			// forward coverage off.
			errs = append(errs, &SegmentError{
				Doc: r,
				Msg: "source root does not exist",
				Why: "The forward coverage of the sources is measured over the documents below this directory. A missing root would silently mean there is nothing to cover.",
				How: "Create the directory, or correct sourceRoots in " + configFileName + ".",
			})
			continue
		}

		walkErr := filepath.WalkDir(abs, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if strings.HasPrefix(d.Name(), ".") && d.Name() != "." {
					return filepath.SkipDir
				}
				return nil
			}
			if IsManifest(path) {
				return nil
			}
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				return relErr
			}
			rel = filepath.ToSlash(rel)

			if _, ok := KindOf(rel); !ok {
				// An unreadable type in a source directory is a defect, not
				// something to pass over. A PDF dropped next to the Markdown
				// carries requirements that nothing will ever ask about.
				errs = append(errs, &SegmentError{
					Doc: rel,
					Msg: "unsupported document type in a source root",
					Why: "speclink segments Markdown by its headings and raster images by a declared region manifest. Anything else contributes nothing to the forward coverage while sitting among documents that do.",
					How: "Convert it to Markdown, supply it as PNG or JPEG with a region manifest, or move it out of the source root.",
				})
				return nil
			}
			if !seen[rel] {
				seen[rel] = true
				docs = append(docs, rel)
			}
			return nil
		})
		if walkErr != nil {
			errs = append(errs, &SegmentError{
				Doc: r,
				Msg: "source root cannot be walked: " + walkErr.Error(),
				How: "Check the directory permissions.",
			})
		}
	}

	sort.Strings(docs)
	return docs, errs
}

// configFileName is repeated here rather than imported. internal/config is a
// consumer of this package's roots, and an import back would be a cycle for the
// sake of one string in one diagnostic.
const configFileName = "speclink.json"
