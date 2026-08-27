package spec

import (
	"encoding/json"
	"sort"
)

// VerifiedVersion is the schema version of the line [Verified] writes.
//
// It needs one for the same reason [DumpVersion] does, and it is the same
// boundary: the target project pins speclink/spec in its go.mod, while the
// developer runs an arbitrary speclink binary. A reader that finds a version it
// does not know must refuse with a clear message rather than quietly record
// nothing, because recording nothing looks exactly like a test that was never
// written.
const VerifiedVersion = 1

// VerifiedMarker opens the line, so a reader can find it among ordinary test
// output without parsing every line as JSON.
//
// The prefix is deliberately ugly. It has to be something no human writes by
// accident and no formatter reflows.
const VerifiedMarker = "speclink-verified:"

// verifiedRecord is the wire shape of one statement.
type verifiedRecord struct {
	Version int      `json:"v"`
	Reqs    []string `json:"reqs"`
}

// Logger is the part of *testing.T that [Verified] needs.
//
// It is an interface so this package does not import testing. testing
// registers command line flags from an init function, and a package that
// production code imports must not do that to the binaries importing it.
type Logger interface {
	Log(args ...any)
}

// Verified states that the surrounding test has just demonstrated the given
// requirements.
//
// It is the one place in speclink where a fact is produced at run time rather
// than read from the source, and that is not the exception it looks like.
// P9 bans constructs that turn *static* facts into dynamic ones — six
// permissions conjured from a prefix, which an analyser could only see by
// reimplementing framework internals. A test result was never a static fact.
// No analysis can derive it, and a claim that a test verifies something is not
// evidence that it does.
//
// So this call has two lives. It is read statically, like every other spec
// call, which is what makes a missing one reportable. And it writes a line when
// it runs, which is what makes a present one believable. Together they separate
// three states that neither could tell apart alone:
//
//   - no call anywhere: nothing claims to verify the requirement.
//   - a call that never ran: written, perhaps behind a condition that is never
//     true, perhaps in a test that fails before reaching it. Claimed, not
//     demonstrated.
//   - a call that ran in a passing test: demonstrated.
//
// Position therefore means something, unlike everywhere else in this language.
// The line is written when control reaches it, so putting it at the end of a
// test says the test got there. Putting it at the top says only that the test
// started.
//
// The requirements are passed as values rather than by name, so the Go compiler
// resolves them. A renamed requirement is a refactoring here as it is
// everywhere else; only the ID travels in the written line, and only as its
// serialisation.
//
// It writes through the test's own logger rather than to standard output, so
// that `go test -json` attributes the line to the test that produced it. That
// attribution is the whole point: without it a record could not be tied to a
// pass or a failure.
func Verified(log Logger, reqs ...Requirement) {
	if log == nil || len(reqs) == 0 {
		return
	}
	// Not part of Logger, because requiring it would exclude anything that is
	// not a *testing.T for no gain. Without it the line reports this file as
	// its origin, which is useless to whoever is reading the test output.
	if h, ok := log.(interface{ Helper() }); ok {
		h.Helper()
	}

	ids := make([]string, 0, len(reqs))
	seen := make(map[string]bool, len(reqs))
	for _, r := range reqs {
		id := string(r.ID)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return
	}
	// Sorted so a test that names the same requirements produces the same line
	// on every run. The record ends up in speclink.lock, whose diff is read by
	// people.
	sort.Strings(ids)

	line, err := json.Marshal(verifiedRecord{Version: VerifiedVersion, Reqs: ids})
	if err != nil {
		// Unreachable for a slice of strings, and there is nothing useful to do
		// with it in a test either way.
		return
	}
	log.Log(VerifiedMarker + string(line))
}
