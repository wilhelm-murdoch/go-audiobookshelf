package audiobookshelf

import (
	"context"
	"encoding/base64"
	"io"
	"net/url"
	"strconv"
)

// UploadFile is one file of an UploadFilesRequest.
type UploadFile struct {
	Name   string
	Reader io.Reader
}

// UploadFilesRequest are the parameters for UploadFiles.
type UploadFilesRequest struct {
	// Title of the new library item (required).
	Title string
	// Author of the new library item.
	Author string
	// Series of the new library item.
	Series string
	// LibraryID of the library to put the item in (required).
	LibraryID string
	// FolderID of the library folder to put the item in (required).
	FolderID string
	// Files to upload.
	Files []UploadFile
}

// RenameResult is the response of RenameTag and RenameGenre.
type RenameResult struct {
	// Merged is true when the new name already existed and the two were
	// merged.
	Merged          bool
	NumItemsUpdated int
}

// UploadFiles uploads files to the server, creating a new library item
// (POST /api/upload). Requires upload permission.
func (c *Client) UploadFiles(ctx context.Context, req *UploadFilesRequest) error {
	fields := map[string]string{
		"title":   req.Title,
		"library": req.LibraryID,
		"folder":  req.FolderID,
	}
	if req.Author != "" {
		fields["author"] = req.Author
	}
	if req.Series != "" {
		fields["series"] = req.Series
	}
	files := make([]multipartFile, 0, len(req.Files))
	for n, f := range req.Files {
		files = append(files, multipartFile{
			field:    strconv.Itoa(n),
			filename: f.Name,
			reader:   f.Reader,
		})
	}
	return c.postMultipart(ctx, "/api/upload", fields, files, nil)
}

// UpdateServerSettings updates server settings (PATCH /api/settings) and
// returns the updated settings. patch maps setting keys (the JSON keys
// of ServerSettings) to their new values, so only the given settings are
// changed. Requires admin.
func (c *Client) UpdateServerSettings(ctx context.Context, patch map[string]any) (*ServerSettings, error) {
	var resp struct {
		Success        bool            `json:"success"`
		ServerSettings *ServerSettings `json:"serverSettings"`
	}
	if err := c.Patch(ctx, "/api/settings", patch, &resp); err != nil {
		return nil, err
	}
	return resp.ServerSettings, nil
}

// Authorize returns the authenticated user and server information for
// the client's token (POST /api/authorize).
func (c *Client) Authorize(ctx context.Context) (*LoginResponse, error) {
	var resp LoginResponse
	if err := c.Post(ctx, "/api/authorize", nil, &resp); err != nil {
		return nil, err
	}
	if resp.User != nil {
		resp.User.client = c
	}
	return &resp, nil
}

// Tags returns all tags in use (GET /api/tags). Requires admin.
func (c *Client) Tags(ctx context.Context) ([]string, error) {
	var resp struct {
		Tags []string `json:"tags"`
	}
	if err := c.Get(ctx, "/api/tags", &resp); err != nil {
		return nil, err
	}
	return resp.Tags, nil
}

// RenameTag renames a tag on all library items (POST /api/tags/rename).
// Requires admin.
func (c *Client) RenameTag(ctx context.Context, tag, newTag string) (*RenameResult, error) {
	body := map[string]string{"tag": tag, "newTag": newTag}
	var resp struct {
		TagMerged       bool `json:"tagMerged"`
		NumItemsUpdated int  `json:"numItemsUpdated"`
	}
	if err := c.Post(ctx, "/api/tags/rename", body, &resp); err != nil {
		return nil, err
	}
	return &RenameResult{Merged: resp.TagMerged, NumItemsUpdated: resp.NumItemsUpdated}, nil
}

// DeleteTag removes a tag from all library items
// (DELETE /api/tags/:tag; the tag is base64-encoded in the path). It
// returns the number of items updated. Requires admin.
func (c *Client) DeleteTag(ctx context.Context, tag string) (int, error) {
	encoded := base64.StdEncoding.EncodeToString([]byte(tag))
	var resp struct {
		NumItemsUpdated int `json:"numItemsUpdated"`
	}
	if err := c.Delete(ctx, "/api/tags/"+url.PathEscape(encoded), &resp); err != nil {
		return 0, err
	}
	return resp.NumItemsUpdated, nil
}

// Genres returns all genres in use (GET /api/genres). Requires admin.
func (c *Client) Genres(ctx context.Context) ([]string, error) {
	var resp struct {
		Genres []string `json:"genres"`
	}
	if err := c.Get(ctx, "/api/genres", &resp); err != nil {
		return nil, err
	}
	return resp.Genres, nil
}

// RenameGenre renames a genre on all library items
// (POST /api/genres/rename). Requires admin.
func (c *Client) RenameGenre(ctx context.Context, genre, newGenre string) (*RenameResult, error) {
	body := map[string]string{"genre": genre, "newGenre": newGenre}
	var resp struct {
		GenreMerged     bool `json:"genreMerged"`
		NumItemsUpdated int  `json:"numItemsUpdated"`
	}
	if err := c.Post(ctx, "/api/genres/rename", body, &resp); err != nil {
		return nil, err
	}
	return &RenameResult{Merged: resp.GenreMerged, NumItemsUpdated: resp.NumItemsUpdated}, nil
}

// DeleteGenre removes a genre from all library items
// (DELETE /api/genres/:genre; the genre is base64-encoded in the path).
// It returns the number of items updated. Requires admin.
func (c *Client) DeleteGenre(ctx context.Context, genre string) (int, error) {
	encoded := base64.StdEncoding.EncodeToString([]byte(genre))
	var resp struct {
		NumItemsUpdated int `json:"numItemsUpdated"`
	}
	if err := c.Delete(ctx, "/api/genres/"+url.PathEscape(encoded), &resp); err != nil {
		return 0, err
	}
	return resp.NumItemsUpdated, nil
}

// ValidateCron validates a cron expression (POST /api/validate-cron). A
// nil error means the expression is valid.
func (c *Client) ValidateCron(ctx context.Context, expression string) error {
	return c.Post(ctx, "/api/validate-cron", map[string]string{"expression": expression}, nil)
}
