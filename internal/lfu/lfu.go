package lfu

import (
	"errors"
	"iter"
)

var ErrKeyNotFound = errors.New("key not found")

const DefaultCapacity = 5

// Cache
// O(capacity) memory
type Cache[K comparable, V any] interface {
	// Get returns the value of the key if the key exists in the cache,
	// otherwise, returns ErrKeyNotFound.
	//
	// O(1), not amortized
	Get(key K) (V, error)

	// Put updates the value of the key if present, or inserts the key if not already present.
	//
	// When the cache reaches its capacity, it should invalidate and remove the least frequently used key
	// before inserting a new item. For this problem, when there is a tie
	// (i.e., two or more keys with the same frequency), the least recently used key would be invalidated.
	//
	// O(1), not amortized
	Put(key K, value V)

	// All returns the iterator in descending order of frequency.
	// If two or more keys have the same frequency, the most recently used key will be listed first.
	//
	// O(capacity)
	All() iter.Seq2[K, V]

	// Size returns the cache size.
	//
	// O(1), not amortized
	Size() int

	// Capacity returns the cache capacity.
	//
	// O(1), not amortized
	Capacity() int

	// GetKeyFrequency returns the element's frequency if the key exists in the cache,
	// otherwise, returns ErrKeyNotFound.
	//
	// O(1), not amortized
	GetKeyFrequency(key K) (int, error)
}

// cacheImpl represents LFU cache implementation
type cacheImpl[K comparable, V any] struct{}

// New initializes the cache with the given capacity.
// If no capacity is provided, the cache will use DefaultCapacity.
func New[K comparable, V any](capacity ...int) *cacheImpl[K, V] {
	return new(cacheImpl[K, V])
}

func (l *cacheImpl[K, V]) Get(key K) (V, error) {
	panic("not implemented")
}

func (l *cacheImpl[K, V]) Put(key K, value V) {
	panic("not implemented")
}

func (l *cacheImpl[K, V]) All() iter.Seq2[K, V] {
	panic("not implemented")
}

func (l *cacheImpl[K, V]) Size() int {
	panic("not implemented")
}

func (l *cacheImpl[K, V]) Capacity() int {
	panic("not implemented")
}

func (l *cacheImpl[K, V]) GetKeyFrequency(key K) (int, error) {
	panic("not implemented")
}
