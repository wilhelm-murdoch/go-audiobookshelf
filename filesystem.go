package audiobookshelf

import "context"

// FilesystemDirectory is a directory on the server's filesystem.
type FilesystemDirectory struct {
	Path     string                `json:"path"`
	Dirname  string                `json:"dirname"`
	FullPath string                `json:"fullPath"`
	Level    int                   `json:"level"`
	Dirs     []FilesystemDirectory `json:"dirs,omitempty"`
}

// Filesystem lists the directories available on the server's filesystem
// (GET /api/filesystem). Requires admin.
func (c *Client) Filesystem(ctx context.Context) ([]FilesystemDirectory, error) {
	var resp struct {
		Directories []FilesystemDirectory `json:"directories"`
	}
	if err := c.Get(ctx, "/api/filesystem", &resp); err != nil {
		return nil, err
	}
	return resp.Directories, nil
}
