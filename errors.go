package audiobookshelf

import "github.com/wilhelm-murdoch/go-audiobookshelf/internal/rest"

// Error is returned for any response with a 4xx or 5xx status code. It is
// an alias for the shared rest.Error so errors.As works against either
// name.
type Error = rest.Error

// IsNotFound reports whether err is an *Error with status 404.
func IsNotFound(err error) bool { return rest.IsNotFound(err) }

// IsUnauthorized reports whether err is an *Error with status 401.
func IsUnauthorized(err error) bool { return rest.IsUnauthorized(err) }

// IsForbidden reports whether err is an *Error with status 403.
func IsForbidden(err error) bool { return rest.IsForbidden(err) }

// IsBadRequest reports whether err is an *Error with status 400.
func IsBadRequest(err error) bool { return rest.IsBadRequest(err) }
