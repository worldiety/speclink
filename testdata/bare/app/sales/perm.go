package sales

import "example.com/bare/foundation/permission"

var (
	PermSubmitQuote = permission.Declare[SubmitQuote]("sales.quote.submit", "Submit quote", "Holders may submit an approved quote.")
	PermFindQuote   = permission.Declare[FindQuote]("sales.quote.find", "Read quote", "Holders may read a quote.")
)
