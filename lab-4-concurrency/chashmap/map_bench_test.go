package chashmap

import (
	"sync/atomic"
	"testing"
)

const benchKeys = 4_096

var benchmarkSink atomic.Int64

func BenchmarkConcurrentHashMapRead(b *testing.B) {
	m := New[int, int](benchKeys)
	for i := 0; i < benchKeys; i++ {
		m.Put(i, i)
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		sum := 0
		for pb.Next() {
			value, _ := m.Get(i & (benchKeys - 1))
			sum += value
			i++
		}
		benchmarkSink.Add(int64(sum))
	})
}

func BenchmarkUnsafeHashMapRead(b *testing.B) {
	m := NewUnsafe[int, int](benchKeys)
	for i := 0; i < benchKeys; i++ {
		m.Put(i, i)
	}

	b.ReportAllocs()
	b.ResetTimer()
	sum := 0
	for i := 0; i < b.N; i++ {
		value, _ := m.Get(i & (benchKeys - 1))
		sum += value
	}
	benchmarkSink.Store(int64(sum))
}

func BenchmarkConcurrentHashMapPutUpdate(b *testing.B) {
	m := New[int, int](benchKeys)
	for i := 0; i < benchKeys; i++ {
		m.Put(i, i)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Put(i&(benchKeys-1), i)
	}
}

func BenchmarkUnsafeHashMapPutUpdate(b *testing.B) {
	m := NewUnsafe[int, int](benchKeys)
	for i := 0; i < benchKeys; i++ {
		m.Put(i, i)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Put(i&(benchKeys-1), i)
	}
}

func BenchmarkConcurrentHashMapMergeHotKey(b *testing.B) {
	m := New[string, int](1)
	key := "counter"

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Merge(key, 1, func(oldValue, newValue int) int {
				return oldValue + newValue
			})
		}
	})
}

func BenchmarkUnsafeHashMapMergeHotKey(b *testing.B) {
	m := NewUnsafe[string, int](1)
	key := "counter"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Merge(key, 1, func(oldValue, newValue int) int {
			return oldValue + newValue
		})
	}
}
