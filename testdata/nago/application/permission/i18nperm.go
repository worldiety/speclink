package permission

import (
	"fmt"

	"github.com/worldiety/i18n"
	"golang.org/x/text/language"
)

// DeclareCreate declares a permission whose texts come from the translation
// catalogue, derived from the entity name of the ubiquitous language.
//
// The CRUD helpers exist so that trivial use cases get translated permission
// texts without every project writing the same two sentences again.
func DeclareCreate[UseCase any](id ID, entityName string) ID {
	return declareI18n[UseCase](id, entityName, "Create %s", "%s erstellen")
}

// DeclareFindByID declares the read permission of a single element.
func DeclareFindByID[UseCase any](id ID, entityName string) ID {
	return declareI18n[UseCase](id, entityName, "Read %s", "%s lesen")
}

func declareI18n[UseCase any](id ID, entityName, en, de string) ID {
	_ = i18n.MustString(
		fmt.Sprintf("%s_perm_name", id),
		i18n.Values{
			language.English: fmt.Sprintf(en, entityName),
			language.German:  fmt.Sprintf(de, entityName),
		},
	).String()
	return id
}
