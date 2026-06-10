package audiobookshelf

import (
	"context"
	"errors"
	"strconv"
)

// MediaProgressUpdate are the parameters for UpdateMyMediaProgress. Only
// set fields are sent.
type MediaProgressUpdate struct {
	// Duration is the total duration of the media in seconds.
	Duration *float64 `json:"duration,omitempty"`
	// Progress is the completion fraction of the media (0..1).
	Progress *float64 `json:"progress,omitempty"`
	// CurrentTime is the playback position in seconds.
	CurrentTime               *float64 `json:"currentTime,omitempty"`
	IsFinished                *bool    `json:"isFinished,omitempty"`
	HideFromContinueListening *bool    `json:"hideFromContinueListening,omitempty"`
	FinishedAt                *int64   `json:"finishedAt,omitempty"`
	StartedAt                 *int64   `json:"startedAt,omitempty"`
}

// MediaProgressBatchUpdate is one entry of BatchUpdateMyMediaProgress.
type MediaProgressBatchUpdate struct {
	LibraryItemID string `json:"libraryItemId"`
	EpisodeID     string `json:"episodeId,omitempty"`
	MediaProgressUpdate
}

// LocalProgressSyncResult is the response of SyncLocalMediaProgress.
type LocalProgressSyncResult struct {
	NumServerProgressUpdates int             `json:"numServerProgressUpdates"`
	LocalProgressUpdates     []MediaProgress `json:"localProgressUpdates,omitempty"`
	ServerProgressUpdates    []MediaProgress `json:"serverProgressUpdates,omitempty"`
}

func myProgressPath(libraryItemID, episodeID string) string {
	return apiPath("me", "progress").Seg(libraryItemID, episodeID).String()
}

// Me returns the authenticated user (GET /api/me).
func (c *Client) Me(ctx context.Context) (*User, error) {
	var user User
	if err := c.Get(ctx, apiPath("me").String(), &user); err != nil {
		return nil, err
	}
	user.client = c
	return &user, nil
}

// MyListeningSessions lists your listening sessions
// (GET /api/me/listening-sessions).
func (c *Client) MyListeningSessions(ctx context.Context, params *SessionListParams) (*SessionsPage, error) {
	var page SessionsPage
	path := apiPath("me", "listening-sessions").Query(params.values()).String()
	if err := c.Get(ctx, path, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// MyListeningStats returns your listening statistics
// (GET /api/me/listening-stats).
func (c *Client) MyListeningStats(ctx context.Context) (*ListeningStats, error) {
	var stats ListeningStats
	if err := c.Get(ctx, apiPath("me", "listening-stats").String(), &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

// RemoveItemFromContinueListening hides a media progress's item from the
// "Continue Listening" shelf
// (GET /api/me/progress/:id/remove-from-continue-listening). progressID
// is a MediaProgress.ID. It returns your updated user.
func (c *Client) RemoveItemFromContinueListening(ctx context.Context, progressID string) (*User, error) {
	var user User
	path := apiPath("me", "progress").Seg(progressID).Lit("remove-from-continue-listening").String()
	if err := c.Get(ctx, path, &user); err != nil {
		return nil, err
	}
	user.client = c
	return &user, nil
}

// MyMediaProgress returns your progress for a library item
// (GET /api/me/progress/:libraryItemId[/:episodeId]). episodeID may be
// empty for books.
func (c *Client) MyMediaProgress(ctx context.Context, libraryItemID, episodeID string) (*MediaProgress, error) {
	var progress MediaProgress
	if err := c.Get(ctx, myProgressPath(libraryItemID, episodeID), &progress); err != nil {
		return nil, err
	}
	return &progress, nil
}

// UpdateMyMediaProgress creates or updates your progress for a library
// item or podcast episode
// (PATCH /api/me/progress/:libraryItemId[/:episodeId]). episodeID may be
// empty for books.
func (c *Client) UpdateMyMediaProgress(ctx context.Context, libraryItemID, episodeID string, update *MediaProgressUpdate) error {
	return c.Patch(ctx, myProgressPath(libraryItemID, episodeID), update, nil)
}

// BatchUpdateMyMediaProgress creates or updates multiple media progress
// entries (PATCH /api/me/progress/batch/update).
func (c *Client) BatchUpdateMyMediaProgress(ctx context.Context, updates []MediaProgressBatchUpdate) error {
	return c.Patch(ctx, apiPath("me", "progress", "batch", "update").String(), updates, nil)
}

// RemoveMyMediaProgress removes one of your media progress entries
// (DELETE /api/me/progress/:id). progressID is a MediaProgress.ID.
func (c *Client) RemoveMyMediaProgress(ctx context.Context, progressID string) error {
	return c.Delete(ctx, apiPath("me", "progress").Seg(progressID).String(), nil)
}

// CreateBookmark creates a bookmark on a book
// (POST /api/me/item/:id/bookmark). time is the position in seconds.
func (c *Client) CreateBookmark(ctx context.Context, libraryItemID string, time int, title string) (*Bookmark, error) {
	body := map[string]any{"time": time, "title": title}
	var bookmark Bookmark
	if err := c.Post(ctx, apiPath("me", "item").Seg(libraryItemID).Lit("bookmark").String(), body, &bookmark); err != nil {
		return nil, err
	}
	return &bookmark, nil
}

// UpdateBookmark changes the title of the bookmark at time
// (PATCH /api/me/item/:id/bookmark).
func (c *Client) UpdateBookmark(ctx context.Context, libraryItemID string, time int, title string) (*Bookmark, error) {
	body := map[string]any{"time": time, "title": title}
	var bookmark Bookmark
	if err := c.Patch(ctx, apiPath("me", "item").Seg(libraryItemID).Lit("bookmark").String(), body, &bookmark); err != nil {
		return nil, err
	}
	return &bookmark, nil
}

// RemoveBookmark removes the bookmark at time
// (DELETE /api/me/item/:id/bookmark/:time).
func (c *Client) RemoveBookmark(ctx context.Context, libraryItemID string, time int) error {
	path := apiPath("me", "item").Seg(libraryItemID).Lit("bookmark", strconv.Itoa(time)).String()
	return c.Delete(ctx, path, nil)
}

// ChangeMyPassword changes your password (PATCH /api/me/password).
func (c *Client) ChangeMyPassword(ctx context.Context, currentPassword, newPassword string) error {
	body := map[string]string{"password": currentPassword, "newPassword": newPassword}
	var resp struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
	if err := c.Patch(ctx, apiPath("me", "password").String(), body, &resp); err != nil {
		return err
	}
	if !resp.Success {
		if resp.Error != "" {
			return errors.New("audiobookshelf: change password: " + resp.Error)
		}
		return errors.New("audiobookshelf: change password failed")
	}
	return nil
}

// SyncLocalMediaProgress syncs media progress from a client device
// (POST /api/me/sync-local-progress).
func (c *Client) SyncLocalMediaProgress(ctx context.Context, localProgress []MediaProgress) (*LocalProgressSyncResult, error) {
	body := map[string]any{"localMediaProgress": localProgress}
	var result LocalProgressSyncResult
	if err := c.Post(ctx, apiPath("me", "sync-local-progress").String(), body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// MyItemsInProgress lists your in-progress library items
// (GET /api/me/items-in-progress). limit caps the number of items (server
// default 25). Podcast items carry RecentEpisode; all items carry
// ProgressLastUpdate.
func (c *Client) MyItemsInProgress(ctx context.Context, limit int) ([]LibraryItem, error) {
	pb := apiPath("me", "items-in-progress")
	if limit > 0 {
		pb.Set("limit", strconv.Itoa(limit))
	}
	var resp struct {
		LibraryItems []LibraryItem `json:"libraryItems"`
	}
	if err := c.Get(ctx, pb.String(), &resp); err != nil {
		return nil, err
	}
	c.setItemClients(resp.LibraryItems)
	return resp.LibraryItems, nil
}

// RemoveSeriesFromContinueListening hides a series from the "Continue
// Series" shelf
// (GET /api/me/series/:id/remove-from-continue-listening). It returns
// your updated user.
func (c *Client) RemoveSeriesFromContinueListening(ctx context.Context, seriesID string) (*User, error) {
	var user User
	path := apiPath("me", "series").Seg(seriesID).Lit("remove-from-continue-listening").String()
	if err := c.Get(ctx, path, &user); err != nil {
		return nil, err
	}
	user.client = c
	return &user, nil
}
