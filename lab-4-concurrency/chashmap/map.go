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
	mu    sync.Mutex
	seed  maphash.Seed
	state atomic.Pointer[table[K, V]]
}

type table[K comparable, V any] struct {
	buckets []*entry[K, V]
	size    int
}

type entry[K comparable, V any] struct {
	key   K
	value V
	next  *entry[K, V]
}

type Iterator[K comparable, V any] struct {
	table       *table[K, V]
	bucketIndex int
	current     *entry[K, V]
	pair        Pair[K, V]
}

func New[K comparable, V any](capacity int) *HashMap[K, V] {
	if capacity < defaultCapacity {
		capacity = defaultCapacity
	}

	m := &HashMap[K, V]{
		seed: maphash.MakeSeed(),
	}
	m.state.Store(&table[K, V]{
		buckets: make([]*entry[K, V], nextPowerOfTwo(capacity)),
	})
	return m
}

func (m *HashMap[K, V]) Put(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()

	current := m.state.Load()
	next, _ := current.withPut(m.seed, key, value)
	m.state.Store(next)
}

func (m *HashMap[K, V]) Get(key K) (V, bool) {
	current := m.state.Load()
	if current == nil {
		var zero V
		return zero, false
	}

	for node := current.buckets[bucketIndex(m.seed, key, len(current.buckets))]; node != nil; node = node.next {
		if node.key == key {
			return node.value, true
		}
	}

	var zero V
	return zero, false
}

func (m *HashMap[K, V]) Size() int {
	current := m.state.Load()
	if current == nil {
		return 0
	}
	return current.size
}

func (m *HashMap[K, V]) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	current := m.state.Load()
	capacity := defaultCapacity
	if current != nil && len(current.buckets) > capacity {
		capacity = len(current.buckets)
	}

	m.state.Store(&table[K, V]{
		buckets: make([]*entry[K, V], capacity),
	})
}

func (m *HashMap[K, V]) Merge(key K, value V, merger func(oldValue V, newValue V) V) V {
	m.mu.Lock()
	defer m.mu.Unlock()

	current := m.state.Load()
	next, merged := current.withMerge(m.seed, key, value, merger)
	m.state.Store(next)
	return merged
}

func (m *HashMap[K, V]) Iterator() Iterator[K, V] {
	return Iterator[K, V]{
		table: m.state.Load(),
	}
}

func (m *HashMap[K, V]) Pairs() []Pair[K, V] {
	current := m.state.Load()
	if current == nil {
		return nil
	}

	pairs := make([]Pair[K, V], 0, current.size)
	for _, bucket := range current.buckets {
		for node := bucket; node != nil; node = node.next {
			pairs = append(pairs, Pair[K, V]{
				Key:   node.key,
				Value: node.value,
			})
		}
	}
	return pairs
}

func (it *Iterator[K, V]) Next() bool {
	if it.table == nil {
		return false
	}

	if it.current != nil {
		it.pair = Pair[K, V]{Key: it.current.key, Value: it.current.value}
		it.current = it.current.next
		return true
	}

	for it.bucketIndex < len(it.table.buckets) {
		node := it.table.buckets[it.bucketIndex]
		it.bucketIndex++
		if node != nil {
			it.pair = Pair[K, V]{Key: node.key, Value: node.value}
			it.current = node.next
			return true
		}
	}

	return false
}

func (it *Iterator[K, V]) Pair() Pair[K, V] {
	return it.pair
}

func (t *table[K, V]) withPut(seed maphash.Seed, key K, value V) (*table[K, V], bool) {
	if t.needsGrow(1) && !t.contains(seed, key) {
		return t.rebuildWith(seed, key, value, false, nil)
	}

	index := bucketIndex(seed, key, len(t.buckets))
	buckets := append([]*entry[K, V](nil), t.buckets...)
	bucket, added := putImmutable(buckets[index], key, value)
	buckets[index] = bucket

	size := t.size
	if added {
		size++
	}
	return &table[K, V]{buckets: buckets, size: size}, added
}

func (t *table[K, V]) withMerge(seed maphash.Seed, key K, value V, merger func(oldValue V, newValue V) V) (*table[K, V], V) {
	if t.needsGrow(1) && !t.contains(seed, key) {
		next, _ := t.rebuildWith(seed, key, value, true, merger)
		return next, value
	}

	index := bucketIndex(seed, key, len(t.buckets))
	buckets := append([]*entry[K, V](nil), t.buckets...)
	bucket, merged, added := mergeImmutable(buckets[index], key, value, merger)
	buckets[index] = bucket

	size := t.size
	if added {
		size++
	}
	return &table[K, V]{buckets: buckets, size: size}, merged
}

func (t *table[K, V]) rebuildWith(seed maphash.Seed, key K, value V, merge bool, merger func(V, V) V) (*table[K, V], bool) {
	next := &table[K, V]{
		buckets: make([]*entry[K, V], len(t.buckets)*2),
		size:    t.size,
	}

	for _, bucket := range t.buckets {
		for node := bucket; node != nil; node = node.next {
			index := bucketIndex(seed, node.key, len(next.buckets))
			next.buckets[index] = &entry[K, V]{
				key:   node.key,
				value: node.value,
				next:  next.buckets[index],
			}
		}
	}

	index := bucketIndex(seed, key, len(next.buckets))
	if merge {
		bucket, _, added := mergeImmutable(next.buckets[index], key, value, merger)
		next.buckets[index] = bucket
		if added {
			next.size++
		}
		return next, added
	}

	bucket, added := putImmutable(next.buckets[index], key, value)
	next.buckets[index] = bucket
	if added {
		next.size++
	}
	return next, added
}

func (t *table[K, V]) needsGrow(added int) bool {
	return (t.size+added)*100 > len(t.buckets)*maxLoadPercent
}

func (t *table[K, V]) contains(seed maphash.Seed, key K) bool {
	for node := t.buckets[bucketIndex(seed, key, len(t.buckets))]; node != nil; node = node.next {
		if node.key == key {
			return true
		}
	}
	return false
}

func putImmutable[K comparable, V any](head *entry[K, V], key K, value V) (*entry[K, V], bool) {
	if head == nil {
		return &entry[K, V]{key: key, value: value}, true
	}

	if head.key == key {
		return &entry[K, V]{key: key, value: value, next: head.next}, false
	}

	next, added := putImmutable(head.next, key, value)
	return &entry[K, V]{key: head.key, value: head.value, next: next}, added
}

func mergeImmutable[K comparable, V any](head *entry[K, V], key K, value V, merger func(oldValue V, newValue V) V) (*entry[K, V], V, bool) {
	if head == nil {
		return &entry[K, V]{key: key, value: value}, value, true
	}

	if head.key == key {
		merged := merger(head.value, value)
		return &entry[K, V]{key: key, value: merged, next: head.next}, merged, false
	}

	next, merged, added := mergeImmutable(head.next, key, value, merger)
	return &entry[K, V]{key: head.key, value: head.value, next: next}, merged, added
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
