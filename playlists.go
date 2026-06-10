package audiobookshelf

import (
	"context"
	"net/url"
)

// CreatePlaylistRequest are the parameters for CreatePlaylist.
type CreatePlaylistRequest struct {
	LibraryID   string         `json:"libraryId"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	CoverPath   string         `json:"coverPath,omitempty"`
	Items       []PlaylistItem `json:"items,omitempty"`
}

// UpdatePlaylistRequest are the parameters for UpdatePlaylist. Nil/zero
// fields are left unchanged.
type UpdatePlaylistRequest struct {
	Name        string  `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	CoverPath   *string `json:"coverPath,omitempty"`
	// Items replaces the playlist's items.
	Items []PlaylistItem `json:"items,omitempty"`
}

func playlistPath(id string, rest ...string) string {
	path := "/api/playlists/" + url.PathEscape(id)
	for _, r := range rest {
		path += "/" + r
	}
	return path
}

// CreatePlaylist creates a playlist (POST /api/playlists).
func (c *Client) CreatePlaylist(ctx context.Context, req *CreatePlaylistRequest) (*Playlist, error) {
	var playlist Playlist
	if err := c.Post(ctx, "/api/playlists", req, &playlist); err != nil {
		return nil, err
	}
	playlist.client = c
	return &playlist, nil
}

// Playlists returns all of your playlists (GET /api/playlists).
func (c *Client) Playlists(ctx context.Context) ([]Playlist, error) {
	var resp struct {
		Playlists []Playlist `json:"playlists"`
	}
	if err := c.Get(ctx, "/api/playlists", &resp); err != nil {
		return nil, err
	}
	for i := range resp.Playlists {
		resp.Playlists[i].client = c
	}
	return resp.Playlists, nil
}

// Playlist returns a playlist (GET /api/playlists/:id).
func (c *Client) Playlist(ctx context.Context, id string) (*Playlist, error) {
	var playlist Playlist
	if err := c.Get(ctx, playlistPath(id), &playlist); err != nil {
		return nil, err
	}
	playlist.client = c
	return &playlist, nil
}

// UpdatePlaylist updates a playlist (PATCH /api/playlists/:id).
func (c *Client) UpdatePlaylist(ctx context.Context, id string, req *UpdatePlaylistRequest) (*Playlist, error) {
	var playlist Playlist
	if err := c.Patch(ctx, playlistPath(id), req, &playlist); err != nil {
		return nil, err
	}
	playlist.client = c
	return &playlist, nil
}

// DeletePlaylist deletes a playlist (DELETE /api/playlists/:id).
func (c *Client) DeletePlaylist(ctx context.Context, id string) error {
	return c.Delete(ctx, playlistPath(id), nil)
}

// AddItemToPlaylist adds an item to a playlist
// (POST /api/playlists/:id/item). episodeID may be empty for books.
func (c *Client) AddItemToPlaylist(ctx context.Context, id, libraryItemID, episodeID string) (*Playlist, error) {
	body := map[string]any{"libraryItemId": libraryItemID}
	if episodeID != "" {
		body["episodeId"] = episodeID
	}
	var playlist Playlist
	if err := c.Post(ctx, playlistPath(id, "item"), body, &playlist); err != nil {
		return nil, err
	}
	playlist.client = c
	return &playlist, nil
}

// RemoveItemFromPlaylist removes an item from a playlist
// (DELETE /api/playlists/:id/item/:libraryItemId[/:episodeId]).
// episodeID may be empty for books.
func (c *Client) RemoveItemFromPlaylist(ctx context.Context, id, libraryItemID, episodeID string) (*Playlist, error) {
	path := playlistPath(id, "item", url.PathEscape(libraryItemID))
	if episodeID != "" {
		path += "/" + url.PathEscape(episodeID)
	}
	var playlist Playlist
	if err := c.Delete(ctx, path, &playlist); err != nil {
		return nil, err
	}
	playlist.client = c
	return &playlist, nil
}

// BatchAddPlaylistItems adds multiple items to a playlist
// (POST /api/playlists/:id/batch/add).
func (c *Client) BatchAddPlaylistItems(ctx context.Context, id string, items []PlaylistItem) (*Playlist, error) {
	var playlist Playlist
	if err := c.Post(ctx, playlistPath(id, "batch", "add"), map[string]any{"items": items}, &playlist); err != nil {
		return nil, err
	}
	playlist.client = c
	return &playlist, nil
}

// BatchRemovePlaylistItems removes multiple items from a playlist
// (POST /api/playlists/:id/batch/remove).
func (c *Client) BatchRemovePlaylistItems(ctx context.Context, id string, items []PlaylistItem) (*Playlist, error) {
	var playlist Playlist
	if err := c.Post(ctx, playlistPath(id, "batch", "remove"), map[string]any{"items": items}, &playlist); err != nil {
		return nil, err
	}
	playlist.client = c
	return &playlist, nil
}

// CreatePlaylistFromCollection creates a playlist with the same items as
// a collection (POST /api/playlists/collection/:collectionId).
func (c *Client) CreatePlaylistFromCollection(ctx context.Context, collectionID string) (*Playlist, error) {
	var playlist Playlist
	if err := c.Post(ctx, "/api/playlists/collection/"+url.PathEscape(collectionID), nil, &playlist); err != nil {
		return nil, err
	}
	playlist.client = c
	return &playlist, nil
}

// CreatePlaylist creates a playlist from the collection. See
// Client.CreatePlaylistFromCollection.
func (col *Collection) CreatePlaylist(ctx context.Context) (*Playlist, error) {
	return col.client.CreatePlaylistFromCollection(ctx, col.ID)
}

// Update updates the playlist and refreshes its fields in place. See
// Client.UpdatePlaylist.
func (p *Playlist) Update(ctx context.Context, req *UpdatePlaylistRequest) error {
	updated, err := p.client.UpdatePlaylist(ctx, p.ID, req)
	if err != nil {
		return err
	}
	*p = *updated
	return nil
}

// Delete deletes the playlist. See Client.DeletePlaylist.
func (p *Playlist) Delete(ctx context.Context) error {
	return p.client.DeletePlaylist(ctx, p.ID)
}

// AddItem adds an item to the playlist and refreshes its fields in
// place. See Client.AddItemToPlaylist.
func (p *Playlist) AddItem(ctx context.Context, libraryItemID, episodeID string) error {
	updated, err := p.client.AddItemToPlaylist(ctx, p.ID, libraryItemID, episodeID)
	if err != nil {
		return err
	}
	*p = *updated
	return nil
}

// RemoveItem removes an item from the playlist and refreshes its fields
// in place. See Client.RemoveItemFromPlaylist.
func (p *Playlist) RemoveItem(ctx context.Context, libraryItemID, episodeID string) error {
	updated, err := p.client.RemoveItemFromPlaylist(ctx, p.ID, libraryItemID, episodeID)
	if err != nil {
		return err
	}
	*p = *updated
	return nil
}

// AddItems adds multiple items to the playlist and refreshes its fields
// in place. See Client.BatchAddPlaylistItems.
func (p *Playlist) AddItems(ctx context.Context, items []PlaylistItem) error {
	updated, err := p.client.BatchAddPlaylistItems(ctx, p.ID, items)
	if err != nil {
		return err
	}
	*p = *updated
	return nil
}

// RemoveItems removes multiple items from the playlist and refreshes its
// fields in place. See Client.BatchRemovePlaylistItems.
func (p *Playlist) RemoveItems(ctx context.Context, items []PlaylistItem) error {
	updated, err := p.client.BatchRemovePlaylistItems(ctx, p.ID, items)
	if err != nil {
		return err
	}
	*p = *updated
	return nil
}
