// Package cache provides a simple in-memory backed cache for storing and
// retrieving images from Gravatar.
package cache

import (
	"container/list"
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
	mu sync.Mutex
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
	c.mu.Lock()
	defer c.mu.Unlock()

	element, ok := c.entries[key]
	if !ok {
		return nil, ErrKeyNotFound
	}

	item, ok := element.Value.(*Entry)
	if !ok {
		return nil, ErrTypeAssertion
	}

	now := time.Now()
	if now.After(item.timestamp.Add(c.expiration.Duration)) {
		delete(c.entries, key)

		c.list.Remove(element)

		return nil, ErrKeyExpired
	}

	item.timestamp = now

	c.list.MoveToFront(element)

	return item.value, nil
}

// Set sets the value for the given key in the cache.
func (c *Cache) Set(key string, value []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	// If the key already exists, update the value and timestamp.
	if element, ok := c.entries[key]; ok {
		item, ok := element.Value.(*Entry)
		if !ok {
			return ErrTypeAssertion
		}

		item.value = value
		item.timestamp = now

		c.list.MoveToFront(element)

		return nil
	}

	// Evict items if necessary.
	if c.list.Len() >= int(c.capacity) {
		element := c.list.Back()

		item, ok := element.Value.(*Entry)
		if !ok {
			return ErrTypeAssertion
		}

		delete(c.entries, item.key)

		c.list.Remove(element)
	}

	// Add the new entry.
	item := &Entry{
		timestamp: now,
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

	element, ok := c.entries[key]
	if !ok {
		return ErrKeyNotFound
	}

	item, ok := element.Value.(*Entry)
	if !ok {
		return ErrTypeAssertion
	}

	delete(c.entries, item.key)

	c.list.Remove(element)

	return nil
}
