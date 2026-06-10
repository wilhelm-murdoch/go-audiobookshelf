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
package audiobookshelf

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

const (
	Version          = "0.1.0"
	DefaultUserAgent = "go-audiobookshelf/" + Version
)

// Client is an Audiobookshelf API client. It is safe for concurrent use once
// configured; SetToken should not be called concurrently with requests.
type Client struct {
	httpClient *http.Client
	baseURL    string
	token      string
	userAgent  string
}

// Option configures a Client.
type Option func(*Client)

// WithToken sets the API token used as a Bearer token on every request. Use
// the user's token or an API key created in the server settings.
func WithToken(token string) Option {
	return func(c *Client) { c.token = token }
}

// WithHTTPClient sets a custom *http.Client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// WithTimeout sets the request timeout on the underlying *http.Client.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.httpClient.Timeout = d }
}

// WithUserAgent overrides the User-Agent header.
func WithUserAgent(ua string) Option {
	return func(c *Client) { c.userAgent = ua }
}

// WithInsecureSkipVerify disables TLS certificate verification.
func WithInsecureSkipVerify() Option {
	return func(c *Client) {
		transport, ok := c.httpClient.Transport.(*http.Transport)
		if !ok || transport == nil {
			transport = http.DefaultTransport.(*http.Transport).Clone()
		}
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{}
		}
		transport.TLSClientConfig.InsecureSkipVerify = true
		c.httpClient.Transport = transport
	}
}

// NewClient returns a Client for the Audiobookshelf server at baseURL.
func NewClient(baseURL string, opts ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: time.Minute},
		baseURL:    strings.TrimRight(baseURL, "/"),
		userAgent:  DefaultUserAgent,
	}

	for _, opt := range opts {
		opt(c)

	}
	return c
}

// BaseURL returns the configured server URL without a trailing slash.
func (c *Client) BaseURL() string { return c.baseURL }

// Token returns the API token currently in use.
func (c *Client) Token() string { return c.token }

// SetToken replaces the API token used for subsequent requests. Login calls
// this automatically.
func (c *Client) SetToken(token string) { c.token = token }

// Get performs a GET request against path (e.g. "/api/libraries") and decodes
// the JSON response into out. Pass nil to discard the response body. The
// typed methods should normally be preferred; Get and friends are escape
// hatches for endpoints or fields this library does not model.
func (c *Client) Get(ctx context.Context, path string, out any) error {
	return c.do(ctx, http.MethodGet, path, nil, out)
}

// Post performs a POST request with an optional JSON body.
func (c *Client) Post(ctx context.Context, path string, body, out any) error {
	return c.do(ctx, http.MethodPost, path, body, out)
}

// Patch performs a PATCH request with an optional JSON body.
func (c *Client) Patch(ctx context.Context, path string, body, out any) error {
	return c.do(ctx, http.MethodPatch, path, body, out)
}

// Put performs a PUT request with an optional JSON body.
func (c *Client) Put(ctx context.Context, path string, body, out any) error {
	return c.do(ctx, http.MethodPut, path, body, out)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string, out any) error {
	return c.do(ctx, http.MethodDelete, path, nil, out)
}

func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("audiobookshelf: building request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return req, nil
}

func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("audiobookshelf: encoding %s %s request: %w", method, path, err)
		}
		reader = bytes.NewReader(buf)
	}

	req, err := c.newRequest(ctx, method, path, reader)
	if err != nil {
		return err

	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")

	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("audiobookshelf: %s %s: %w", method, path, err)
	}

	defer resp.Body.Close()

	if err := checkResponse(resp, method, path); err != nil {
		return err
	}

	if out == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		if errors.Is(err, io.EOF) {
			return nil // success with an empty body
		}
		return fmt.Errorf("audiobookshelf: decoding %s %s response: %w", method, path, err)
	}
	return nil
}

// getBinary performs a GET request and returns the raw response body. The
// caller must close the returned ReadCloser. The second return value is the
// response Content-Type.
func (c *Client) getBinary(ctx context.Context, path string) (io.ReadCloser, string, error) {
	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, "", err
	}

	req.Header.Del("Accept")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("audiobookshelf: GET %s: %w", path, err)
	}

	if err := checkResponse(resp, http.MethodGet, path); err != nil {
		resp.Body.Close()
		return nil, "", err
	}

	return resp.Body, resp.Header.Get("Content-Type"), nil
}

// multipartFile is one file part of a multipart upload.
type multipartFile struct {
	field    string
	filename string
	reader   io.Reader
}

// postMultipart performs a multipart/form-data POST with the given form fields
// and files, streaming the request body.
func (c *Client) postMultipart(ctx context.Context, path string, fields map[string]string, files []multipartFile, out any) error {
	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)
	go func() {
		err := func() error {
			for key, value := range fields {
				if err := mw.WriteField(key, value); err != nil {
					return err
				}
			}

			for _, f := range files {
				part, err := mw.CreateFormFile(f.field, f.filename)
				if err != nil {
					return err
				}
				if _, err := io.Copy(part, f.reader); err != nil {
					return err
				}
			}

			return mw.Close()
		}()

		pw.CloseWithError(err)
	}()

	req, err := c.newRequest(ctx, http.MethodPost, path, pr)
	if err != nil {
		return err

	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("audiobookshelf: POST %s: %w", path, err)
	}

	defer resp.Body.Close()

	if err := checkResponse(resp, http.MethodPost, path); err != nil {
		return err
	}

	if out == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return fmt.Errorf("audiobookshelf: decoding POST %s response: %w", path, err)
	}

	return nil
}
