package audiobookshelf

import (
	"context"
	"strings"
)

// UpdateSeriesRequest are the parameters for UpdateSeries. Nil fields
// are left unchanged.
type UpdateSeriesRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// Series returns a series (GET /api/series/:id). include lists extras to
// include: "progress" and/or "rssfeed".
func (c *Client) Series(ctx context.Context, id string, include ...string) (*Series, error) {
	pb := apiPath("series").Seg(id)
	if len(include) > 0 {
		pb.Set("include", strings.Join(include, ","))
	}
	var series Series
	if err := c.Get(ctx, pb.String(), &series); err != nil {
		return nil, err
	}
	series.client = c
	return &series, nil
}

// UpdateSeries updates a series (PATCH /api/series/:id).
func (c *Client) UpdateSeries(ctx context.Context, id string, req *UpdateSeriesRequest) (*Series, error) {
	var series Series
	if err := c.Patch(ctx, apiPath("series").Seg(id).String(), req, &series); err != nil {
		return nil, err
	}
	series.client = c
	return &series, nil
}

// Update updates the series and refreshes its fields in place. See
// Client.UpdateSeries.
func (s *Series) Update(ctx context.Context, req *UpdateSeriesRequest) error {
	updated, err := s.client.UpdateSeries(ctx, s.ID, req)
	if err != nil {
		return err
	}
	*s = *updated
	return nil
}
