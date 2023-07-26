// Package cache provides a simple in-memory backed cache for storing and
// retrieving images from Gravatar.
package cache

import (
	"container/list"
	"fmt"
	"sync"
	"time"

	"git.sr.ht/~jamesponddotco/privytar/internal/timeutil"
	"git.sr.ht/~jamesponddotco/xstd-go/xerrors"
)

const (
	// ErrTypeAssertion is returned when a type assertion fails.
	ErrTypeAssertion xerrors.Error = "type assertion failed"

	// ErrKeyNotFound is returned when a key is not found in the cache.
	ErrKeyNotFound xerrors.Error = "key not found"

	// ErrKeyExpired is returned when a key has expired in the cache.
	ErrKeyExpired xerrors.Error = "key expired"
)

// Entry represents an entry in the cache.
type Entry struct {
	// timestamp is the time the entry was added to the cache or last accessed.
	timestamp time.Time

	// key is the cache key for the entry.
	key string

	// value is the value of the entry.
	value []byte
}

// Cache represents an in-memory LRU cache for Gravatar images.
type Cache struct {
	// entries is a map of cache keys to cache entries.
	entries map[string]*list.Element

	// list is a doubly linked list of cache entries.
	list *list.List

	// capacity is the maximum number of entries the cache can hold.
	capacity uint

	// expiration is the expiration time for cache entries.
	expiration timeutil.CacheDuration

	// mu is a read-write mutex to ensure thread safety.
	mu sync.RWMutex
}

// New creates a new cache with the given capacity and expiration time.
func New(capacity uint, expiration timeutil.CacheDuration) *Cache {
	return &Cache{
		entries:    make(map[string]*list.Element, capacity),
		list:       list.New(),
		capacity:   capacity,
		expiration: expiration,
	}
}

// Get retrieves the value for the given key from the cache.
func (c *Cache) Get(key string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if element, ok := c.entries[key]; ok {
		item, ok := element.Value.(*Entry)
		if !ok {
			return nil, fmt.Errorf("%w: GET %s", ErrTypeAssertion, key)
		}

		if time.Now().After(item.timestamp.Add(c.expiration.Duration)) {
			c.mu.RUnlock()

			if err := c.Delete(key); err != nil {
				return nil, err
			}

			c.mu.RLock()

			return nil, fmt.Errorf("%w: GET %s", ErrKeyExpired, key)
		}

		c.list.MoveToFront(element)

		return item.value, nil
	}

	return nil, fmt.Errorf("%w: GET %s", ErrKeyNotFound, key)
}

// Set sets the value for the given key in the cache.
func (c *Cache) Set(key string, value []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, ok := c.entries[key]; ok {
		c.list.MoveToFront(element)

		item, ok := element.Value.(*Entry)
		if !ok {
			return fmt.Errorf("%w: SET %s", ErrTypeAssertion, key)
		}

		item.value = value
		item.timestamp = time.Now()

		return nil
	}

	if c.list.Len() == int(c.capacity) {
		element := c.list.Back()

		c.list.Remove(element)

		item, ok := element.Value.(*Entry)
		if !ok {
			return fmt.Errorf("%w: SET %s", ErrTypeAssertion, key)
		}

		delete(c.entries, item.key)
	}

	item := &Entry{
		timestamp: time.Now(),
		key:       key,
		value:     value,
	}

	element := c.list.PushFront(item)

	c.entries[key] = element

	return nil
}

// Delete removes the entry with the given key from the cache.
func (c *Cache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, ok := c.entries[key]; ok {
		item, ok := element.Value.(*Entry)
		if !ok {
			return fmt.Errorf("%w: DELETE %s", ErrTypeAssertion, key)
		}

		c.list.Remove(element)

		delete(c.entries, item.key)
	}

	return nil
}
