package audiobookshelf

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestErrorHelpers(t *testing.T) {
	tests := []struct {
		code        int
		notFound    bool
		unauthrized bool
		forbidden   bool
		badRequest  bool
	}{
		{code: http.StatusNotFound, notFound: true},
		{code: http.StatusUnauthorized, unauthrized: true},
		{code: http.StatusForbidden, forbidden: true},
		{code: http.StatusBadRequest, badRequest: true},
	}

	for _, tt := range tests {
		err := &Error{StatusCode: tt.code}
		if IsNotFound(err) != tt.notFound {
			t.Errorf("IsNotFound(%d) = %v", tt.code, IsNotFound(err))
		}
		if IsUnauthorized(err) != tt.unauthrized {
			t.Errorf("IsUnauthorized(%d) = %v", tt.code, IsUnauthorized(err))
		}
		if IsForbidden(err) != tt.forbidden {
			t.Errorf("IsForbidden(%d) = %v", tt.code, IsForbidden(err))
		}
		if IsBadRequest(err) != tt.badRequest {
			t.Errorf("IsBadRequest(%d) = %v", tt.code, IsBadRequest(err))
		}
	}

	// A non-API error matches none of the helpers.
	if IsNotFound(errors.New("boom")) {
		t.Error("IsNotFound(non-API) = true")
	}
}

func TestErrorMessageFallsBackToStatusText(t *testing.T) {
	err := &Error{Prefix: "audiobookshelf", Method: "GET", Path: "/x", StatusCode: http.StatusNotFound}
	if got, want := err.Error(), "audiobookshelf: GET /x: 404 Not Found"; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}

	withMsg := &Error{Prefix: "audiobookshelf", Method: "GET", Path: "/x", StatusCode: 500, Message: "kaboom"}
	if got, want := withMsg.Error(), "audiobookshelf: GET /x: 500 kaboom"; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestGenericVerbs(t *testing.T) {
	var cap captured
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		cap.method = r.Method
		cap.path = r.URL.Path
		if _, err := w.Write([]byte(`{"ok":true}`)); err != nil {
			t.Errorf("write: %v", err)
		}
	})

	ctx := context.Background()
	var out map[string]any

	if err := client.Put(ctx, "/api/thing", map[string]string{"a": "b"}, &out); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if cap.method != http.MethodPut || cap.path != "/api/thing" {
		t.Errorf("Put recorded %s %s", cap.method, cap.path)
	}
	if out["ok"] != true {
		t.Errorf("Put decoded %v", out)
	}
}

func TestClientOptions(t *testing.T) {
	base := NewClient("https://abs.example.com/", WithToken("tok"))

	if base.BaseURL() != "https://abs.example.com" {
		t.Errorf("BaseURL trailing slash not trimmed: %q", base.BaseURL())
	}
	if base.Token() != "tok" {
		t.Errorf("Token = %q", base.Token())
	}

	base.SetToken("tok2")
	if base.Token() != "tok2" {
		t.Errorf("SetToken not applied: %q", base.Token())
	}
}

// TestUserAgentAndTimeoutOptions checks the configuration options take
// effect on the wire and that a request still succeeds with them set.
func TestUserAgentAndTimeoutOptions(t *testing.T) {
	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	client := NewClient(srv.URL,
		WithUserAgent("custom/1.0"),
		WithTimeout(5*time.Second),
	)

	if err := client.Ping(context.Background()); err != nil {
		t.Fatalf("Ping: %v", err)
	}

	if gotUA != "custom/1.0" {
		t.Errorf("User-Agent = %q, want custom/1.0", gotUA)
	}
}

func TestWithHTTPClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	custom := &http.Client{Timeout: time.Minute}
	client := NewClient(srv.URL, WithHTTPClient(custom))

	if err := client.Ping(context.Background()); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}
