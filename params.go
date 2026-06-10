package audiobookshelf

import (
	"net/url"
	"strconv"
)

// appendQuery appends an encoded query string to path if q is non-empty.
func appendQuery(path string, q url.Values) string {
	if len(q) == 0 {
		return path
	}
	return path + "?" + q.Encode()
}

// PageParams are the limit/page parameters used by paginated list
// endpoints. Pages are 0-indexed. A zero Limit applies no limit.
type PageParams struct {
	Limit int
	Page  int
}

func (p *PageParams) values() url.Values {
	q := url.Values{}
	if p == nil {
		return q
	}
	if p.Limit > 0 {
		q.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.Page > 0 {
		q.Set("page", strconv.Itoa(p.Page))
	}
	return q
}

// SessionListParams are the parameters for listening-session list
// endpoints, which paginate with itemsPerPage/page. Pages are 0-indexed.
type SessionListParams struct {
	// User filters sessions by user ID. Only honored by Client.Sessions
	// (admin).
	User string
	// ItemsPerPage is the number of sessions per page (server default 10).
	ItemsPerPage int
	Page         int
}

func (p *SessionListParams) values() url.Values {
	q := url.Values{}
	if p == nil {
		return q
	}
	if p.User != "" {
		q.Set("user", p.User)
	}
	if p.ItemsPerPage > 0 {
		q.Set("itemsPerPage", strconv.Itoa(p.ItemsPerPage))
	}
	if p.Page > 0 {
		q.Set("page", strconv.Itoa(p.Page))
	}
	return q
}

// ImageParams control the size and format of cover and author images.
type ImageParams struct {
	// Width of the image (server default 400).
	Width int
	// Height of the image. If zero, the image is scaled proportionally.
	Height int
	// Format requests "webp" or "jpeg". The server default depends on the
	// request headers.
	Format string
	// Raw requests the original image file instead of a scaled version.
	Raw bool
}

func (p *ImageParams) values() url.Values {
	q := url.Values{}
	if p == nil {
		return q
	}
	if p.Width > 0 {
		q.Set("width", strconv.Itoa(p.Width))
	}
	if p.Height > 0 {
		q.Set("height", strconv.Itoa(p.Height))
	}
	if p.Format != "" {
		q.Set("format", p.Format)
	}
	if p.Raw {
		q.Set("raw", "1")
	}
	return q
}

// Page is the envelope returned by paginated library list endpoints.
type Page[T any] struct {
	Results        []T    `json:"results"`
	Total          int    `json:"total"`
	Limit          int    `json:"limit"`
	Page           int    `json:"page"`
	SortBy         string `json:"sortBy,omitempty"`
	SortDesc       bool   `json:"sortDesc,omitempty"`
	FilterBy       string `json:"filterBy,omitempty"`
	MediaType      string `json:"mediaType,omitempty"`
	Minified       bool   `json:"minified,omitempty"`
	CollapseSeries bool   `json:"collapseseries,omitempty"`
	Include        string `json:"include,omitempty"`
}
