package audiobookshelf

import (
	"encoding/json"
	"testing"
)

// TestShelfEntities covers the typed accessors that decode a shelf's
// polymorphic entities according to its Type.
func TestShelfEntities(t *testing.T) {
	t.Run("book", func(t *testing.T) {
		s := Shelf{Type: "book", Entities: json.RawMessage(`[{"id":"li_1"},{"id":"li_2"}]`)}
		items, err := s.LibraryItemEntities()
		if err != nil {
			t.Fatalf("LibraryItemEntities: %v", err)
		}
		if len(items) != 2 || items[0].ID != "li_1" {
			t.Errorf("items = %+v", items)
		}
	})

	t.Run("series", func(t *testing.T) {
		s := Shelf{Type: "series", Entities: json.RawMessage(`[{"id":"se_1","name":"A"}]`)}
		series, err := s.SeriesEntities()
		if err != nil {
			t.Fatalf("SeriesEntities: %v", err)
		}
		if len(series) != 1 || series[0].Name != "A" {
			t.Errorf("series = %+v", series)
		}
	})

	t.Run("authors", func(t *testing.T) {
		s := Shelf{Type: "authors", Entities: json.RawMessage(`[{"id":"au_1","name":"Herbert"}]`)}
		authors, err := s.AuthorEntities()
		if err != nil {
			t.Fatalf("AuthorEntities: %v", err)
		}
		if len(authors) != 1 || authors[0].Name != "Herbert" {
			t.Errorf("authors = %+v", authors)
		}
	})
}
