package audiobookshelf

import (
	"context"
)

// CreateUserRequest are the parameters for CreateUser.
type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	// Type is "guest", "user", or "admin".
	Type                            string           `json:"type"`
	IsActive                        *bool            `json:"isActive,omitempty"`
	IsLocked                        *bool            `json:"isLocked,omitempty"`
	Permissions                     *UserPermissions `json:"permissions,omitempty"`
	LibrariesAccessible             []string         `json:"librariesAccessible,omitempty"`
	ItemTagsAccessible              []string         `json:"itemTagsAccessible,omitempty"`
	SeriesHideFromContinueListening []string         `json:"seriesHideFromContinueListening,omitempty"`
	MediaProgress                   []MediaProgress  `json:"mediaProgress,omitempty"`
	Bookmarks                       []Bookmark       `json:"bookmarks,omitempty"`
}

// UpdateUserRequest are the parameters for UpdateUser. Nil/zero fields
// are left unchanged.
type UpdateUserRequest struct {
	Username                        string           `json:"username,omitempty"`
	Password                        string           `json:"password,omitempty"`
	Type                            string           `json:"type,omitempty"`
	IsActive                        *bool            `json:"isActive,omitempty"`
	Permissions                     *UserPermissions `json:"permissions,omitempty"`
	LibrariesAccessible             []string         `json:"librariesAccessible,omitempty"`
	ItemTagsAccessible              []string         `json:"itemTagsAccessible,omitempty"`
	SeriesHideFromContinueListening []string         `json:"seriesHideFromContinueListening,omitempty"`
}

// OnlineUsers is the response of OnlineUsers.
type OnlineUsers struct {
	UsersOnline  []User            `json:"usersOnline"`
	OpenSessions []PlaybackSession `json:"openSessions"`
}

// SessionsPage is one page of playback sessions.
type SessionsPage struct {
	Total        int               `json:"total"`
	NumPages     int               `json:"numPages"`
	Page         int               `json:"page"`
	ItemsPerPage int               `json:"itemsPerPage"`
	Sessions     []PlaybackSession `json:"sessions"`
	// UserFilter is set when the request filtered by user.
	UserFilter string `json:"userFilter,omitempty"`
}

func userPath(id string, rest ...string) string {
	return apiPath("users").Seg(id).Lit(rest...).String()
}

// CreateUser creates a user (POST /api/users). Requires admin.
func (c *Client) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
	var resp struct {
		User User `json:"user"`
	}
	if err := c.Post(ctx, apiPath("users").String(), req, &resp); err != nil {
		return nil, err
	}
	resp.User.client = c
	return &resp.User, nil
}

// Users returns all users (GET /api/users). Requires admin.
func (c *Client) Users(ctx context.Context) ([]User, error) {
	var resp struct {
		Users []User `json:"users"`
	}
	if err := c.Get(ctx, apiPath("users").String(), &resp); err != nil {
		return nil, err
	}
	for i := range resp.Users {
		resp.Users[i].client = c
	}
	return resp.Users, nil
}

// OnlineUsers returns the users currently online and the open playback
// sessions (GET /api/users/online). Requires admin.
func (c *Client) OnlineUsers(ctx context.Context) (*OnlineUsers, error) {
	var resp OnlineUsers
	if err := c.Get(ctx, apiPath("users", "online").String(), &resp); err != nil {
		return nil, err
	}
	for i := range resp.UsersOnline {
		resp.UsersOnline[i].client = c
	}
	return &resp, nil
}

// User returns a user (GET /api/users/:id). Requires admin.
func (c *Client) User(ctx context.Context, id string) (*User, error) {
	var user User
	if err := c.Get(ctx, userPath(id), &user); err != nil {
		return nil, err
	}
	user.client = c
	return &user, nil
}

// UpdateUser updates a user (PATCH /api/users/:id). Requires admin.
func (c *Client) UpdateUser(ctx context.Context, id string, req *UpdateUserRequest) (*User, error) {
	var resp struct {
		Success bool `json:"success"`
		User    User `json:"user"`
	}
	if err := c.Patch(ctx, userPath(id), req, &resp); err != nil {
		return nil, err
	}
	resp.User.client = c
	return &resp.User, nil
}

// DeleteUser deletes a user (DELETE /api/users/:id). Requires admin.
func (c *Client) DeleteUser(ctx context.Context, id string) error {
	return c.Delete(ctx, userPath(id), nil)
}

// UserListeningSessions lists a user's listening sessions
// (GET /api/users/:id/listening-sessions). Requires admin.
func (c *Client) UserListeningSessions(ctx context.Context, id string, params *SessionListParams) (*SessionsPage, error) {
	var page SessionsPage
	path := apiPath("users").Seg(id).Lit("listening-sessions").Query(params.values()).String()
	if err := c.Get(ctx, path, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// UserListeningStats returns a user's listening statistics
// (GET /api/users/:id/listening-stats). Requires admin.
func (c *Client) UserListeningStats(ctx context.Context, id string) (*ListeningStats, error) {
	var stats ListeningStats
	if err := c.Get(ctx, userPath(id, "listening-stats"), &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

// PurgeUserMediaProgress removes all media progress of a user
// (POST /api/users/:id/purge-media-progress). Requires admin.
func (c *Client) PurgeUserMediaProgress(ctx context.Context, id string) (*User, error) {
	var user User
	if err := c.Post(ctx, userPath(id, "purge-media-progress"), nil, &user); err != nil {
		return nil, err
	}
	user.client = c
	return &user, nil
}

// Update updates the user and refreshes its fields in place. See
// Client.UpdateUser.
func (u *User) Update(ctx context.Context, req *UpdateUserRequest) error {
	updated, err := u.client.UpdateUser(ctx, u.ID, req)
	if err != nil {
		return err
	}
	client := u.client
	*u = *updated
	u.client = client
	return nil
}

// Delete deletes the user. See Client.DeleteUser.
func (u *User) Delete(ctx context.Context) error {
	return u.client.DeleteUser(ctx, u.ID)
}

// ListeningSessions lists the user's listening sessions. See
// Client.UserListeningSessions.
func (u *User) ListeningSessions(ctx context.Context, params *SessionListParams) (*SessionsPage, error) {
	return u.client.UserListeningSessions(ctx, u.ID, params)
}

// ListeningStats returns the user's listening statistics. See
// Client.UserListeningStats.
func (u *User) ListeningStats(ctx context.Context) (*ListeningStats, error) {
	return u.client.UserListeningStats(ctx, u.ID)
}

// PurgeMediaProgress removes all media progress of the user. See
// Client.PurgeUserMediaProgress.
func (u *User) PurgeMediaProgress(ctx context.Context) error {
	_, err := u.client.PurgeUserMediaProgress(ctx, u.ID)
	return err
}
