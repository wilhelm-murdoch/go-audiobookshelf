package audiobookshelf

import (
	"context"
	"io"
	"net/url"
	"strings"
)

// AuthorParams are the optional query parameters for Client.Author.
type AuthorParams struct {
	Include   []string
	LibraryID string
}

func (ap *AuthorParams) values() url.Values {
	q := url.Values{}
	if ap == nil {
		return q
	}

	if len(ap.Include) > 0 {
		q.Set("include", strings.Join(ap.Include, ","))
	}

	if ap.LibraryID != "" {
		q.Set("library", ap.LibraryID)
	}

	return q
}

// UpdateAuthorRequest are the parameters for UpdateAuthor. Nil fields are
// left unchanged.
type UpdateAuthorRequest struct {
	ASIN        *string `json:"asin,omitempty"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	ImagePath   *string `json:"imagePath,omitempty"`
}

// AuthorUpdateResult is the response of UpdateAuthor. If the new name matched
// an existing author, Merged is true and Author is the author that was merged
// into.
type AuthorUpdateResult struct {
	Author  *Author `json:"author"`
	Merged  bool    `json:"merged,omitempty"`
	Updated bool    `json:"updated,omitempty"`
}

// MatchAuthorRequest are the parameters for MatchAuthor. Provide either the
// ASIN or the author name to search for.
type MatchAuthorRequest struct {
	ASIN string `json:"asin,omitempty"`
	Q    string `json:"q,omitempty"`
}

// AuthorMatchResult is the response of MatchAuthor.
type AuthorMatchResult struct {
	Updated bool    `json:"updated"`
	Author  *Author `json:"author"`
}

func authorPath(id string, rest ...string) string {
	path := "/api/authors/" + url.PathEscape(id)
	for _, r := range rest {
		path += "/" + r
	}

	return path
}

// Author returns an author.
// GET /api/authors/:id
func (c *Client) Author(ctx context.Context, id string, params *AuthorParams) (*Author, error) {
	var author Author

	if err := c.Get(ctx, appendQuery(authorPath(id), params.values()), &author); err != nil {
		return nil, err
	}

	author.client = c

	return &author, nil
}

// UpdateAuthor updates an author. Renaming an author to an existing author's
// name merges the two.
// PATCH /api/authors/:id
func (c *Client) UpdateAuthor(ctx context.Context, id string, req *UpdateAuthorRequest) (*AuthorUpdateResult, error) {
	var result AuthorUpdateResult

	if err := c.Patch(ctx, authorPath(id), req, &result); err != nil {
		return nil, err
	}

	if result.Author != nil {
		result.Author.client = c
	}

	return &result, nil
}

// MatchAuthor matches an author against Audnexus and updates their details.
// POST /api/authors/:id/match
func (c *Client) MatchAuthor(ctx context.Context, id string, req *MatchAuthorRequest) (*AuthorMatchResult, error) {
	var result AuthorMatchResult
	if err := c.Post(ctx, authorPath(id, "match"), req, &result); err != nil {
		return nil, err
	}
	if result.Author != nil {
		result.Author.client = c
	}
	return &result, nil
}

// AuthorImage fetches an author's image. The caller must close the reader.
// The string result is the image's Content-Type.
// GET /api/authors/:id/image
func (c *Client) AuthorImage(ctx context.Context, id string, params *ImageParams) (io.ReadCloser, string, error) {
	return c.getBinary(ctx, appendQuery(authorPath(id, "image"), params.values()))
}

// Update updates the author.
func (a *Author) Update(ctx context.Context, req *UpdateAuthorRequest) (*AuthorUpdateResult, error) {
	return a.client.UpdateAuthor(ctx, a.ID, req)
}

// Match matches the author against Audnexus.
func (a *Author) Match(ctx context.Context, req *MatchAuthorRequest) (*AuthorMatchResult, error) {
	return a.client.MatchAuthor(ctx, a.ID, req)
}

// Image fetches the author's image.
func (a *Author) Image(ctx context.Context, params *ImageParams) (io.ReadCloser, string, error) {
	return a.client.AuthorImage(ctx, a.ID, params)
}
