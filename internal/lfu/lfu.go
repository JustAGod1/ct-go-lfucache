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
	// O(1)
	Get(key K) (V, error)

	// Put updates the value of the key if present, or inserts the key if not already present.
	//
	// When the cache reaches its capacity, it should invalidate and remove the least frequently used key
	// before inserting a new item. For this problem, when there is a tie
	// (i.e., two or more keys with the same frequency), the least recently used key would be invalidated.
	//
	// O(1)
	Put(key K, value V)

	// All returns the iterator in descending order of frequency.
	// If two or more keys have the same frequency, the most recently used key will be listed first.
	//
	// O(capacity)
	All() iter.Seq2[K, V]

	// Size returns the cache size.
	//
	// O(1)
	Size() int

	// Capacity returns the cache capacity.
	//
	// O(1)
	Capacity() int

	// GetKeyFrequency returns the element's frequency if the key exists in the cache,
	// otherwise, returns ErrKeyNotFound.
	//
	// O(1)
	GetKeyFrequency(key K) (int, error)
}

type cacheData[K, V any] struct {
	key   K
	value V
	// link to node in which we are contained
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
	// we need to add this first container for least accessed ones to simplify Put logic
	r.sequence.add(sameFreqContainer[K, V]{freq: 1})

	return r
}

func (l *cacheImpl[K, V]) touch(n *linkedListNode[cacheData[K, V]]) {
	newFreq := n.data.container.data.freq + 1

	n.data.container.data.entries.remove(n)

	var target *linkedListNode[sameFreqContainer[K, V]]
	if n.data.container.next == nil || n.data.container.next.data.freq > newFreq {
		// as container is sorted, and we started with container with our frequency this container is greater than we
		// need for at least 1 so we can insert a new one
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
		// I dunno I'm stupid
		var v V
		return v, ErrKeyNotFound
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

	// here comes deletion
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

// 1. What did the horse say after it tripped?
// Help! I’ve fallen and I can’t giddyup!
//
// 2. Why can’t you hear a pterodactyl going to the bathroom?
//
// Because the “P” is silent.
//
// 3. What do you call a well-balanced horse?
//
// Stable.
//
// 4. What do you call an angry carrot?
//
// A steamed veggie.
//
// 5. Where do polar bears keep their money?
//
// In a snowbank.
//
// 6. How do you make an egg-roll?
//
// You push it!
//
// 7. What would bears be without bees?
//
// Ears.
//
// 8. What do you call a pile of cats?
//
// A meow-ntain.
//
// 9. Why do cows wear bells?
//
// Because their horns don’t work.
//
// 10. Why did the bicycle fall over?
//
// Because it was two tired.
//
// 11. What did the triangle say to the circle?
//
// You’re pointless.
//
// 12. RIP, boiling water.
//
// You will be mist.
//
// 13. Time flies like an arrow.
//
// Fruit flies like a banana.
//
// 14. I ordered a chicken and an egg online.
//
// I’ll let you know what comes first.
//
// 15. Why was Cinderella so bad at soccer?
//
// She kept running away from the ball!
//
// 16. What do lawyers wear to court?
//
// Lawsuits.
