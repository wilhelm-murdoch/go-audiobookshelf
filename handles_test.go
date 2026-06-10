package audiobookshelf

import (
	"context"
	"strings"
	"testing"
)

// TestResourceHandles exercises the resource-handle wrapper methods,
// confirming each delegates to the right client call (and therefore path)
// and that in-place refreshers do not panic.
func TestResourceHandles(t *testing.T) {
	tests := []struct {
		name   string
		resp   string
		call   func(ctx context.Context, c *Client) error
		method string
		path   string
	}{
		// Library handle.
		{name: "Library.Items", method: "GET", path: "/api/libraries/lib_1/items",
			call: func(ctx context.Context, c *Client) error {
				_, err := (&Library{client: c, ID: "lib_1"}).Items(ctx, nil)
				return err
			}},
		{name: "Library.Update", method: "PATCH", path: "/api/libraries/lib_1",
			call: func(ctx context.Context, c *Client) error {
				return (&Library{client: c, ID: "lib_1"}).Update(ctx, &UpdateLibraryRequest{})
			}},
		{name: "Library.Delete", method: "DELETE", path: "/api/libraries/lib_1",
			call: func(ctx context.Context, c *Client) error {
				return (&Library{client: c, ID: "lib_1"}).Delete(ctx)
			}},
		{name: "Library.Series", method: "GET", path: "/api/libraries/lib_1/series",
			call: func(ctx context.Context, c *Client) error {
				_, err := (&Library{client: c, ID: "lib_1"}).Series(ctx, nil)
				return err
			}},
		{name: "Library.Collections", method: "GET", path: "/api/libraries/lib_1/collections",
			call: func(ctx context.Context, c *Client) error {
				_, err := (&Library{client: c, ID: "lib_1"}).Collections(ctx, nil)
				return err
			}},
		{name: "Library.Playlists", method: "GET", path: "/api/libraries/lib_1/playlists",
			call: func(ctx context.Context, c *Client) error {
				_, err := (&Library{client: c, ID: "lib_1"}).Playlists(ctx, nil)
				return err
			}},
		{name: "Library.Personalized", resp: "[]", method: "GET", path: "/api/libraries/lib_1/personalized",
			call: func(ctx context.Context, c *Client) error {
				_, err := (&Library{client: c, ID: "lib_1"}).Personalized(ctx, 0, "")
				return err
			}},
		{name: "Library.FilterData", method: "GET", path: "/api/libraries/lib_1/filterdata",
			call: func(ctx context.Context, c *Client) error {
				_, err := (&Library{client: c, ID: "lib_1"}).FilterData(ctx)
				return err
			}},
		{name: "Library.Search", method: "GET", path: "/api/libraries/lib_1/search",
			call: func(ctx context.Context, c *Client) error {
				_, err := (&Library{client: c, ID: "lib_1"}).Search(ctx, "x", 0)
				return err
			}},
		{name: "Library.Stats", method: "GET", path: "/api/libraries/lib_1/stats",
			call: func(ctx context.Context, c *Client) error {
				_, err := (&Library{client: c, ID: "lib_1"}).Stats(ctx)
				return err
			}},
		{name: "Library.Authors", method: "GET", path: "/api/libraries/lib_1/authors",
			call: func(ctx context.Context, c *Client) error {
				_, err := (&Library{client: c, ID: "lib_1"}).Authors(ctx)
				return err
			}},
		{name: "Library.MatchAll", method: "GET", path: "/api/libraries/lib_1/matchall",
			call: func(ctx context.Context, c *Client) error {
				return (&Library{client: c, ID: "lib_1"}).MatchAll(ctx)
			}},
		{name: "Library.Scan", method: "POST", path: "/api/libraries/lib_1/scan",
			call: func(ctx context.Context, c *Client) error {
				return (&Library{client: c, ID: "lib_1"}).Scan(ctx, false)
			}},
		{name: "Library.RecentEpisodes", method: "GET", path: "/api/libraries/lib_1/recent-episodes",
			call: func(ctx context.Context, c *Client) error {
				_, err := (&Library{client: c, ID: "lib_1"}).RecentEpisodes(ctx, nil)
				return err
			}},
		{name: "Library.RemoveIssues", method: "DELETE", path: "/api/libraries/lib_1/issues",
			call: func(ctx context.Context, c *Client) error {
				return (&Library{client: c, ID: "lib_1"}).RemoveIssues(ctx)
			}},
		{name: "Library.EpisodeDownloads", method: "GET", path: "/api/libraries/lib_1/episode-downloads",
			call: func(ctx context.Context, c *Client) error {
				_, err := (&Library{client: c, ID: "lib_1"}).EpisodeDownloads(ctx)
				return err
			}},

		// LibraryItem handle.
		{name: "LibraryItem.Delete", method: "DELETE", path: "/api/items/li_1",
			call: func(ctx context.Context, c *Client) error {
				return (&LibraryItem{client: c, ID: "li_1"}).Delete(ctx, false)
			}},
		{name: "LibraryItem.UpdateMedia", method: "PATCH", path: "/api/items/li_1/media",
			call: func(ctx context.Context, c *Client) error {
				_, err := (&LibraryItem{client: c, ID: "li_1"}).UpdateMedia(ctx, &MediaUpdate{})
				return err
			}},
		{name: "LibraryItem.Cover", method: "GET", path: "/api/items/li_1/cover",
			call: func(ctx context.Context, c *Client) error {
				rc, _, err := (&LibraryItem{client: c, ID: "li_1"}).Cover(ctx, nil)
				if rc != nil {
					_ = rc.Close()
				}
				return err
			}},
		{name: "LibraryItem.UploadCover", method: "POST", path: "/api/items/li_1/cover",
			call: func(ctx context.Context, c *Client) error {
				return (&LibraryItem{client: c, ID: "li_1"}).UploadCover(ctx, "c.jpg", strings.NewReader("data"))
			}},
		{name: "LibraryItem.RemoveCover", method: "DELETE", path: "/api/items/li_1/cover",
			call: func(ctx context.Context, c *Client) error {
				return (&LibraryItem{client: c, ID: "li_1"}).RemoveCover(ctx)
			}},
		{name: "LibraryItem.Match", method: "POST", path: "/api/items/li_1/match",
			call: func(ctx context.Context, c *Client) error {
				_, err := (&LibraryItem{client: c, ID: "li_1"}).Match(ctx, &MatchLibraryItemRequest{})
				return err
			}},
		{name: "LibraryItem.Play", method: "POST", path: "/api/items/li_1/play",
			call: func(ctx context.Context, c *Client) error {
				_, err := (&LibraryItem{client: c, ID: "li_1"}).Play(ctx, nil)
				return err
			}},
		{name: "LibraryItem.PlayEpisode", method: "POST", path: "/api/items/li_1/play/ep_1",
			call: func(ctx context.Context, c *Client) error {
				_, err := (&LibraryItem{client: c, ID: "li_1"}).PlayEpisode(ctx, "ep_1", nil)
				return err
			}},
		{name: "LibraryItem.Scan", method: "POST", path: "/api/items/li_1/scan",
			call: func(ctx context.Context, c *Client) error {
				_, err := (&LibraryItem{client: c, ID: "li_1"}).Scan(ctx)
				return err
			}},
		{name: "LibraryItem.UpdateChapters", method: "POST", path: "/api/items/li_1/chapters",
			call: func(ctx context.Context, c *Client) error {
				_, err := (&LibraryItem{client: c, ID: "li_1"}).UpdateChapters(ctx, nil)
				return err
			}},
		{name: "LibraryItem.OpenRSSFeed", method: "POST", path: "/api/feeds/item/li_1/open",
			call: func(ctx context.Context, c *Client) error {
				_, err := (&LibraryItem{client: c, ID: "li_1"}).OpenRSSFeed(ctx, &OpenRSSFeedRequest{})
				return err
			}},

		// Collection handle.
		{name: "Collection.Update", method: "PATCH", path: "/api/collections/co_1",
			call: func(ctx context.Context, c *Client) error {
				return (&Collection{client: c, ID: "co_1"}).Update(ctx, &UpdateCollectionRequest{})
			}},
		{name: "Collection.Delete", method: "DELETE", path: "/api/collections/co_1",
			call: func(ctx context.Context, c *Client) error {
				return (&Collection{client: c, ID: "co_1"}).Delete(ctx)
			}},
		{name: "Collection.AddBook", method: "POST", path: "/api/collections/co_1/book",
			call: func(ctx context.Context, c *Client) error {
				return (&Collection{client: c, ID: "co_1"}).AddBook(ctx, "li_1")
			}},
		{name: "Collection.RemoveBook", method: "DELETE", path: "/api/collections/co_1/book/li_1",
			call: func(ctx context.Context, c *Client) error {
				return (&Collection{client: c, ID: "co_1"}).RemoveBook(ctx, "li_1")
			}},
		{name: "Collection.AddBooks", method: "POST", path: "/api/collections/co_1/batch/add",
			call: func(ctx context.Context, c *Client) error {
				return (&Collection{client: c, ID: "co_1"}).AddBooks(ctx, nil)
			}},
		{name: "Collection.RemoveBooks", method: "POST", path: "/api/collections/co_1/batch/remove",
			call: func(ctx context.Context, c *Client) error {
				return (&Collection{client: c, ID: "co_1"}).RemoveBooks(ctx, nil)
			}},
		{name: "Collection.CreatePlaylist", method: "POST", path: "/api/playlists/collection/co_1",
			call: func(ctx context.Context, c *Client) error {
				_, err := (&Collection{client: c, ID: "co_1"}).CreatePlaylist(ctx)
				return err
			}},
		{name: "Collection.OpenRSSFeed", method: "POST", path: "/api/feeds/collection/co_1/open",
			call: func(ctx context.Context, c *Client) error {
				_, err := (&Collection{client: c, ID: "co_1"}).OpenRSSFeed(ctx, &OpenRSSFeedRequest{})
				return err
			}},

		// Playlist handle.
		{name: "Playlist.Update", method: "PATCH", path: "/api/playlists/pl_1",
			call: func(ctx context.Context, c *Client) error {
				return (&Playlist{client: c, ID: "pl_1"}).Update(ctx, &UpdatePlaylistRequest{})
			}},
		{name: "Playlist.Delete", method: "DELETE", path: "/api/playlists/pl_1",
			call: func(ctx context.Context, c *Client) error {
				return (&Playlist{client: c, ID: "pl_1"}).Delete(ctx)
			}},
		{name: "Playlist.AddItem", method: "POST", path: "/api/playlists/pl_1/item",
			call: func(ctx context.Context, c *Client) error {
				return (&Playlist{client: c, ID: "pl_1"}).AddItem(ctx, "li_1", "")
			}},
		{name: "Playlist.RemoveItem", method: "DELETE", path: "/api/playlists/pl_1/item/li_1",
			call: func(ctx context.Context, c *Client) error {
				return (&Playlist{client: c, ID: "pl_1"}).RemoveItem(ctx, "li_1", "")
			}},
		{name: "Playlist.AddItems", method: "POST", path: "/api/playlists/pl_1/batch/add",
			call: func(ctx context.Context, c *Client) error {
				return (&Playlist{client: c, ID: "pl_1"}).AddItems(ctx, nil)
			}},
		{name: "Playlist.RemoveItems", method: "POST", path: "/api/playlists/pl_1/batch/remove",
			call: func(ctx context.Context, c *Client) error {
				return (&Playlist{client: c, ID: "pl_1"}).RemoveItems(ctx, nil)
			}},

		// User handle.
		{name: "User.Update", method: "PATCH", path: "/api/users/us_1",
			call: func(ctx context.Context, c *Client) error {
				return (&User{client: c, ID: "us_1"}).Update(ctx, &UpdateUserRequest{})
			}},
		{name: "User.Delete", method: "DELETE", path: "/api/users/us_1",
			call: func(ctx context.Context, c *Client) error {
				return (&User{client: c, ID: "us_1"}).Delete(ctx)
			}},
		{name: "User.ListeningSessions", method: "GET", path: "/api/users/us_1/listening-sessions",
			call: func(ctx context.Context, c *Client) error {
				_, err := (&User{client: c, ID: "us_1"}).ListeningSessions(ctx, nil)
				return err
			}},
		{name: "User.ListeningStats", method: "GET", path: "/api/users/us_1/listening-stats",
			call: func(ctx context.Context, c *Client) error {
				_, err := (&User{client: c, ID: "us_1"}).ListeningStats(ctx)
				return err
			}},
		{name: "User.PurgeMediaProgress", method: "POST", path: "/api/users/us_1/purge-media-progress",
			call: func(ctx context.Context, c *Client) error {
				return (&User{client: c, ID: "us_1"}).PurgeMediaProgress(ctx)
			}},

		// Author handle.
		{name: "Author.Update", method: "PATCH", path: "/api/authors/au_1",
			call: func(ctx context.Context, c *Client) error {
				_, err := (&Author{client: c, ID: "au_1"}).Update(ctx, &UpdateAuthorRequest{})
				return err
			}},
		{name: "Author.Match", method: "POST", path: "/api/authors/au_1/match",
			call: func(ctx context.Context, c *Client) error {
				_, err := (&Author{client: c, ID: "au_1"}).Match(ctx, &MatchAuthorRequest{})
				return err
			}},
		{name: "Author.Image", method: "GET", path: "/api/authors/au_1/image",
			call: func(ctx context.Context, c *Client) error {
				rc, _, err := (&Author{client: c, ID: "au_1"}).Image(ctx, nil)
				if rc != nil {
					_ = rc.Close()
				}
				return err
			}},

		// Series handle.
		{name: "Series.Update", method: "PATCH", path: "/api/series/se_1",
			call: func(ctx context.Context, c *Client) error {
				return (&Series{client: c, ID: "se_1"}).Update(ctx, &UpdateSeriesRequest{})
			}},
		{name: "Series.OpenRSSFeed", method: "POST", path: "/api/feeds/series/se_1/open",
			call: func(ctx context.Context, c *Client) error {
				_, err := (&Series{client: c, ID: "se_1"}).OpenRSSFeed(ctx, &OpenRSSFeedRequest{})
				return err
			}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, cap := recordingClient(t, tt.resp)

			if err := tt.call(context.Background(), client); err != nil {
				t.Fatalf("%s: %v", tt.name, err)
			}

			if cap.method != tt.method {
				t.Errorf("method = %s, want %s", cap.method, tt.method)
			}

			if cap.path != tt.path {
				t.Errorf("path = %s, want %s", cap.path, tt.path)
			}
		})
	}
}
