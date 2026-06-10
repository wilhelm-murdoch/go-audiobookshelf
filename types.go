package audiobookshelf

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
)

// The Audiobookshelf API returns most schemas in several variants (base,
// minified, expanded). Rather than tripling the type count, this library
// uses superset structs: fields only present in some variants are tagged
// omitempty and are zero-valued when not returned. Timestamp fields use
// the Millis type (milliseconds since the Unix epoch) and duration fields
// use the Seconds type; both are numeric and expose Time/Duration helpers
// (see timestamps.go).

// Library is a content library on the server.
type Library struct {
	client *Client

	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Folders      []Folder `json:"folders,omitempty"`
	DisplayOrder int      `json:"displayOrder,omitempty"`
	Icon         string   `json:"icon,omitempty"`
	// MediaType is "book" or "podcast".
	MediaType  string           `json:"mediaType,omitempty"`
	Provider   string           `json:"provider,omitempty"`
	Settings   *LibrarySettings `json:"settings,omitempty"`
	CreatedAt  Millis           `json:"createdAt,omitempty"`
	LastUpdate Millis           `json:"lastUpdate,omitempty"`
}

// Folder is a folder of a library on the server. Only FullPath should be
// set when creating one.
type Folder struct {
	ID        string `json:"id,omitempty"`
	FullPath  string `json:"fullPath"`
	LibraryID string `json:"libraryId,omitempty"`
	AddedAt   Millis `json:"addedAt,omitempty"`
}

// LibrarySettings are the settings of a library.
type LibrarySettings struct {
	// CoverAspectRatio is 1 for square covers, 0 for standard.
	CoverAspectRatio          int     `json:"coverAspectRatio"`
	DisableWatcher            bool    `json:"disableWatcher"`
	SkipMatchingMediaWithASIN bool    `json:"skipMatchingMediaWithAsin"`
	SkipMatchingMediaWithISBN bool    `json:"skipMatchingMediaWithIsbn"`
	AutoScanCronExpression    *string `json:"autoScanCronExpression"`
}

// LibraryFilterData is the filter data of a library, used to populate
// filter pickers.
type LibraryFilterData struct {
	Authors   []Author `json:"authors"`
	Genres    []string `json:"genres"`
	Tags      []string `json:"tags"`
	Series    []Series `json:"series"`
	Narrators []string `json:"narrators"`
	Languages []string `json:"languages"`
}

// LibraryItem is an item (book or podcast) in a library.
type LibraryItem struct {
	client *Client

	ID          string `json:"id"`
	Ino         string `json:"ino,omitempty"`
	LibraryID   string `json:"libraryId,omitempty"`
	FolderID    string `json:"folderId,omitempty"`
	Path        string `json:"path,omitempty"`
	RelPath     string `json:"relPath,omitempty"`
	IsFile      bool   `json:"isFile,omitempty"`
	MtimeMs     Millis `json:"mtimeMs,omitempty"`
	CtimeMs     Millis `json:"ctimeMs,omitempty"`
	BirthtimeMs Millis `json:"birthtimeMs,omitempty"`
	AddedAt     Millis `json:"addedAt,omitempty"`
	UpdatedAt   Millis `json:"updatedAt,omitempty"`
	LastScan    Millis `json:"lastScan,omitempty"`
	ScanVersion string `json:"scanVersion,omitempty"`
	IsMissing   bool   `json:"isMissing,omitempty"`
	IsInvalid   bool   `json:"isInvalid,omitempty"`
	// MediaType is "book" or "podcast".
	MediaType    string        `json:"mediaType,omitempty"`
	Media        *Media        `json:"media,omitempty"`
	LibraryFiles []LibraryFile `json:"libraryFiles,omitempty"`

	// Minified/expanded variants only.
	NumFiles int   `json:"numFiles,omitempty"`
	Size     int64 `json:"size,omitempty"`

	// Optional includes (see LibraryItemParams and list endpoints).
	UserMediaProgress   *MediaProgress           `json:"userMediaProgress,omitempty"`
	RSSFeed             *RSSFeed                 `json:"rssFeed,omitempty"`
	EpisodesDownloading []PodcastEpisodeDownload `json:"episodesDownloading,omitempty"`
	CollapsedSeries     *Series                  `json:"collapsedSeries,omitempty"`

	// Set on results of Client.MyItemsInProgress.
	RecentEpisode      *PodcastEpisode `json:"recentEpisode,omitempty"`
	ProgressLastUpdate Millis          `json:"progressLastUpdate,omitempty"`
}

// Media is the media of a library item. It is a superset of the Book and
// Podcast schemas; which fields are populated depends on the library
// item's MediaType and on whether the server returned the base, minified,
// or expanded variant.
type Media struct {
	ID            string         `json:"id,omitempty"`
	LibraryItemID string         `json:"libraryItemId,omitempty"`
	Metadata      *MediaMetadata `json:"metadata,omitempty"`
	CoverPath     string         `json:"coverPath,omitempty"`
	Tags          []string       `json:"tags,omitempty"`

	// Book fields.
	AudioFiles []AudioFile  `json:"audioFiles,omitempty"`
	Chapters   []Chapter    `json:"chapters,omitempty"`
	Tracks     []AudioTrack `json:"tracks,omitempty"`
	EbookFile  *EBookFile   `json:"ebookFile,omitempty"`
	// Book minified fields.
	NumTracks     int    `json:"numTracks,omitempty"`
	NumAudioFiles int    `json:"numAudioFiles,omitempty"`
	NumChapters   int    `json:"numChapters,omitempty"`
	EbookFormat   string `json:"ebookFormat,omitempty"`

	// Podcast fields.
	Episodes                 []PodcastEpisode `json:"episodes,omitempty"`
	NumEpisodes              int              `json:"numEpisodes,omitempty"`
	AutoDownloadEpisodes     bool             `json:"autoDownloadEpisodes,omitempty"`
	AutoDownloadSchedule     string           `json:"autoDownloadSchedule,omitempty"`
	LastEpisodeCheck         Millis           `json:"lastEpisodeCheck,omitempty"`
	MaxEpisodesToKeep        int              `json:"maxEpisodesToKeep,omitempty"`
	MaxNewEpisodesToDownload int              `json:"maxNewEpisodesToDownload,omitempty"`

	// Minified/expanded variants only.
	Duration Seconds `json:"duration,omitempty"`
	Size     int64   `json:"size,omitempty"`
}

// MediaMetadata is the metadata of a book or podcast. It is a superset of
// the Book Metadata and Podcast Metadata schemas.
type MediaMetadata struct {
	Title             string   `json:"title,omitempty"`
	TitleIgnorePrefix string   `json:"titleIgnorePrefix,omitempty"`
	Subtitle          string   `json:"subtitle,omitempty"`
	Description       string   `json:"description,omitempty"`
	Genres            []string `json:"genres,omitempty"`
	Language          string   `json:"language,omitempty"`
	Explicit          bool     `json:"explicit,omitempty"`

	// Book fields.
	Authors       []Author        `json:"authors,omitempty"`
	Narrators     []string        `json:"narrators,omitempty"`
	Series        SeriesSequences `json:"series,omitempty"`
	PublishedYear string          `json:"publishedYear,omitempty"`
	PublishedDate string          `json:"publishedDate,omitempty"`
	Publisher     string          `json:"publisher,omitempty"`
	ISBN          string          `json:"isbn,omitempty"`
	ASIN          string          `json:"asin,omitempty"`
	// Book minified/expanded fields.
	AuthorName   string `json:"authorName,omitempty"`
	AuthorNameLF string `json:"authorNameLF,omitempty"`
	NarratorName string `json:"narratorName,omitempty"`
	SeriesName   string `json:"seriesName,omitempty"`

	// Podcast fields.
	Author         string `json:"author,omitempty"`
	ReleaseDate    string `json:"releaseDate,omitempty"`
	FeedURL        string `json:"feedUrl,omitempty"`
	ImageURL       string `json:"imageUrl,omitempty"`
	ITunesPageURL  string `json:"itunesPageUrl,omitempty"`
	ITunesID       int64  `json:"itunesId,omitempty"`
	ITunesArtistID int64  `json:"itunesArtistId,omitempty"`
	// Type is "episodic" or "serial" for podcasts.
	Type string `json:"type,omitempty"`
}

// UnmarshalJSON tolerates Audiobookshelf's loose typing of a few metadata
// fields: itunesId and itunesArtistId arrive as either a JSON number or a
// quoted string, and publishedYear as either a string or a number.
func (m *MediaMetadata) UnmarshalJSON(data []byte) error {
	type alias MediaMetadata
	aux := struct {
		ITunesID       json.RawMessage `json:"itunesId"`
		ITunesArtistID json.RawMessage `json:"itunesArtistId"`
		PublishedYear  json.RawMessage `json:"publishedYear"`
		*alias
	}{alias: (*alias)(m)}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	m.ITunesID = flexInt64(aux.ITunesID)
	m.ITunesArtistID = flexInt64(aux.ITunesArtistID)
	m.PublishedYear = flexString(aux.PublishedYear)

	return nil
}

// flexInt64 decodes a JSON number or quoted number into an int64. Any
// other value (null, empty, or non-numeric) yields 0.
func flexInt64(raw json.RawMessage) int64 {
	if len(raw) == 0 {
		return 0
	}

	var n int64
	if err := json.Unmarshal(raw, &n); err == nil {
		return n
	}

	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		if v, err := strconv.ParseInt(s, 10, 64); err == nil {
			return v
		}
	}

	return 0
}

// flexString decodes a JSON string or number into a string. A null or
// empty value yields "".
func flexString(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}

	return strings.TrimSpace(string(raw))
}

// SeriesSequence is a series a book belongs to, with the book's position
// in the series.
type SeriesSequence struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Sequence string `json:"sequence,omitempty"`
}

// SeriesSequences unmarshals from either a JSON array of series (the
// usual shape) or a single series object (returned when filtering by
// series or including author items).
type SeriesSequences []SeriesSequence

// UnmarshalJSON implements json.Unmarshaler.
func (s *SeriesSequences) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) > 0 && trimmed[0] == '{' {
		var single SeriesSequence
		if err := json.Unmarshal(trimmed, &single); err != nil {
			return err
		}
		*s = SeriesSequences{single}
		return nil
	}
	var many []SeriesSequence
	if err := json.Unmarshal(trimmed, &many); err != nil {
		return err
	}
	*s = many
	return nil
}

// Chapter is a chapter of a book.
type Chapter struct {
	ID    int     `json:"id"`
	Start Seconds `json:"start"`
	End   Seconds `json:"end"`
	Title string  `json:"title"`
}

// AudioFile is an audio file of a book or podcast episode.
type AudioFile struct {
	Index                int            `json:"index,omitempty"`
	Ino                  string         `json:"ino,omitempty"`
	Metadata             *FileMetadata  `json:"metadata,omitempty"`
	AddedAt              Millis         `json:"addedAt,omitempty"`
	UpdatedAt            Millis         `json:"updatedAt,omitempty"`
	TrackNumFromMeta     *int           `json:"trackNumFromMeta,omitempty"`
	DiscNumFromMeta      *int           `json:"discNumFromMeta,omitempty"`
	TrackNumFromFilename *int           `json:"trackNumFromFilename,omitempty"`
	DiscNumFromFilename  *int           `json:"discNumFromFilename,omitempty"`
	ManuallyVerified     bool           `json:"manuallyVerified,omitempty"`
	Exclude              bool           `json:"exclude,omitempty"`
	Error                string         `json:"error,omitempty"`
	Format               string         `json:"format,omitempty"`
	Duration             Seconds        `json:"duration,omitempty"`
	BitRate              int            `json:"bitRate,omitempty"`
	Language             string         `json:"language,omitempty"`
	Codec                string         `json:"codec,omitempty"`
	TimeBase             string         `json:"timeBase,omitempty"`
	Channels             int            `json:"channels,omitempty"`
	ChannelLayout        string         `json:"channelLayout,omitempty"`
	Chapters             []Chapter      `json:"chapters,omitempty"`
	EmbeddedCoverArt     string         `json:"embeddedCoverArt,omitempty"`
	MetaTags             *AudioMetaTags `json:"metaTags,omitempty"`
	MimeType             string         `json:"mimeType,omitempty"`
}

// AudioMetaTags are ID3 metadata tags pulled from an audio file. Only
// non-null tags are returned by the server.
type AudioMetaTags struct {
	TagAlbum                    string `json:"tagAlbum,omitempty"`
	TagArtist                   string `json:"tagArtist,omitempty"`
	TagGenre                    string `json:"tagGenre,omitempty"`
	TagTitle                    string `json:"tagTitle,omitempty"`
	TagSeries                   string `json:"tagSeries,omitempty"`
	TagSeriesPart               string `json:"tagSeriesPart,omitempty"`
	TagTrack                    string `json:"tagTrack,omitempty"`
	TagDisc                     string `json:"tagDisc,omitempty"`
	TagSubtitle                 string `json:"tagSubtitle,omitempty"`
	TagAlbumArtist              string `json:"tagAlbumArtist,omitempty"`
	TagDate                     string `json:"tagDate,omitempty"`
	TagComposer                 string `json:"tagComposer,omitempty"`
	TagPublisher                string `json:"tagPublisher,omitempty"`
	TagComment                  string `json:"tagComment,omitempty"`
	TagDescription              string `json:"tagDescription,omitempty"`
	TagEncoder                  string `json:"tagEncoder,omitempty"`
	TagEncodedBy                string `json:"tagEncodedBy,omitempty"`
	TagISBN                     string `json:"tagIsbn,omitempty"`
	TagLanguage                 string `json:"tagLanguage,omitempty"`
	TagASIN                     string `json:"tagASIN,omitempty"`
	TagOverdriveMediaMarker     string `json:"tagOverdriveMediaMarker,omitempty"`
	TagOriginalYear             string `json:"tagOriginalYear,omitempty"`
	TagReleaseCountry           string `json:"tagReleaseCountry,omitempty"`
	TagReleaseType              string `json:"tagReleaseType,omitempty"`
	TagReleaseStatus            string `json:"tagReleaseStatus,omitempty"`
	TagISRC                     string `json:"tagISRC,omitempty"`
	TagMusicBrainzTrackID       string `json:"tagMusicBrainzTrackId,omitempty"`
	TagMusicBrainzAlbumID       string `json:"tagMusicBrainzAlbumId,omitempty"`
	TagMusicBrainzAlbumArtistID string `json:"tagMusicBrainzAlbumArtistId,omitempty"`
	TagMusicBrainzArtistID      string `json:"tagMusicBrainzArtistId,omitempty"`
}

// AudioTrack is a playable audio track derived from an audio file.
type AudioTrack struct {
	Index       int           `json:"index,omitempty"`
	StartOffset Seconds       `json:"startOffset,omitempty"`
	Duration    Seconds       `json:"duration,omitempty"`
	Title       string        `json:"title,omitempty"`
	ContentURL  string        `json:"contentUrl,omitempty"`
	MimeType    string        `json:"mimeType,omitempty"`
	Metadata    *FileMetadata `json:"metadata,omitempty"`
}

// EBookFile is the ebook file of a book.
type EBookFile struct {
	Ino         string        `json:"ino,omitempty"`
	Metadata    *FileMetadata `json:"metadata,omitempty"`
	EbookFormat string        `json:"ebookFormat,omitempty"`
	AddedAt     Millis        `json:"addedAt,omitempty"`
	UpdatedAt   Millis        `json:"updatedAt,omitempty"`
}

// LibraryFile is a file belonging to a library item.
type LibraryFile struct {
	Ino       string        `json:"ino,omitempty"`
	Metadata  *FileMetadata `json:"metadata,omitempty"`
	AddedAt   Millis        `json:"addedAt,omitempty"`
	UpdatedAt Millis        `json:"updatedAt,omitempty"`
	FileType  string        `json:"fileType,omitempty"`
}

// FileMetadata is the filesystem metadata of a file.
type FileMetadata struct {
	Filename    string `json:"filename,omitempty"`
	Ext         string `json:"ext,omitempty"`
	Path        string `json:"path,omitempty"`
	RelPath     string `json:"relPath,omitempty"`
	Size        int64  `json:"size,omitempty"`
	MtimeMs     Millis `json:"mtimeMs,omitempty"`
	CtimeMs     Millis `json:"ctimeMs,omitempty"`
	BirthtimeMs Millis `json:"birthtimeMs,omitempty"`
}

// PodcastEpisode is a downloaded episode of a podcast.
type PodcastEpisode struct {
	LibraryItemID string                   `json:"libraryItemId,omitempty"`
	ID            string                   `json:"id,omitempty"`
	Index         int                      `json:"index,omitempty"`
	Season        string                   `json:"season,omitempty"`
	Episode       string                   `json:"episode,omitempty"`
	EpisodeType   string                   `json:"episodeType,omitempty"`
	Title         string                   `json:"title,omitempty"`
	Subtitle      string                   `json:"subtitle,omitempty"`
	Description   string                   `json:"description,omitempty"`
	Enclosure     *PodcastEpisodeEnclosure `json:"enclosure,omitempty"`
	PubDate       string                   `json:"pubDate,omitempty"`
	AudioFile     *AudioFile               `json:"audioFile,omitempty"`
	AudioTrack    *AudioTrack              `json:"audioTrack,omitempty"`
	PublishedAt   Millis                   `json:"publishedAt,omitempty"`
	AddedAt       Millis                   `json:"addedAt,omitempty"`
	UpdatedAt     Millis                   `json:"updatedAt,omitempty"`
	Duration      Seconds                  `json:"duration,omitempty"`
	Size          int64                    `json:"size,omitempty"`
}

// PodcastEpisodeEnclosure is the download information of a podcast
// episode.
type PodcastEpisodeEnclosure struct {
	URL  string `json:"url,omitempty"`
	Type string `json:"type,omitempty"`
	// Length is the reported size in bytes. The API returns it as a
	// string.
	Length string `json:"length,omitempty"`
}

// PodcastEpisodeDownload is a queued or completed podcast episode
// download.
type PodcastEpisodeDownload struct {
	ID                  string `json:"id,omitempty"`
	EpisodeDisplayTitle string `json:"episodeDisplayTitle,omitempty"`
	URL                 string `json:"url,omitempty"`
	LibraryItemID       string `json:"libraryItemId,omitempty"`
	LibraryID           string `json:"libraryId,omitempty"`
	IsFinished          bool   `json:"isFinished,omitempty"`
	Failed              bool   `json:"failed,omitempty"`
	StartedAt           Millis `json:"startedAt,omitempty"`
	CreatedAt           Millis `json:"createdAt,omitempty"`
	FinishedAt          Millis `json:"finishedAt,omitempty"`
	PodcastTitle        string `json:"podcastTitle,omitempty"`
	PodcastExplicit     bool   `json:"podcastExplicit,omitempty"`
	Season              string `json:"season,omitempty"`
	Episode             string `json:"episode,omitempty"`
	EpisodeType         string `json:"episodeType,omitempty"`
	PublishedAt         Millis `json:"publishedAt,omitempty"`
}

// PodcastFeed is podcast data fetched from an RSS feed.
type PodcastFeed struct {
	Metadata    *PodcastFeedMetadata `json:"metadata,omitempty"`
	Episodes    []PodcastFeedEpisode `json:"episodes,omitempty"`
	NumEpisodes int                  `json:"numEpisodes,omitempty"`
}

// PodcastFeedMetadata is the metadata of a podcast from its RSS feed.
type PodcastFeedMetadata struct {
	Image            string   `json:"image,omitempty"`
	Categories       []string `json:"categories,omitempty"`
	FeedURL          string   `json:"feedUrl,omitempty"`
	Description      string   `json:"description,omitempty"`
	DescriptionPlain string   `json:"descriptionPlain,omitempty"`
	Title            string   `json:"title,omitempty"`
	Language         string   `json:"language,omitempty"`
	// Explicit is reported by feeds as a string, usually "true" or
	// "false".
	Explicit string `json:"explicit,omitempty"`
	Author   string `json:"author,omitempty"`
	PubDate  string `json:"pubDate,omitempty"`
	Link     string `json:"link,omitempty"`
}

// PodcastFeedEpisode is an episode of a podcast from its RSS feed.
type PodcastFeedEpisode struct {
	Title            string `json:"title,omitempty"`
	Subtitle         string `json:"subtitle,omitempty"`
	Description      string `json:"description,omitempty"`
	DescriptionPlain string `json:"descriptionPlain,omitempty"`
	PubDate          string `json:"pubDate,omitempty"`
	EpisodeType      string `json:"episodeType,omitempty"`
	Season           string `json:"season,omitempty"`
	Episode          string `json:"episode,omitempty"`
	Author           string `json:"author,omitempty"`
	// Duration as reported by the RSS feed, e.g. "21:02".
	Duration    string                   `json:"duration,omitempty"`
	Explicit    string                   `json:"explicit,omitempty"`
	PublishedAt Millis                   `json:"publishedAt,omitempty"`
	Enclosure   *PodcastEpisodeEnclosure `json:"enclosure,omitempty"`
}

// Author is the author of books in a library.
type Author struct {
	client *Client

	ID          string `json:"id"`
	ASIN        string `json:"asin,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	ImagePath   string `json:"imagePath,omitempty"`
	AddedAt     Millis `json:"addedAt,omitempty"`
	UpdatedAt   Millis `json:"updatedAt,omitempty"`
	// Expanded variant only.
	NumBooks int `json:"numBooks,omitempty"`

	// Optional includes (see Client.Author).
	LibraryItems []LibraryItem  `json:"libraryItems,omitempty"`
	Series       []AuthorSeries `json:"series,omitempty"`
}

// AuthorSeries is a series with the author's books in it, returned when
// requesting an author with items and series included.
type AuthorSeries struct {
	ID    string        `json:"id"`
	Name  string        `json:"name"`
	Items []LibraryItem `json:"items"`
}

// Series is a series of books. The populated fields depend on the variant
// the server returns (base, num-books, books, or sequence).
type Series struct {
	client *Client

	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	AddedAt     Millis `json:"addedAt,omitempty"`
	UpdatedAt   Millis `json:"updatedAt,omitempty"`

	// Variant-specific fields.
	NameIgnorePrefix     string        `json:"nameIgnorePrefix,omitempty"`
	NameIgnorePrefixSort string        `json:"nameIgnorePrefixSort,omitempty"`
	Type                 string        `json:"type,omitempty"`
	LibraryItemIDs       []string      `json:"libraryItemIds,omitempty"`
	NumBooks             int           `json:"numBooks,omitempty"`
	Books                []LibraryItem `json:"books,omitempty"`
	TotalDuration        Seconds       `json:"totalDuration,omitempty"`
	Sequence             string        `json:"sequence,omitempty"`
	// SeriesSequenceList is set on collapsed subseries in library item
	// lists.
	SeriesSequenceList string `json:"seriesSequenceList,omitempty"`

	// Optional includes (see Client.Series).
	Progress *SeriesProgress `json:"progress,omitempty"`
	RSSFeed  *RSSFeed        `json:"rssFeed,omitempty"`
}

// SeriesProgress is the user's progress through a series.
type SeriesProgress struct {
	LibraryItemIDs         []string `json:"libraryItemIds"`
	LibraryItemIDsFinished []string `json:"libraryItemIdsFinished"`
	IsFinished             bool     `json:"isFinished"`
}

// Collection is a collection of book library items.
type Collection struct {
	client *Client

	ID          string        `json:"id"`
	LibraryID   string        `json:"libraryId,omitempty"`
	UserID      string        `json:"userId,omitempty"`
	Name        string        `json:"name,omitempty"`
	Description string        `json:"description,omitempty"`
	Books       []LibraryItem `json:"books,omitempty"`
	LastUpdate  Millis        `json:"lastUpdate,omitempty"`
	CreatedAt   Millis        `json:"createdAt,omitempty"`

	// Optional include (see Client.Collection).
	RSSFeed *RSSFeed `json:"rssFeed,omitempty"`
}

// Playlist is a user's playlist of library items and podcast episodes.
type Playlist struct {
	client *Client

	ID          string         `json:"id"`
	LibraryID   string         `json:"libraryId,omitempty"`
	UserID      string         `json:"userId,omitempty"`
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	CoverPath   string         `json:"coverPath,omitempty"`
	Items       []PlaylistItem `json:"items,omitempty"`
	LastUpdate  Millis         `json:"lastUpdate,omitempty"`
	CreatedAt   Millis         `json:"createdAt,omitempty"`
}

// PlaylistItem is an item of a playlist. Only LibraryItemID and EpisodeID
// are needed when adding items; the expanded fields are filled in by the
// server.
type PlaylistItem struct {
	LibraryItemID string          `json:"libraryItemId"`
	EpisodeID     string          `json:"episodeId,omitempty"`
	Episode       *PodcastEpisode `json:"episode,omitempty"`
	LibraryItem   *LibraryItem    `json:"libraryItem,omitempty"`
}

// MediaProgress is a user's playback progress for a book or podcast
// episode.
type MediaProgress struct {
	// ID is the library item ID for books, or
	// "<libraryItemId>-<episodeId>" for podcast episodes.
	ID                        string  `json:"id,omitempty"`
	LibraryItemID             string  `json:"libraryItemId,omitempty"`
	EpisodeID                 string  `json:"episodeId,omitempty"`
	Duration                  Seconds `json:"duration,omitempty"`
	Progress                  float64 `json:"progress,omitempty"`
	CurrentTime               Seconds `json:"currentTime,omitempty"`
	IsFinished                bool    `json:"isFinished,omitempty"`
	HideFromContinueListening bool    `json:"hideFromContinueListening,omitempty"`
	LastUpdate                Millis  `json:"lastUpdate,omitempty"`
	StartedAt                 Millis  `json:"startedAt,omitempty"`
	FinishedAt                Millis  `json:"finishedAt,omitempty"`

	// "With media" variant only.
	Media   *Media          `json:"media,omitempty"`
	Episode *PodcastEpisode `json:"episode,omitempty"`
}

// Play methods of a PlaybackSession.
const (
	PlayMethodDirectPlay   = 0
	PlayMethodDirectStream = 1
	PlayMethodTranscode    = 2
	PlayMethodLocal        = 3
)

// PlaybackSession is a playback session of a library item or podcast
// episode.
type PlaybackSession struct {
	ID            string         `json:"id,omitempty"`
	UserID        string         `json:"userId,omitempty"`
	LibraryID     string         `json:"libraryId,omitempty"`
	LibraryItemID string         `json:"libraryItemId,omitempty"`
	EpisodeID     string         `json:"episodeId,omitempty"`
	MediaType     string         `json:"mediaType,omitempty"`
	MediaMetadata *MediaMetadata `json:"mediaMetadata,omitempty"`
	Chapters      []Chapter      `json:"chapters,omitempty"`
	DisplayTitle  string         `json:"displayTitle,omitempty"`
	DisplayAuthor string         `json:"displayAuthor,omitempty"`
	CoverPath     string         `json:"coverPath,omitempty"`
	Duration      Seconds        `json:"duration,omitempty"`
	// PlayMethod is one of the PlayMethod constants.
	PlayMethod    int         `json:"playMethod,omitempty"`
	MediaPlayer   string      `json:"mediaPlayer,omitempty"`
	DeviceInfo    *DeviceInfo `json:"deviceInfo,omitempty"`
	ServerVersion string      `json:"serverVersion,omitempty"`
	Date          string      `json:"date,omitempty"`
	DayOfWeek     string      `json:"dayOfWeek,omitempty"`
	TimeListening Seconds     `json:"timeListening,omitempty"`
	StartTime     Seconds     `json:"startTime,omitempty"`
	CurrentTime   Seconds     `json:"currentTime,omitempty"`
	StartedAt     Millis      `json:"startedAt,omitempty"`
	UpdatedAt     Millis      `json:"updatedAt,omitempty"`

	// Expanded variant only.
	AudioTracks []AudioTrack `json:"audioTracks,omitempty"`
	LibraryItem *LibraryItem `json:"libraryItem,omitempty"`

	// Set by Client.Sessions when no user filter is given.
	User *SessionUser `json:"user,omitempty"`
}

// SessionUser identifies the user of a playback session in admin session
// listings.
type SessionUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// DeviceInfo describes the client device of a playback session. When
// starting playback, only the request fields (DeviceID, ClientName,
// ClientVersion, Manufacturer, Model, SDKVersion) should be set.
type DeviceInfo struct {
	ID             string `json:"id,omitempty"`
	UserID         string `json:"userId,omitempty"`
	DeviceID       string `json:"deviceId,omitempty"`
	IPAddress      string `json:"ipAddress,omitempty"`
	BrowserName    string `json:"browserName,omitempty"`
	BrowserVersion string `json:"browserVersion,omitempty"`
	OSName         string `json:"osName,omitempty"`
	OSVersion      string `json:"osVersion,omitempty"`
	DeviceName     string `json:"deviceName,omitempty"`
	DeviceType     string `json:"deviceType,omitempty"`
	Manufacturer   string `json:"manufacturer,omitempty"`
	Model          string `json:"model,omitempty"`
	SDKVersion     int    `json:"sdkVersion,omitempty"`
	ClientName     string `json:"clientName,omitempty"`
	ClientVersion  string `json:"clientVersion,omitempty"`
}

// User types.
const (
	UserTypeRoot  = "root"
	UserTypeAdmin = "admin"
	UserTypeUser  = "user"
	UserTypeGuest = "guest"
)

// User is a user account on the server.
type User struct {
	client *Client

	ID       string `json:"id"`
	Username string `json:"username,omitempty"`
	// Type is one of the UserType constants.
	Type                            string           `json:"type,omitempty"`
	Token                           string           `json:"token,omitempty"`
	MediaProgress                   []MediaProgress  `json:"mediaProgress,omitempty"`
	SeriesHideFromContinueListening []string         `json:"seriesHideFromContinueListening,omitempty"`
	Bookmarks                       []Bookmark       `json:"bookmarks,omitempty"`
	IsActive                        bool             `json:"isActive,omitempty"`
	IsLocked                        bool             `json:"isLocked,omitempty"`
	LastSeen                        Millis           `json:"lastSeen,omitempty"`
	CreatedAt                       Millis           `json:"createdAt,omitempty"`
	Permissions                     *UserPermissions `json:"permissions,omitempty"`
	LibrariesAccessible             []string         `json:"librariesAccessible,omitempty"`
	ItemTagsAccessible              []string         `json:"itemTagsAccessible,omitempty"`

	// "With session" variant (online users) only.
	Session *PlaybackSession `json:"session,omitempty"`
}

// UserPermissions are the permissions of a user.
type UserPermissions struct {
	Download              bool `json:"download"`
	Update                bool `json:"update"`
	Delete                bool `json:"delete"`
	Upload                bool `json:"upload"`
	AccessAllLibraries    bool `json:"accessAllLibraries"`
	AccessAllTags         bool `json:"accessAllTags"`
	AccessExplicitContent bool `json:"accessExplicitContent"`
}

// Bookmark is an audio bookmark of a user.
type Bookmark struct {
	LibraryItemID string `json:"libraryItemId"`
	Title         string `json:"title,omitempty"`
	// Time is the position of the bookmark in seconds.
	Time      int    `json:"time"`
	CreatedAt Millis `json:"createdAt,omitempty"`
}

// Backup is a backup of the server.
type Backup struct {
	ID                   string `json:"id"`
	BackupMetadataCovers bool   `json:"backupMetadataCovers,omitempty"`
	BackupDirPath        string `json:"backupDirPath,omitempty"`
	DatePretty           string `json:"datePretty,omitempty"`
	FullPath             string `json:"fullPath,omitempty"`
	Path                 string `json:"path,omitempty"`
	Filename             string `json:"filename,omitempty"`
	FileSize             int64  `json:"fileSize,omitempty"`
	CreatedAt            Millis `json:"createdAt,omitempty"`
	ServerVersion        string `json:"serverVersion,omitempty"`
}

// NotificationSettings are the server's notification settings.
type NotificationSettings struct {
	ID                   string         `json:"id,omitempty"`
	AppriseType          string         `json:"appriseType,omitempty"`
	AppriseAPIURL        string         `json:"appriseApiUrl,omitempty"`
	Notifications        []Notification `json:"notifications,omitempty"`
	MaxFailedAttempts    int            `json:"maxFailedAttempts,omitempty"`
	MaxNotificationQueue int            `json:"maxNotificationQueue,omitempty"`
	NotificationDelay    int            `json:"notificationDelay,omitempty"`
}

// Notification is a configured notification.
type Notification struct {
	ID                           string   `json:"id,omitempty"`
	LibraryID                    string   `json:"libraryId,omitempty"`
	EventName                    string   `json:"eventName,omitempty"`
	URLs                         []string `json:"urls,omitempty"`
	TitleTemplate                string   `json:"titleTemplate,omitempty"`
	BodyTemplate                 string   `json:"bodyTemplate,omitempty"`
	Enabled                      bool     `json:"enabled,omitempty"`
	Type                         string   `json:"type,omitempty"`
	LastFiredAt                  Millis   `json:"lastFiredAt,omitempty"`
	LastAttemptFailed            bool     `json:"lastAttemptFailed,omitempty"`
	NumConsecutiveFailedAttempts int      `json:"numConsecutiveFailedAttempts,omitempty"`
	NumTimesFired                int      `json:"numTimesFired,omitempty"`
	CreatedAt                    Millis   `json:"createdAt,omitempty"`
}

// NotificationEvent describes an event a notification can fire on.
type NotificationEvent struct {
	Name             string   `json:"name"`
	RequiresLibrary  bool     `json:"requiresLibrary,omitempty"`
	LibraryMediaType string   `json:"libraryMediaType,omitempty"`
	Description      string   `json:"description,omitempty"`
	Variables        []string `json:"variables,omitempty"`
	Defaults         struct {
		Title string `json:"title,omitempty"`
		Body  string `json:"body,omitempty"`
	} `json:"defaults"`
	// TestData holds example values used when firing a test
	// notification. Values are usually strings but may be numbers, so it
	// is decoded as map[string]any.
	TestData map[string]any `json:"testData,omitempty"`
}

// ServerSettings are the settings of the server.
type ServerSettings struct {
	ID                           string   `json:"id,omitempty"`
	ScannerFindCovers            bool     `json:"scannerFindCovers"`
	ScannerCoverProvider         string   `json:"scannerCoverProvider,omitempty"`
	ScannerParseSubtitle         bool     `json:"scannerParseSubtitle"`
	ScannerPreferMatchedMetadata bool     `json:"scannerPreferMatchedMetadata"`
	ScannerDisableWatcher        bool     `json:"scannerDisableWatcher"`
	StoreCoverWithItem           bool     `json:"storeCoverWithItem"`
	StoreMetadataWithItem        bool     `json:"storeMetadataWithItem"`
	MetadataFileFormat           string   `json:"metadataFileFormat,omitempty"`
	RateLimitLoginRequests       int      `json:"rateLimitLoginRequests,omitempty"`
	RateLimitLoginWindow         int      `json:"rateLimitLoginWindow,omitempty"`
	BackupSchedule               string   `json:"backupSchedule,omitempty"`
	BackupsToKeep                int      `json:"backupsToKeep,omitempty"`
	MaxBackupSize                int      `json:"maxBackupSize,omitempty"`
	LoggerDailyLogsToKeep        int      `json:"loggerDailyLogsToKeep,omitempty"`
	LoggerScannerLogsToKeep      int      `json:"loggerScannerLogsToKeep,omitempty"`
	HomeBookshelfView            int      `json:"homeBookshelfView"`
	BookshelfView                int      `json:"bookshelfView"`
	SortingIgnorePrefix          bool     `json:"sortingIgnorePrefix"`
	SortingPrefixes              []string `json:"sortingPrefixes,omitempty"`
	ChromecastEnabled            bool     `json:"chromecastEnabled"`
	DateFormat                   string   `json:"dateFormat,omitempty"`
	TimeFormat                   string   `json:"timeFormat,omitempty"`
	Language                     string   `json:"language,omitempty"`
	LogLevel                     int      `json:"logLevel,omitempty"`
	Version                      string   `json:"version,omitempty"`
}

// UnmarshalJSON tolerates Audiobookshelf returning backupSchedule as a
// boolean (false when auto-backups are disabled) rather than a cron
// string. A boolean is treated as an empty schedule.
func (s *ServerSettings) UnmarshalJSON(data []byte) error {
	type alias ServerSettings
	aux := struct {
		BackupSchedule json.RawMessage `json:"backupSchedule"`
		*alias
	}{alias: (*alias)(s)}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	s.BackupSchedule = ""
	if len(aux.BackupSchedule) > 0 {
		var sched string
		if err := json.Unmarshal(aux.BackupSchedule, &sched); err == nil {
			s.BackupSchedule = sched
		}
	}

	return nil
}

// RSSFeed is an open RSS feed for a library item, collection, or series.
type RSSFeed struct {
	ID            string           `json:"id,omitempty"`
	Slug          string           `json:"slug,omitempty"`
	UserID        string           `json:"userId,omitempty"`
	EntityType    string           `json:"entityType,omitempty"`
	EntityID      string           `json:"entityId,omitempty"`
	CoverPath     string           `json:"coverPath,omitempty"`
	ServerAddress string           `json:"serverAddress,omitempty"`
	FeedURL       string           `json:"feedUrl,omitempty"`
	Meta          *RSSFeedMetadata `json:"meta,omitempty"`
	Episodes      []RSSFeedEpisode `json:"episodes,omitempty"`
	CreatedAt     Millis           `json:"createdAt,omitempty"`
	UpdatedAt     Millis           `json:"updatedAt,omitempty"`
}

// RSSFeedMetadata is the metadata of an open RSS feed.
type RSSFeedMetadata struct {
	Title           string `json:"title,omitempty"`
	Description     string `json:"description,omitempty"`
	Author          string `json:"author,omitempty"`
	ImageURL        string `json:"imageUrl,omitempty"`
	FeedURL         string `json:"feedUrl,omitempty"`
	Link            string `json:"link,omitempty"`
	Explicit        bool   `json:"explicit,omitempty"`
	Type            string `json:"type,omitempty"`
	Language        string `json:"language,omitempty"`
	PreventIndexing bool   `json:"preventIndexing,omitempty"`
	OwnerName       string `json:"ownerName,omitempty"`
	OwnerEmail      string `json:"ownerEmail,omitempty"`
}

// RSSFeedEpisode is an episode of an open RSS feed.
type RSSFeedEpisode struct {
	ID          string `json:"id,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Enclosure   *struct {
		URL  string `json:"url,omitempty"`
		Type string `json:"type,omitempty"`
		Size int64  `json:"size,omitempty"`
	} `json:"enclosure,omitempty"`
	PubDate       string  `json:"pubDate,omitempty"`
	Link          string  `json:"link,omitempty"`
	Author        string  `json:"author,omitempty"`
	Explicit      bool    `json:"explicit,omitempty"`
	Duration      Seconds `json:"duration,omitempty"`
	Season        string  `json:"season,omitempty"`
	Episode       string  `json:"episode,omitempty"`
	EpisodeType   string  `json:"episodeType,omitempty"`
	LibraryItemID string  `json:"libraryItemId,omitempty"`
	EpisodeID     string  `json:"episodeId,omitempty"`
	TrackIndex    int     `json:"trackIndex,omitempty"`
	FullPath      string  `json:"fullPath,omitempty"`
}

// ListeningStats are listening statistics for a user.
type ListeningStats struct {
	// TotalTime is the total listening time in seconds.
	TotalTime Seconds `json:"totalTime"`
	// Items maps library item IDs to per-item stats.
	Items map[string]ItemListeningStats `json:"items,omitempty"`
	// Days maps days (YYYY-MM-DD) to listening time in seconds.
	Days map[string]Seconds `json:"days,omitempty"`
	// DayOfWeek maps weekday names to listening time in seconds.
	DayOfWeek map[string]Seconds `json:"dayOfWeek,omitempty"`
	// Today is the listening time today in seconds.
	Today          Seconds           `json:"today"`
	RecentSessions []PlaybackSession `json:"recentSessions,omitempty"`
}

// ItemListeningStats are listening statistics for one library item.
type ItemListeningStats struct {
	ID            string         `json:"id"`
	TimeListening Seconds        `json:"timeListening"`
	MediaMetadata *MediaMetadata `json:"mediaMetadata,omitempty"`
}
