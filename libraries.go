package audiobookshelf

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
)

// CreateLibraryRequest are the parameters for CreateLibrary.
type CreateLibraryRequest struct {
	Name    string   `json:"name"`
	Folders []Folder `json:"folders,omitempty"`
	Icon    string   `json:"icon,omitempty"`
	// MediaType is "book" (default) or "podcast".
	MediaType string           `json:"mediaType,omitempty"`
	Provider  string           `json:"provider,omitempty"`
	Settings  *LibrarySettings `json:"settings,omitempty"`
}

// UpdateLibraryRequest are the parameters for UpdateLibrary. Nil/zero
// fields are left unchanged.
type UpdateLibraryRequest struct {
	Name         string           `json:"name,omitempty"`
	Folders      []Folder         `json:"folders,omitempty"`
	DisplayOrder int              `json:"displayOrder,omitempty"`
	Icon         string           `json:"icon,omitempty"`
	Provider     string           `json:"provider,omitempty"`
	Settings     *LibrarySettings `json:"settings,omitempty"`
}

// LibraryItemListParams are the optional query parameters for
// LibraryItems.
type LibraryItemListParams struct {
	// Limit per page; 0 means no limit. Page is 0-indexed.
	Limit int
	Page  int
	// Sort is the attribute to sort by in JavaScript object notation,
	// e.g. "media.metadata.title".
	Sort string
	// Desc reverses the sort order.
	Desc bool
	// Filter filters the results, e.g. "authors.<base64 author id>". See
	// the Filtering section of the API documentation.
	Filter string
	// Minified requests minified library items.
	Minified bool
	// CollapseSeries collapses books of a series into a single entry.
	CollapseSeries bool
	// Include is a comma-separated list of extras; the only current
	// option is "rssfeed".
	Include string
}

func (p *LibraryItemListParams) values() url.Values {
	q := url.Values{}
	if p == nil {
		return q
	}
	if p.Limit > 0 {
		q.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.Page > 0 {
		q.Set("page", strconv.Itoa(p.Page))
	}
	if p.Sort != "" {
		q.Set("sort", p.Sort)
	}
	if p.Desc {
		q.Set("desc", "1")
	}
	if p.Filter != "" {
		q.Set("filter", p.Filter)
	}
	if p.Minified {
		q.Set("minified", "1")
	}
	if p.CollapseSeries {
		q.Set("collapseseries", "1")
	}
	if p.Include != "" {
		q.Set("include", p.Include)
	}
	return q
}

// Shelf is one shelf of a library's personalized view. Entities is left
// raw because its element type depends on Type; use the typed accessors.
type Shelf struct {
	ID             string `json:"id"`
	Label          string `json:"label"`
	LabelStringKey string `json:"labelStringKey,omitempty"`
	// Type is "book", "podcast", "episode", "series", or "authors".
	Type     string          `json:"type"`
	Category string          `json:"category,omitempty"`
	Entities json.RawMessage `json:"entities"`
	Total    int             `json:"total,omitempty"`
}

// LibraryItemEntities decodes the shelf's entities as library items
// (shelf types "book", "podcast", and "episode").
func (s *Shelf) LibraryItemEntities() ([]LibraryItem, error) {
	var items []LibraryItem
	err := json.Unmarshal(s.Entities, &items)
	return items, err
}

// SeriesEntities decodes the shelf's entities as series (shelf type
// "series").
func (s *Shelf) SeriesEntities() ([]Series, error) {
	var series []Series
	err := json.Unmarshal(s.Entities, &series)
	return series, err
}

// AuthorEntities decodes the shelf's entities as authors (shelf type
// "authors").
func (s *Shelf) AuthorEntities() ([]Author, error) {
	var authors []Author
	err := json.Unmarshal(s.Entities, &authors)
	return authors, err
}

// LibrarySearchResults is the response of SearchLibrary.
type LibrarySearchResults struct {
	Book    []LibraryItemSearchResult `json:"book,omitempty"`
	Podcast []LibraryItemSearchResult `json:"podcast,omitempty"`
	Tags    []string                  `json:"tags,omitempty"`
	Authors []Author                  `json:"authors,omitempty"`
	Series  []SeriesSearchResult      `json:"series,omitempty"`
}

// LibraryItemSearchResult is one matched library item of a library
// search.
type LibraryItemSearchResult struct {
	LibraryItem *LibraryItem `json:"libraryItem"`
	MatchKey    string       `json:"matchKey,omitempty"`
	MatchText   string       `json:"matchText,omitempty"`
}

// SeriesSearchResult is one matched series of a library search.
type SeriesSearchResult struct {
	Series *Series       `json:"series"`
	Books  []LibraryItem `json:"books,omitempty"`
}

// LibraryStats are the statistics of a library.
type LibraryStats struct {
	TotalItems       int                `json:"totalItems"`
	TotalAuthors     int                `json:"totalAuthors"`
	TotalGenres      int                `json:"totalGenres"`
	TotalDuration    float64            `json:"totalDuration"`
	NumAudioTracks   int                `json:"numAudioTracks"`
	TotalSize        int64              `json:"totalSize"`
	LongestItems     []LibraryStatsItem `json:"longestItems,omitempty"`
	LargestItems     []LibraryStatsItem `json:"largestItems,omitempty"`
	AuthorsWithCount []struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Count int    `json:"count"`
	} `json:"authorsWithCount,omitempty"`
	GenresWithCount []struct {
		Genre string `json:"genre"`
		Count int    `json:"count"`
	} `json:"genresWithCount,omitempty"`
}

// LibraryStatsItem is a longest/largest item entry of LibraryStats.
type LibraryStatsItem struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Duration float64 `json:"duration,omitempty"`
	Size     int64   `json:"size,omitempty"`
}

// EpisodeDownloadQueue is the podcast episode download state of a
// library.
type EpisodeDownloadQueue struct {
	CurrentDownload *PodcastEpisodeDownload  `json:"currentDownload,omitempty"`
	Queue           []PodcastEpisodeDownload `json:"queue"`
}

// LibraryOrder sets the display position of one library for
// ReorderLibraries.
type LibraryOrder struct {
	ID       string `json:"id"`
	NewOrder int    `json:"newOrder"`
}

func (c *Client) setLibraryClients(libraries []Library) {
	for i := range libraries {
		libraries[i].client = c
	}
}

func (c *Client) setItemClients(items []LibraryItem) {
	for i := range items {
		items[i].client = c
	}
}

// CreateLibrary creates a new library (POST /api/libraries).
func (c *Client) CreateLibrary(ctx context.Context, req *CreateLibraryRequest) (*Library, error) {
	var library Library
	if err := c.Post(ctx, "/api/libraries", req, &library); err != nil {
		return nil, err
	}
	library.client = c
	return &library, nil
}

// Libraries returns all libraries accessible to the user
// (GET /api/libraries).
func (c *Client) Libraries(ctx context.Context) ([]Library, error) {
	var resp struct {
		Libraries []Library `json:"libraries"`
	}
	if err := c.Get(ctx, "/api/libraries", &resp); err != nil {
		return nil, err
	}
	c.setLibraryClients(resp.Libraries)
	return resp.Libraries, nil
}

// Library returns the library with the given ID (GET /api/libraries/:id).
func (c *Client) Library(ctx context.Context, id string) (*Library, error) {
	var library Library
	if err := c.Get(ctx, "/api/libraries/"+url.PathEscape(id), &library); err != nil {
		return nil, err
	}
	library.client = c
	return &library, nil
}

// UpdateLibrary updates a library (PATCH /api/libraries/:id) and returns
// the updated library.
func (c *Client) UpdateLibrary(ctx context.Context, id string, req *UpdateLibraryRequest) (*Library, error) {
	var library Library
	if err := c.Patch(ctx, "/api/libraries/"+url.PathEscape(id), req, &library); err != nil {
		return nil, err
	}
	library.client = c
	return &library, nil
}

// DeleteLibrary deletes a library (DELETE /api/libraries/:id). Library
// folders and files are not deleted.
func (c *Client) DeleteLibrary(ctx context.Context, id string) error {
	return c.Delete(ctx, "/api/libraries/"+url.PathEscape(id), nil)
}

// LibraryItems lists the items of a library
// (GET /api/libraries/:id/items).
func (c *Client) LibraryItems(ctx context.Context, libraryID string, params *LibraryItemListParams) (*Page[LibraryItem], error) {
	var page Page[LibraryItem]
	path := appendQuery("/api/libraries/"+url.PathEscape(libraryID)+"/items", params.values())
	if err := c.Get(ctx, path, &page); err != nil {
		return nil, err
	}
	c.setItemClients(page.Results)
	return &page, nil
}

// RemoveLibraryIssues removes items with issues (missing or invalid) from
// a library (DELETE /api/libraries/:id/issues). No files are deleted.
func (c *Client) RemoveLibraryIssues(ctx context.Context, libraryID string) error {
	return c.Delete(ctx, "/api/libraries/"+url.PathEscape(libraryID)+"/issues", nil)
}

// LibraryEpisodeDownloads returns the podcast episode download queue of a
// library (GET /api/libraries/:id/episode-downloads).
func (c *Client) LibraryEpisodeDownloads(ctx context.Context, libraryID string) (*EpisodeDownloadQueue, error) {
	var queue EpisodeDownloadQueue
	if err := c.Get(ctx, "/api/libraries/"+url.PathEscape(libraryID)+"/episode-downloads", &queue); err != nil {
		return nil, err
	}
	return &queue, nil
}

// LibrarySeries lists the series of a library
// (GET /api/libraries/:id/series).
func (c *Client) LibrarySeries(ctx context.Context, libraryID string, params *LibraryItemListParams) (*Page[Series], error) {
	var page Page[Series]
	path := appendQuery("/api/libraries/"+url.PathEscape(libraryID)+"/series", params.values())
	if err := c.Get(ctx, path, &page); err != nil {
		return nil, err
	}
	for i := range page.Results {
		page.Results[i].client = c
	}
	return &page, nil
}

// LibraryCollections lists the collections of a library
// (GET /api/libraries/:id/collections).
func (c *Client) LibraryCollections(ctx context.Context, libraryID string, params *PageParams) (*Page[Collection], error) {
	var page Page[Collection]
	path := appendQuery("/api/libraries/"+url.PathEscape(libraryID)+"/collections", params.values())
	if err := c.Get(ctx, path, &page); err != nil {
		return nil, err
	}
	for i := range page.Results {
		page.Results[i].client = c
	}
	return &page, nil
}

// LibraryPlaylists lists the user's playlists of a library
// (GET /api/libraries/:id/playlists).
func (c *Client) LibraryPlaylists(ctx context.Context, libraryID string, params *PageParams) (*Page[Playlist], error) {
	var page Page[Playlist]
	path := appendQuery("/api/libraries/"+url.PathEscape(libraryID)+"/playlists", params.values())
	if err := c.Get(ctx, path, &page); err != nil {
		return nil, err
	}
	for i := range page.Results {
		page.Results[i].client = c
	}
	return &page, nil
}

// LibraryPersonalized returns the personalized ("home page") view of a
// library (GET /api/libraries/:id/personalized). limit caps the number of
// entities per shelf (server default 10); include may be "rssfeed" or
// empty.
func (c *Client) LibraryPersonalized(ctx context.Context, libraryID string, limit int, include string) ([]Shelf, error) {
	q := url.Values{}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if include != "" {
		q.Set("include", include)
	}
	var shelves []Shelf
	path := appendQuery("/api/libraries/"+url.PathEscape(libraryID)+"/personalized", q)
	if err := c.Get(ctx, path, &shelves); err != nil {
		return nil, err
	}
	return shelves, nil
}

// LibraryFilterData returns the filter data of a library
// (GET /api/libraries/:id/filterdata).
func (c *Client) LibraryFilterData(ctx context.Context, libraryID string) (*LibraryFilterData, error) {
	var data LibraryFilterData
	if err := c.Get(ctx, "/api/libraries/"+url.PathEscape(libraryID)+"/filterdata", &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// SearchLibrary searches a library (GET /api/libraries/:id/search). limit
// caps the results per category (server default 12).
func (c *Client) SearchLibrary(ctx context.Context, libraryID, query string, limit int) (*LibrarySearchResults, error) {
	q := url.Values{"q": []string{query}}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	var results LibrarySearchResults
	path := appendQuery("/api/libraries/"+url.PathEscape(libraryID)+"/search", q)
	if err := c.Get(ctx, path, &results); err != nil {
		return nil, err
	}
	for i := range results.Book {
		if results.Book[i].LibraryItem != nil {
			results.Book[i].LibraryItem.client = c
		}
	}
	for i := range results.Podcast {
		if results.Podcast[i].LibraryItem != nil {
			results.Podcast[i].LibraryItem.client = c
		}
	}
	return &results, nil
}

// LibraryStats returns the statistics of a library
// (GET /api/libraries/:id/stats).
func (c *Client) LibraryStats(ctx context.Context, libraryID string) (*LibraryStats, error) {
	var stats LibraryStats
	if err := c.Get(ctx, "/api/libraries/"+url.PathEscape(libraryID)+"/stats", &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

// LibraryAuthors returns the authors of a library
// (GET /api/libraries/:id/authors).
func (c *Client) LibraryAuthors(ctx context.Context, libraryID string) ([]Author, error) {
	var resp struct {
		Authors []Author `json:"authors"`
	}
	if err := c.Get(ctx, "/api/libraries/"+url.PathEscape(libraryID)+"/authors", &resp); err != nil {
		return nil, err
	}
	for i := range resp.Authors {
		resp.Authors[i].client = c
	}
	return resp.Authors, nil
}

// MatchAllLibraryItems starts a quick match of all items of a library
// (GET /api/libraries/:id/matchall). Requires admin.
func (c *Client) MatchAllLibraryItems(ctx context.Context, libraryID string) error {
	return c.Get(ctx, "/api/libraries/"+url.PathEscape(libraryID)+"/matchall", nil)
}

// ScanLibrary starts a scan of a library's folders
// (POST /api/libraries/:id/scan). force rescans all items even when
// unchanged. Requires admin.
func (c *Client) ScanLibrary(ctx context.Context, libraryID string, force bool) error {
	q := url.Values{}
	if force {
		q.Set("force", "1")
	}
	return c.Post(ctx, appendQuery("/api/libraries/"+url.PathEscape(libraryID)+"/scan", q), nil, nil)
}

// LibraryRecentEpisodes lists recent podcast episodes of a library
// (GET /api/libraries/:id/recent-episodes).
func (c *Client) LibraryRecentEpisodes(ctx context.Context, libraryID string, params *PageParams) (*Page[PodcastEpisode], error) {
	var page struct {
		Episodes []PodcastEpisode `json:"episodes"`
		Total    int              `json:"total"`
		Limit    int              `json:"limit"`
		Page     int              `json:"page"`
	}
	path := appendQuery("/api/libraries/"+url.PathEscape(libraryID)+"/recent-episodes", params.values())
	if err := c.Get(ctx, path, &page); err != nil {
		return nil, err
	}
	return &Page[PodcastEpisode]{
		Results: page.Episodes,
		Total:   page.Total,
		Limit:   page.Limit,
		Page:    page.Page,
	}, nil
}

// ReorderLibraries changes the display order of libraries
// (POST /api/libraries/order) and returns all libraries in their new
// order. Requires admin.
func (c *Client) ReorderLibraries(ctx context.Context, order []LibraryOrder) ([]Library, error) {
	var resp struct {
		Libraries []Library `json:"libraries"`
	}
	if err := c.Post(ctx, "/api/libraries/order", order, &resp); err != nil {
		return nil, err
	}
	c.setLibraryClients(resp.Libraries)
	return resp.Libraries, nil
}

// Items lists the items of the library. See Client.LibraryItems.
func (l *Library) Items(ctx context.Context, params *LibraryItemListParams) (*Page[LibraryItem], error) {
	return l.client.LibraryItems(ctx, l.ID, params)
}

// Update updates the library and refreshes its fields in place. See
// Client.UpdateLibrary.
func (l *Library) Update(ctx context.Context, req *UpdateLibraryRequest) error {
	updated, err := l.client.UpdateLibrary(ctx, l.ID, req)
	if err != nil {
		return err
	}
	*l = *updated
	return nil
}

// Delete deletes the library. See Client.DeleteLibrary.
func (l *Library) Delete(ctx context.Context) error {
	return l.client.DeleteLibrary(ctx, l.ID)
}

// Series lists the series of the library. See Client.LibrarySeries.
func (l *Library) Series(ctx context.Context, params *LibraryItemListParams) (*Page[Series], error) {
	return l.client.LibrarySeries(ctx, l.ID, params)
}

// Collections lists the collections of the library. See
// Client.LibraryCollections.
func (l *Library) Collections(ctx context.Context, params *PageParams) (*Page[Collection], error) {
	return l.client.LibraryCollections(ctx, l.ID, params)
}

// Playlists lists the user's playlists of the library. See
// Client.LibraryPlaylists.
func (l *Library) Playlists(ctx context.Context, params *PageParams) (*Page[Playlist], error) {
	return l.client.LibraryPlaylists(ctx, l.ID, params)
}

// Personalized returns the personalized view of the library. See
// Client.LibraryPersonalized.
func (l *Library) Personalized(ctx context.Context, limit int, include string) ([]Shelf, error) {
	return l.client.LibraryPersonalized(ctx, l.ID, limit, include)
}

// FilterData returns the filter data of the library. See
// Client.LibraryFilterData.
func (l *Library) FilterData(ctx context.Context) (*LibraryFilterData, error) {
	return l.client.LibraryFilterData(ctx, l.ID)
}

// Search searches the library. See Client.SearchLibrary.
func (l *Library) Search(ctx context.Context, query string, limit int) (*LibrarySearchResults, error) {
	return l.client.SearchLibrary(ctx, l.ID, query, limit)
}

// Stats returns the statistics of the library. See Client.LibraryStats.
func (l *Library) Stats(ctx context.Context) (*LibraryStats, error) {
	return l.client.LibraryStats(ctx, l.ID)
}

// Authors returns the authors of the library. See Client.LibraryAuthors.
func (l *Library) Authors(ctx context.Context) ([]Author, error) {
	return l.client.LibraryAuthors(ctx, l.ID)
}

// MatchAll quick-matches all items of the library. See
// Client.MatchAllLibraryItems.
func (l *Library) MatchAll(ctx context.Context) error {
	return l.client.MatchAllLibraryItems(ctx, l.ID)
}

// Scan starts a folder scan of the library. See Client.ScanLibrary.
func (l *Library) Scan(ctx context.Context, force bool) error {
	return l.client.ScanLibrary(ctx, l.ID, force)
}

// RecentEpisodes lists recent podcast episodes of the library. See
// Client.LibraryRecentEpisodes.
func (l *Library) RecentEpisodes(ctx context.Context, params *PageParams) (*Page[PodcastEpisode], error) {
	return l.client.LibraryRecentEpisodes(ctx, l.ID, params)
}

// RemoveIssues removes items with issues from the library. See
// Client.RemoveLibraryIssues.
func (l *Library) RemoveIssues(ctx context.Context) error {
	return l.client.RemoveLibraryIssues(ctx, l.ID)
}

// EpisodeDownloads returns the episode download queue of the library. See
// Client.LibraryEpisodeDownloads.
func (l *Library) EpisodeDownloads(ctx context.Context) (*EpisodeDownloadQueue, error) {
	return l.client.LibraryEpisodeDownloads(ctx, l.ID)
}
