package audiobookshelf

import "github.com/wilhelm-murdoch/go-audiobookshelf/internal/rest"

// apiPath starts a path builder rooted at "/api" extended by the given
// fixed segments, e.g. apiPath("libraries") yields "/api/libraries".
func apiPath(root ...string) *rest.Path {
	return rest.NewPath("/api").Lit(root...)
}

// rawPath starts a path builder at a trusted, verbatim prefix served from
// the server root, e.g. rawPath("/login").
func rawPath(prefix string) *rest.Path {
	return rest.NewPath(prefix)
}
