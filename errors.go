package audiobookshelf

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Error is returned for any response with a 4xx or 5xx status code.
type Error struct {
	Method     string
	Path       string
	StatusCode int
	Message    string
}

func (e *Error) Error() string {
	msg := e.Message
	if msg == "" {
		msg = http.StatusText(e.StatusCode)
	}

	return fmt.Sprintf("audiobookshelf: %s %s: %d %s", e.Method, e.Path, e.StatusCode, msg)
}

// errorBodyLimit caps how much of an error response body is kept.
const errorBodyLimit = 4 << 10

func checkResponse(resp *http.Response, method, path string) error {
	if resp.StatusCode < 400 {
		return nil
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, errorBodyLimit))
	return &Error{
		Method:     method,
		Path:       path,
		StatusCode: resp.StatusCode,
		Message:    strings.TrimSpace(string(body)),
	}
}

func hasStatus(err error, code int) bool {
	var apiErr *Error
	return errors.As(err, &apiErr) && apiErr.StatusCode == code
}

// IsNotFound reports whether err is an *Error with status 404.
func IsNotFound(err error) bool { return hasStatus(err, http.StatusNotFound) }

// IsUnauthorized reports whether err is an *Error with status 401.
func IsUnauthorized(err error) bool { return hasStatus(err, http.StatusUnauthorized) }

// IsForbidden reports whether err is an *Error with status 403.
func IsForbidden(err error) bool { return hasStatus(err, http.StatusForbidden) }

// IsBadRequest reports whether err is an *Error with status 400.
func IsBadRequest(err error) bool { return hasStatus(err, http.StatusBadRequest) }
