package audiobookshelf

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// captured records the request a client method produced.
type captured struct {
	method string
	path   string
	query  url.Values
	body   map[string]any
}

// recordingClient returns a client whose server records the request and
// replies with respBody (or "{}" when empty). The recorder is shared by
// reference so assertions can inspect it after the call.
func recordingClient(t *testing.T, respBody string) (*Client, *captured) {
	t.Helper()

	cap := &captured{}
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		cap.method = r.Method
		cap.path = r.URL.Path
		cap.query = r.URL.Query()

		if raw, _ := io.ReadAll(r.Body); len(raw) > 0 {
			_ = json.Unmarshal(raw, &cap.body)
		}

		if respBody == "" {
			respBody = "{}"
		}

		if _, err := io.WriteString(w, respBody); err != nil {
			t.Errorf("writing response: %v", err)
		}
	})

	return client, cap
}

// TestEndpointRequests drives every Client method and asserts the HTTP
// method, escaped path, and any query parameters it produced. This locks
// in the path construction for the whole API surface.
func TestEndpointRequests(t *testing.T) {
	tests := []struct {
		name   string
		resp   string
		call   func(ctx context.Context, c *Client) error
		method string
		path   string
		query  map[string]string
	}{
		// libraries.go
		{name: "CreateLibrary", method: "POST", path: "/api/libraries",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.CreateLibrary(ctx, &CreateLibraryRequest{Name: "x"})
				return err
			}},
		{name: "Library", method: "GET", path: "/api/libraries/lib_1",
			call: func(ctx context.Context, c *Client) error { _, err := c.Library(ctx, "lib_1"); return err }},
		{name: "UpdateLibrary", method: "PATCH", path: "/api/libraries/lib_1",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.UpdateLibrary(ctx, "lib_1", &UpdateLibraryRequest{Name: "y"})
				return err
			}},
		{name: "DeleteLibrary", method: "DELETE", path: "/api/libraries/lib_1",
			call: func(ctx context.Context, c *Client) error { return c.DeleteLibrary(ctx, "lib_1") }},
		{name: "LibrarySeries", method: "GET", path: "/api/libraries/lib_1/series",
			call: func(ctx context.Context, c *Client) error { _, err := c.LibrarySeries(ctx, "lib_1", nil); return err }},
		{name: "LibraryCollections", method: "GET", path: "/api/libraries/lib_1/collections",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.LibraryCollections(ctx, "lib_1", nil)
				return err
			}},
		{name: "LibraryPlaylists", method: "GET", path: "/api/libraries/lib_1/playlists",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.LibraryPlaylists(ctx, "lib_1", nil)
				return err
			}},
		{name: "LibraryPersonalized", resp: "[]", method: "GET", path: "/api/libraries/lib_1/personalized",
			query: map[string]string{"limit": "5", "include": "rssfeed"},
			call: func(ctx context.Context, c *Client) error {
				_, err := c.LibraryPersonalized(ctx, "lib_1", 5, "rssfeed")
				return err
			}},
		{name: "LibraryFilterData", method: "GET", path: "/api/libraries/lib_1/filterdata",
			call: func(ctx context.Context, c *Client) error { _, err := c.LibraryFilterData(ctx, "lib_1"); return err }},
		{name: "SearchLibrary", method: "GET", path: "/api/libraries/lib_1/search",
			query: map[string]string{"q": "dune", "limit": "3"},
			call: func(ctx context.Context, c *Client) error {
				_, err := c.SearchLibrary(ctx, "lib_1", "dune", 3)
				return err
			}},
		{name: "LibraryStats", method: "GET", path: "/api/libraries/lib_1/stats",
			call: func(ctx context.Context, c *Client) error { _, err := c.LibraryStats(ctx, "lib_1"); return err }},
		{name: "LibraryAuthors", method: "GET", path: "/api/libraries/lib_1/authors",
			call: func(ctx context.Context, c *Client) error { _, err := c.LibraryAuthors(ctx, "lib_1"); return err }},
		{name: "MatchAllLibraryItems", method: "GET", path: "/api/libraries/lib_1/matchall",
			call: func(ctx context.Context, c *Client) error { return c.MatchAllLibraryItems(ctx, "lib_1") }},
		{name: "ScanLibrary", method: "POST", path: "/api/libraries/lib_1/scan",
			query: map[string]string{"force": "1"},
			call:  func(ctx context.Context, c *Client) error { return c.ScanLibrary(ctx, "lib_1", true) }},
		{name: "LibraryRecentEpisodes", method: "GET", path: "/api/libraries/lib_1/recent-episodes",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.LibraryRecentEpisodes(ctx, "lib_1", nil)
				return err
			}},
		{name: "LibraryEpisodeDownloads", method: "GET", path: "/api/libraries/lib_1/episode-downloads",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.LibraryEpisodeDownloads(ctx, "lib_1")
				return err
			}},
		{name: "RemoveLibraryIssues", method: "DELETE", path: "/api/libraries/lib_1/issues",
			call: func(ctx context.Context, c *Client) error { return c.RemoveLibraryIssues(ctx, "lib_1") }},
		{name: "ReorderLibraries", method: "POST", path: "/api/libraries/order",
			call: func(ctx context.Context, c *Client) error { _, err := c.ReorderLibraries(ctx, nil); return err }},

		// items.go
		{name: "DeleteAllLibraryItems", method: "DELETE", path: "/api/items/all",
			call: func(ctx context.Context, c *Client) error { return c.DeleteAllLibraryItems(ctx) }},
		{name: "LibraryItem", method: "GET", path: "/api/items/li_1",
			query: map[string]string{"expanded": "1"},
			call: func(ctx context.Context, c *Client) error {
				_, err := c.LibraryItem(ctx, "li_1", &LibraryItemParams{Expanded: true})
				return err
			}},
		{name: "DeleteLibraryItem", method: "DELETE", path: "/api/items/li_1",
			query: map[string]string{"hard": "1"},
			call:  func(ctx context.Context, c *Client) error { return c.DeleteLibraryItem(ctx, "li_1", true) }},
		{name: "UpdateLibraryItemMedia", method: "PATCH", path: "/api/items/li_1/media",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.UpdateLibraryItemMedia(ctx, "li_1", &MediaUpdate{})
				return err
			}},
		{name: "SetLibraryItemCoverFromURL", method: "POST", path: "/api/items/li_1/cover",
			call: func(ctx context.Context, c *Client) error {
				return c.SetLibraryItemCoverFromURL(ctx, "li_1", "http://x/c.jpg")
			}},
		{name: "UpdateLibraryItemCover", method: "PATCH", path: "/api/items/li_1/cover",
			call: func(ctx context.Context, c *Client) error { return c.UpdateLibraryItemCover(ctx, "li_1", "/c.jpg") }},
		{name: "RemoveLibraryItemCover", method: "DELETE", path: "/api/items/li_1/cover",
			call: func(ctx context.Context, c *Client) error { return c.RemoveLibraryItemCover(ctx, "li_1") }},
		{name: "MatchLibraryItem", method: "POST", path: "/api/items/li_1/match",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.MatchLibraryItem(ctx, "li_1", &MatchLibraryItemRequest{})
				return err
			}},
		{name: "PlayLibraryItem", method: "POST", path: "/api/items/li_1/play",
			call: func(ctx context.Context, c *Client) error { _, err := c.PlayLibraryItem(ctx, "li_1", nil); return err }},
		{name: "PlayPodcastEpisode", method: "POST", path: "/api/items/li_1/play/ep_1",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.PlayPodcastEpisode(ctx, "li_1", "ep_1", nil)
				return err
			}},
		{name: "UpdateLibraryItemTracks", method: "PATCH", path: "/api/items/li_1/tracks",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.UpdateLibraryItemTracks(ctx, "li_1", nil)
				return err
			}},
		{name: "ScanLibraryItem", method: "POST", path: "/api/items/li_1/scan",
			call: func(ctx context.Context, c *Client) error { _, err := c.ScanLibraryItem(ctx, "li_1"); return err }},
		{name: "LibraryItemToneObject", method: "GET", path: "/api/items/li_1/tone-object",
			call: func(ctx context.Context, c *Client) error { _, err := c.LibraryItemToneObject(ctx, "li_1"); return err }},
		{name: "UpdateLibraryItemChapters", method: "POST", path: "/api/items/li_1/chapters",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.UpdateLibraryItemChapters(ctx, "li_1", nil)
				return err
			}},
		{name: "ToneScanLibraryItem", method: "POST", path: "/api/items/li_1/tone-scan/2",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.ToneScanLibraryItem(ctx, "li_1", 2)
				return err
			}},
		{name: "BatchDeleteLibraryItems", method: "POST", path: "/api/items/batch/delete",
			call: func(ctx context.Context, c *Client) error { return c.BatchDeleteLibraryItems(ctx, []string{"li_1"}) }},
		{name: "BatchUpdateLibraryItems", method: "POST", path: "/api/items/batch/update",
			call: func(ctx context.Context, c *Client) error { _, err := c.BatchUpdateLibraryItems(ctx, nil); return err }},
		{name: "BatchGetLibraryItems", method: "POST", path: "/api/items/batch/get",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.BatchGetLibraryItems(ctx, []string{"li_1"})
				return err
			}},
		{name: "BatchQuickMatchLibraryItems", method: "POST", path: "/api/items/batch/quickmatch",
			call: func(ctx context.Context, c *Client) error {
				return c.BatchQuickMatchLibraryItems(ctx, []string{"li_1"}, nil)
			}},

		// collections.go
		{name: "CreateCollection", method: "POST", path: "/api/collections",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.CreateCollection(ctx, &CreateCollectionRequest{Name: "x"})
				return err
			}},
		{name: "Collections", method: "GET", path: "/api/collections",
			call: func(ctx context.Context, c *Client) error { _, err := c.Collections(ctx); return err }},
		{name: "Collection", method: "GET", path: "/api/collections/co_1",
			query: map[string]string{"include": "rssfeed"},
			call:  func(ctx context.Context, c *Client) error { _, err := c.Collection(ctx, "co_1", "rssfeed"); return err }},
		{name: "UpdateCollection", method: "PATCH", path: "/api/collections/co_1",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.UpdateCollection(ctx, "co_1", &UpdateCollectionRequest{})
				return err
			}},
		{name: "DeleteCollection", method: "DELETE", path: "/api/collections/co_1",
			call: func(ctx context.Context, c *Client) error { return c.DeleteCollection(ctx, "co_1") }},
		{name: "AddBookToCollection", method: "POST", path: "/api/collections/co_1/book",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.AddBookToCollection(ctx, "co_1", "li_1")
				return err
			}},
		{name: "RemoveBookFromCollection", method: "DELETE", path: "/api/collections/co_1/book/li_1",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.RemoveBookFromCollection(ctx, "co_1", "li_1")
				return err
			}},
		{name: "BatchAddBooksToCollection", method: "POST", path: "/api/collections/co_1/batch/add",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.BatchAddBooksToCollection(ctx, "co_1", nil)
				return err
			}},
		{name: "BatchRemoveBooksFromCollection", method: "POST", path: "/api/collections/co_1/batch/remove",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.BatchRemoveBooksFromCollection(ctx, "co_1", nil)
				return err
			}},

		// playlists.go
		{name: "CreatePlaylist", method: "POST", path: "/api/playlists",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.CreatePlaylist(ctx, &CreatePlaylistRequest{Name: "x"})
				return err
			}},
		{name: "Playlists", method: "GET", path: "/api/playlists",
			call: func(ctx context.Context, c *Client) error { _, err := c.Playlists(ctx); return err }},
		{name: "Playlist", method: "GET", path: "/api/playlists/pl_1",
			call: func(ctx context.Context, c *Client) error { _, err := c.Playlist(ctx, "pl_1"); return err }},
		{name: "UpdatePlaylist", method: "PATCH", path: "/api/playlists/pl_1",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.UpdatePlaylist(ctx, "pl_1", &UpdatePlaylistRequest{})
				return err
			}},
		{name: "DeletePlaylist", method: "DELETE", path: "/api/playlists/pl_1",
			call: func(ctx context.Context, c *Client) error { return c.DeletePlaylist(ctx, "pl_1") }},
		{name: "AddItemToPlaylist", method: "POST", path: "/api/playlists/pl_1/item",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.AddItemToPlaylist(ctx, "pl_1", "li_1", "")
				return err
			}},
		{name: "RemoveItemFromPlaylist", method: "DELETE", path: "/api/playlists/pl_1/item/li_1/ep_1",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.RemoveItemFromPlaylist(ctx, "pl_1", "li_1", "ep_1")
				return err
			}},
		{name: "BatchAddPlaylistItems", method: "POST", path: "/api/playlists/pl_1/batch/add",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.BatchAddPlaylistItems(ctx, "pl_1", nil)
				return err
			}},
		{name: "BatchRemovePlaylistItems", method: "POST", path: "/api/playlists/pl_1/batch/remove",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.BatchRemovePlaylistItems(ctx, "pl_1", nil)
				return err
			}},
		{name: "CreatePlaylistFromCollection", method: "POST", path: "/api/playlists/collection/co_1",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.CreatePlaylistFromCollection(ctx, "co_1")
				return err
			}},

		// users.go
		{name: "CreateUser", method: "POST", path: "/api/users",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.CreateUser(ctx, &CreateUserRequest{Username: "u"})
				return err
			}},
		{name: "Users", method: "GET", path: "/api/users",
			call: func(ctx context.Context, c *Client) error { _, err := c.Users(ctx); return err }},
		{name: "OnlineUsers", method: "GET", path: "/api/users/online",
			call: func(ctx context.Context, c *Client) error { _, err := c.OnlineUsers(ctx); return err }},
		{name: "User", method: "GET", path: "/api/users/us_1",
			call: func(ctx context.Context, c *Client) error { _, err := c.User(ctx, "us_1"); return err }},
		{name: "UpdateUser", method: "PATCH", path: "/api/users/us_1",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.UpdateUser(ctx, "us_1", &UpdateUserRequest{})
				return err
			}},
		{name: "DeleteUser", method: "DELETE", path: "/api/users/us_1",
			call: func(ctx context.Context, c *Client) error { return c.DeleteUser(ctx, "us_1") }},
		{name: "UserListeningSessions", method: "GET", path: "/api/users/us_1/listening-sessions",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.UserListeningSessions(ctx, "us_1", nil)
				return err
			}},
		{name: "UserListeningStats", method: "GET", path: "/api/users/us_1/listening-stats",
			call: func(ctx context.Context, c *Client) error { _, err := c.UserListeningStats(ctx, "us_1"); return err }},
		{name: "PurgeUserMediaProgress", method: "POST", path: "/api/users/us_1/purge-media-progress",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.PurgeUserMediaProgress(ctx, "us_1")
				return err
			}},

		// podcasts.go
		{name: "CreatePodcast", method: "POST", path: "/api/podcasts",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.CreatePodcast(ctx, &CreatePodcastRequest{})
				return err
			}},
		{name: "PodcastFeed", method: "POST", path: "/api/podcasts/feed",
			call: func(ctx context.Context, c *Client) error { _, err := c.PodcastFeed(ctx, "http://x/f"); return err }},
		{name: "PodcastFeedsFromOPML", method: "POST", path: "/api/podcasts/opml",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.PodcastFeedsFromOPML(ctx, "<opml/>")
				return err
			}},
		{name: "CheckNewPodcastEpisodes", method: "GET", path: "/api/podcasts/li_1/checknew",
			query: map[string]string{"limit": "4"},
			call: func(ctx context.Context, c *Client) error {
				_, err := c.CheckNewPodcastEpisodes(ctx, "li_1", 4)
				return err
			}},
		{name: "PodcastEpisodeDownloads", method: "GET", path: "/api/podcasts/li_1/downloads",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.PodcastEpisodeDownloads(ctx, "li_1")
				return err
			}},
		{name: "ClearPodcastEpisodeDownloadQueue", method: "GET", path: "/api/podcasts/li_1/clear-queue",
			call: func(ctx context.Context, c *Client) error { return c.ClearPodcastEpisodeDownloadQueue(ctx, "li_1") }},
		{name: "SearchPodcastFeedForEpisodes", method: "GET", path: "/api/podcasts/li_1/search-episode",
			query: map[string]string{"title": "ep"},
			call: func(ctx context.Context, c *Client) error {
				_, err := c.SearchPodcastFeedForEpisodes(ctx, "li_1", "ep")
				return err
			}},
		{name: "DownloadPodcastEpisodes", method: "POST", path: "/api/podcasts/li_1/download-episodes",
			call: func(ctx context.Context, c *Client) error { return c.DownloadPodcastEpisodes(ctx, "li_1", nil) }},
		{name: "MatchPodcastEpisodes", method: "POST", path: "/api/podcasts/li_1/match-episodes",
			query: map[string]string{"override": "1"},
			call: func(ctx context.Context, c *Client) error {
				_, err := c.MatchPodcastEpisodes(ctx, "li_1", true)
				return err
			}},
		{name: "PodcastEpisode", method: "GET", path: "/api/podcasts/li_1/episode/ep_1",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.PodcastEpisode(ctx, "li_1", "ep_1")
				return err
			}},
		{name: "UpdatePodcastEpisode", method: "PATCH", path: "/api/podcasts/li_1/episode/ep_1",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.UpdatePodcastEpisode(ctx, "li_1", "ep_1", &PodcastEpisodeUpdate{})
				return err
			}},
		{name: "DeletePodcastEpisode", method: "DELETE", path: "/api/podcasts/li_1/episode/ep_1",
			query: map[string]string{"hard": "1"},
			call: func(ctx context.Context, c *Client) error {
				_, err := c.DeletePodcastEpisode(ctx, "li_1", "ep_1", true)
				return err
			}},

		// authors.go
		{name: "Author", method: "GET", path: "/api/authors/au_1",
			query: map[string]string{"library": "lib_1", "include": "items,series"},
			call: func(ctx context.Context, c *Client) error {
				_, err := c.Author(ctx, "au_1", &AuthorParams{LibraryID: "lib_1", Include: []string{"items", "series"}})
				return err
			}},
		{name: "UpdateAuthor", method: "PATCH", path: "/api/authors/au_1",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.UpdateAuthor(ctx, "au_1", &UpdateAuthorRequest{})
				return err
			}},
		{name: "MatchAuthor", method: "POST", path: "/api/authors/au_1/match",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.MatchAuthor(ctx, "au_1", &MatchAuthorRequest{})
				return err
			}},

		// series.go
		{name: "Series", method: "GET", path: "/api/series/se_1",
			query: map[string]string{"include": "progress,rssfeed"},
			call: func(ctx context.Context, c *Client) error {
				_, err := c.Series(ctx, "se_1", "progress", "rssfeed")
				return err
			}},
		{name: "UpdateSeries", method: "PATCH", path: "/api/series/se_1",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.UpdateSeries(ctx, "se_1", &UpdateSeriesRequest{})
				return err
			}},

		// sessions.go
		{name: "Sessions", method: "GET", path: "/api/sessions",
			query: map[string]string{"user": "us_1"},
			call: func(ctx context.Context, c *Client) error {
				_, err := c.Sessions(ctx, &SessionListParams{User: "us_1"})
				return err
			}},
		{name: "DeleteSession", method: "DELETE", path: "/api/sessions/ps_1",
			call: func(ctx context.Context, c *Client) error { return c.DeleteSession(ctx, "ps_1") }},
		{name: "SyncLocalSession", method: "POST", path: "/api/session/local",
			call: func(ctx context.Context, c *Client) error { return c.SyncLocalSession(ctx, &PlaybackSession{}) }},
		{name: "SyncLocalSessions", method: "POST", path: "/api/session/local-all",
			call: func(ctx context.Context, c *Client) error { _, err := c.SyncLocalSessions(ctx, nil); return err }},
		{name: "OpenSession", method: "GET", path: "/api/session/ps_1",
			call: func(ctx context.Context, c *Client) error { _, err := c.OpenSession(ctx, "ps_1"); return err }},
		{name: "SyncOpenSession", method: "POST", path: "/api/session/ps_1/sync",
			call: func(ctx context.Context, c *Client) error { return c.SyncOpenSession(ctx, "ps_1", &SessionSync{}) }},
		{name: "CloseOpenSession", method: "POST", path: "/api/session/ps_1/close",
			call: func(ctx context.Context, c *Client) error { return c.CloseOpenSession(ctx, "ps_1", nil) }},

		// search.go
		{name: "SearchCovers", method: "GET", path: "/api/search/covers",
			query: map[string]string{"title": "dune", "podcast": "1"},
			call: func(ctx context.Context, c *Client) error {
				_, err := c.SearchCovers(ctx, &CoverSearchParams{Title: "dune", Podcast: true})
				return err
			}},
		{name: "SearchBooks", resp: "[]", method: "GET", path: "/api/search/books",
			query: map[string]string{"title": "dune", "provider": "google"},
			call: func(ctx context.Context, c *Client) error {
				_, err := c.SearchBooks(ctx, &BookSearchParams{Title: "dune", Provider: "google"})
				return err
			}},
		{name: "SearchPodcasts", resp: "[]", method: "GET", path: "/api/search/podcast",
			query: map[string]string{"term": "tech"},
			call:  func(ctx context.Context, c *Client) error { _, err := c.SearchPodcasts(ctx, "tech"); return err }},
		{name: "SearchAuthors", method: "GET", path: "/api/search/authors",
			query: map[string]string{"q": "herbert"},
			call:  func(ctx context.Context, c *Client) error { _, err := c.SearchAuthors(ctx, "herbert"); return err }},
		{name: "SearchChapters", method: "GET", path: "/api/search/chapters",
			query: map[string]string{"asin": "B01", "region": "us"},
			call:  func(ctx context.Context, c *Client) error { _, err := c.SearchChapters(ctx, "B01", "us"); return err }},

		// notifications.go
		{name: "NotificationSettings", method: "GET", path: "/api/notifications",
			call: func(ctx context.Context, c *Client) error { _, err := c.NotificationSettings(ctx); return err }},
		{name: "UpdateNotificationSettings", method: "PATCH", path: "/api/notifications",
			call: func(ctx context.Context, c *Client) error {
				return c.UpdateNotificationSettings(ctx, &UpdateNotificationSettingsRequest{})
			}},
		{name: "NotificationEvents", method: "GET", path: "/api/notificationdata",
			call: func(ctx context.Context, c *Client) error { _, err := c.NotificationEvents(ctx); return err }},
		{name: "FireTestNotificationEvent", method: "GET", path: "/api/notifications/test",
			query: map[string]string{"fail": "1"},
			call:  func(ctx context.Context, c *Client) error { return c.FireTestNotificationEvent(ctx, true) }},
		{name: "CreateNotification", method: "POST", path: "/api/notifications",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.CreateNotification(ctx, &NotificationRequest{})
				return err
			}},
		{name: "DeleteNotification", method: "DELETE", path: "/api/notifications/no_1",
			call: func(ctx context.Context, c *Client) error { _, err := c.DeleteNotification(ctx, "no_1"); return err }},
		{name: "UpdateNotification", method: "PATCH", path: "/api/notifications/no_1",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.UpdateNotification(ctx, &NotificationRequest{ID: "no_1"})
				return err
			}},
		{name: "SendTestNotification", method: "GET", path: "/api/notifications/no_1/test",
			call: func(ctx context.Context, c *Client) error { return c.SendTestNotification(ctx, "no_1") }},

		// tools.go
		{name: "EncodeM4B", method: "POST", path: "/api/tools/item/li_1/encode-m4b",
			query: map[string]string{"bitrate": "128k"},
			call: func(ctx context.Context, c *Client) error {
				return c.EncodeM4B(ctx, "li_1", &EncodeM4BParams{Bitrate: "128k"})
			}},
		{name: "CancelM4BEncode", method: "DELETE", path: "/api/tools/item/li_1/encode-m4b",
			call: func(ctx context.Context, c *Client) error { return c.CancelM4BEncode(ctx, "li_1") }},
		{name: "EmbedMetadata", method: "POST", path: "/api/tools/item/li_1/embed-metadata",
			query: map[string]string{"forceEmbedChapters": "1"},
			call: func(ctx context.Context, c *Client) error {
				return c.EmbedMetadata(ctx, "li_1", &EmbedMetadataParams{ForceEmbedChapters: true})
			}},

		// feeds.go
		{name: "OpenLibraryItemRSSFeed", method: "POST", path: "/api/feeds/item/li_1/open",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.OpenLibraryItemRSSFeed(ctx, "li_1", &OpenRSSFeedRequest{})
				return err
			}},
		{name: "OpenCollectionRSSFeed", method: "POST", path: "/api/feeds/collection/co_1/open",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.OpenCollectionRSSFeed(ctx, "co_1", &OpenRSSFeedRequest{})
				return err
			}},
		{name: "OpenSeriesRSSFeed", method: "POST", path: "/api/feeds/series/se_1/open",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.OpenSeriesRSSFeed(ctx, "se_1", &OpenRSSFeedRequest{})
				return err
			}},
		{name: "CloseRSSFeed", method: "POST", path: "/api/feeds/fe_1/close",
			call: func(ctx context.Context, c *Client) error { return c.CloseRSSFeed(ctx, "fe_1") }},

		// cache.go
		{name: "PurgeCache", method: "POST", path: "/api/cache/purge",
			call: func(ctx context.Context, c *Client) error { return c.PurgeCache(ctx) }},
		{name: "PurgeItemsCache", method: "POST", path: "/api/cache/items/purge",
			call: func(ctx context.Context, c *Client) error { return c.PurgeItemsCache(ctx) }},

		// filesystem.go
		{name: "Filesystem", method: "GET", path: "/api/filesystem",
			call: func(ctx context.Context, c *Client) error { _, err := c.Filesystem(ctx); return err }},

		// misc.go
		{name: "UpdateServerSettings", method: "PATCH", path: "/api/settings",
			call: func(ctx context.Context, c *Client) error { _, err := c.UpdateServerSettings(ctx, nil); return err }},
		{name: "Authorize", method: "POST", path: "/api/authorize",
			call: func(ctx context.Context, c *Client) error { _, err := c.Authorize(ctx); return err }},
		{name: "Tags", method: "GET", path: "/api/tags",
			call: func(ctx context.Context, c *Client) error { _, err := c.Tags(ctx); return err }},
		{name: "RenameTag", method: "POST", path: "/api/tags/rename",
			call: func(ctx context.Context, c *Client) error { _, err := c.RenameTag(ctx, "a", "b"); return err }},
		{name: "Genres", method: "GET", path: "/api/genres",
			call: func(ctx context.Context, c *Client) error { _, err := c.Genres(ctx); return err }},
		{name: "RenameGenre", method: "POST", path: "/api/genres/rename",
			call: func(ctx context.Context, c *Client) error { _, err := c.RenameGenre(ctx, "a", "b"); return err }},
		{name: "DeleteGenre", method: "DELETE", path: "/api/genres/U2NpLUZp",
			call: func(ctx context.Context, c *Client) error { _, err := c.DeleteGenre(ctx, "Sci-Fi"); return err }},
		{name: "ValidateCron", method: "POST", path: "/api/validate-cron",
			call: func(ctx context.Context, c *Client) error { return c.ValidateCron(ctx, "0 0 * * *") }},

		// me.go
		{name: "Me", method: "GET", path: "/api/me",
			call: func(ctx context.Context, c *Client) error { _, err := c.Me(ctx); return err }},
		{name: "MyListeningSessions", method: "GET", path: "/api/me/listening-sessions",
			call: func(ctx context.Context, c *Client) error { _, err := c.MyListeningSessions(ctx, nil); return err }},
		{name: "MyListeningStats", method: "GET", path: "/api/me/listening-stats",
			call: func(ctx context.Context, c *Client) error { _, err := c.MyListeningStats(ctx); return err }},
		{name: "RemoveItemFromContinueListening", method: "GET", path: "/api/me/progress/mp_1/remove-from-continue-listening",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.RemoveItemFromContinueListening(ctx, "mp_1")
				return err
			}},
		{name: "RemoveSeriesFromContinueListening", method: "GET", path: "/api/me/series/se_1/remove-from-continue-listening",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.RemoveSeriesFromContinueListening(ctx, "se_1")
				return err
			}},
		{name: "BatchUpdateMyMediaProgress", method: "PATCH", path: "/api/me/progress/batch/update",
			call: func(ctx context.Context, c *Client) error { return c.BatchUpdateMyMediaProgress(ctx, nil) }},
		{name: "RemoveMyMediaProgress", method: "DELETE", path: "/api/me/progress/mp_1",
			call: func(ctx context.Context, c *Client) error { return c.RemoveMyMediaProgress(ctx, "mp_1") }},
		{name: "CreateBookmark", method: "POST", path: "/api/me/item/li_1/bookmark",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.CreateBookmark(ctx, "li_1", 30, "t")
				return err
			}},
		{name: "UpdateBookmark", method: "PATCH", path: "/api/me/item/li_1/bookmark",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.UpdateBookmark(ctx, "li_1", 30, "t")
				return err
			}},
		{name: "RemoveBookmark", method: "DELETE", path: "/api/me/item/li_1/bookmark/30",
			call: func(ctx context.Context, c *Client) error { return c.RemoveBookmark(ctx, "li_1", 30) }},
		{name: "ChangeMyPassword", resp: `{"success":true}`, method: "PATCH", path: "/api/me/password",
			call: func(ctx context.Context, c *Client) error { return c.ChangeMyPassword(ctx, "old", "new") }},
		{name: "SyncLocalMediaProgress", method: "POST", path: "/api/me/sync-local-progress",
			call: func(ctx context.Context, c *Client) error { _, err := c.SyncLocalMediaProgress(ctx, nil); return err }},
		{name: "MyItemsInProgress", method: "GET", path: "/api/me/items-in-progress",
			query: map[string]string{"limit": "5"},
			call:  func(ctx context.Context, c *Client) error { _, err := c.MyItemsInProgress(ctx, 5); return err }},

		// backups.go
		{name: "Backups", method: "GET", path: "/api/backups",
			call: func(ctx context.Context, c *Client) error { _, err := c.Backups(ctx); return err }},
		{name: "CreateBackup", method: "POST", path: "/api/backups",
			call: func(ctx context.Context, c *Client) error { _, err := c.CreateBackup(ctx); return err }},
		{name: "DeleteBackup", method: "DELETE", path: "/api/backups/bk_1",
			call: func(ctx context.Context, c *Client) error { _, err := c.DeleteBackup(ctx, "bk_1"); return err }},
		{name: "ApplyBackup", method: "GET", path: "/api/backups/bk_1/apply",
			call: func(ctx context.Context, c *Client) error { return c.ApplyBackup(ctx, "bk_1") }},
		{name: "UpdateBackupPath", method: "PATCH", path: "/api/backups/path",
			call: func(ctx context.Context, c *Client) error { return c.UpdateBackupPath(ctx, "/backups") }},
		{name: "UploadBackup", method: "POST", path: "/api/backups/upload",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.UploadBackup(ctx, "b.audiobookshelf", strings.NewReader("data"))
				return err
			}},

		// server.go
		{name: "Logout", method: "POST", path: "/logout",
			call: func(ctx context.Context, c *Client) error { return c.Logout(ctx, "") }},
		{name: "InitServer", method: "POST", path: "/init",
			call: func(ctx context.Context, c *Client) error { return c.InitServer(ctx, "root", "pw") }},
		{name: "Status", method: "GET", path: "/status",
			call: func(ctx context.Context, c *Client) error { _, err := c.Status(ctx); return err }},
		{name: "Ping", method: "GET", path: "/ping",
			call: func(ctx context.Context, c *Client) error { return c.Ping(ctx) }},
		{name: "Healthcheck", method: "GET", path: "/healthcheck",
			call: func(ctx context.Context, c *Client) error { return c.Healthcheck(ctx) }},
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

			for k, want := range tt.query {
				if got := cap.query.Get(k); got != want {
					t.Errorf("query[%s] = %q, want %q (raw %q)", k, got, want, cap.query.Encode())
				}
			}
		})
	}
}
