package audiobookshelf_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	audiobookshelf "github.com/wilhelm-murdoch/go-audiobookshelf"
)

// Log in with a username and password, then list libraries and their
// items. Login stores the returned token on the client for subsequent
// requests.
func Example() {
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
	}
}

// Authenticate up front with an existing user token or API key instead of
// logging in, and adjust the HTTP behaviour with options.
func ExampleNewClient() {
	client := audiobookshelf.NewClient("https://abs.example.com",
		audiobookshelf.WithToken("your-token-or-api-key"),
		audiobookshelf.WithTimeout(30*time.Second),
		audiobookshelf.WithUserAgent("my-app/1.0"),
	)

	_ = client
}

// Resources returned by the client carry the client with them, so
// follow-up calls chain without re-passing IDs.
func ExampleClient_Library() {
	ctx := context.Background()
	client := audiobookshelf.NewClient("https://abs.example.com", audiobookshelf.WithToken("***"))

	library, err := client.Library(ctx, "lib_c1u6t4p45c35rf0nzd")
	if err != nil {
		log.Fatal(err)
	}

	results, err := library.Search(ctx, "goodkind", 5)
	if err != nil {
		log.Fatal(err)
	}

	for _, hit := range results.Book {
		fmt.Println(hit.LibraryItem.Media.Metadata.Title)
	}
}

// List endpoints return a Page[T] envelope. Pages are zero-indexed, and a
// zero Limit lets the server apply its own default.
func ExampleClient_LibraryItems() {
	ctx := context.Background()
	client := audiobookshelf.NewClient("https://abs.example.com", audiobookshelf.WithToken("***"))

	page, err := client.LibraryItems(ctx, "lib_c1u6t4p45c35rf0nzd", &audiobookshelf.LibraryItemListParams{
		Limit: 50,
		Page:  1, // the second page
		Sort:  "media.metadata.title",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("page %d, %d items total\n", page.Page, page.Total)
	for _, item := range page.Results {
		fmt.Println(item.Media.Metadata.Title)
	}
}

// Any 4xx/5xx response is an *audiobookshelf.Error. Use the status
// helpers for common cases or errors.As for the full detail.
func ExampleError() {
	ctx := context.Background()
	client := audiobookshelf.NewClient("https://abs.example.com", audiobookshelf.WithToken("***"))

	_, err := client.LibraryItem(ctx, "li_missing", nil)

	if audiobookshelf.IsNotFound(err) {
		fmt.Println("no such item")
	}

	var apiErr *audiobookshelf.Error
	if errors.As(err, &apiErr) {
		fmt.Printf("%s %s -> %d: %s\n", apiErr.Method, apiErr.Path, apiErr.StatusCode, apiErr.Message)
	}
}

// Timestamps decode as Millis and durations as Seconds, each with helpers
// that convert to the standard library types.
func ExampleMillis() {
	ctx := context.Background()
	client := audiobookshelf.NewClient("https://abs.example.com", audiobookshelf.WithToken("***"))

	item, err := client.LibraryItem(ctx, "li_8gch9ve09orgn4fdz8", nil)
	if err != nil {
		log.Fatal(err)
	}

	added := item.AddedAt.Time() // time.Time in UTC, zero when unset
	fmt.Println("added:", added.Format(time.RFC3339))

	if item.Media != nil {
		length := item.Media.Duration.Duration() // time.Duration
		fmt.Println("length:", length)
	}
}
