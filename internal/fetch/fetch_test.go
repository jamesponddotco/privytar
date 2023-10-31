package fetch_test

import (
	"context"
	"errors"
	"testing"

	"git.sr.ht/~jamesponddotco/privytar/internal/fetch"
)

func TestClient_Remote(t *testing.T) {
	t.Parallel()

	client := fetch.New("TestService", "test@example.com")

	tests := []struct {
		name          string
		uri           string
		expectedError error
	}{
		{
			name: "Successful fetch",
			uri:  "https://s.gravatar.com/avatar/c9fb2194c5e620c85b10840bc63121fd984ed86e91bc819d5dad7baf1168e5c1",
		},
		{
			name:          "Unsuccessful fetch",
			uri:           "https://s.gravatar.com/avat",
			expectedError: fetch.ErrFetchData,
		},
		{
			name:          "Invalid URL",
			uri:           "://invalid.url",
			expectedError: fetch.ErrFetchData,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			data, err := client.Remote(ctx, tt.uri)

			if tt.expectedError != nil {
				if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}

				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if len(data) == 0 {
				t.Errorf("expected non-empty data, got empty")
			}
		})
	}
}
