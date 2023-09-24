package cache

import (
	"errors"
	"testing"
	"time"

	"git.sr.ht/~jamesponddotco/privytar/internal/timeutil"
)

func TestCache_Get_TypeAssertionError(t *testing.T) {
	t.Parallel()

	c := New(2, timeutil.CacheDuration{Duration: 1 * time.Hour})

	// Deliberately insert incorrect type into the internal list.
	c.entries["badKey"] = c.list.PushFront("badType")

	// When we attempt to get this key, we should hit the type assertion error.
	_, err := c.Get("badKey")
	if !errors.Is(err, ErrTypeAssertion) {
		t.Errorf("Expected error: %v, got: %v", ErrTypeAssertion, err)
	}
}

func TestCache_Set_TypeAssertionError(t *testing.T) {
	t.Parallel()

	c := New(2, timeutil.CacheDuration{Duration: 1 * time.Hour})

	c.entries["badKey"] = c.list.PushFront("badType")

	err := c.Set("badKey", []byte("value"))
	if !errors.Is(err, ErrTypeAssertion) {
		t.Errorf("Expected error: %v, got: %v", ErrTypeAssertion, err)
	}

	err = c.Set("key1", []byte("value1"))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Make "key1" the most recently used item by accessing it.
	if _, err = c.Get("key1"); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Add another entry to trigger eviction of the least recently used item
	// (badKey) This should trigger the second type assertion error.
	err = c.Set("key2", []byte("value2"))
	if !errors.Is(err, ErrTypeAssertion) {
		t.Errorf("Expected error: %v, got: %v", ErrTypeAssertion, err)
	}
}

func TestCache_Delete_TypeAssertionError(t *testing.T) {
	t.Parallel()

	c := New(2, timeutil.CacheDuration{Duration: 1 * time.Hour})

	c.entries["badKey"] = c.list.PushFront("badType")

	err := c.Delete("badKey")
	if !errors.Is(err, ErrTypeAssertion) {
		t.Errorf("Expected error: %v, got: %v", ErrTypeAssertion, err)
	}
}
