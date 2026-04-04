package tests_test

import (
	"bdrs/lab-1-perfect/perfect"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"testing"
)

func ExamplePerfectHashTable() {
	ht, err := perfect.BuildPerfectHash([]perfect.Pair{
		{Key: "abcd", Value: 1},
		{Key: "abc", Value: 2},
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(ht.Get("abc"))
	fmt.Println(ht.Get("abd"))
	// Output:
	// 2 true
	// 0 false
}

func TestPerfectHashEmptyTable(t *testing.T) {
	ht, err := perfect.NewPerfectTable()
	if err != nil {
		t.Fatalf("new table: %v", err)
	}

	if ht.Len() != 0 {
		t.Fatalf("expected empty table, got len=%d", ht.Len())
	}

	if _, ok := ht.Get("missing"); ok {
		t.Fatalf("missing key must not be found in empty table")
	}

	primaryBuckets, secondarySlots := ht.Stats()
	if primaryBuckets == 0 {
		t.Fatalf("expected at least one primary bucket")
	}
	if secondarySlots != 0 {
		t.Fatalf("expected zero secondary slots for empty table, got %d", secondarySlots)
	}
}

func TestPerfectHashBuildAndLookup(t *testing.T) {
	pairs := make([]perfect.Pair, 0, 100)
	for i := 0; i < 100; i++ {
		pairs = append(pairs, perfect.Pair{
			Key:   "key" + strconv.Itoa(i),
			Value: float64(i),
		})
	}

	ht, err := perfect.BuildPerfectHash(pairs)
	if err != nil {
		t.Fatalf("build perfect hash: %v", err)
	}

	if ht.Len() != 100 {
		t.Fatalf("unexpected len: got=%d want=100", ht.Len())
	}

	for i := 0; i < 100; i++ {
		v, ok := ht.Get("key" + strconv.Itoa(i))
		if !ok || v != float64(i) {
			t.Fatalf("lookup key%d: got=(%v,%v) want=(%v,true)", i, v, ok, float64(i))
		}
	}

	if _, ok := ht.Get("missing"); ok {
		t.Fatalf("missing key must not be found")
	}
}

func TestPerfectHashSetUpdateRebuildsCorrectly(t *testing.T) {
	ht, err := perfect.NewPerfectTable()
	if err != nil {
		t.Fatalf("new table: %v", err)
	}

	if err := ht.Set("key10", 100.0); err != nil {
		t.Fatalf("first Set failed: %v", err)
	}
	if err := ht.Set("key10", 999.0); err != nil {
		t.Fatalf("update Set failed: %v", err)
	}
	if err := ht.Set("key20", 200.0); err != nil {
		t.Fatalf("second key Set failed: %v", err)
	}

	v, ok := ht.Get("key10")
	if !ok || v != 999.0 {
		t.Fatalf("expected updated value=999.0, got value=%f ok=%v", v, ok)
	}

	v, ok = ht.Get("key20")
	if !ok || v != 200.0 {
		t.Fatalf("expected value=200.0, got value=%f ok=%v", v, ok)
	}

	if ht.Len() != 2 {
		t.Fatalf("expected len=2, got %d", ht.Len())
	}
}

func TestPerfectHashRandomizedAgainstMap(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	ref := make(map[string]float64)

	ht, err := perfect.NewPerfectTable()
	if err != nil {
		t.Fatalf("new table: %v", err)
	}

	for i := 0; i < 500; i++ {
		key := "key-" + strconv.Itoa(rng.Intn(200))
		value := float64(rng.Intn(100000)) / 10

		ref[key] = value
		if err := ht.Set(key, value); err != nil {
			t.Fatalf("set %q: %v", key, err)
		}
	}

	if ht.Len() != len(ref) {
		t.Fatalf("len mismatch: got=%d want=%d", ht.Len(), len(ref))
	}

	for key, want := range ref {
		got, ok := ht.Get(key)
		if !ok || got != want {
			t.Fatalf("key %q got=(%v,%v) want=(%v,true)", key, got, ok, want)
		}
	}

	if _, ok := ht.Get("definitely-missing"); ok {
		t.Fatalf("missing key must not be found")
	}
}

func FuzzPerfectHashTableSetGet(f *testing.F) {
	f.Add("key1", 10.0)
	f.Add("key2", 20.0)
	f.Add("", 0.0)
	f.Add("very_long_key_with_many_characters_123456789", 999.99)

	f.Fuzz(func(t *testing.T, key string, value float64) {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			t.Skip()
		}

		ht, err := perfect.NewPerfectTable()
		if err != nil {
			t.Fatalf("failed to create table: %v", err)
		}

		if err := ht.Set(key, value); err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		got, ok := ht.Get(key)
		if !ok {
			t.Fatalf("Get failed for key=%q after Set", key)
		}
		if got != value {
			t.Fatalf("value mismatch: got=%f want=%f", got, value)
		}
	})
}
