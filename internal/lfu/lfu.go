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

// Just, you know, a bidirectional linked list
type linkedListNode[Data any] struct {
	data Data
	next *linkedListNode[Data]
	prev *linkedListNode[Data]
}

type linkedList[Data any] struct {
	head *linkedListNode[Data]
	tail *linkedListNode[Data]
}

func (l *linkedList[Data]) add(data Data) *linkedListNode[Data] {
	node := new(linkedListNode[Data])
	node.data = data

	return l.addNode(node)
}

func (l *linkedList[Data]) addNode(node *linkedListNode[Data]) *linkedListNode[Data] {
	node.prev = l.tail
	node.next = nil
	if l.tail == nil {
		l.head = node
	} else {
		l.tail.next = node
	}
	l.tail = node

	return node
}

func (l *linkedList[Data]) insertAfter(pivot *linkedListNode[Data], data Data) *linkedListNode[Data] {
	pNext := pivot.next

	node := new(linkedListNode[Data])
	node.data = data

	pivot.next = node
	node.prev = pivot
	node.next = pNext
	if pNext != nil {
		pNext.prev = node
	} else {
		l.tail = node
	}

	return node
}

func (l *linkedList[Data]) isEmpty() bool {
	return l.head == nil
}

func (l *linkedList[Data]) remove(node *linkedListNode[Data]) {
	if l.head == node && l.tail == node {
		l.head = nil
		l.tail = nil
		return
	}

	nPrev := node.prev
	nNext := node.next

	if nPrev != nil {
		nPrev.next = nNext
	} else {
		l.head = nNext
	}

	if nNext != nil {
		nNext.prev = nPrev
	} else {
		l.tail = nPrev
	}

}

type cacheData[K, V any] struct {
	key       K
	value     V
	container *linkedListNode[sameFreqContainer[K, V]]
}

type sameFreqContainer[K, V any] struct {
	freq    int
	entries linkedList[cacheData[K, V]]
}

// So we have some non-trivial logic here
//
//
// ┌──────────────────────────────────────────────────────────────────────────────┐
// │ Frequency sorted in ascending order                                          │
// │ ┌───────────────────────────────┐          ┌───────────────────────────────┐ │
// │ │  sameFreqContainer:           │          │  sameFreqContainer:           │ │
// │ │  - freq: 1                    │          │  - freq: 3                    │ │
// │ │  - entries:                   │          │  - entries:                   │ │
// │ │  ┌─────────────────────────┐  │          │  ┌─────────────────────────┐  │ │
// │ │  │  Used long time ago     │  │          │  │                         │  │ │
// │ │  │  ┌───────────────────┐  │  │          │  │  ┌───────────────────┐  │  │ │
// │ │  │  │ cacheData:        │  │  │          │  │  │ cacheData:        │  │  │ │
// │ │  │  │ - key             │  │  │          │  │  │ - key             │  │  │ │
// │ │  │  │ - value           │  │  │          │  │  │ - value           │  │  │ │
// │ │  │  └────────┬──────────┘  │  │          │  │  └────────┬──────────┘  │  │ │
// │ │  │           │             │  ├─────────►│  │           │             │  │ │
// │ │  │   Recently│used         │  │          │  │           │             │  │ │
// │ │  │  ┌────────▼──────────┐  │  │          │  │  ┌────────▼──────────┐  │  │ │
// │ │  │  │ cacheData:        │  │  │          │  │  │ cacheData:        │  │  │ │
// │ │  │  │ - key             │  │  │          │  │  │ - key             │  │  │ │
// │ │  │  │ - value           │  │  │          │  │  │ - value           │  │  │ │
// │ │  │  └───────────────────┘  │  │          │  │  └───────────────────┘  │  │ │
// │ │  └─────────────────────────┘  │          │  └─────────────────────────┘  │ │
// │ └───────────────────────────────┘          └───────────────────────────────┘ │
// └──────────────────────────────────────────────────────────────────────────────┘
//
// So yeah we have a linked list with linked lists.

// cacheImpl represents LFU cache implementation
type cacheImpl[K comparable, V any] struct {
	index    map[K]*linkedListNode[cacheData[K, V]]
	sequence linkedList[sameFreqContainer[K, V]]
	capacity int
}

// New initializes the cache with the given capacity.
// If no capacity is provided, the cache will use DefaultCapacity.
func New[K comparable, V any](capacity ...int) *cacheImpl[K, V] {
	r := new(cacheImpl[K, V])
	if len(capacity) > 1 {
		panic("wtf")
	}
	if len(capacity) == 1 {
		r.capacity = capacity[0]
	} else {
		r.capacity = DefaultCapacity
	}
	if r.capacity < 0 {
		panic("negative capacity")
	}
	r.index = make(map[K]*linkedListNode[cacheData[K, V]], r.capacity)
	r.sequence.add(sameFreqContainer[K, V]{freq: 1})

	return r
}

func (l *cacheImpl[K, V]) touch(n *linkedListNode[cacheData[K, V]]) {
	newFreq := n.data.container.data.freq + 1

	n.data.container.data.entries.remove(n)

	var target *linkedListNode[sameFreqContainer[K, V]]
	if n.data.container.next == nil || n.data.container.next.data.freq > newFreq {
		target = l.sequence.insertAfter(n.data.container, sameFreqContainer[K, V]{freq: newFreq})
	} else {
		// Here we found a node with exactly our target freq because our list is sorted in ascending order.
		target = n.data.container.next
	}

	target.data.entries.addNode(n)

	// We don't delete node with freq 1 because all new elements goes there
	if n.data.container.data.entries.isEmpty() && n.data.container.data.freq > 1 {
		l.sequence.remove(n.data.container)
	}

	n.data.container = target
}
func (l *cacheImpl[K, V]) Get(key K) (V, error) {
	n, ok := l.index[key]
	if !ok {
		return (make(map[K]V))[key], ErrKeyNotFound
	}
	l.touch(n)

	return n.data.value, nil
}

func (l *cacheImpl[K, V]) Put(key K, value V) {
	n, ok := l.index[key]
	if ok {
		l.touch(n)
		n.data.value = value
		return
	}

	if l.Size()+1 > l.Capacity() {
		cur := l.sequence.head

		// actually only first node can be empty but nah
		for cur.data.entries.isEmpty() {
			cur = cur.next
		}
		head := cur.data.entries.head

		// head is always the oldest one
		cur.data.entries.remove(head)
		delete(l.index, head.data.key)
	}

	node := l.sequence.head.data.entries.add(cacheData[K, V]{key, value, l.sequence.head})
	l.index[key] = node
}

func (l *cacheImpl[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		curFreq := l.sequence.tail
		for curFreq != nil {
			curEntry := curFreq.data.entries.tail
			for curEntry != nil {
				if !yield(curEntry.data.key, curEntry.data.value) {
					return
				}
				curEntry = curEntry.prev
			}
			curFreq = curFreq.prev
		}
	}
}

func (l *cacheImpl[K, V]) Size() int {
	return len(l.index)
}

func (l *cacheImpl[K, V]) Capacity() int {
	return l.capacity
}

func (l *cacheImpl[K, V]) GetKeyFrequency(key K) (int, error) {
	n, ok := l.index[key]
	if !ok {
		return 0, ErrKeyNotFound
	}

	return n.data.container.data.freq, nil
}
