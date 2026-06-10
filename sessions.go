package audiobookshelf

import (
	"context"
)

// SessionSync are the playback position parameters for SyncOpenSession
// and CloseOpenSession.
type SessionSync struct {
	// CurrentTime is the playback position in seconds.
	CurrentTime Seconds `json:"currentTime"`
	// TimeListened is the seconds listened since the last sync.
	TimeListened Seconds `json:"timeListened"`
	// Duration is the total duration of the playing item in seconds.
	Duration Seconds `json:"duration"`
}

// SessionSyncResult is the per-session result of SyncLocalSessions.
type SessionSyncResult struct {
	ID             string `json:"id"`
	Success        bool   `json:"success"`
	Error          string `json:"error,omitempty"`
	ProgressSynced bool   `json:"progressSynced,omitempty"`
}

// Sessions lists playback sessions on the server (GET /api/sessions).
// Requires admin. Use params.User to filter by user.
func (c *Client) Sessions(ctx context.Context, params *SessionListParams) (*SessionsPage, error) {
	var page SessionsPage
	if err := c.Get(ctx, apiPath("sessions").Query(params.values()).String(), &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// DeleteSession deletes a playback session (DELETE /api/sessions/:id).
func (c *Client) DeleteSession(ctx context.Context, id string) error {
	return c.Delete(ctx, apiPath("sessions").Seg(id).String(), nil)
}

// SyncLocalSession syncs a playback session from a client device
// (POST /api/session/local).
func (c *Client) SyncLocalSession(ctx context.Context, session *PlaybackSession) error {
	return c.Post(ctx, apiPath("session", "local").String(), session, nil)
}

// SyncLocalSessions syncs multiple playback sessions from a client
// device (POST /api/session/local-all).
func (c *Client) SyncLocalSessions(ctx context.Context, sessions []PlaybackSession) ([]SessionSyncResult, error) {
	body := map[string]any{"sessions": sessions}
	var resp struct {
		Results []SessionSyncResult `json:"results"`
	}
	if err := c.Post(ctx, apiPath("session", "local-all").String(), body, &resp); err != nil {
		return nil, err
	}
	return resp.Results, nil
}

// OpenSession returns one of your open playback sessions
// (GET /api/session/:id).
func (c *Client) OpenSession(ctx context.Context, id string) (*PlaybackSession, error) {
	var session PlaybackSession
	if err := c.Get(ctx, apiPath("session").Seg(id).String(), &session); err != nil {
		return nil, err
	}
	return &session, nil
}

// SyncOpenSession reports playback progress for one of your open
// playback sessions (POST /api/session/:id/sync).
func (c *Client) SyncOpenSession(ctx context.Context, id string, sync *SessionSync) error {
	return c.Post(ctx, apiPath("session").Seg(id).Lit("sync").String(), sync, nil)
}

// CloseOpenSession closes one of your open playback sessions
// (POST /api/session/:id/close). sync is optional; when given, the
// session is synced before closing.
func (c *Client) CloseOpenSession(ctx context.Context, id string, sync *SessionSync) error {
	var body any
	if sync != nil {
		body = sync
	}
	return c.Post(ctx, apiPath("session").Seg(id).Lit("close").String(), body, nil)
}
