// Package rest adapts use cases to the standard library's router.
//
// The presentation receives use cases and nothing else, so this is the only
// shape it ever needs: work out who is calling, decode the input, call, encode
// the output. Everything specific to a context is that context's routes.
package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	"example.com/bare/foundation/auth"
)

// Authenticator works out who is calling.
//
// It is a parameter rather than a package level hook, because a test for a
// route should be able to say who is calling without arranging global state.
type Authenticator func(*http.Request) (auth.Subject, error)

// Handle adapts a use case to an http.Handler.
func Handle[In, Out any](who Authenticator, uc func(auth.Subject, In) (Out, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		subject, err := who(r)
		if err != nil {
			Fail(w, err)
			return
		}

		in, err := Decode[In](r)
		if err != nil {
			http.Error(w, "malformed request", http.StatusBadRequest)
			return
		}

		out, err := uc(auth.WithContext(subject, r.Context()), in)
		if err != nil {
			Fail(w, err)
			return
		}
		Write(w, out)
	}
}

// Decode fills In from the request body.
//
// An empty body is the zero value rather than an error: a route that takes no
// input is the commonest kind, and demanding "{}" of every caller would be a
// requirement of the transport that the use case never made.
func Decode[In any](r *http.Request) (In, error) {
	var in In
	if r.Body == nil || r.ContentLength == 0 {
		return in, nil
	}
	err := json.NewDecoder(r.Body).Decode(&in)
	return in, err
}

// Write encodes the result.
func Write(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// Fail maps an error to a status code.
//
// Only the errors this layer is entitled to recognise. Anything else is 500:
// a handler that guessed would leak the domain's vocabulary into the protocol,
// and a status code is a promise to a client that cannot be taken back.
func Fail(w http.ResponseWriter, err error) {
	if errors.Is(err, auth.ErrPermissionDenied) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	http.Error(w, "internal error", http.StatusInternalServerError)
}

// Log records that a request happened.
//
// The shape every wrapper has: a handler in, a handler out, the original called
// somewhere in the middle. It is here so that at least one route in this
// project is mounted behind something, because a recogniser that has only ever
// seen bare handlers is a recogniser nobody has tested.
func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
