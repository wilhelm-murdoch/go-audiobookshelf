package audiobookshelf

import (
	"context"
	"strconv"
)

// CreatePodcastRequest are the parameters for CreatePodcast.
type CreatePodcastRequest struct {
	// Path of the new podcast folder on the server.
	Path string `json:"path"`
	// FolderID of the library folder to put the podcast in.
	FolderID string `json:"folderId"`
	// LibraryID of the library the podcast will belong to.
	LibraryID string `json:"libraryId"`
	// Media holds the podcast's media: Metadata, CoverPath,
	// AutoDownloadEpisodes, and AutoDownloadSchedule are honored.
	Media *Media `json:"media,omitempty"`
	// EpisodesToDownload are feed episodes to download right away.
	EpisodesToDownload []PodcastFeedEpisode `json:"episodesToDownload,omitempty"`
}

// PodcastEpisodeUpdate are the parameters for UpdatePodcastEpisode. Only
// set fields are sent.
type PodcastEpisodeUpdate struct {
	Index       *int                     `json:"index,omitempty"`
	Season      *string                  `json:"season,omitempty"`
	Episode     *string                  `json:"episode,omitempty"`
	EpisodeType *string                  `json:"episodeType,omitempty"`
	Title       *string                  `json:"title,omitempty"`
	Subtitle    *string                  `json:"subtitle,omitempty"`
	Description *string                  `json:"description,omitempty"`
	Enclosure   *PodcastEpisodeEnclosure `json:"enclosure,omitempty"`
	PubDate     *string                  `json:"pubDate,omitempty"`
	PublishedAt *Millis                  `json:"publishedAt,omitempty"`
}

// PodcastSearchEpisode is one match of SearchPodcastFeedForEpisodes.
type PodcastSearchEpisode struct {
	Episode *PodcastFeedEpisode `json:"episode"`
	// Levenshtein is the edit distance between the search title and the
	// episode's title.
	Levenshtein int `json:"levenshtein"`
}

func podcastPath(id string, rest ...string) string {
	return apiPath("podcasts").Seg(id).Lit(rest...).String()
}

// CreatePodcast creates a podcast library item (POST /api/podcasts).
// Requires upload permission.
func (c *Client) CreatePodcast(ctx context.Context, req *CreatePodcastRequest) (*LibraryItem, error) {
	var item LibraryItem
	if err := c.Post(ctx, apiPath("podcasts").String(), req, &item); err != nil {
		return nil, err
	}
	item.client = c
	return &item, nil
}

// PodcastFeed fetches podcast data from an RSS feed URL
// (POST /api/podcasts/feed).
func (c *Client) PodcastFeed(ctx context.Context, feedURL string) (*PodcastFeed, error) {
	var resp struct {
		Podcast *PodcastFeed `json:"podcast"`
	}
	if err := c.Post(ctx, apiPath("podcasts", "feed").String(), map[string]string{"rssFeed": feedURL}, &resp); err != nil {
		return nil, err
	}
	return resp.Podcast, nil
}

// PodcastFeedsFromOPML fetches podcast feeds for all RSS feeds in OPML
// text (POST /api/podcasts/opml).
func (c *Client) PodcastFeedsFromOPML(ctx context.Context, opmlText string) ([]PodcastFeed, error) {
	var resp struct {
		Feeds []PodcastFeed `json:"feeds"`
	}
	if err := c.Post(ctx, apiPath("podcasts", "opml").String(), map[string]string{"opmlText": opmlText}, &resp); err != nil {
		return nil, err
	}
	return resp.Feeds, nil
}

// CheckNewPodcastEpisodes checks a podcast's feed for new episodes and
// downloads them (GET /api/podcasts/:id/checknew). limit caps the number
// of episodes downloaded (server default 3, 0 keeps the default). It
// returns the episodes that will be downloaded.
func (c *Client) CheckNewPodcastEpisodes(ctx context.Context, id string, limit int) ([]PodcastFeedEpisode, error) {
	pb := apiPath("podcasts").Seg(id).Lit("checknew")
	if limit > 0 {
		pb.Set("limit", strconv.Itoa(limit))
	}
	var resp struct {
		Episodes []PodcastFeedEpisode `json:"episodes"`
	}
	if err := c.Get(ctx, pb.String(), &resp); err != nil {
		return nil, err
	}
	return resp.Episodes, nil
}

// PodcastEpisodeDownloads returns the episode download queue of a
// podcast (GET /api/podcasts/:id/downloads).
func (c *Client) PodcastEpisodeDownloads(ctx context.Context, id string) ([]PodcastEpisodeDownload, error) {
	var resp struct {
		Downloads []PodcastEpisodeDownload `json:"downloads"`
	}
	if err := c.Get(ctx, podcastPath(id, "downloads"), &resp); err != nil {
		return nil, err
	}
	return resp.Downloads, nil
}

// ClearPodcastEpisodeDownloadQueue clears the episode download queue of
// a podcast (GET /api/podcasts/:id/clear-queue).
func (c *Client) ClearPodcastEpisodeDownloadQueue(ctx context.Context, id string) error {
	return c.Get(ctx, podcastPath(id, "clear-queue"), nil)
}

// SearchPodcastFeedForEpisodes searches a podcast's feed for episodes by
// title (GET /api/podcasts/:id/search-episode).
func (c *Client) SearchPodcastFeedForEpisodes(ctx context.Context, id, title string) ([]PodcastSearchEpisode, error) {
	var resp struct {
		Episodes []PodcastSearchEpisode `json:"episodes"`
	}
	path := apiPath("podcasts").Seg(id).Lit("search-episode").Set("title", title).String()
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp.Episodes, nil
}

// DownloadPodcastEpisodes queues feed episodes of a podcast for download
// (POST /api/podcasts/:id/download-episodes).
func (c *Client) DownloadPodcastEpisodes(ctx context.Context, id string, episodes []PodcastFeedEpisode) error {
	return c.Post(ctx, podcastPath(id, "download-episodes"), episodes, nil)
}

// MatchPodcastEpisodes matches a podcast's episodes against its feed and
// updates their details (POST /api/podcasts/:id/match-episodes).
// override overwrites existing details. It returns the number of
// episodes updated.
func (c *Client) MatchPodcastEpisodes(ctx context.Context, id string, override bool) (int, error) {
	var resp struct {
		NumEpisodesUpdated int `json:"numEpisodesUpdated"`
	}
	path := apiPath("podcasts").Seg(id).Lit("match-episodes").Flag("override", override).String()
	if err := c.Post(ctx, path, nil, &resp); err != nil {
		return 0, err
	}
	return resp.NumEpisodesUpdated, nil
}

// PodcastEpisode returns an episode of a podcast
// (GET /api/podcasts/:id/episode/:episodeId).
func (c *Client) PodcastEpisode(ctx context.Context, id, episodeID string) (*PodcastEpisode, error) {
	var episode PodcastEpisode
	path := apiPath("podcasts").Seg(id).Lit("episode").Seg(episodeID).String()
	if err := c.Get(ctx, path, &episode); err != nil {
		return nil, err
	}
	return &episode, nil
}

// UpdatePodcastEpisode updates an episode of a podcast
// (PATCH /api/podcasts/:id/episode/:episodeId) and returns the updated
// library item.
func (c *Client) UpdatePodcastEpisode(ctx context.Context, id, episodeID string, update *PodcastEpisodeUpdate) (*LibraryItem, error) {
	var item LibraryItem
	path := apiPath("podcasts").Seg(id).Lit("episode").Seg(episodeID).String()
	if err := c.Patch(ctx, path, update, &item); err != nil {
		return nil, err
	}
	item.client = c
	return &item, nil
}

// DeletePodcastEpisode deletes an episode of a podcast
// (DELETE /api/podcasts/:id/episode/:episodeId). With hard, the episode's
// audio file is also deleted from the filesystem. It returns the updated
// library item.
func (c *Client) DeletePodcastEpisode(ctx context.Context, id, episodeID string, hard bool) (*LibraryItem, error) {
	var item LibraryItem
	path := apiPath("podcasts").Seg(id).Lit("episode").Seg(episodeID).Flag("hard", hard).String()
	if err := c.Delete(ctx, path, &item); err != nil {
		return nil, err
	}
	item.client = c
	return &item, nil
}
