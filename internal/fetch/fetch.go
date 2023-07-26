// Package fetch provides a cache client that can fetch data from a URL.
package fetch

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"git.sr.ht/~jamesponddotco/httpx-go"
	"git.sr.ht/~jamesponddotco/imgdiet-go"
	"git.sr.ht/~jamesponddotco/privytar/internal/meta"
	"git.sr.ht/~jamesponddotco/xstd-go/xerrors"
	"golang.org/x/time/rate"
)

// ErrFetchData is returned when the client fails to fetch data from a URL.
const ErrFetchData xerrors.Error = "failed to fetch data"

// Client represents a client that can fetch data from a URL.
type Client struct {
	// httpc is the underlying HTTP client used to fetch data.
	httpc *httpx.Client
}

// New creates a new client that can fetch data from a URL.
func New(serviceName, serviceEmail string) *Client {
	return &Client{
		httpc: &httpx.Client{
			RateLimiter: rate.NewLimiter(rate.Limit(2), 1),
			RetryPolicy: httpx.DefaultRetryPolicy(),
			UserAgent: &httpx.UserAgent{
				Token:   serviceName,
				Version: meta.Version,
				Comment: []string{serviceEmail},
			},
			Cache: nil,
		},
	}
}

// Remote fetches data from a URL, optimizes it to reduce its size, and returns
// it as a byte slice.
func (c *Client) Remote(ctx context.Context, uri string) ([]byte, error) {
	resp, err := c.httpc.Get(ctx, uri)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFetchData, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %s", ErrFetchData, resp.Status)
	}

	imgdiet.Start(nil)
	defer imgdiet.Stop()

	data, err := imgdiet.Open(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFetchData, err)
	}
	defer data.Close()

	image, err := data.Optimize(imgdiet.DefaultOptions())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFetchData, err)
	}

	if data.Saved() < resp.ContentLength {
		return image, nil
	}

	image, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFetchData, err)
	}

	return image, nil
}
