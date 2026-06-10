package audiobookshelf

import (
	"context"
)

// OpenRSSFeedRequest are the parameters for the open-RSS-feed endpoints.
type OpenRSSFeedRequest struct {
	ServerAddress string `json:"serverAddress"`
	Slug          string `json:"slug"`
}

func (c *Client) openFeed(ctx context.Context, path string, req *OpenRSSFeedRequest) (*RSSFeed, error) {
	var resp struct {
		Feed *RSSFeed `json:"feed"`
	}

	if err := c.Post(ctx, path, req, &resp); err != nil {
		return nil, err
	}

	return resp.Feed, nil
}

// OpenLibraryItemRSSFeed opens an RSS feed for a library item
// (POST /api/feeds/item/:itemId/open). Requires admin.
func (c *Client) OpenLibraryItemRSSFeed(ctx context.Context, libraryItemID string, req *OpenRSSFeedRequest) (*RSSFeed, error) {
	return c.openFeed(ctx, apiPath("feeds", "item").Seg(libraryItemID).Lit("open").String(), req)
}

// OpenCollectionRSSFeed opens an RSS feed for a collection
// (POST /api/feeds/collection/:collectionId/open). Requires admin.
func (c *Client) OpenCollectionRSSFeed(ctx context.Context, collectionID string, req *OpenRSSFeedRequest) (*RSSFeed, error) {
	return c.openFeed(ctx, apiPath("feeds", "collection").Seg(collectionID).Lit("open").String(), req)
}

// OpenSeriesRSSFeed opens an RSS feed for a series
// (POST /api/feeds/series/:seriesId/open). Requires admin.
func (c *Client) OpenSeriesRSSFeed(ctx context.Context, seriesID string, req *OpenRSSFeedRequest) (*RSSFeed, error) {
	return c.openFeed(ctx, apiPath("feeds", "series").Seg(seriesID).Lit("open").String(), req)
}

// CloseRSSFeed closes an open RSS feed (POST /api/feeds/:id/close).
// Requires admin.
func (c *Client) CloseRSSFeed(ctx context.Context, feedID string) error {
	return c.Post(ctx, apiPath("feeds").Seg(feedID).Lit("close").String(), nil, nil)
}

// OpenRSSFeed opens an RSS feed for the library item. See
// Client.OpenLibraryItemRSSFeed.
func (i *LibraryItem) OpenRSSFeed(ctx context.Context, req *OpenRSSFeedRequest) (*RSSFeed, error) {
	return i.client.OpenLibraryItemRSSFeed(ctx, i.ID, req)
}

// OpenRSSFeed opens an RSS feed for the collection. See
// Client.OpenCollectionRSSFeed.
func (col *Collection) OpenRSSFeed(ctx context.Context, req *OpenRSSFeedRequest) (*RSSFeed, error) {
	return col.client.OpenCollectionRSSFeed(ctx, col.ID, req)
}

// OpenRSSFeed opens an RSS feed for the series. See
// Client.OpenSeriesRSSFeed.
func (s *Series) OpenRSSFeed(ctx context.Context, req *OpenRSSFeedRequest) (*RSSFeed, error) {
	return s.client.OpenSeriesRSSFeed(ctx, s.ID, req)
}
