// Package audiobookshelf is a client for the Audiobookshelf API
//
// A Client is created with NewClient and configured with functional options.
// Authentication is either a pre-existing API token or a username/password
// login:
//
//	client := audiobookshelf.NewClient("https://abs.example.com")
//	if _, err := client.Login(ctx, "user", "pass"); err != nil {
//		// handle error
//	}
//
//	libraries, err := client.Libraries(ctx)
//
// Methods are grouped by API resource, one file per group, mirroring the
// sections of the official API documentation. Resources returned by the
// client carry the client with them so follow-up calls can be chained:
//
//	library, _ := client.Library(ctx, "lib_...")
//	page, _ := library.Items(ctx, nil)
//
// The transport, path building, error model, and authentication live in
// the internal/rest toolkit; this package is the typed Audiobookshelf
// layer on top of it.
package audiobookshelf

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/wilhelm-murdoch/go-audiobookshelf/internal/rest"
)

const (
	Version          = "0.1.0"
	DefaultUserAgent = "go-audiobookshelf/" + Version

	// TestedServerVersion is the Audiobookshelf release this version of
	// the client is verified against in CI. Other server versions usually
	// work; a mismatch is the first thing to check when a response fails
	// to decode.
	TestedServerVersion = "2.35.1"

	errorPrefix = "audiobookshelf"
)

// Client is an Audiobookshelf API client. It is safe for concurrent use
// once configured; SetToken should not be called concurrently with
// requests.
type Client struct {
	rest     *rest.Client
	auth     *rest.BearerToken
	restOpts []rest.Option
}

// Option configures a Client.
type Option func(*Client)

// WithToken sets the API token used as a Bearer token on every request. Use
// the user's token or an API key created in the server settings.
func WithToken(token string) Option {
	return func(c *Client) { c.auth.SetToken(token) }
}

// WithHTTPClient sets a custom *http.Client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.restOpts = append(c.restOpts, rest.WithHTTPClient(hc)) }
}

// WithTimeout sets the request timeout on the underlying *http.Client.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.restOpts = append(c.restOpts, rest.WithTimeout(d)) }
}

// WithUserAgent overrides the User-Agent header.
func WithUserAgent(ua string) Option {
	return func(c *Client) { c.restOpts = append(c.restOpts, rest.WithUserAgent(ua)) }
}

// WithInsecureSkipVerify disables TLS certificate verification.
func WithInsecureSkipVerify() Option {
	return func(c *Client) { c.restOpts = append(c.restOpts, rest.WithInsecureSkipVerify()) }
}

// WithDebug logs every request and response — method, URL, headers, and
// body — to w. It is a debugging aid for inspecting what the server
// actually sends; it is not a production logger.
//
// Security: the Authorization and cookie headers are redacted, but bodies
// are printed verbatim and may contain secrets — most notably the
// username and password sent by Login. Never enable it against a server
// on an untrusted network, and scrub the output before sharing it.
func WithDebug(w io.Writer) Option {
	return func(c *Client) { c.restOpts = append(c.restOpts, rest.WithDebug(w)) }
}

// NewClient returns a Client for the Audiobookshelf server at baseURL.
func NewClient(baseURL string, opts ...Option) *Client {
	c := &Client{auth: rest.NewBearerToken("")}

	for _, opt := range opts {
		opt(c)
	}

	restOpts := append([]rest.Option{
		rest.WithUserAgent(DefaultUserAgent),
		rest.WithAuthenticator(c.auth),
		rest.WithErrorPrefix(errorPrefix),
	}, c.restOpts...)

	c.rest = rest.New(strings.TrimRight(baseURL, "/"), restOpts...)

	return c
}

// BaseURL returns the configured server URL without a trailing slash.
func (c *Client) BaseURL() string { return c.rest.BaseURL() }

// Token returns the API token currently in use.
func (c *Client) Token() string { return c.auth.Token() }

// SetToken replaces the API token used for subsequent requests. Login calls
// this automatically.
func (c *Client) SetToken(token string) { c.auth.SetToken(token) }

// Get performs a GET request against path (e.g. "/api/libraries") and decodes
// the JSON response into out. Pass nil to discard the response body. The
// typed methods should normally be preferred; Get and friends are escape
// hatches for endpoints or fields this library does not model.
func (c *Client) Get(ctx context.Context, path string, out any) error {
	return c.rest.Get(ctx, path, out)
}

// Post performs a POST request with an optional JSON body.
func (c *Client) Post(ctx context.Context, path string, body, out any) error {
	return c.rest.Post(ctx, path, body, out)
}

// Patch performs a PATCH request with an optional JSON body.
func (c *Client) Patch(ctx context.Context, path string, body, out any) error {
	return c.rest.Patch(ctx, path, body, out)
}

// Put performs a PUT request with an optional JSON body.
func (c *Client) Put(ctx context.Context, path string, body, out any) error {
	return c.rest.Put(ctx, path, body, out)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string, out any) error {
	return c.rest.Delete(ctx, path, out)
}

func (c *Client) getBinary(ctx context.Context, path string) (io.ReadCloser, string, error) {
	return c.rest.GetBinary(ctx, path)
}

// multipartFile is one file part of a multipart upload.
type multipartFile struct {
	field    string
	filename string
	reader   io.Reader
}

func (c *Client) postMultipart(ctx context.Context, path string, fields map[string]string, files []multipartFile, out any) error {
	parts := make([]rest.MultipartFile, len(files))
	for i, f := range files {
		parts[i] = rest.MultipartFile{Field: f.field, Filename: f.filename, Reader: f.reader}
	}

	return c.rest.PostMultipart(ctx, path, fields, parts, out)
}
