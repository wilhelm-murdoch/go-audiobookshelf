package audiobookshelf

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"testing"
)

func TestLibrariesAndItems(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/libraries":
			encoder := json.NewEncoder(w)

			err := encoder.Encode(map[string]any{
				"libraries": []map[string]any{
					{"id": "lib_1", "name": "Main", "mediaType": "book"},
				},
			})
			if err != nil {
				t.Errorf("encoding libraries: %v", err)
			}
		case "/api/libraries/lib_1/items":
			q := r.URL.Query()
			if q.Get("limit") != "10" || q.Get("minified") != "1" || q.Get("sort") != "media.metadata.title" {
				t.Errorf("unexpected query: %s", r.URL.RawQuery)
			}

			encoder := json.NewEncoder(w)

			err := encoder.Encode(map[string]any{
				"results": []map[string]any{
					{"id": "li_1", "mediaType": "book", "media": map[string]any{
						"metadata": map[string]any{"title": "Wizards First Rule"},
					}},
				},
				"total": 1,
				"limit": 10,
				"page":  0,
			})
			if err != nil {
				t.Errorf("encoding results: %v", err)
			}
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	})

	ctx := context.Background()
	libraries, err := client.Libraries(ctx)
	if err != nil {
		t.Fatalf("Libraries: %v", err)
	}

	if len(libraries) != 1 || libraries[0].Name != "Main" {
		t.Fatalf("libraries = %+v", libraries)
	}

	// The handle carries the client, so chained calls work.
	page, err := libraries[0].Items(ctx, &LibraryItemListParams{
		Limit:    10,
		Sort:     "media.metadata.title",
		Minified: true,
	})
	if err != nil {
		t.Fatalf("Items: %v", err)
	}

	if page.Total != 1 || len(page.Results) != 1 {
		t.Fatalf("page = %+v", page)
	}

	item := page.Results[0]
	if item.Media == nil || item.Media.Metadata.Title != "Wizards First Rule" {
		t.Errorf("item = %+v", item)
	}

	if item.client != client {
		t.Error("item is missing its client handle")
	}
}

func TestDeleteTagEncodesBase64(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte("The Best"))
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		want := "/api/tags/" + encoded
		if r.URL.EscapedPath() != want {
			t.Errorf("path = %s, want %s", r.URL.EscapedPath(), want)
		}

		encoder := json.NewEncoder(w)

		err := encoder.Encode(map[string]int{"numItemsUpdated": 2})
		if err != nil {
			t.Errorf("encoding numItemsUpdated: %v", err)
		}
	})

	n, err := client.DeleteTag(context.Background(), "The Best")
	if err != nil {
		t.Fatalf("DeleteTag: %v", err)
	}

	if n != 2 {
		t.Errorf("numItemsUpdated = %d, want 2", n)
	}
}

func TestMyMediaProgressPaths(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/me/progress/li_1/ep_1":
			encoder := json.NewEncoder(w)

			err := encoder.Encode(map[string]any{
				"id": "li_1-ep_1", "libraryItemId": "li_1", "episodeId": "ep_1", "progress": 0.5,
			})
			if err != nil {
				t.Errorf("encoding progress: %v", err)
			}
		case r.Method == http.MethodPatch && r.URL.Path == "/api/me/progress/li_1":
			var body map[string]any

			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&body); err != nil {
				t.Errorf("decoding body: %v", err)
			}

			if body["isFinished"] != true {
				t.Errorf("body = %v", body)
			}

			w.WriteHeader(http.StatusOK)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	})

	ctx := context.Background()

	progress, err := client.MyMediaProgress(ctx, "li_1", "ep_1")
	if err != nil {
		t.Fatalf("MyMediaProgress: %v", err)
	}

	if progress.Progress != 0.5 {
		t.Errorf("progress = %v", progress.Progress)
	}

	finished := true
	if err = client.UpdateMyMediaProgress(ctx, "li_1", "", &MediaProgressUpdate{IsFinished: &finished}); err != nil {
		t.Fatalf("UpdateMyMediaProgress: %v", err)
	}
}
