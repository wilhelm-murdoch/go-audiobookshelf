package audiobookshelf

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWithDebugWiresThroughAndRedacts(t *testing.T) {
	var buf bytes.Buffer

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	client := NewClient(srv.URL, WithToken("test-token"), WithDebug(&buf))
	if err := client.Ping(context.Background()); err != nil {
		t.Fatalf("Ping: %v", err)
	}

	s := buf.String()
	if !strings.Contains(s, ">> GET") || !strings.Contains(s, "/ping") {
		t.Errorf("request not logged:\n%s", s)
	}
	if !strings.Contains(s, "<redacted>") {
		t.Errorf("Authorization not redacted:\n%s", s)
	}
	if strings.Contains(s, "test-token") {
		t.Errorf("token leaked into debug output:\n%s", s)
	}
}
