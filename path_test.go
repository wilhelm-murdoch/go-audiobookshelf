package audiobookshelf

import "testing"

func TestPathBuilderBasic(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"api root only", apiPath().String(), "/api"},
		{"api single root", apiPath("libraries").String(), "/api/libraries"},
		{"api multi root", apiPath("items", "batch", "delete").String(), "/api/items/batch/delete"},
		{"raw root", rawPath("/login").String(), "/login"},
		{"seg", apiPath("libraries").Seg("lib_1").String(), "/api/libraries/lib_1"},
		{"seg then lit", apiPath("libraries").Seg("lib_1").Lit("items").String(), "/api/libraries/lib_1/items"},
		{"multi seg", apiPath("me", "progress").Seg("li_1", "ep_1").String(), "/api/me/progress/li_1/ep_1"},
		{"multi lit", apiPath("playlists").Seg("pl_1").Lit("batch", "add").String(), "/api/playlists/pl_1/batch/add"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestPathBuilderSegEscapes(t *testing.T) {
	got := apiPath("tags").Seg("The Best/Worst").String()
	want := "/api/tags/The%20Best%2FWorst"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPathBuilderSegSkipsEmpty(t *testing.T) {
	// An empty segment is dropped, so optional IDs can be passed
	// unconditionally.
	got := apiPath("me", "progress").Seg("li_1", "").String()
	want := "/api/me/progress/li_1"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	got = apiPath("playlists").Seg("pl_1").Lit("item").Seg("li_1", "").String()
	want = "/api/playlists/pl_1/item/li_1"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPathBuilderLitSkipsEmpty(t *testing.T) {
	got := apiPath("items").Seg("li_1").Lit("", "tone-scan").String()
	want := "/api/items/li_1/tone-scan"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPathBuilderFlag(t *testing.T) {
	if got := apiPath("items").Seg("li_1").Flag("hard", true).String(); got != "/api/items/li_1?hard=1" {
		t.Errorf("flag on: got %q", got)
	}

	if got := apiPath("items").Seg("li_1").Flag("hard", false).String(); got != "/api/items/li_1" {
		t.Errorf("flag off: got %q", got)
	}
}

func TestPathBuilderSetAndQueryEncoding(t *testing.T) {
	got := apiPath("search", "books").Set("title", "a b").Set("author", "c&d").String()
	// Encode sorts keys: author before title.
	want := "/api/search/books?author=c%26d&title=a+b"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPathBuilderQueryMerge(t *testing.T) {
	q := (&PageParams{Limit: 10, Page: 2}).values()
	got := apiPath("libraries").Seg("lib_1").Lit("items").Query(q).String()
	want := "/api/libraries/lib_1/items?limit=10&page=2"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPathBuilderNilQueryNoop(t *testing.T) {
	var p *PageParams // nil receiver; values() must be safe
	got := apiPath("libraries").Seg("lib_1").Lit("items").Query(p.values()).String()
	if got != "/api/libraries/lib_1/items" {
		t.Errorf("got %q", got)
	}
}
