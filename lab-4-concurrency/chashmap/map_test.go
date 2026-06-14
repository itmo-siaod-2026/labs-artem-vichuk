package chashmap

import (
	"sync"
	"testing"
)

func TestPutGetSizeAndClear(t *testing.T) {
	m := New[string, int](2)

	m.Put("one", 1)
	m.Put("two", 2)
	m.Put("one", 11)

	if got, ok := m.Get("one"); !ok || got != 11 {
		t.Fatalf("Get(one) = %d, %t; want 11, true", got, ok)
	}
	if got := m.Size(); got != 2 {
		t.Fatalf("Size() = %d; want 2", got)
	}

	m.Clear()
	if got := m.Size(); got != 0 {
		t.Fatalf("Size() after Clear = %d; want 0", got)
	}
	if _, ok := m.Get("one"); ok {
		t.Fatal("Get(one) after Clear returned value")
	}
}

func TestMerge(t *testing.T) {
	m := New[string, int](1)

	if got := m.Merge("x", 1, func(oldValue, newValue int) int {
		return oldValue + newValue
	}); got != 1 {
		t.Fatalf("Merge insert = %d; want 1", got)
	}

	if got := m.Merge("x", 2, func(oldValue, newValue int) int {
		return oldValue + newValue
	}); got != 3 {
		t.Fatalf("Merge update = %d; want 3", got)
	}
	if got, _ := m.Get("x"); got != 3 {
		t.Fatalf("Get(x) = %d; want 3", got)
	}
}

func TestIteratorUsesSnapshot(t *testing.T) {
	m := New[int, int](4)
	for i := 0; i < 10; i++ {
		m.Put(i, i)
	}

	it := m.Iterator()
	m.Clear()
	m.Put(100, 100)

	seen := make(map[int]int)
	for it.Next() {
		pair := it.Pair()
		seen[pair.Key] = pair.Value
	}

	if len(seen) != 10 {
		t.Fatalf("iterator yielded %d pairs; want 10", len(seen))
	}
	for i := 0; i < 10; i++ {
		if seen[i] != i {
			t.Fatalf("iterator snapshot missing %d", i)
		}
	}
}

func TestConcurrentMergeIsLinearizableForSingleKey(t *testing.T) {
	const goroutines = 16
	const increments = 1_000

	m := New[string, int](1)
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < increments; j++ {
				m.Merge("counter", 1, func(oldValue, newValue int) int {
					return oldValue + newValue
				})
			}
		}()
	}
	wg.Wait()

	got, ok := m.Get("counter")
	if !ok {
		t.Fatal("counter is absent")
	}
	if want := goroutines * increments; got != want {
		t.Fatalf("counter = %d; want %d", got, want)
	}
}

func TestConcurrentReadersObserveCompletedWrites(t *testing.T) {
	const keys = 1_000
	m := New[int, int](keys)

	var writer sync.WaitGroup
	writer.Add(1)
	go func() {
		defer writer.Done()
		for i := 0; i < keys; i++ {
			m.Put(i, i*i)
		}
	}()

	var readers sync.WaitGroup
	for r := 0; r < 8; r++ {
		readers.Add(1)
		go func() {
			defer readers.Done()
			for i := 0; i < keys; i++ {
				if value, ok := m.Get(i); ok && value != i*i {
					t.Errorf("Get(%d) = %d; want %d", i, value, i*i)
				}
				_ = m.Size()
			}
		}()
	}

	writer.Wait()
	readers.Wait()

	for i := 0; i < keys; i++ {
		if value, ok := m.Get(i); !ok || value != i*i {
			t.Fatalf("completed Put(%d) is not visible: %d, %t", i, value, ok)
		}
	}
}
