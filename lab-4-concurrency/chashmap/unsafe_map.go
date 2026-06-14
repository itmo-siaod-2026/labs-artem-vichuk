package chashmap

import "hash/maphash"

type UnsafeHashMap[K comparable, V any] struct {
	seed    maphash.Seed
	buckets []*entry[K, V]
	size    int
}

func NewUnsafe[K comparable, V any](capacity int) *UnsafeHashMap[K, V] {
	if capacity < defaultCapacity {
		capacity = defaultCapacity
	}
	return &UnsafeHashMap[K, V]{
		seed:    maphash.MakeSeed(),
		buckets: make([]*entry[K, V], nextPowerOfTwo(capacity)),
	}
}

func (m *UnsafeHashMap[K, V]) Put(key K, value V) {
	if (m.size+1)*100 > len(m.buckets)*maxLoadPercent && !m.contains(key) {
		m.grow()
	}

	index := bucketIndex(m.seed, key, len(m.buckets))
	for node := m.buckets[index]; node != nil; node = node.next {
		if node.key == key {
			node.value = value
			return
		}
	}

	m.buckets[index] = &entry[K, V]{key: key, value: value, next: m.buckets[index]}
	m.size++
}

func (m *UnsafeHashMap[K, V]) Get(key K) (V, bool) {
	for node := m.buckets[bucketIndex(m.seed, key, len(m.buckets))]; node != nil; node = node.next {
		if node.key == key {
			return node.value, true
		}
	}

	var zero V
	return zero, false
}

func (m *UnsafeHashMap[K, V]) Size() int {
	return m.size
}

func (m *UnsafeHashMap[K, V]) Clear() {
	m.buckets = make([]*entry[K, V], len(m.buckets))
	m.size = 0
}

func (m *UnsafeHashMap[K, V]) Merge(key K, value V, merger func(V, V) V) V {
	if (m.size+1)*100 > len(m.buckets)*maxLoadPercent && !m.contains(key) {
		m.grow()
	}

	index := bucketIndex(m.seed, key, len(m.buckets))
	for node := m.buckets[index]; node != nil; node = node.next {
		if node.key == key {
			node.value = merger(node.value, value)
			return node.value
		}
	}

	m.buckets[index] = &entry[K, V]{key: key, value: value, next: m.buckets[index]}
	m.size++
	return value
}

func (m *UnsafeHashMap[K, V]) Pairs() []Pair[K, V] {
	pairs := make([]Pair[K, V], 0, m.size)
	for _, bucket := range m.buckets {
		for node := bucket; node != nil; node = node.next {
			pairs = append(pairs, Pair[K, V]{Key: node.key, Value: node.value})
		}
	}
	return pairs
}

func (m *UnsafeHashMap[K, V]) contains(key K) bool {
	for node := m.buckets[bucketIndex(m.seed, key, len(m.buckets))]; node != nil; node = node.next {
		if node.key == key {
			return true
		}
	}
	return false
}

func (m *UnsafeHashMap[K, V]) grow() {
	next := make([]*entry[K, V], len(m.buckets)*2)
	for _, bucket := range m.buckets {
		for node := bucket; node != nil; node = node.next {
			index := bucketIndex(m.seed, node.key, len(next))
			next[index] = &entry[K, V]{key: node.key, value: node.value, next: next[index]}
		}
	}
	m.buckets = next
}
