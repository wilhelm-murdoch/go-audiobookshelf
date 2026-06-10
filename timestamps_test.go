package audiobookshelf

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMillisTime(t *testing.T) {
	m := Millis(1668388200000) // 2022-11-14T01:10:00Z
	got := m.Time()

	want := time.Date(2022, 11, 14, 1, 10, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("Time() = %s, want %s", got, want)
	}

	if got.Location() != time.UTC {
		t.Errorf("Time() location = %s, want UTC", got.Location())
	}
}

func TestMillisZeroIsUnset(t *testing.T) {
	var m Millis
	if !m.IsZero() {
		t.Error("zero Millis IsZero() = false")
	}

	if !m.Time().IsZero() {
		t.Error("zero Millis Time() is not the zero time")
	}

	if Millis(1).IsZero() {
		t.Error("non-zero Millis IsZero() = true")
	}
}

func TestMillisFromTimeRoundTrip(t *testing.T) {
	want := time.Date(2023, 6, 1, 12, 0, 0, 0, time.UTC)
	if got := MillisFromTime(want).Time(); !got.Equal(want) {
		t.Errorf("round trip = %s, want %s", got, want)
	}

	if MillisFromTime(time.Time{}) != 0 {
		t.Error("MillisFromTime(zero) != 0")
	}
}

func TestSecondsDuration(t *testing.T) {
	if got := Seconds(1.5).Duration(); got != 1500*time.Millisecond {
		t.Errorf("Duration() = %s, want 1.5s", got)
	}

	if got := SecondsFromDuration(90 * time.Second); got != 90 {
		t.Errorf("SecondsFromDuration = %v, want 90", got)
	}
}

// TestMillisSecondsJSON confirms the named types stay wire-compatible:
// they marshal as plain JSON numbers and decode from them.
func TestMillisSecondsJSON(t *testing.T) {
	type sample struct {
		At  Millis  `json:"at"`
		Dur Seconds `json:"dur"`
	}

	out, err := json.Marshal(sample{At: 1668388200000, Dur: 3.5})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if string(out) != `{"at":1668388200000,"dur":3.5}` {
		t.Errorf("marshal = %s", out)
	}

	var got sample
	if err := json.Unmarshal([]byte(`{"at":1668388200000,"dur":3.5}`), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.At != 1668388200000 || got.Dur != 3.5 {
		t.Errorf("unmarshal = %+v", got)
	}
}

// TestDecodedTimestampsAreUsable exercises the types on a real decoded
// response: a library item's duration and timestamps.
func TestDecodedTimestampsAreUsable(t *testing.T) {
	var item LibraryItem
	body := `{"id":"li_1","addedAt":1668388200000,"media":{"duration":3661.0}}`
	if err := json.Unmarshal([]byte(body), &item); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if item.AddedAt.Time().Year() != 2022 {
		t.Errorf("AddedAt year = %d", item.AddedAt.Time().Year())
	}

	if item.Media == nil || item.Media.Duration.Duration() != 3661*time.Second {
		t.Errorf("media duration = %v", item.Media)
	}
}
