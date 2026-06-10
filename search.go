package audiobookshelf

import (
	"context"
	"encoding/json"
)

// CoverSearchParams are the parameters for SearchCovers.
type CoverSearchParams struct {
	// Title is required.
	Title  string
	Author string
	// Provider is the metadata provider for book covers (server default
	// "google").
	Provider string
	// Podcast searches podcast covers; only Title is used then.
	Podcast bool
}

// BookSearchParams are the parameters for SearchBooks.
type BookSearchParams struct {
	// Title to search for. With the Audible provider this can also be an
	// ASIN.
	Title  string
	Author string
	// Provider is the metadata provider (server default "google").
	Provider string
}

// BookSearchResult is one result of SearchBooks. The populated fields
// depend on the metadata provider; ID is left raw because providers
// return both strings and numbers.
type BookSearchResult struct {
	ID            json.RawMessage  `json:"id,omitempty"`
	Title         string           `json:"title,omitempty"`
	Subtitle      string           `json:"subtitle,omitempty"`
	Author        string           `json:"author,omitempty"`
	Narrator      string           `json:"narrator,omitempty"`
	Publisher     string           `json:"publisher,omitempty"`
	PublishedYear json.RawMessage  `json:"publishedYear,omitempty"`
	Description   string           `json:"description,omitempty"`
	Cover         string           `json:"cover,omitempty"`
	Covers        []string         `json:"covers,omitempty"`
	Genres        []string         `json:"genres,omitempty"`
	Tags          string           `json:"tags,omitempty"`
	Series        []SeriesSequence `json:"series,omitempty"`
	Language      string           `json:"language,omitempty"`
	ISBN          string           `json:"isbn,omitempty"`
	ASIN          string           `json:"asin,omitempty"`
	// Duration in minutes (Audible provider).
	Duration int    `json:"duration,omitempty"`
	Region   string `json:"region,omitempty"`
	Rating   string `json:"rating,omitempty"`
}

// PodcastSearchResult is one result of SearchPodcasts (iTunes).
type PodcastSearchResult struct {
	ID               int64    `json:"id,omitempty"`
	ArtistID         int64    `json:"artistId,omitempty"`
	Title            string   `json:"title,omitempty"`
	ArtistName       string   `json:"artistName,omitempty"`
	Description      string   `json:"description,omitempty"`
	DescriptionPlain string   `json:"descriptionPlain,omitempty"`
	ReleaseDate      string   `json:"releaseDate,omitempty"`
	Genres           []string `json:"genres,omitempty"`
	Cover            string   `json:"cover,omitempty"`
	TrackCount       int      `json:"trackCount,omitempty"`
	FeedURL          string   `json:"feedUrl,omitempty"`
	PageURL          string   `json:"pageUrl,omitempty"`
	Explicit         bool     `json:"explicit,omitempty"`
}

// AuthorSearchResult is the result of SearchAuthors (Audnexus).
type AuthorSearchResult struct {
	ASIN        string `json:"asin,omitempty"`
	Description string `json:"description,omitempty"`
	Image       string `json:"image,omitempty"`
	Name        string `json:"name,omitempty"`
}

// ChapterSearchResult is the result of SearchChapters (Audnexus).
type ChapterSearchResult struct {
	ASIN                 string `json:"asin,omitempty"`
	BrandIntroDurationMs int    `json:"brandIntroDurationMs,omitempty"`
	BrandOutroDurationMs int    `json:"brandOutroDurationMs,omitempty"`
	IsAccurate           bool   `json:"isAccurate,omitempty"`
	RuntimeLengthMs      int64  `json:"runtimeLengthMs,omitempty"`
	RuntimeLengthSec     int64  `json:"runtimeLengthSec,omitempty"`
	Chapters             []struct {
		LengthMs       int64  `json:"lengthMs"`
		StartOffsetMs  int64  `json:"startOffsetMs"`
		StartOffsetSec int64  `json:"startOffsetSec"`
		Title          string `json:"title"`
	} `json:"chapters,omitempty"`
}

// SearchCovers searches metadata providers for cover images
// (GET /api/search/covers) and returns image URLs.
func (c *Client) SearchCovers(ctx context.Context, params *CoverSearchParams) ([]string, error) {
	pb := apiPath("search", "covers").Set("title", params.Title)
	if params.Author != "" {
		pb.Set("author", params.Author)
	}
	if params.Provider != "" {
		pb.Set("provider", params.Provider)
	}
	if params.Podcast {
		pb.Flag("podcast", true)
	}
	var resp struct {
		Results []string `json:"results"`
	}
	if err := c.Get(ctx, pb.String(), &resp); err != nil {
		return nil, err
	}
	return resp.Results, nil
}

// SearchBooks searches a metadata provider for books
// (GET /api/search/books).
func (c *Client) SearchBooks(ctx context.Context, params *BookSearchParams) ([]BookSearchResult, error) {
	pb := apiPath("search", "books")
	if params.Title != "" {
		pb.Set("title", params.Title)
	}
	if params.Author != "" {
		pb.Set("author", params.Author)
	}
	if params.Provider != "" {
		pb.Set("provider", params.Provider)
	}
	var results []BookSearchResult
	if err := c.Get(ctx, pb.String(), &results); err != nil {
		return nil, err
	}
	return results, nil
}

// SearchPodcasts searches iTunes for podcasts
// (GET /api/search/podcast).
func (c *Client) SearchPodcasts(ctx context.Context, term string) ([]PodcastSearchResult, error) {
	var results []PodcastSearchResult
	if err := c.Get(ctx, apiPath("search", "podcast").Set("term", term).String(), &results); err != nil {
		return nil, err
	}
	return results, nil
}

// SearchAuthors searches Audnexus for an author
// (GET /api/search/authors). The name must match exactly to get a
// result.
func (c *Client) SearchAuthors(ctx context.Context, name string) (*AuthorSearchResult, error) {
	var result AuthorSearchResult
	if err := c.Get(ctx, apiPath("search", "authors").Set("q", name).String(), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SearchChapters searches Audnexus for a book's chapters by ASIN
// (GET /api/search/chapters). region is e.g. "us" (the server default).
func (c *Client) SearchChapters(ctx context.Context, asin, region string) (*ChapterSearchResult, error) {
	pb := apiPath("search", "chapters").Set("asin", asin)
	if region != "" {
		pb.Set("region", region)
	}
	var result ChapterSearchResult
	if err := c.Get(ctx, pb.String(), &result); err != nil {
		return nil, err
	}
	return &result, nil
}
