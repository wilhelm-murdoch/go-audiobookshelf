package audiobookshelf

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadBackupTo(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/backups/2022-11-14T0130/download" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/octet-stream")

		if _, err := w.Write([]byte("backup-bytes")); err != nil {
			t.Errorf("writing file: %v", err)
		}
	})

	ctx := context.Background()
	dir := t.TempDir()

	// Destination is a directory: file is named after the backup ID.
	path, err := client.DownloadBackupTo(ctx, "2022-11-14T0130", dir)
	if err != nil {
		t.Fatalf("DownloadBackupTo(dir): %v", err)
	}

	want := filepath.Join(dir, "2022-11-14T0130.audiobookshelf")
	if path != want {
		t.Errorf("path = %q, want %q", path, want)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading downloaded backup: %v", err)
	}

	if string(data) != "backup-bytes" {
		t.Errorf("contents = %q", data)
	}

	// Destination is an explicit file path.
	explicit := filepath.Join(dir, "my-backup.bin")
	path, err = client.DownloadBackupTo(ctx, "2022-11-14T0130", explicit)
	if err != nil {
		t.Fatalf("DownloadBackupTo(file): %v", err)
	}

	if path != explicit {
		t.Errorf("path = %q, want %q", path, explicit)
	}

	if data, _ := os.ReadFile(explicit); string(data) != "backup-bytes" {
		t.Errorf("contents = %q", data)
	}
}
