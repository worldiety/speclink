package golang

import "strings"

// Style is the part of an architecture that is a convention rather than a rule.
//
// The rules under K5 ask whether a use case lives in the right file and has a
// constructor. That two projects can answer those differently while following
// the same architecture is not a loophole — it is the observation that "one
// file per use case, named after it" and "uc_submit_quote.go" are two different
// statements, and only the first is architectural.
//
// The distinction was invisible while there was one style. The reference ERP
// names its use case constructors with a UC suffix rather than New<Type>, and
// lays out 45 files to match, which under a fixed convention reads as 238
// defects. It is not 238 defects and it never was; it is a second convention,
// and there was nowhere to write it down.
//
// What is deliberately not here is a switch to turn a rule off. A style says
// how a thing is spelled, not whether it has to be there — the escape from a
// rule is spec.Waive with a reason, or the scope, and adding a third would be
// severities under another name.
type Style struct {
	// Name identifies the convention in diagnostics.
	Name string
	// UseCaseFile returns the file a use case of this type name belongs in.
	UseCaseFile func(typeName string) string
	// Constructor returns the name of the function that builds it.
	Constructor func(typeName string) string
	// PermissionVar returns the name of the permission guarding it.
	PermissionVar func(typeName string) string
}

// DDD1 is the convention of the go_nago_ddd1 profile.
//
// One use case per file, named uc_<snake_case>.go; a constructor New<Type>
// beside it; a permission Perm<Type>. The framework's own code follows it,
// which is what made it the convention before anybody wrote it down.
var DDD1 = Style{
	Name:          "ddd1",
	UseCaseFile:   func(name string) string { return "uc_" + snakeCase(name) + ".go" },
	Constructor:   func(name string) string { return "New" + name },
	PermissionVar: func(name string) string { return "Perm" + name },
}

// snakeCase converts a Go identifier into the file name convention for use
// cases: FindByID becomes find_by_id, SubmitQuote becomes submit_quote.
//
// Consecutive capitals are treated as one word, which is what makes acronyms
// come out right. The framework's own file names follow this: uc_find_by_id.go
// for FindByID, uc_list_permissions.go for ListPermissions.
func snakeCase(name string) string {
	runes := []rune(name)
	var b strings.Builder

	for i, r := range runes {
		if !isUpper(r) {
			b.WriteRune(r)
			continue
		}
		// A capital starts a new word unless it continues a run of capitals,
		// in which case only the last one does — "IDToken" splits into
		// id_token, not i_d_token.
		if i > 0 {
			prevLower := !isUpper(runes[i-1])
			nextLower := i+1 < len(runes) && !isUpper(runes[i+1])
			if prevLower || nextLower {
				b.WriteByte('_')
			}
		}
		b.WriteRune(toLower(r))
	}
	return b.String()
}

func isUpper(r rune) bool { return r >= 'A' && r <= 'Z' }

func toLower(r rune) rune {
	if isUpper(r) {
		return r + ('a' - 'A')
	}
	return r
}
