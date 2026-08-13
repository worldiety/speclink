// Package i18n is a stub of the worldiety translation catalogue, reduced to the
// surface the speclink permission rule matches on.
package i18n

import "golang.org/x/text/language"

// Values maps a language to its translation.
type Values map[language.Tag]string

// Resource is a translated string.
type Resource struct{ key string }

// String returns the translation for the active language.
func (r Resource) String() string { return r.key }

// MustString registers a translated string under a stable key.
func MustString(key string, values Values) Resource { return Resource{key: key} }
