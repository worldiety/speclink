// Package main exposes the quotation context over HTTP.
package main

import (
	"io"
	"net/http"
	"strings"

	"example.com/erp/app/sales"

	"go.wdy.de/nago/application/hapi"
	"go.wdy.de/nago/application/token"
	"go.wdy.de/nago/auth"
)

// SubmitQuoteRequest is what a caller sends to submit a quote.
//
// It is deliberately not the command the use case takes. A presentation layer
// that maps a wire shape onto a domain command is the normal case, and it is
// the reason a wire type is only ever recorded where the framework states it
// rather than inferred from the use case's own signature.
type SubmitQuoteRequest struct {
	Subject auth.Subject
	QuoteID string `json:"quoteId"`
	Title   string `json:"title"`
}

// SubmitQuoteBody is the JSON a caller posts.
type SubmitQuoteBody struct {
	QuoteID string `json:"quoteId"`
	Title   string `json:"title"`
}

// SubmitQuoteResponse is what the caller gets back.
type SubmitQuoteResponse struct {
	Sequence uint64 `json:"sequence"`
}

// ListQuotesRequest carries nothing but the caller's identity.
type ListQuotesRequest struct {
	Subject auth.Subject
}

// ListQuotesResponse is the list a caller reads.
type ListQuotesResponse struct {
	Quotes []sales.QuoteOverview `json:"quotes"`
}

// configureAPI mounts the quotation routes.
func configureAPI(api *hapi.API, authenticate token.AuthenticateSubject, uc sales.UseCases) {
	hapi.Post[SubmitQuoteRequest](api, hapi.Operation{
		Path:    "/api/v1/quotes",
		Summary: "Submit a quote",
	}).
		Request(
			hapi.BearerAuth[SubmitQuoteRequest](authenticate, func(dst *SubmitQuoteRequest, subject auth.Subject) error {
				dst.Subject = subject
				return nil
			}),
			hapi.JSONFromBody(func(dst *SubmitQuoteRequest, body SubmitQuoteBody) error {
				dst.QuoteID = body.QuoteID
				dst.Title = body.Title
				return nil
			}),
		).
		Response(
			hapi.ToJSON[SubmitQuoteRequest, SubmitQuoteResponse](func(in SubmitQuoteRequest) (SubmitQuoteResponse, error) {
				seq, err := uc.SubmitQuote(in.Subject, sales.SubmitQuoteCmd{
					QuoteID: in.QuoteID,
					Title:   in.Title,
				})
				if err != nil {
					return SubmitQuoteResponse{}, err
				}
				return SubmitQuoteResponse{Sequence: uint64(seq)}, nil
			}),
		)

	hapi.Get[ListQuotesRequest](api, hapi.Operation{
		Path:    "/api/v1/quotes",
		Summary: "List quotes",
	}).
		Request(
			hapi.BearerAuth[ListQuotesRequest](authenticate, func(dst *ListQuotesRequest, subject auth.Subject) error {
				dst.Subject = subject
				return nil
			}),
		).
		Response(
			hapi.ToJSON[ListQuotesRequest, ListQuotesResponse](func(in ListQuotesRequest) (ListQuotesResponse, error) {
				quotes, err := uc.ListQuotes(in.Subject)
				if err != nil {
					return ListQuotesResponse{}, err
				}
				return ListQuotesResponse{Quotes: quotes}, nil
			}),
		)
}

// WithdrawQuoteRequest identifies the quote to take back.
type WithdrawQuoteRequest struct {
	Subject auth.Subject
	QuoteID string
	Reason  string
}

// configureWithdrawal mounts the withdrawal route through the general form.
//
// The verb is not in the name of the call here, so it can only come from the
// operation. A recogniser that defaulted to the framework's fallback would
// print a method the code never stated.
func configureWithdrawal(api *hapi.API, authenticate token.AuthenticateSubject, uc sales.UseCases) {
	hapi.Endpoint[WithdrawQuoteRequest](api, hapi.Operation{
		Method:  http.MethodDelete,
		Path:    "/api/v1/quotes/{quoteId}",
		Summary: "Withdraw a quote",
	}).
		Request(
			hapi.BearerAuth[WithdrawQuoteRequest](authenticate, func(dst *WithdrawQuoteRequest, subject auth.Subject) error {
				dst.Subject = subject
				return nil
			}),
		).
		Response(
			hapi.ToBinary[WithdrawQuoteRequest](func(in WithdrawQuoteRequest) (io.Reader, error) {
				if _, err := uc.WithdrawQuote(in.Subject, sales.WithdrawQuoteCmd{
					QuoteID: in.QuoteID,
					Reason:  in.Reason,
				}); err != nil {
					return nil, err
				}
				return strings.NewReader("withdrawn"), nil
			}),
		)
}
