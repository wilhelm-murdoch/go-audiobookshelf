package audiobookshelf

import "context"

// PurgeCache purges the server's whole cache directory
// POST /api/cache/purge
func (c *Client) PurgeCache(ctx context.Context) error {
	return c.Post(ctx, apiPath("cache", "purge").String(), nil, nil)
}

// PurgeItemsCache purges the items cache directory
// POST /api/cache/items/purge
func (c *Client) PurgeItemsCache(ctx context.Context) error {
	return c.Post(ctx, apiPath("cache", "items", "purge").String(), nil, nil)
}
