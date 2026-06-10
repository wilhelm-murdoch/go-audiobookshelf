# go-audiobookshelf

A practical, idiomatic Go client for the [Audiobookshelf](https://www.audiobookshelf.org/) API, covering the full API surface documented at [api.audiobookshelf.org](https://api.audiobookshelf.org/). 

```sh
go get git.nightcity.network/nightcity-network/go-audiobookshelf
```

## Compatibility

This client is versioned on its own [SemVer](https://semver.org/) — the version reflects changes to the Go API, not the server. Each release is verified in CI against a specific Audiobookshelf version (`TestedServerVersion`); other server versions usually work, and a mismatch is the first thing to check when a response fails to decode.

| go-audiobookshelf | Tested against Audiobookshelf |
| ----------------- | ----------------------------- |
| v0.1.x            | 2.35.1                        |

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"log"

	audiobookshelf "git.nightcity.network/nightcity-network/go-audiobookshelf"
)

func main() {
	ctx := context.Background()

	client := audiobookshelf.NewClient("https://abs.example.com")
	if _, err := client.Login(ctx, "root", "password"); err != nil {
		log.Fatal(err)
	}

	libraries, err := client.Libraries(ctx)
	if err != nil {
		log.Fatal(err)
	}
	
	for _, library := range libraries {
		page, err := library.Items(ctx, &audiobookshelf.LibraryItemListParams{
			Limit:    25,
			Sort:     "media.metadata.title",
			Minified: true,
		})
		
		if err != nil {
			log.Fatal(err)
		}
		
		fmt.Printf("%s: %d items\n", library.Name, page.Total)
		for _, item := range page.Results {
			fmt.Printf("  %s\n", item.Media.Metadata.Title)
		}
	}
}
```

## Authentication

Audiobookshelf authenticates with a Bearer token. Either log in with a username and password or pass an existing user token or API key up front:

```go
client := audiobookshelf.NewClient("https://abs.example.com",
	audiobookshelf.WithToken("***"),
)
```

Other options: 
- `WithHTTPClient`
- `WithTimeout`
- `WithUserAgent`
- `WithInsecureSkipVerify`.

## Resource handles

Resources returned by the client carry the client with them, so follow-up operations chain naturally:

```go
library, _ := client.Library(ctx, "lib_c1u6t4p45c35rf0nzd")
results, _ := library.Search(ctx, "goodkind", 5)

item, _ := client.LibraryItem(ctx, "li_8gch9ve09orgn4fdz8", nil)
session, _ := item.Play(ctx, nil)
```

Every operation is also available directly on the client, so a fetched handle is never required when you already have an ID.

## Error handling

Any 4xx/5xx response is returned as an `*audiobookshelf.Error` carrying the method, path, status code, and response body. Helpers cover the common cases:

```go
item, err := client.LibraryItem(ctx, id, nil)
if audiobookshelf.IsNotFound(err) {
	// 404
}
```

## API coverage

All endpoint groups of the official documentation are implemented:

| Group | Methods (selection) |
| --- | --- |
| Server | `Login`, `Logout`, `InitServer`, `Status`, `Ping`, `Healthcheck`, `Authorize` |
| Libraries | `Libraries`, `Library`, `CreateLibrary`, `UpdateLibrary`, `DeleteLibrary`, `LibraryItems`, `LibrarySeries`, `LibraryCollections`, `LibraryPlaylists`, `LibraryPersonalized`, `LibraryFilterData`, `SearchLibrary`, `LibraryStats`, `LibraryAuthors`, `MatchAllLibraryItems`, `ScanLibrary`, `LibraryRecentEpisodes`, `LibraryEpisodeDownloads`, `RemoveLibraryIssues`, `ReorderLibraries` |
| Library items | `LibraryItem`, `DeleteLibraryItem`, `DeleteAllLibraryItems`, `UpdateLibraryItemMedia`, cover get/upload/set-from-URL/update/remove, `MatchLibraryItem`, `PlayLibraryItem`, `PlayPodcastEpisode`, `UpdateLibraryItemTracks`, `ScanLibraryItem`, `LibraryItemToneObject`, `UpdateLibraryItemChapters`, `ToneScanLibraryItem`, batch delete/update/get/quickmatch |
| Users | `CreateUser`, `Users`, `OnlineUsers`, `User`, `UpdateUser`, `DeleteUser`, `UserListeningSessions`, `UserListeningStats`, `PurgeUserMediaProgress` |
| Me | `Me`, `MyListeningSessions`, `MyListeningStats`, `MyMediaProgress`, `UpdateMyMediaProgress`, `BatchUpdateMyMediaProgress`, `RemoveMyMediaProgress`, bookmarks, `ChangeMyPassword`, `SyncLocalMediaProgress`, `MyItemsInProgress`, continue-listening removal |
| Collections | full CRUD plus single/batch book add and remove |
| Playlists | full CRUD plus single/batch item add and remove, `CreatePlaylistFromCollection` |
| Sessions | `Sessions`, `DeleteSession`, `SyncLocalSession(s)`, `OpenSession`, `SyncOpenSession`, `CloseOpenSession` |
| Podcasts | `CreatePodcast`, `PodcastFeed`, `PodcastFeedsFromOPML`, `CheckNewPodcastEpisodes`, `PodcastEpisodeDownloads`, `ClearPodcastEpisodeDownloadQueue`, `SearchPodcastFeedForEpisodes`, `DownloadPodcastEpisodes`, `MatchPodcastEpisodes`, episode get/update/delete |
| Authors / Series | `Author`, `UpdateAuthor`, `MatchAuthor`, `AuthorImage`, `Series`, `UpdateSeries` |
| Backups | `Backups`, `CreateBackup`, `DeleteBackup`, `DownloadBackup`, `DownloadBackupTo`, `ApplyBackup`, `UploadBackup`, `UpdateBackupPath` |
| Notifications | settings get/update, `NotificationEvents`, notification CRUD, test endpoints |
| Search | `SearchCovers`, `SearchBooks`, `SearchPodcasts`, `SearchAuthors`, `SearchChapters` |
| RSS feeds | open for item/collection/series, `CloseRSSFeed` |
| Tools | `EncodeM4B`, `CancelM4BEncode`, `EmbedMetadata` |
| Cache / Filesystem / Misc | `PurgeCache`, `PurgeItemsCache`, `Filesystem`, `UploadFiles`, `UpdateServerSettings`, tags, genres, `ValidateCron` |

For anything not modeled, the generic escape hatches `client.Get`, `Post`, `Patch`, `Put`, and `Delete` take any path and decode into any type.

## Notes on types

The API returns most schemas in base, minified, and expanded variants. This library uses superset structs instead of tripling the type count. Fields absent from a variant are simply zero-valued. Timestamps use the `Millis` type and durations use the `Seconds` type; both are plain JSON numbers on the wire and carry helpers:

```go
item, _ := client.LibraryItem(ctx, id, nil)
added := item.AddedAt.Time()            // time.Time (UTC), zero when unset
length := item.Media.Duration.Duration() // time.Duration

// Building requests works the other way:
cur := audiobookshelf.SecondsFromDuration(90 * time.Second)
_ = client.UpdateMyMediaProgress(ctx, id, "", &audiobookshelf.MediaProgressUpdate{CurrentTime: &cur})
```
