package golang

import "strings"

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

// useCaseFileName returns the file a use case of the given name must live in.
func useCaseFileName(typeName string) string {
	return "uc_" + snakeCase(typeName) + ".go"
}

func isUpper(r rune) bool { return r >= 'A' && r <= 'Z' }

func toLower(r rune) rune {
	if isUpper(r) {
		return r + ('a' - 'A')
	}
	return r
}
