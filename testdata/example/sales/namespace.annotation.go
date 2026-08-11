package sales

import (
	"example.com/erp/anforderungen/fun/quote"
	"github.com/worldiety/speclink/spec"
)

var _ = spec.ForDecl(Namespace,
	spec.Satisfies(quote.RQuoteApprove),
	spec.Rationale("The resource type is the anchor of every access rule on a quote."),
)
