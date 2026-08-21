// Package permtext routes permission texts through the translation catalogue.
//
// Writing the catalogue call out at every declaration is a dozen lines each and
// hundreds across a system, so it is factored into one place. A rule that only
// accepted the inlined form would punish that.
package permtext

import (
	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/permission"
	"golang.org/x/text/language"
)

// Name returns the translatable display name of a permission.
func Name(id permission.ID, en string) string {
	return i18n.MustString(
		i18n.Key(string(id)+"_perm_name"),
		i18n.Values{language.English: en},
	).String()
}
