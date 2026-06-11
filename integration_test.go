//go:build integration

// Package audiobookshelf integration tests run against a live
// Audiobookshelf server (a container in CI). They are excluded from the
// normal unit suite by the "integration" build tag and only run when
// ABS_BASE_URL is set:
//
//	ABS_BASE_URL=http://localhost:13378 go test -tags=integration -count=1 .
//
// In CI a server is provided as a Woodpecker service; see
// .woodpecker/workflow.yaml. These tests cover the slice of the API that
// needs no seeded media (server lifecycle, users, libraries, collections,
// playlists, tags/genres, settings). Media-dependent endpoints (items,
// covers, play sessions, scans, podcast downloads) are intentionally out
// of scope here - they require mounted fixtures and async scan polling.
package audiobookshelf

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

const (
	itRootUser = "root"
	itRootPass = "integration-pass"
)

// itClient is the shared, authenticated client built in TestMain.
var itClient *Client

func TestMain(m *testing.M) {
	baseURL := os.Getenv("ABS_BASE_URL")
	if baseURL == "" {
		fmt.Println("ABS_BASE_URL not set; skipping integration tests")
		// Exit 0 so a tagged run without a server is a no-op, not a failure.
		os.Exit(0)
	}

	client := NewClient(baseURL)

	if err := bootstrap(client); err != nil {
		fmt.Printf("integration bootstrap failed: %v\n", err)
		os.Exit(1)
	}

	itClient = client
	os.Exit(m.Run())
}

// bootstrap waits for the server, initializes the root user on a fresh
// server, and logs in. It is idempotent so the suite can re-run against an
// already-initialized server.
func bootstrap(c *Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	if err := waitReady(ctx, c); err != nil {
		return err
	}

	status, err := c.Status(ctx)
	if err != nil {
		return fmt.Errorf("status: %w", err)
	}

	if !status.IsInit {
		if err := c.InitServer(ctx, itRootUser, itRootPass); err != nil {
			return fmt.Errorf("init server: %w", err)
		}

		// The server may restart its services right after init; give the
		// login endpoint a moment to come back.
		if err := waitReady(ctx, c); err != nil {
			return err
		}
	}

	if _, err := c.Login(ctx, itRootUser, itRootPass); err != nil {
		return fmt.Errorf("login: %w", err)
	}

	return nil
}

// waitReady polls the health check until the server responds or ctx is
// done. Audiobookshelf takes a few seconds to boot, and Woodpecker does
// not wait for service containers to become healthy.
func waitReady(ctx context.Context, c *Client) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		if err := c.Healthcheck(ctx); err == nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("server not ready: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}

func TestIntegrationServerLifecycle(t *testing.T) {
	ctx := context.Background()

	status, err := itClient.Status(ctx)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if !status.IsInit {
		t.Error("server reports IsInit=false after bootstrap")
	}

	if itClient.Token() == "" {
		t.Error("client has no token after login")
	}

	me, err := itClient.Me(ctx)
	if err != nil {
		t.Fatalf("Me: %v", err)
	}
	if me.Username != itRootUser {
		t.Errorf("Me().Username = %q, want %q", me.Username, itRootUser)
	}

	if _, err := itClient.Authorize(ctx); err != nil {
		t.Errorf("Authorize: %v", err)
	}
}

func TestIntegrationUserLifecycle(t *testing.T) {
	ctx := context.Background()
	username := "it-user-" + uniqueSuffix()

	created, err := itClient.CreateUser(ctx, &CreateUserRequest{
		Username: username,
		Password: "pw-" + uniqueSuffix(),
		Type:     UserTypeUser,
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	t.Cleanup(func() {
		if err := itClient.DeleteUser(context.Background(), created.ID); err != nil {
			t.Errorf("cleanup DeleteUser: %v", err)
		}
	})

	if created.Username != username {
		t.Errorf("created username = %q, want %q", created.Username, username)
	}

	users, err := itClient.Users(ctx)
	if err != nil {
		t.Fatalf("Users: %v", err)
	}
	if !containsUser(users, created.ID) {
		t.Errorf("created user %s not in Users() listing", created.ID)
	}
}

func TestIntegrationLibraryAndPlaylist(t *testing.T) {
	ctx := context.Background()

	// /metadata is a standard, always-present directory in the official
	// image, so the library can be created without seeding media.
	lib, err := itClient.CreateLibrary(ctx, &CreateLibraryRequest{
		Name:    "it-lib-" + uniqueSuffix(),
		Folders: []Folder{{FullPath: "/metadata"}},
	})
	if err != nil {
		t.Fatalf("CreateLibrary: %v", err)
	}
	t.Cleanup(func() {
		if err := itClient.DeleteLibrary(context.Background(), lib.ID); err != nil {
			t.Errorf("cleanup DeleteLibrary: %v", err)
		}
	})

	fetched, err := itClient.Library(ctx, lib.ID)
	if err != nil {
		t.Fatalf("Library: %v", err)
	}
	if fetched.ID != lib.ID {
		t.Errorf("Library().ID = %q, want %q", fetched.ID, lib.ID)
	}

	// NOTE: collections cannot be created empty - Audiobookshelf rejects
	// them with "No books" - so collection CRUD is deferred to the
	// media-dependent (tier 2) suite. Playlists may be created empty.
	pl, err := itClient.CreatePlaylist(ctx, &CreatePlaylistRequest{
		LibraryID: lib.ID,
		Name:      "it-playlist",
	})
	if err != nil {
		t.Fatalf("CreatePlaylist: %v", err)
	}
	t.Cleanup(func() {
		if err := itClient.DeletePlaylist(context.Background(), pl.ID); err != nil {
			t.Errorf("cleanup DeletePlaylist: %v", err)
		}
	})

	libs, err := itClient.Libraries(ctx)
	if err != nil {
		t.Fatalf("Libraries: %v", err)
	}
	if !containsLibrary(libs, lib.ID) {
		t.Errorf("created library %s not in Libraries() listing", lib.ID)
	}
}

func TestIntegrationServerReads(t *testing.T) {
	ctx := context.Background()

	if _, err := itClient.Tags(ctx); err != nil {
		t.Errorf("Tags: %v", err)
	}
	if _, err := itClient.Genres(ctx); err != nil {
		t.Errorf("Genres: %v", err)
	}
	if _, err := itClient.NotificationSettings(ctx); err != nil {
		t.Errorf("NotificationSettings: %v", err)
	}
	if _, err := itClient.Filesystem(ctx); err != nil {
		t.Errorf("Filesystem: %v", err)
	}
}

func uniqueSuffix() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func containsUser(users []User, id string) bool {
	for i := range users {
		if users[i].ID == id {
			return true
		}
	}
	return false
}

func containsLibrary(libs []Library, id string) bool {
	for i := range libs {
		if libs[i].ID == id {
			return true
		}
	}
	return false
}
