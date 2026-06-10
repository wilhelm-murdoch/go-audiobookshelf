package rest

import "net/http"

// Authenticator applies authentication to an outgoing request. Implement
// it to support Bearer tokens, API-key headers, signed requests, and so
// on.
type Authenticator interface {
	Authenticate(req *http.Request)
}

// AuthenticatorFunc adapts a plain function to the Authenticator
// interface.
type AuthenticatorFunc func(req *http.Request)

// Authenticate calls f(req).
func (f AuthenticatorFunc) Authenticate(req *http.Request) { f(req) }

// BearerToken authenticates requests with an "Authorization: Bearer"
// header. The token may be replaced with SetToken (for example after a
// login). It is not safe to call SetToken concurrently with requests.
type BearerToken struct {
	token string
}

// NewBearerToken returns a BearerToken seeded with token (which may be
// empty).
func NewBearerToken(token string) *BearerToken {
	return &BearerToken{token: token}
}

// Authenticate adds the Authorization header when a token is set.
func (b *BearerToken) Authenticate(req *http.Request) {
	if b.token != "" {
		req.Header.Set("Authorization", "Bearer "+b.token)
	}
}

// Token returns the current token.
func (b *BearerToken) Token() string { return b.token }

// SetToken replaces the token used for subsequent requests.
func (b *BearerToken) SetToken(token string) { b.token = token }
