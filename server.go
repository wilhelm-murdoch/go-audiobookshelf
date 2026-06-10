package audiobookshelf

import "context"

// LoginResponse is the response of Login and Authorize.
type LoginResponse struct {
	User                 *User           `json:"user"`
	UserDefaultLibraryID string          `json:"userDefaultLibraryId"`
	ServerSettings       *ServerSettings `json:"serverSettings"`
	// Source is the server's installation source, e.g. "docker".
	Source string `json:"Source"`
}

// ServerStatus is the initialization status of the server.
type ServerStatus struct {
	IsInit   bool   `json:"isInit"`
	Language string `json:"language"`
	// ConfigPath and MetadataPath are only set while IsInit is false.
	ConfigPath   string `json:"ConfigPath,omitempty"`
	MetadataPath string `json:"MetadataPath,omitempty"`
}

// Login authenticates with a username and password (POST /login). On
// success the user's token is stored on the client for subsequent
// requests.
func (c *Client) Login(ctx context.Context, username, password string) (*LoginResponse, error) {
	body := map[string]string{"username": username, "password": password}
	var resp LoginResponse
	if err := c.Post(ctx, rawPath("/login").String(), body, &resp); err != nil {
		return nil, err
	}
	if resp.User != nil {
		resp.User.client = c
		if resp.User.Token != "" {
			c.SetToken(resp.User.Token)
		}
	}
	return &resp, nil
}

// Logout logs the client out of the server (POST /logout). socketID is
// optional and removes the socket from the server's client list.
func (c *Client) Logout(ctx context.Context, socketID string) error {
	var body any
	if socketID != "" {
		body = map[string]string{"socketId": socketID}
	}
	return c.Post(ctx, rawPath("/logout").String(), body, nil)
}

// InitServer initializes a brand-new server with a root user
// (POST /init).
func (c *Client) InitServer(ctx context.Context, rootUsername, rootPassword string) error {
	body := map[string]map[string]string{
		"newRoot": {"username": rootUsername, "password": rootPassword},
	}
	return c.Post(ctx, rawPath("/init").String(), body, nil)
}

// Status reports the server's initialization status (GET /status). It
// does not require authentication.
func (c *Client) Status(ctx context.Context) (*ServerStatus, error) {
	var status ServerStatus
	if err := c.Get(ctx, rawPath("/status").String(), &status); err != nil {
		return nil, err
	}
	return &status, nil
}

// Ping checks that the server is up and responding with JSON
// (GET /ping). It does not require authentication.
func (c *Client) Ping(ctx context.Context) error {
	return c.Get(ctx, rawPath("/ping").String(), nil)
}

// Healthcheck checks that the server is operating (GET /healthcheck). It
// does not require authentication.
func (c *Client) Healthcheck(ctx context.Context) error {
	return c.Get(ctx, rawPath("/healthcheck").String(), nil)
}
