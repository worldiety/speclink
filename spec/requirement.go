package spec

// Source names where a requirement originates. This is the outer edge of the
// specification chain and the one link that is traditionally free text and
// therefore unverified (error pattern M2 of the concept).
//
// Exactly one of Doc and Extern must be set.
//
// Doc is a repository-relative path below requirements/_sources/ and must
// exist. The raw source itself is never modified; requirements point at it.
//
// Anchor is the slug of a heading in the target document: lower case, spaces
// turned into "-", punctuation removed. "## 8.1 Angebot (Kopf)" yields
// "81-angebot-kopf". speclink verifies that a heading with this slug exists.
// Leave empty for non-text documents.
//
// Extern carries laws and standards that have no document in the repository,
// e.g. "HGB §§ 383 ff." or "ISO 27001 A.5.9".
//
// Note describes which part of an image is meant. It is not verifiable and is
// the deliberately accepted residual gap of M2.
type Source struct {
	Doc    string
	Anchor string
	Extern string
	Note   string
}

// Attachment is accompanying material. Unlike [Source] it is not the origin of
// the requirement but its companion.
//
// Path is relative to the attachment folder of the requirement, or
// repository-relative for material shared between requirements.
//
// Material that covers exactly one requirement belongs in that requirement's
// attachment folder. Material covering several belongs either to a common
// parent node (preferred, the DAG then carries it) or, if no such parent
// exists, to the group's _material/ folder. Found material, such as a
// screenshot of a legacy system, is never an attachment: it is a [Source].
type Attachment struct {
	Path string
	Role Role
	Note string
}

// Requirement declares a single requirement. It may only appear in *.spec.go
// files of the requirement tree, never in annotation files: a requirement is
// owned by the domain side and outlives any implementation.
//
// Text is normative and short — one sentence, used in lists, matrices and
// diagnostics. Long form (tables, acceptance criteria, process descriptions)
// belongs in the Markdown file named by Detail.
type Requirement struct {
	ID         RequirementID
	Kind       Kind
	Discipline Discipline
	Status     Status
	Title      string
	Text       string
	Detail     string // Markdown file in the attachment folder, optional
	Rationale  string // mandatory when Kind is Decision
	// Consequences is what this ruling costs, and is mandatory when Kind is
	// Decision.
	//
	// # Why it is not part of the rationale
	//
	// Because it is the half nobody writes unprompted. A justification is
	// pleasant to write and people write it at length; admitting what the
	// decision makes worse is not, and in a single field it quietly
	// disappears. Every decision record worth the name has carried this
	// separately since Nygard's original, and the reason is the same: without
	// it a record reads as advocacy for the choice rather than as an account
	// of it.
	//
	// It is what stops the reader in three years from beginning an
	// improvement that was already considered and paid for. "The registry is
	// a single writer and will not scale past one machine" is the sentence
	// that saves that week.
	Consequences string
	DerivedFrom  []Requirement
	Supersedes   []Requirement
	Sources      []Source
	Attachments  []Attachment
	// Topics are the themes this requirement belongs to, and the chapters it
	// appears in. Optional: what carries none is gathered in a chapter of its
	// own rather than dropped.
	Topics []Topic
}

// Glossary declares a domain term. Like [Requirement] it is a declaration and
// belongs in the requirement tree; annotation files reference it via [Term].
type Glossary struct {
	ID         TermID
	Title      string
	Definition string
}
