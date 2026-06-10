package audiobookshelf

import (
	"context"
	"net/url"
)

// NotificationData bundles the notification settings and available
// events.
type NotificationData struct {
	Data struct {
		Events []NotificationEvent `json:"events"`
	} `json:"data"`
	Settings *NotificationSettings `json:"settings"`
}

// UpdateNotificationSettingsRequest are the parameters for
// UpdateNotificationSettings.
type UpdateNotificationSettingsRequest struct {
	AppriseAPIURL        *string `json:"appriseApiUrl,omitempty"`
	MaxFailedAttempts    int     `json:"maxFailedAttempts,omitempty"`
	MaxNotificationQueue int     `json:"maxNotificationQueue,omitempty"`
}

// NotificationRequest are the parameters for CreateNotification and
// UpdateNotification. ID is required for updates and ignored on
// creation.
type NotificationRequest struct {
	ID            string   `json:"id,omitempty"`
	LibraryID     string   `json:"libraryId,omitempty"`
	EventName     string   `json:"eventName,omitempty"`
	URLs          []string `json:"urls,omitempty"`
	TitleTemplate string   `json:"titleTemplate,omitempty"`
	BodyTemplate  string   `json:"bodyTemplate,omitempty"`
	Enabled       bool     `json:"enabled,omitempty"`
	Type          string   `json:"type,omitempty"`
}

// NotificationSettings returns the server's notification settings and
// event data (GET /api/notifications). Requires admin.
func (c *Client) NotificationSettings(ctx context.Context) (*NotificationData, error) {
	var data NotificationData
	if err := c.Get(ctx, "/api/notifications", &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// UpdateNotificationSettings updates the server's notification settings
// (PATCH /api/notifications). Requires admin.
func (c *Client) UpdateNotificationSettings(ctx context.Context, req *UpdateNotificationSettingsRequest) error {
	return c.Patch(ctx, "/api/notifications", req, nil)
}

// NotificationEvents returns the available notification events
// (GET /api/notificationdata). Requires admin.
func (c *Client) NotificationEvents(ctx context.Context) ([]NotificationEvent, error) {
	var resp struct {
		Events []NotificationEvent `json:"events"`
	}
	if err := c.Get(ctx, "/api/notificationdata", &resp); err != nil {
		return nil, err
	}
	return resp.Events, nil
}

// FireTestNotificationEvent fires the test notification event
// (GET /api/notifications/test). fail makes the notification fail on
// purpose. Requires admin.
func (c *Client) FireTestNotificationEvent(ctx context.Context, fail bool) error {
	q := url.Values{}
	if fail {
		q.Set("fail", "1")
	}
	return c.Get(ctx, appendQuery("/api/notifications/test", q), nil)
}

// CreateNotification creates a notification (POST /api/notifications)
// and returns the updated notification settings. Requires admin.
func (c *Client) CreateNotification(ctx context.Context, req *NotificationRequest) (*NotificationSettings, error) {
	var resp struct {
		Settings *NotificationSettings `json:"settings"`
	}
	if err := c.Post(ctx, "/api/notifications", req, &resp); err != nil {
		return nil, err
	}
	return resp.Settings, nil
}

// DeleteNotification deletes a notification
// (DELETE /api/notifications/:id) and returns the updated notification
// settings. Requires admin.
func (c *Client) DeleteNotification(ctx context.Context, id string) (*NotificationSettings, error) {
	var resp struct {
		Settings *NotificationSettings `json:"settings"`
	}
	if err := c.Delete(ctx, "/api/notifications/"+url.PathEscape(id), &resp); err != nil {
		return nil, err
	}
	return resp.Settings, nil
}

// UpdateNotification updates a notification
// (PATCH /api/notifications/:id). req.ID must be set. Requires admin.
func (c *Client) UpdateNotification(ctx context.Context, req *NotificationRequest) (*NotificationSettings, error) {
	var resp struct {
		Settings *NotificationSettings `json:"settings"`
	}
	if err := c.Patch(ctx, "/api/notifications/"+url.PathEscape(req.ID), req, &resp); err != nil {
		return nil, err
	}
	return resp.Settings, nil
}

// SendTestNotification sends a test of a configured notification
// (GET /api/notifications/:id/test). Requires admin.
func (c *Client) SendTestNotification(ctx context.Context, id string) error {
	return c.Get(ctx, "/api/notifications/"+url.PathEscape(id)+"/test", nil)
}
