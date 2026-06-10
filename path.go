package audiobookshelf

import (
	"net/url"
	"strings"
)

// pathBuilder assembles an Audiobookshelf request path together with its
// query string. Start one with apiPath for the common "/api/..." routes
// or rawPath for the few endpoints served from the server root (e.g.
// "/login"). Append dynamic, user-supplied values with Seg (each is
// escaped with url.PathEscape) and fixed sub-resource names with Lit
// (used verbatim). Add query parameters with Query, Set, or Flag. Render
// the final result with String.
//
// Empty segments are skipped, so optional identifiers (such as a podcast
// episode ID) can be passed unconditionally.
type pathBuilder struct {
	b     strings.Builder
	query url.Values
}

// apiPath starts a builder at "/api" extended by the given fixed root
// segments, e.g. apiPath("libraries") yields "/api/libraries". Root
// segments are trusted and are not escaped.
func apiPath(root ...string) *pathBuilder {
	pb := &pathBuilder{}
	pb.b.WriteString("/api")
	pb.lit(root)
	return pb
}

// rawPath starts a builder at a trusted, verbatim prefix that is used as
// given, e.g. rawPath("/login").
func rawPath(prefix string) *pathBuilder {
	pb := &pathBuilder{}
	pb.b.WriteString(prefix)
	return pb
}

// Seg appends one escaped path segment per non-empty value. Values are
// escaped with url.PathEscape, so never pre-escape user input.
func (pb *pathBuilder) Seg(segments ...string) *pathBuilder {
	for _, s := range segments {
		if s == "" {
			continue
		}
		pb.b.WriteByte('/')
		pb.b.WriteString(url.PathEscape(s))
	}
	return pb
}

// Lit appends one verbatim path segment per non-empty value. Use it only
// for trusted, fixed sub-resource names such as "items" or "cover".
func (pb *pathBuilder) Lit(segments ...string) *pathBuilder {
	pb.lit(segments)
	return pb
}

func (pb *pathBuilder) lit(segments []string) {
	for _, s := range segments {
		if s == "" {
			continue
		}
		pb.b.WriteByte('/')
		pb.b.WriteString(s)
	}
}

// Query merges the given values into the query string. A nil or empty q
// is a no-op.
func (pb *pathBuilder) Query(q url.Values) *pathBuilder {
	for key, values := range q {
		for _, v := range values {
			pb.add(key, v)
		}
	}
	return pb
}

// Set adds a single query parameter.
func (pb *pathBuilder) Set(key, value string) *pathBuilder {
	pb.add(key, value)
	return pb
}

// Flag adds key=1 when on, matching the server's convention for boolean
// query parameters, and does nothing otherwise.
func (pb *pathBuilder) Flag(key string, on bool) *pathBuilder {
	if on {
		pb.add(key, "1")
	}
	return pb
}

func (pb *pathBuilder) add(key, value string) {
	if pb.query == nil {
		pb.query = url.Values{}
	}
	pb.query.Add(key, value)
}

// String renders the full path, appending "?"+encoded query when any
// query parameters were set.
func (pb *pathBuilder) String() string {
	if len(pb.query) == 0 {
		return pb.b.String()
	}

	return pb.b.String() + "?" + pb.query.Encode()
}
