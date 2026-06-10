package audiobookshelf

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return NewClient(server.URL, WithToken("test-token"))
}

func TestRequestHeaders(t *testing.T) {
	var got *http.Request
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		got = r.Clone(context.Background())
		if _, err := w.Write([]byte(`{}`)); err != nil {
			t.Errorf("writing file: %v", err)
		}
	})

	if err := client.Get(context.Background(), "/api/me", nil); err != nil {
		t.Fatalf("Get: %v", err)
	}

	if auth := got.Header.Get("Authorization"); auth != "Bearer test-token" {
		t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
	}

	if ua := got.Header.Get("User-Agent"); ua != DefaultUserAgent {
		t.Errorf("User-Agent = %q, want %q", ua, DefaultUserAgent)
	}
}

func TestErrorResponse(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not found", http.StatusNotFound)
	})

	err := client.Get(context.Background(), "/api/items/li_missing", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("error is %T, want *Error", err)
	}

	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want 404", apiErr.StatusCode)
	}

	if apiErr.Message != "Not found" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "Not found")
	}

	if !IsNotFound(err) {
		t.Error("IsNotFound = false, want true")
	}

	if IsUnauthorized(err) {
		t.Error("IsUnauthorized = true, want false")
	}
}

func TestEmptySuccessBody(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	var out struct{}
	if err := client.Get(context.Background(), "/ping", &out); err != nil {
		t.Fatalf("Get with empty body: %v", err)
	}
}

func TestLoginStoresToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/login" {
			t.Errorf("got %s %s, want POST /login", r.Method, r.URL.Path)
		}

		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decoding login body: %v", err)
		}

		if body["username"] != "root" || body["password"] != "secret" {
			t.Errorf("login body = %v", body)
		}

		encoder := json.NewEncoder(w)

		err := encoder.Encode(map[string]any{
			"user": map[string]any{
				"id":       "root",
				"username": "root",
				"token":    "new-token",
			},
			"userDefaultLibraryId": "lib_1",
		})
		if err != nil {
			t.Errorf("encoding user: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := NewClient(server.URL)
	resp, err := client.Login(context.Background(), "root", "secret")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	if resp.User == nil || resp.User.Username != "root" {
		t.Fatalf("unexpected user: %+v", resp.User)
	}

	if client.Token() != "new-token" {
		t.Errorf("Token = %q, want %q", client.Token(), "new-token")
	}

	if resp.UserDefaultLibraryID != "lib_1" {
		t.Errorf("UserDefaultLibraryID = %q, want lib_1", resp.UserDefaultLibraryID)
	}
}

func TestServerSettingsBackupScheduleFlexible(t *testing.T) {
	// Disabled auto-backups: Audiobookshelf returns a boolean.
	var disabled ServerSettings
	if err := json.Unmarshal([]byte(`{"backupSchedule":false,"backupsToKeep":2}`), &disabled); err != nil {
		t.Fatalf("bool schedule: %v", err)
	}
	if disabled.BackupSchedule != "" {
		t.Errorf("BackupSchedule = %q, want empty", disabled.BackupSchedule)
	}
	if disabled.BackupsToKeep != 2 {
		t.Errorf("BackupsToKeep = %d, want 2 (other fields must still decode)", disabled.BackupsToKeep)
	}

	// Enabled: a cron string.
	var enabled ServerSettings
	if err := json.Unmarshal([]byte(`{"backupSchedule":"30 1 * * *"}`), &enabled); err != nil {
		t.Fatalf("string schedule: %v", err)
	}
	if enabled.BackupSchedule != "30 1 * * *" {
		t.Errorf("BackupSchedule = %q", enabled.BackupSchedule)
	}
}

func TestNotificationEventTestDataMixedTypes(t *testing.T) {
	// Audiobookshelf sends testData values as a mix of strings and
	// numbers, so the map must tolerate any JSON scalar.
	var ev NotificationEvent
	body := `{"name":"onTest","testData":{"libraryItemId":"li_1","episodeIndex":3}}`
	if err := json.Unmarshal([]byte(body), &ev); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if ev.TestData["libraryItemId"] != "li_1" {
		t.Errorf("testData string = %v", ev.TestData["libraryItemId"])
	}
	if ev.TestData["episodeIndex"].(float64) != 3 {
		t.Errorf("testData number = %v", ev.TestData["episodeIndex"])
	}
}

func TestSeriesSequencesFlexibleUnmarshal(t *testing.T) {
	var fromArray MediaMetadata
	if err := json.Unmarshal([]byte(`{"series":[{"id":"ser_1","name":"A","sequence":"1"}]}`), &fromArray); err != nil {
		t.Fatalf("array unmarshal: %v", err)
	}

	if len(fromArray.Series) != 1 || fromArray.Series[0].ID != "ser_1" {
		t.Errorf("array form: %+v", fromArray.Series)
	}

	var fromObject MediaMetadata
	if err := json.Unmarshal([]byte(`{"series":{"id":"ser_2","name":"B","sequence":"3"}}`), &fromObject); err != nil {
		t.Fatalf("object unmarshal: %v", err)
	}

	if len(fromObject.Series) != 1 || fromObject.Series[0].Sequence != "3" {
		t.Errorf("object form: %+v", fromObject.Series)
	}
}

func TestGetBinary(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/items/li_1/cover" {
			t.Errorf("path = %s", r.URL.Path)
		}

		if r.URL.Query().Get("width") != "200" {
			t.Errorf("width = %s, want 200", r.URL.Query().Get("width"))
		}

		w.Header().Set("Content-Type", "image/jpeg")

		if _, err := w.Write([]byte("jpeg-bytes")); err != nil {
			t.Errorf("writing file: %v", err)
		}
	})

	body, contentType, err := client.LibraryItemCover(context.Background(), "li_1", &ImageParams{Width: 200})
	if err != nil {
		t.Fatalf("LibraryItemCover: %v", err)
	}

	defer func() { _ = body.Close() }()

	data, _ := io.ReadAll(body)
	if string(data) != "jpeg-bytes" {
		t.Errorf("body = %q", data)
	}

	if contentType != "image/jpeg" {
		t.Errorf("contentType = %q", contentType)
	}
}

func TestMultipartUpload(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("parsing multipart form: %v", err)
		}

		if got := r.FormValue("title"); got != "A Book" {
			t.Errorf("title = %q", got)
		}

		file, header, err := r.FormFile("0")
		if err != nil {
			t.Fatalf("form file: %v", err)
		}

		defer func() { _ = file.Close() }()

		if header.Filename != "book.m4b" {
			t.Errorf("filename = %q", header.Filename)
		}

		data, _ := io.ReadAll(file)
		if string(data) != "audio-data" {
			t.Errorf("file contents = %q", data)
		}

		w.WriteHeader(http.StatusOK)
	})

	err := client.UploadFiles(context.Background(), &UploadFilesRequest{
		Title:     "A Book",
		LibraryID: "lib_1",
		FolderID:  "fol_1",
		Files: []UploadFile{
			{Name: "book.m4b", Reader: strings.NewReader("audio-data")},
		},
	})

	if err != nil {
		t.Fatalf("UploadFiles: %v", err)
	}
}
