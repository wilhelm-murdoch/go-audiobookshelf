package audiobookshelf

import (
	"context"
	"io"
	"net/url"
	"strconv"
	"strings"
)

// LibraryItemParams are the optional query parameters for LibraryItem.
type LibraryItemParams struct {
	Expanded bool
	Include  []string
	Episode  string
}

func (p *LibraryItemParams) values() url.Values {
	q := url.Values{}
	if p == nil {
		return q
	}

	if p.Expanded {
		q.Set("expanded", "1")
	}

	if len(p.Include) > 0 {
		q.Set("include", strings.Join(p.Include, ","))
	}

	if p.Episode != "" {
		q.Set("episode", p.Episode)
	}

	return q
}

// MediaUpdate is the payload for UpdateLibraryItemMedia and
// BatchUpdateLibraryItems. It is a superset of the book and podcast
// media parameters; only set fields are sent.
type MediaUpdate struct {
	Metadata                 map[string]any `json:"metadata,omitempty"`
	CoverPath                *string        `json:"coverPath,omitempty"`
	Tags                     []string       `json:"tags,omitempty"`
	Chapters                 []Chapter      `json:"chapters,omitempty"`
	AutoDownloadEpisodes     *bool          `json:"autoDownloadEpisodes,omitempty"`
	AutoDownloadSchedule     *string        `json:"autoDownloadSchedule,omitempty"`
	LastEpisodeCheck         *int64         `json:"lastEpisodeCheck,omitempty"`
	MaxEpisodesToKeep        *int           `json:"maxEpisodesToKeep,omitempty"`
	MaxNewEpisodesToDownload *int           `json:"maxNewEpisodesToDownload,omitempty"`
}

// UpdateMediaResult is the response of UpdateLibraryItemMedia.
type UpdateMediaResult struct {
	Updated     bool         `json:"updated"`
	LibraryItem *LibraryItem `json:"libraryItem"`
}

// MatchLibraryItemRequest are the parameters for MatchLibraryItem. Empty
// fields fall back to the library item's own details.
type MatchLibraryItemRequest struct {
	Provider         string `json:"provider,omitempty"`
	Title            string `json:"title,omitempty"`
	Author           string `json:"author,omitempty"`
	ISBN             string `json:"isbn,omitempty"`
	ASIN             string `json:"asin,omitempty"`
	OverrideDefaults bool   `json:"overrideDefaults,omitempty"`
}

// MatchResult is the response of MatchLibraryItem.
type MatchResult struct {
	Updated     bool         `json:"updated"`
	LibraryItem *LibraryItem `json:"libraryItem"`
}

// PlayRequest are the parameters for PlayLibraryItem and
// PlayPodcastEpisode. All fields are optional.
type PlayRequest struct {
	DeviceInfo         *DeviceInfo `json:"deviceInfo,omitempty"`
	ForceDirectPlay    bool        `json:"forceDirectPlay,omitempty"`
	ForceTranscode     bool        `json:"forceTranscode,omitempty"`
	SupportedMimeTypes []string    `json:"supportedMimeTypes,omitempty"`
	MediaPlayer        string      `json:"mediaPlayer,omitempty"`
}

// TrackOrder identifies one audio file for UpdateLibraryItemTracks. The
// order of the slice becomes the new track order.
type TrackOrder struct {
	Ino     string `json:"ino"`
	Exclude bool   `json:"exclude,omitempty"`
}

// BatchUpdateItem is one update of BatchUpdateLibraryItems.
type BatchUpdateItem struct {
	ID           string       `json:"id"`
	MediaPayload *MediaUpdate `json:"mediaPayload"`
}

// QuickMatchOptions are the options for BatchQuickMatchLibraryItems.
type QuickMatchOptions struct {
	Provider         string `json:"provider,omitempty"`
	OverrideDefaults bool   `json:"overrideDefaults,omitempty"`
}

func itemPath(id string, rest ...string) (string, error) {
	return basePathBuilder("/api/items/", id, rest...)
}

// DeleteAllLibraryItems deletes ALL library items from the database
// (DELETE /api/items/all). No files are deleted. Requires the root user.
func (c *Client) DeleteAllLibraryItems(ctx context.Context) error {
	return c.Delete(ctx, "/api/items/all", nil)
}

// LibraryItem returns a library item (GET /api/items/:id).
func (c *Client) LibraryItem(ctx context.Context, id string, params *LibraryItemParams) (*LibraryItem, error) {
	path, err := itemPath(id)
	if err != nil {
		return nil, err
	}

	var item LibraryItem
	if err := c.Get(ctx, appendQuery(path, params.values()), &item); err != nil {
		return nil, err
	}

	item.client = c

	return &item, nil
}

// DeleteLibraryItem deletes a library item from the database
// (DELETE /api/items/:id). With hard, the item's files are also deleted
// from the filesystem.
func (c *Client) DeleteLibraryItem(ctx context.Context, id string, hard bool) error {
	q := url.Values{}
	if hard {
		q.Set("hard", "1")
	}

	path, err := itemPath(id)
	if err != nil {
		return err
	}

	return c.Delete(ctx, appendQuery(path, q), nil)
}

// UpdateLibraryItemMedia updates a library item's media
// (PATCH /api/items/:id/media).
func (c *Client) UpdateLibraryItemMedia(ctx context.Context, id string, update *MediaUpdate) (*UpdateMediaResult, error) {
	path, err := itemPath(id, "media")
	if err != nil {
		return nil, err
	}

	var result UpdateMediaResult
	if err := c.Patch(ctx, path, update, &result); err != nil {
		return nil, err
	}

	if result.LibraryItem != nil {
		result.LibraryItem.client = c
	}

	return &result, nil
}

// LibraryItemCover fetches a library item's cover image
// (GET /api/items/:id/cover). The caller must close the reader. The
// string result is the image's Content-Type.
func (c *Client) LibraryItemCover(ctx context.Context, id string, params *ImageParams) (io.ReadCloser, string, error) {
	path, err := itemPath(id, "cover")
	if err != nil {
		return nil, "", err
	}

	return c.getBinary(ctx, appendQuery(path, params.values()))
}

// UploadLibraryItemCover uploads a cover image for a library item
// (POST /api/items/:id/cover).
func (c *Client) UploadLibraryItemCover(ctx context.Context, id, filename string, cover io.Reader) error {
	files := []multipartFile{{field: "cover", filename: filename, reader: cover}}

	path, err := itemPath(id, "cover")
	if err != nil {
		return err
	}

	return c.postMultipart(ctx, path, nil, files, nil)
}

// SetLibraryItemCoverFromURL has the server download a cover image for a
// library item (POST /api/items/:id/cover with a url payload).
func (c *Client) SetLibraryItemCoverFromURL(ctx context.Context, id, coverURL string) error {
	path, err := itemPath(id, "cover")
	if err != nil {
		return err
	}

	return c.Post(ctx, path, map[string]string{"url": coverURL}, nil)
}

// UpdateLibraryItemCover points a library item's cover at an image file
// already on the server (PATCH /api/items/:id/cover).
func (c *Client) UpdateLibraryItemCover(ctx context.Context, id, coverPath string) error {
	path, err := itemPath(id, "cover")
	if err != nil {
		return err
	}

	return c.Patch(ctx, path, map[string]string{"cover": coverPath}, nil)
}

// RemoveLibraryItemCover removes a library item's cover
// (DELETE /api/items/:id/cover).
func (c *Client) RemoveLibraryItemCover(ctx context.Context, id string) error {
	path, err := itemPath(id, "cover")
	if err != nil {
		return err
	}

	return c.Delete(ctx, path, nil)
}

// MatchLibraryItem matches a library item against a metadata provider and
// updates its details (POST /api/items/:id/match).
func (c *Client) MatchLibraryItem(ctx context.Context, id string, req *MatchLibraryItemRequest) (*MatchResult, error) {
	path, err := itemPath(id, "match")
	if err != nil {
		return nil, err
	}

	var result MatchResult
	if err := c.Post(ctx, path, req, &result); err != nil {
		return nil, err
	}

	if result.LibraryItem != nil {
		result.LibraryItem.client = c
	}

	return &result, nil
}

// PlayLibraryItem starts a playback session for a library item
// (POST /api/items/:id/play). req may be nil.
func (c *Client) PlayLibraryItem(ctx context.Context, id string, req *PlayRequest) (*PlaybackSession, error) {
	if req == nil {
		req = &PlayRequest{}
	}

	path, err := itemPath(id, "play")
	if err != nil {
		return nil, err
	}

	var session PlaybackSession
	if err := c.Post(ctx, path, req, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

// PlayPodcastEpisode starts a playback session for a podcast episode
// (POST /api/items/:id/play/:episodeId). req may be nil.
func (c *Client) PlayPodcastEpisode(ctx context.Context, id, episodeID string, req *PlayRequest) (*PlaybackSession, error) {
	if req == nil {
		req = &PlayRequest{}
	}

	path, err := itemPath(id, "play", url.PathEscape(episodeID))
	if err != nil {
		return nil, err
	}

	var session PlaybackSession
	if err := c.Post(ctx, path, req, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

// UpdateLibraryItemTracks sets the track order of a book's audio files
// (PATCH /api/items/:id/tracks) and returns the updated library item.
func (c *Client) UpdateLibraryItemTracks(ctx context.Context, id string, order []TrackOrder) (*LibraryItem, error) {
	body := map[string]any{"orderedFileData": order}

	path, err := itemPath(id, "tracks")
	if err != nil {
		return nil, err
	}

	var item LibraryItem
	if err := c.Patch(ctx, path, body, &item); err != nil {
		return nil, err
	}

	item.client = c

	return &item, nil
}

// ScanLibraryItem rescans a library item's files
// (POST /api/items/:id/scan) and returns the scan result: "NOTHING",
// "ADDED", "UPDATED", "REMOVED", or "UPTODATE". Requires admin.
func (c *Client) ScanLibraryItem(ctx context.Context, id string) (string, error) {
	var resp struct {
		Result string `json:"result"`
	}

	path, err := itemPath(id, "scan")
	if err != nil {
		return "", err
	}

	if err := c.Post(ctx, path, nil, &resp); err != nil {
		return "", err
	}

	return resp.Result, nil
}

// LibraryItemToneObject returns the tone metadata object of a library
// item (GET /api/items/:id/tone-object). Requires admin.
func (c *Client) LibraryItemToneObject(ctx context.Context, id string) (map[string]any, error) {
	path, err := itemPath(id, "tone-object")
	if err != nil {
		return nil, err
	}

	var tone map[string]any
	if err := c.Get(ctx, path, &tone); err != nil {
		return nil, err
	}

	return tone, nil
}

// UpdateLibraryItemChapters replaces the chapters of a book
// (POST /api/items/:id/chapters). It returns whether the chapters were
// actually changed.
func (c *Client) UpdateLibraryItemChapters(ctx context.Context, id string, chapters []Chapter) (bool, error) {
	body := map[string]any{"chapters": chapters}

	var resp struct {
		Success bool `json:"success"`
		Updated bool `json:"updated"`
	}

	path, err := itemPath(id, "chapters")
	if err != nil {
		return false, err
	}

	if err := c.Post(ctx, path, body, &resp); err != nil {
		return false, err
	}

	return resp.Updated, nil
}

// ToneScanLibraryItem probes a library item's audio file with tone
// (POST /api/items/:id/tone-scan/:index). index selects the audio file
// (1-based); pass 0 for the first file. Requires admin.
func (c *Client) ToneScanLibraryItem(ctx context.Context, id string, index int) (map[string]any, error) {
	path, err := itemPath(id, "tone-scan")
	if err != nil {
		return nil, err
	}

	if index > 0 {
		path += "/" + strconv.Itoa(index)
	}

	var result map[string]any
	if err := c.Post(ctx, path, nil, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// BatchDeleteLibraryItems deletes multiple library items from the
// database (POST /api/items/batch/delete). No files are deleted.
func (c *Client) BatchDeleteLibraryItems(ctx context.Context, ids []string) error {
	return c.Post(ctx, "/api/items/batch/delete", map[string]any{"libraryItemIds": ids}, nil)
}

// BatchUpdateLibraryItems updates the media of multiple library items
// (POST /api/items/batch/update). It returns the number of items actually
// changed.
func (c *Client) BatchUpdateLibraryItems(ctx context.Context, updates []BatchUpdateItem) (int, error) {
	var resp struct {
		Success bool `json:"success"`
		Updates int  `json:"updates"`
	}

	if err := c.Post(ctx, "/api/items/batch/update", updates, &resp); err != nil {
		return 0, err
	}

	return resp.Updates, nil
}

// BatchGetLibraryItems returns multiple library items
// (POST /api/items/batch/get).
func (c *Client) BatchGetLibraryItems(ctx context.Context, ids []string) ([]LibraryItem, error) {
	var resp struct {
		LibraryItems []LibraryItem `json:"libraryItems"`
	}

	if err := c.Post(ctx, "/api/items/batch/get", map[string]any{"libraryItemIds": ids}, &resp); err != nil {
		return nil, err
	}

	c.setItemClients(resp.LibraryItems)

	return resp.LibraryItems, nil
}

// BatchQuickMatchLibraryItems quick-matches multiple library items
// (POST /api/items/batch/quickmatch). opts may be nil.
func (c *Client) BatchQuickMatchLibraryItems(ctx context.Context, ids []string, opts *QuickMatchOptions) error {
	body := map[string]any{"libraryItemIds": ids}
	if opts != nil {
		body["options"] = opts
	}

	return c.Post(ctx, "/api/items/batch/quickmatch", body, nil)
}

// Delete deletes the library item. See Client.DeleteLibraryItem.
func (i *LibraryItem) Delete(ctx context.Context, hard bool) error {
	return i.client.DeleteLibraryItem(ctx, i.ID, hard)
}

// UpdateMedia updates the item's media. See
// Client.UpdateLibraryItemMedia.
func (i *LibraryItem) UpdateMedia(ctx context.Context, update *MediaUpdate) (*UpdateMediaResult, error) {
	return i.client.UpdateLibraryItemMedia(ctx, i.ID, update)
}

// Cover fetches the item's cover image. See Client.LibraryItemCover.
func (i *LibraryItem) Cover(ctx context.Context, params *ImageParams) (io.ReadCloser, string, error) {
	return i.client.LibraryItemCover(ctx, i.ID, params)
}

// UploadCover uploads a cover image. See Client.UploadLibraryItemCover.
func (i *LibraryItem) UploadCover(ctx context.Context, filename string, cover io.Reader) error {
	return i.client.UploadLibraryItemCover(ctx, i.ID, filename, cover)
}

// RemoveCover removes the item's cover. See
// Client.RemoveLibraryItemCover.
func (i *LibraryItem) RemoveCover(ctx context.Context) error {
	return i.client.RemoveLibraryItemCover(ctx, i.ID)
}

// Match matches the item against a metadata provider. See
// Client.MatchLibraryItem.
func (i *LibraryItem) Match(ctx context.Context, req *MatchLibraryItemRequest) (*MatchResult, error) {
	return i.client.MatchLibraryItem(ctx, i.ID, req)
}

// Play starts a playback session for the item. See
// Client.PlayLibraryItem.
func (i *LibraryItem) Play(ctx context.Context, req *PlayRequest) (*PlaybackSession, error) {
	return i.client.PlayLibraryItem(ctx, i.ID, req)
}

// PlayEpisode starts a playback session for a podcast episode of the
// item. See Client.PlayPodcastEpisode.
func (i *LibraryItem) PlayEpisode(ctx context.Context, episodeID string, req *PlayRequest) (*PlaybackSession, error) {
	return i.client.PlayPodcastEpisode(ctx, i.ID, episodeID, req)
}

// Scan rescans the item's files. See Client.ScanLibraryItem.
func (i *LibraryItem) Scan(ctx context.Context) (string, error) {
	return i.client.ScanLibraryItem(ctx, i.ID)
}

// UpdateChapters replaces the item's chapters. See
// Client.UpdateLibraryItemChapters.
func (i *LibraryItem) UpdateChapters(ctx context.Context, chapters []Chapter) (bool, error) {
	return i.client.UpdateLibraryItemChapters(ctx, i.ID, chapters)
}
