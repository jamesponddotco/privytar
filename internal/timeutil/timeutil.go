// Package timeutil provides utility functions and types for time.
package timeutil

import (
	"encoding/json"
	"fmt"
	"time"
)

// CacheDuration is a wrapper around time.Duration that is used to indicate that
// a value is a cache duration.
type CacheDuration struct {
	time.Duration
}

// MarshalJSON implements the json.Marshaler interface.
func (d CacheDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String()) //nolint:wrapcheck // safe to ignore
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (d *CacheDuration) UnmarshalJSON(b []byte) error {
	var s string

	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("failed to unmarshal cache duration: %w", err)
	}

	dur, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("failed to parse cache duration: %w", err)
	}

	d.Duration = dur

	return nil
}
