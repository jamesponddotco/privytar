package cache_test

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"git.sr.ht/~jamesponddotco/privytar/internal/cache"
	"git.sr.ht/~jamesponddotco/privytar/internal/timeutil"
)

func TestCache_Get(t *testing.T) {
	t.Parallel()

	var (
		cacheCapacity = uint(100)
		cacheDuration = timeutil.CacheDuration{
			Duration: 100 * time.Millisecond,
		}
	)

	tests := []struct {
		name        string
		preloadKeys map[string][]byte
		getKey      string
		want        []byte
		wantErr     bool
	}{
		{
			name: "test get valid key",
			preloadKeys: map[string][]byte{
				"validKey": []byte("validValue"),
			},
			getKey:  "validKey",
			want:    []byte("validValue"),
			wantErr: false,
		},
		{
			name: "test get invalid key",
			preloadKeys: map[string][]byte{
				"validKey": []byte("validValue"),
			},
			getKey:  "invalidKey",
			want:    nil,
			wantErr: true,
		},
		{
			name: "test get expired key",
			preloadKeys: map[string][]byte{
				"expiredKey": []byte("expiredValue"),
			},
			getKey:  "expiredKey",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := cache.New(cacheCapacity, cacheDuration)

			for key, value := range tt.preloadKeys {
				_ = c.Set(key, value)
			}

			// Manually expire a key if needed
			if _, ok := tt.preloadKeys["expiredKey"]; ok {
				time.Sleep(cacheDuration.Duration + 10*time.Millisecond)
			}

			got, err := c.Get(tt.getKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("Cache.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !bytes.Equal(got, tt.want) {
				t.Errorf("Cache.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCache_Set(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		capacity   uint
		expiration timeutil.CacheDuration
		entries    map[string][]byte
		key        string
		value      []byte
		wantErr    bool
		expectLRU  string
	}{
		{
			name:       "New entry",
			capacity:   5,
			expiration: timeutil.CacheDuration{Duration: 5 * time.Minute},
			entries:    map[string][]byte{},
			key:        "test",
			value:      []byte("value"),
			wantErr:    false,
			expectLRU:  "",
		},
		{
			name:       "Overwrite existing entry",
			capacity:   5,
			expiration: timeutil.CacheDuration{Duration: 5 * time.Minute},
			entries:    map[string][]byte{"test": []byte("old")},
			key:        "test",
			value:      []byte("value"),
			wantErr:    false,
			expectLRU:  "",
		},
		{
			name:       "Cache overcapacity",
			capacity:   2,
			expiration: timeutil.CacheDuration{Duration: 5 * time.Minute},
			entries:    map[string][]byte{"key1": []byte("val1"), "key2": []byte("val2")},
			key:        "test",
			value:      []byte("value"),
			wantErr:    false,
			expectLRU:  "key1",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := cache.New(tt.capacity, tt.expiration)

			for k, v := range tt.entries {
				c.Set(k, v)
			}

			if err := c.Set(tt.key, tt.value); (err != nil) != tt.wantErr {
				t.Errorf("Cache.Set() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify by Get.
			got, _ := c.Get(tt.key)
			if !bytes.Equal(got, tt.value) {
				t.Errorf("Cache.Set() = %v, want %v", got, tt.value)
			}

			// Check LRU eviction.
			if tt.expectLRU != "" {
				_, err := c.Get(tt.expectLRU)
				if !errors.Is(err, cache.ErrKeyNotFound) {
					t.Errorf("Expected key %s to be evicted", tt.expectLRU)
				}
			}
		})
	}
}

func TestCache_Delete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		capacity       uint
		expiration     timeutil.CacheDuration
		setup          func(*cache.Cache)
		keyToDelete    string
		expectedError  error
		verifyPostCond func(*testing.T, *cache.Cache)
	}{
		{
			name:       "Delete existing key",
			capacity:   2,
			expiration: timeutil.CacheDuration{Duration: 1 * time.Hour},
			setup: func(c *cache.Cache) {
				if err := c.Set("foo", []byte("bar")); err != nil {
					t.Fatalf("Setup error: %v", err)
				}
			},
			keyToDelete: "foo",
			verifyPostCond: func(t *testing.T, c *cache.Cache) {
				t.Helper()

				_, err := c.Get("foo")
				if !errors.Is(err, cache.ErrKeyNotFound) {
					t.Errorf("Expected error: %v, got: %v", cache.ErrKeyNotFound, err)
				}
			},
		},
		{
			name:        "Delete non-existing key",
			capacity:    2,
			expiration:  timeutil.CacheDuration{Duration: 1 * time.Hour},
			keyToDelete: "foo",
			verifyPostCond: func(t *testing.T, c *cache.Cache) {
				t.Helper()

				_, err := c.Get("foo")
				if !errors.Is(err, cache.ErrKeyNotFound) {
					t.Errorf("Expected error: %v, got: %v", cache.ErrKeyNotFound, err)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := cache.New(tt.capacity, tt.expiration)
			if tt.setup != nil {
				tt.setup(c)
			}

			err := c.Delete(tt.keyToDelete)

			if tt.expectedError != nil {
				if err == nil || !errors.Is(err, tt.expectedError) {
					t.Errorf("Expected error: %v, got: %v", tt.expectedError, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if tt.verifyPostCond != nil {
				tt.verifyPostCond(t, c)
			}
		})
	}
}
