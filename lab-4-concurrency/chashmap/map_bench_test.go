package chashmap

import (
	"sync/atomic"
	"testing"
	"time"
)

const benchKeys = 4_096
const benchInsertBatch = 4_096

var benchmarkSink atomic.Int64

type latencyRecorder struct {
	count   atomic.Uint64
	totalNS atomic.Uint64
}

func newLatencyRecorder() *latencyRecorder {
	return &latencyRecorder{}
}

func (r *latencyRecorder) observe(duration time.Duration) {
	ns := uint64(duration.Nanoseconds())
	r.totalNS.Add(ns)
	r.count.Add(1)
}

func (r *latencyRecorder) report(b *testing.B) {
	count := r.count.Load()
	if count == 0 {
		return
	}

	avg := float64(r.totalNS.Load()) / float64(count)
	b.ReportMetric(avg, "ns/op")
}

func BenchmarkConcurrentHashMapReadLatencySequential(b *testing.B) {
	m := New[int, int](benchKeys)
	for i := 0; i < benchKeys; i++ {
		m.Put(i, i)
	}

	latency := newLatencyRecorder()
	b.ReportAllocs()
	b.ResetTimer()
	sum := 0
	for i := 0; i < b.N; i++ {
		start := time.Now()
		value, _ := m.Get(i & (benchKeys - 1))
		latency.observe(time.Since(start))
		sum += value
	}
	benchmarkSink.Store(int64(sum))
	latency.report(b)
}

func BenchmarkUnsafeHashMapReadLatencySequential(b *testing.B) {
	m := NewUnsafe[int, int](benchKeys)
	for i := 0; i < benchKeys; i++ {
		m.Put(i, i)
	}

	latency := newLatencyRecorder()
	b.ReportAllocs()
	b.ResetTimer()
	sum := 0
	for i := 0; i < b.N; i++ {
		start := time.Now()
		value, _ := m.Get(i & (benchKeys - 1))
		latency.observe(time.Since(start))
		sum += value
	}
	benchmarkSink.Store(int64(sum))
	latency.report(b)
}

func BenchmarkConcurrentHashMapPutUpdateLatencySequential(b *testing.B) {
	m := New[int, int](benchKeys)
	for i := 0; i < benchKeys; i++ {
		m.Put(i, i)
	}

	latency := newLatencyRecorder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		m.Put(i&(benchKeys-1), i)
		latency.observe(time.Since(start))
	}
	latency.report(b)
}

func BenchmarkConcurrentHashMapPutUpdateLatencyParallel(b *testing.B) {
	m := New[int, int](benchKeys)
	for i := 0; i < benchKeys; i++ {
		m.Put(i, i)
	}

	var nextStart atomic.Uint64
	latency := newLatencyRecorder()
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := int(nextStart.Add(benchKeys))
		for pb.Next() {
			start := time.Now()
			m.Put(i&(benchKeys-1), i)
			latency.observe(time.Since(start))
			i++
		}
	})
	latency.report(b)
}

func BenchmarkUnsafeHashMapPutUpdateLatencySequential(b *testing.B) {
	m := NewUnsafe[int, int](benchKeys)
	for i := 0; i < benchKeys; i++ {
		m.Put(i, i)
	}

	latency := newLatencyRecorder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		m.Put(i&(benchKeys-1), i)
		latency.observe(time.Since(start))
	}
	latency.report(b)
}

func BenchmarkConcurrentHashMapPutInsertNoGrowLatency(b *testing.B) {
	m := New[int, int](benchInsertBatch * 2)
	key := 0

	latency := newLatencyRecorder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if key == benchInsertBatch {
			b.StopTimer()
			m = New[int, int](benchInsertBatch * 2)
			key = 0
			b.StartTimer()
		}
		start := time.Now()
		m.Put(key, key)
		latency.observe(time.Since(start))
		key++
	}
	latency.report(b)
}

func BenchmarkUnsafeHashMapPutInsertNoGrowLatency(b *testing.B) {
	m := NewUnsafe[int, int](benchInsertBatch * 2)
	key := 0

	latency := newLatencyRecorder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if key == benchInsertBatch {
			b.StopTimer()
			m = NewUnsafe[int, int](benchInsertBatch * 2)
			key = 0
			b.StartTimer()
		}
		start := time.Now()
		m.Put(key, key)
		latency.observe(time.Since(start))
		key++
	}
	latency.report(b)
}

func BenchmarkConcurrentHashMapPutInsertWithGrowLatency(b *testing.B) {
	m := New[int, int](1)
	key := 0

	latency := newLatencyRecorder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if key == benchInsertBatch {
			b.StopTimer()
			m = New[int, int](1)
			key = 0
			b.StartTimer()
		}
		start := time.Now()
		m.Put(key, key)
		latency.observe(time.Since(start))
		key++
	}
	latency.report(b)
}

func BenchmarkUnsafeHashMapPutInsertWithGrowLatency(b *testing.B) {
	m := NewUnsafe[int, int](1)
	key := 0

	latency := newLatencyRecorder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if key == benchInsertBatch {
			b.StopTimer()
			m = NewUnsafe[int, int](1)
			key = 0
			b.StartTimer()
		}
		start := time.Now()
		m.Put(key, key)
		latency.observe(time.Since(start))
		key++
	}
	latency.report(b)
}

func BenchmarkConcurrentHashMapReadMostlyLatencyParallel(b *testing.B) {
	m := New[int, int](benchKeys)
	for i := 0; i < benchKeys; i++ {
		m.Put(i, i)
	}

	var ops atomic.Uint64
	latency := newLatencyRecorder()
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		sum := 0
		for pb.Next() {
			op := ops.Add(1)
			key := int(op) & (benchKeys - 1)
			if op&15 == 0 {
				start := time.Now()
				m.Put(key, int(op))
				latency.observe(time.Since(start))
				continue
			}
			start := time.Now()
			value, _ := m.Get(key)
			latency.observe(time.Since(start))
			sum += value
		}
		benchmarkSink.Add(int64(sum))
	})
	latency.report(b)
}

func BenchmarkConcurrentHashMapMergeHotKeyLatencySequential(b *testing.B) {
	m := New[string, int](1)
	key := "counter"

	latency := newLatencyRecorder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		m.Merge(key, 1, func(oldValue, newValue int) int {
			return oldValue + newValue
		})
		latency.observe(time.Since(start))
	}
	latency.report(b)
}

func BenchmarkUnsafeHashMapMergeHotKeyLatencySequential(b *testing.B) {
	m := NewUnsafe[string, int](1)
	key := "counter"

	latency := newLatencyRecorder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		m.Merge(key, 1, func(oldValue, newValue int) int {
			return oldValue + newValue
		})
		latency.observe(time.Since(start))
	}
	latency.report(b)
}

func BenchmarkConcurrentHashMapPairsSnapshotLatency(b *testing.B) {
	m := New[int, int](benchKeys)
	for i := 0; i < benchKeys; i++ {
		m.Put(i, i)
	}

	latency := newLatencyRecorder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		benchmarkSink.Add(int64(len(m.Pairs())))
		latency.observe(time.Since(start))
	}
	latency.report(b)
}

func BenchmarkUnsafeHashMapPairsSnapshotLatency(b *testing.B) {
	m := NewUnsafe[int, int](benchKeys)
	for i := 0; i < benchKeys; i++ {
		m.Put(i, i)
	}

	latency := newLatencyRecorder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		benchmarkSink.Add(int64(len(m.Pairs())))
		latency.observe(time.Since(start))
	}
	latency.report(b)
}

func BenchmarkConcurrentHashMapReadThroughputParallel(b *testing.B) {
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
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "ops/s")
}

func BenchmarkConcurrentHashMapPutUpdateThroughputParallel(b *testing.B) {
	m := New[int, int](benchKeys)
	for i := 0; i < benchKeys; i++ {
		m.Put(i, i)
	}

	var nextStart atomic.Uint64
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := int(nextStart.Add(benchKeys))
		for pb.Next() {
			m.Put(i&(benchKeys-1), i)
			i++
		}
	})
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "ops/s")
}

func BenchmarkConcurrentHashMapReadMostlyThroughputParallel(b *testing.B) {
	m := New[int, int](benchKeys)
	for i := 0; i < benchKeys; i++ {
		m.Put(i, i)
	}

	var ops atomic.Uint64
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		sum := 0
		for pb.Next() {
			op := ops.Add(1)
			key := int(op) & (benchKeys - 1)
			if op&15 == 0 {
				m.Put(key, int(op))
				continue
			}
			value, _ := m.Get(key)
			sum += value
		}
		benchmarkSink.Add(int64(sum))
	})
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "ops/s")
}

func BenchmarkConcurrentHashMapMergeHotKeyThroughputParallel(b *testing.B) {
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
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "ops/s")
}
