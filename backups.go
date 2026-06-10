package audiobookshelf

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
)

// Backups returns all backups on the server (GET /api/backups). Requires
// admin.
func (c *Client) Backups(ctx context.Context) ([]Backup, error) {
	var resp struct {
		Backups []Backup `json:"backups"`
	}
	if err := c.Get(ctx, "/api/backups", &resp); err != nil {
		return nil, err
	}
	return resp.Backups, nil
}

// CreateBackup creates a backup (POST /api/backups) and returns all
// backups, including the new one. Requires admin.
func (c *Client) CreateBackup(ctx context.Context) ([]Backup, error) {
	var resp struct {
		Backups []Backup `json:"backups"`
	}
	if err := c.Post(ctx, "/api/backups", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Backups, nil
}

// DeleteBackup deletes a backup (DELETE /api/backups/:id) and returns
// the remaining backups. Requires admin.
func (c *Client) DeleteBackup(ctx context.Context, id string) ([]Backup, error) {
	var resp struct {
		Backups []Backup `json:"backups"`
	}
	if err := c.Delete(ctx, "/api/backups/"+url.PathEscape(id), &resp); err != nil {
		return nil, err
	}
	return resp.Backups, nil
}

// DownloadBackup streams a backup file
// (GET /api/backups/:id/download; present on the server but not in the
// API docs). The caller must close the reader. Requires the root user.
func (c *Client) DownloadBackup(ctx context.Context, id string) (io.ReadCloser, error) {
	body, _, err := c.getBinary(ctx, "/api/backups/"+url.PathEscape(id)+"/download")
	return body, err
}

// DownloadBackupTo downloads a backup file to dest. If dest is an
// existing directory, the backup is saved inside it as
// "<id>.audiobookshelf"; otherwise dest is used as the file path. It
// returns the path written. Requires the root user.
func (c *Client) DownloadBackupTo(ctx context.Context, id, dest string) (string, error) {
	body, err := c.DownloadBackup(ctx, id)
	if err != nil {
		return "", err
	}

	defer func() { _ = body.Close() }()

	path := dest
	if info, err := os.Stat(dest); err == nil && info.IsDir() {
		path = filepath.Join(dest, id+".audiobookshelf")
	}

	file, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("audiobookshelf: creating backup file: %w", err)
	}

	if _, err := io.Copy(file, body); err != nil {
		if err := file.Close(); err != nil {
			return "", fmt.Errorf("audiobookshelf: closing file: %w", err)
		}

		if err := os.Remove(path); err != nil {
			return "", fmt.Errorf("audiobookshelf: removing file: %w", err)
		}

		return "", fmt.Errorf("audiobookshelf: writing backup file: %w", err)
	}

	if err := file.Close(); err != nil {
		if err := os.Remove(path); err != nil {
			return "", fmt.Errorf("audiobookshelf: removing file: %w", err)
		}

		return "", fmt.Errorf("audiobookshelf: writing backup file: %w", err)
	}

	return path, nil
}

// ApplyBackup restores a backup (GET /api/backups/:id/apply). Requires
// admin.
func (c *Client) ApplyBackup(ctx context.Context, id string) error {
	return c.Get(ctx, "/api/backups/"+url.PathEscape(id)+"/apply", nil)
}

// UpdateBackupPath changes the directory on the server where backups
// are stored (PATCH /api/backups/path; present on the server but not in
// the API docs). The directory is created if it does not exist. If the
// server's BACKUP_PATH environment variable is set, the change reverts
// on restart. Requires the root user.
func (c *Client) UpdateBackupPath(ctx context.Context, path string) error {
	return c.Patch(ctx, "/api/backups/path", map[string]string{"path": path}, nil)
}

// UploadBackup uploads a backup file (POST /api/backups/upload) and
// returns all backups, including the uploaded one. Requires admin.
func (c *Client) UploadBackup(ctx context.Context, filename string, file io.Reader) ([]Backup, error) {
	var resp struct {
		Backups []Backup `json:"backups"`
	}

	files := []multipartFile{{field: "file", filename: filename, reader: file}}
	if err := c.postMultipart(ctx, "/api/backups/upload", nil, files, &resp); err != nil {
		return nil, err
	}

	return resp.Backups, nil
}
