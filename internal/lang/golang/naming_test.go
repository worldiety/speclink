package golang

import "testing"

func TestSnakeCase(t *testing.T) {
	tests := []struct{ in, want string }{
		// The framework's own file names, which the rule has to reproduce.
		{"Create", "create"},
		{"FindByID", "find_by_id"},
		{"ListPermissions", "list_permissions"},
		{"UpdatePermissions", "update_permissions"},
		{"FindMyRoles", "find_my_roles"},
		{"SubmitQuoteUC", "submit_quote_uc"},
		// Acronyms must not fall apart into single letters.
		{"IDToken", "id_token"},
		{"HTTPServer", "http_server"},
		{"ParseURL", "parse_url"},
		{"A", "a"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := snakeCase(tt.in); got != tt.want {
			t.Errorf("snakeCase(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestUseCaseFileName(t *testing.T) {
	if got, want := useCaseFileName("FindByID"), "uc_find_by_id.go"; got != want {
		t.Errorf("useCaseFileName(FindByID) = %q, want %q", got, want)
	}
}
