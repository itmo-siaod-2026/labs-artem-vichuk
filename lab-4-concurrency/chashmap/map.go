package chashmap

import (
	"hash/maphash"
	"sync"
	"sync/atomic"
)

const (
	defaultCapacity = 16
	maxLoadPercent  = 75
)

type Pair[K comparable, V any] struct {
	Key   K
	Value V
}

type HashMap[K comparable, V any] struct {
	resizeMu sync.RWMutex
	seed     maphash.Seed
	size     atomic.Int64
	state    atomic.Pointer[table[K, V]]
}

type table[K comparable, V any] struct {
	buckets []bucket[K, V]
}

type bucket[K comparable, V any] struct {
	mu   sync.RWMutex
	head *entry[K, V]
}

type entry[K comparable, V any] struct {
	key   K
	value V
	next  *entry[K, V]
}

type Iterator[K comparable, V any] struct {
	pairs []Pair[K, V]
	index int
	pair  Pair[K, V]
}

func New[K comparable, V any](capacity int) *HashMap[K, V] {
	if capacity < defaultCapacity {
		capacity = defaultCapacity
	}

	m := &HashMap[K, V]{
		seed: maphash.MakeSeed(),
	}
	m.state.Store(&table[K, V]{
		buckets: make([]bucket[K, V], nextPowerOfTwo(capacity)),
	})
	return m
}

func (m *HashMap[K, V]) Put(key K, value V) {
	for {
		current := m.state.Load()
		m.resizeMu.RLock()
		if current != m.state.Load() {
			m.resizeMu.RUnlock()
			continue
		}

		index := bucketIndex(m.seed, key, len(current.buckets))
		b := &current.buckets[index]
		b.mu.Lock()
		added := putMutable(&b.head, key, value)
		var size int64
		if added {
			size = m.size.Add(1)
		} else {
			size = m.size.Load()
		}
		needsGrow := added && int(size)*100 > len(current.buckets)*maxLoadPercent
		b.mu.Unlock()
		m.resizeMu.RUnlock()

		if needsGrow {
			m.grow(current)
		}
		return
	}
}

func (m *HashMap[K, V]) Get(key K) (V, bool) {
	m.resizeMu.RLock()
	defer m.resizeMu.RUnlock()

	current := m.state.Load()
	b := &current.buckets[bucketIndex(m.seed, key, len(current.buckets))]
	b.mu.RLock()
	defer b.mu.RUnlock()

	for node := b.head; node != nil; node = node.next {
		if node.key == key {
			return node.value, true
		}
	}

	var zero V
	return zero, false
}

func (m *HashMap[K, V]) Size() int {
	return int(m.size.Load())
}

func (m *HashMap[K, V]) Clear() {
	m.resizeMu.Lock()
	defer m.resizeMu.Unlock()

	current := m.state.Load()
	capacity := defaultCapacity
	if current != nil && len(current.buckets) > capacity {
		capacity = len(current.buckets)
	}

	m.state.Store(&table[K, V]{
		buckets: make([]bucket[K, V], capacity),
	})
	m.size.Store(0)
}

func (m *HashMap[K, V]) Merge(key K, value V, merger func(oldValue V, newValue V) V) V {
	for {
		current := m.state.Load()
		m.resizeMu.RLock()
		if current != m.state.Load() {
			m.resizeMu.RUnlock()
			continue
		}

		index := bucketIndex(m.seed, key, len(current.buckets))
		b := &current.buckets[index]
		b.mu.Lock()
		merged, added := mergeMutable(&b.head, key, value, merger)
		var size int64
		if added {
			size = m.size.Add(1)
		} else {
			size = m.size.Load()
		}
		needsGrow := added && int(size)*100 > len(current.buckets)*maxLoadPercent
		b.mu.Unlock()
		m.resizeMu.RUnlock()

		if needsGrow {
			m.grow(current)
		}
		return merged
	}
}

func (m *HashMap[K, V]) Iterator() Iterator[K, V] {
	return Iterator[K, V]{
		pairs: m.Pairs(),
	}
}

func (m *HashMap[K, V]) Pairs() []Pair[K, V] {
	m.resizeMu.RLock()
	defer m.resizeMu.RUnlock()

	current := m.state.Load()

	pairs := make([]Pair[K, V], 0, m.Size())
	for i := range current.buckets {
		current.buckets[i].mu.RLock()
	}
	defer func() {
		for i := len(current.buckets) - 1; i >= 0; i-- {
			current.buckets[i].mu.RUnlock()
		}
	}()

	for i := range current.buckets {
		for node := current.buckets[i].head; node != nil; node = node.next {
			pairs = append(pairs, Pair[K, V]{
				Key:   node.key,
				Value: node.value,
			})
		}
	}
	return pairs
}

func (it *Iterator[K, V]) Next() bool {
	if it.index >= len(it.pairs) {
		return false
	}
	it.pair = it.pairs[it.index]
	it.index++
	return true
}

func (it *Iterator[K, V]) Pair() Pair[K, V] {
	return it.pair
}

func (m *HashMap[K, V]) grow(expected *table[K, V]) {
	m.resizeMu.Lock()
	defer m.resizeMu.Unlock()

	current := m.state.Load()
	if current != expected || int(m.size.Load())*100 <= len(current.buckets)*maxLoadPercent {
		return
	}

	next := &table[K, V]{
		buckets: make([]bucket[K, V], len(current.buckets)*2),
	}
	for i := range current.buckets {
		for node := current.buckets[i].head; node != nil; node = node.next {
			index := bucketIndex(m.seed, node.key, len(next.buckets))
			next.buckets[index].head = &entry[K, V]{
				key:   node.key,
				value: node.value,
				next:  next.buckets[index].head,
			}
		}
	}
	m.state.Store(next)
}

func putMutable[K comparable, V any](head **entry[K, V], key K, value V) bool {
	for node := *head; node != nil; node = node.next {
		if node.key == key {
			node.value = value
			return false
		}
	}

	*head = &entry[K, V]{key: key, value: value, next: *head}
	return true
}

func mergeMutable[K comparable, V any](head **entry[K, V], key K, value V, merger func(oldValue V, newValue V) V) (V, bool) {
	for node := *head; node != nil; node = node.next {
		if node.key == key {
			node.value = merger(node.value, value)
			return node.value, false
		}
	}

	*head = &entry[K, V]{key: key, value: value, next: *head}
	return value, true
}

func bucketIndex[K comparable](seed maphash.Seed, key K, buckets int) int {
	return int(maphash.Comparable(seed, key) & uint64(buckets-1))
}

func nextPowerOfTwo(value int) int {
	power := 1
	for power < value {
		power <<= 1
	}
	return power
}
