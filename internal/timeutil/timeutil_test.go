package timeutil_test

import (
	"bytes"
	"testing"
	"time"

	"git.sr.ht/~jamesponddotco/privytar/internal/timeutil"
)

func TestCacheDuration_MarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   timeutil.CacheDuration
		want    []byte
		wantErr bool
	}{
		{
			name:    "valid duration",
			input:   timeutil.CacheDuration{Duration: 5 * time.Minute},
			want:    []byte(`"5m0s"`),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.input.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Fatalf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !bytes.Equal(got, tt.want) {
				t.Fatalf("MarshalJSON() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}

func TestCacheDuration_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   []byte
		want    timeutil.CacheDuration
		wantErr bool
	}{
		{
			name:    "valid duration string",
			input:   []byte(`"5m0s"`),
			want:    timeutil.CacheDuration{Duration: 5 * time.Minute},
			wantErr: false,
		},
		{
			name:    "invalid duration string",
			input:   []byte(`"5x"`),
			wantErr: true,
		},
		{
			name:    "invalid json",
			input:   []byte(`5m0s`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var got timeutil.CacheDuration

			err := got.UnmarshalJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			if got != tt.want {
				t.Fatalf("UnmarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}
