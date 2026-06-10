package audiobookshelf

import "time"

// Millis is a Unix timestamp in milliseconds since the epoch, the form
// the Audiobookshelf API uses for every timestamp. It is a plain int64
// and marshals to and from JSON as a number, so it is a drop-in for the
// raw numeric fields it replaces. The API uses 0 to mean "unset".
type Millis int64

// MillisFromTime converts t to milliseconds since the Unix epoch. A zero
// t maps to 0 (the API's "unset" value).
func MillisFromTime(t time.Time) Millis {
	if t.IsZero() {
		return 0
	}

	return Millis(t.UnixMilli())
}

// Time returns the timestamp as a time.Time in UTC. A zero (unset) Millis
// returns the zero time.Time; test for it with IsZero.
func (m Millis) Time() time.Time {
	if m == 0 {
		return time.Time{}
	}

	return time.UnixMilli(int64(m)).UTC()
}

// IsZero reports whether the timestamp is unset.
func (m Millis) IsZero() bool { return m == 0 }

// Seconds is a duration in (possibly fractional) seconds, the form the
// Audiobookshelf API uses for media durations and playback positions. It
// is a plain float64 and marshals to and from JSON as a number.
type Seconds float64

// SecondsFromDuration converts d to fractional seconds.
func SecondsFromDuration(d time.Duration) Seconds {
	return Seconds(d.Seconds())
}

// Duration returns the value as a time.Duration, rounding to the nearest
// nanosecond.
func (s Seconds) Duration() time.Duration {
	return time.Duration(float64(s) * float64(time.Second))
}
