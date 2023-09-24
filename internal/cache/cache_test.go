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
		name          string
		capacity      uint
		expiration    timeutil.CacheDuration
		initialKeys   map[string][]byte
		setKey        string
		setValue      []byte
		expectErr     bool
		expectedValue []byte
		evictedKey    string
	}{
		{
			name:          "Add new entry",
			capacity:      5,
			expiration:    timeutil.CacheDuration{Duration: 5 * time.Minute},
			initialKeys:   map[string][]byte{},
			setKey:        "key",
			setValue:      []byte("value"),
			expectedValue: []byte("value"),
		},
		{
			name:          "Overwrite existing entry",
			capacity:      5,
			expiration:    timeutil.CacheDuration{Duration: 5 * time.Minute},
			initialKeys:   map[string][]byte{"key": []byte("oldValue")},
			setKey:        "key",
			setValue:      []byte("newValue"),
			expectedValue: []byte("newValue"),
		},
		{
			name:       "Cache overcapacity",
			capacity:   2,
			expiration: timeutil.CacheDuration{Duration: 5 * time.Minute},
			initialKeys: map[string][]byte{
				"key1": []byte("val1"),
				"key2": []byte("val2"),
			},
			setKey:        "key3",
			setValue:      []byte("val3"),
			expectedValue: []byte("val3"), // Expected value for key3 after Set.
			evictedKey:    "key1",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := cache.New(tt.capacity, tt.expiration)

			// Preload initial keys.
			for key, value := range tt.initialKeys {
				if err := c.Set(key, value); err != nil {
					t.Fatalf("Setup error: %v", err)
				}
			}

			// Call Set method.
			if err := c.Set(tt.setKey, tt.setValue); (err != nil) != tt.expectErr {
				t.Errorf("Cache.Set() error = %v, expectErr %v", err, tt.expectErr)
			}

			// Verify set value.
			value, err := c.Get(tt.setKey)
			if err != nil {
				t.Errorf("Failed to retrieve set key: %v", err)
			}
			if !bytes.Equal(value, tt.expectedValue) {
				t.Errorf("Cache.Get() = %v, expected %v", value, tt.expectedValue)
			}

			// Check if any key was evicted.
			if tt.evictedKey != "" {
				_, err := c.Get(tt.evictedKey)
				if err == nil || !errors.Is(err, cache.ErrKeyNotFound) {
					t.Errorf("Expected key %s to be evicted", tt.evictedKey)
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
		preloadKeys    map[string][]byte
		keyToDelete    string
		expectedError  error
		expectKeyAfter bool
	}{
		{
			name:        "Delete existing key",
			capacity:    2,
			expiration:  timeutil.CacheDuration{Duration: 1 * time.Hour},
			preloadKeys: map[string][]byte{"foo": []byte("bar")},
			keyToDelete: "foo",
		},
		{
			name:           "Delete non-existing key",
			capacity:       2,
			expiration:     timeutil.CacheDuration{Duration: 1 * time.Hour},
			preloadKeys:    map[string][]byte{},
			keyToDelete:    "foo",
			expectedError:  cache.ErrKeyNotFound,
			expectKeyAfter: false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := cache.New(tt.capacity, tt.expiration)
			for k, v := range tt.preloadKeys {
				if err := c.Set(k, v); err != nil {
					t.Fatalf("Setup error: %v", err)
				}
			}

			err := c.Delete(tt.keyToDelete)
			if tt.expectedError != nil {
				if err == nil || !errors.Is(err, tt.expectedError) {
					t.Errorf("Expected error: %v, got: %v", tt.expectedError, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			_, err = c.Get(tt.keyToDelete)
			if tt.expectKeyAfter && errors.Is(err, cache.ErrKeyNotFound) {
				t.Errorf("Expected key %s to be present", tt.keyToDelete)
			} else if !tt.expectKeyAfter && !errors.Is(err, cache.ErrKeyNotFound) {
				t.Errorf("Expected key %s to be absent", tt.keyToDelete)
			}
		})
	}
}
