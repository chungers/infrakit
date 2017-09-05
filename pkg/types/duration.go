package types

import (
	"sort"
	"strings"
	"time"
)

// Duration is a wrapper for time.Duration that adds JSON parsing
type Duration time.Duration

func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

// DurationSorter is a wrapper to make Duration sortable
type durationSorter []Duration

func (d durationSorter) Len() int           { return len(d) }
func (d durationSorter) Less(i, j int) bool { return time.Duration(d[i]) < time.Duration(d[j]) }
func (d durationSorter) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }

// SortDurations sorts the durations
func SortDurations(d []Duration) {
	sort.Sort(durationSorter(d))
}

// DurationFromString returns either the parsed duration or a default value.
func DurationFromString(s string, defaultValue time.Duration) Duration {
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return Duration(defaultValue)
	}
	return Duration(parsed)
}

// MarshalJSON returns the json representation
func (d Duration) MarshalJSON() ([]byte, error) {
	return []byte("\"" + time.Duration(d).String() + "\""), nil
}

// UnmarshalJSON unmarshals the buffer to this struct
func (d *Duration) UnmarshalJSON(buff []byte) error {
	parsed, err := time.ParseDuration(strings.Trim(string(buff), "\""))
	if err != nil {
		return err
	}
	*d = Duration(parsed)
	return nil
}
