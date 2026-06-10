package audiobookshelf

import (
	"context"
	"net/url"
	"strconv"
)

// EncodeM4BParams are the optional ffmpeg parameters for EncodeM4B.
type EncodeM4BParams struct {
	// Bitrate, e.g. "64k" (the server default).
	Bitrate string
	// Codec, e.g. "aac" (the server default).
	Codec string
	// Channels, e.g. 2 (the server default).
	Channels int
}

func (p *EncodeM4BParams) values() url.Values {
	q := url.Values{}
	if p == nil {
		return q
	}
	if p.Bitrate != "" {
		q.Set("bitrate", p.Bitrate)
	}
	if p.Codec != "" {
		q.Set("codec", p.Codec)
	}
	if p.Channels > 0 {
		q.Set("channels", strconv.Itoa(p.Channels))
	}
	return q
}

// EmbedMetadataParams are the optional parameters for EmbedMetadata.
type EmbedMetadataParams struct {
	// SkipBackup skips backing up the original audio files to
	// /metadata/cache/items.
	SkipBackup bool
	// ForceEmbedChapters embeds chapters in every audio file of
	// multi-track audiobooks.
	ForceEmbedChapters bool
}

func (p *EmbedMetadataParams) values() url.Values {
	q := url.Values{}
	if p == nil {
		return q
	}
	if p.SkipBackup {
		q.Set("backup", "0")
	}
	if p.ForceEmbedChapters {
		q.Set("forceEmbedChapters", "1")
	}
	return q
}

// EncodeM4B starts a task encoding a book's audio files into a single
// M4B audiobook file (POST /api/tools/item/:id/encode-m4b). Requires
// admin.
func (c *Client) EncodeM4B(ctx context.Context, libraryItemID string, params *EncodeM4BParams) error {
	path := apiPath("tools", "item").Seg(libraryItemID).Lit("encode-m4b").Query(params.values()).String()
	return c.Post(ctx, path, nil, nil)
}

// CancelM4BEncode cancels a running M4B encode task
// (DELETE /api/tools/item/:id/encode-m4b). Requires admin.
func (c *Client) CancelM4BEncode(ctx context.Context, libraryItemID string) error {
	return c.Delete(ctx, apiPath("tools", "item").Seg(libraryItemID).Lit("encode-m4b").String(), nil)
}

// EmbedMetadata starts a task embedding metadata into a library item's
// audio files (POST /api/tools/item/:id/embed-metadata). Requires admin.
func (c *Client) EmbedMetadata(ctx context.Context, libraryItemID string, params *EmbedMetadataParams) error {
	path := apiPath("tools", "item").Seg(libraryItemID).Lit("embed-metadata").Query(params.values()).String()
	return c.Post(ctx, path, nil, nil)
}
