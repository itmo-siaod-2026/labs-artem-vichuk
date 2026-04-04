package perfect

import (
	"errors"
	"fmt"
)

type Pair struct {
	Key   string
	Value float64
}

type entry struct {
	key   string
	value float64
	used  bool
}

type bucket struct {
	seed  uint64
	slots []entry
	count int
}

type PerfectHashTable struct {
	pairs       map[string]float64
	baseHash    baseHashFunc
	primarySeed uint64
	buckets     []bucket
	size        uint64
}

func NewPerfectTable() (*PerfectHashTable, error) {
	ht := &PerfectHashTable{
		pairs:    make(map[string]float64),
		baseHash: randomBaseHashFunc(),
		size:     1,
	}
	if err := ht.rebuild(); err != nil {
		return nil, err
	}
	return ht, nil
}

func BuildPerfectHash(pairs []Pair) (*PerfectHashTable, error) {
	ht := &PerfectHashTable{
		pairs:    make(map[string]float64, len(pairs)),
		baseHash: randomBaseHashFunc(),
		size:     uint64(len(pairs)),
	}
	if ht.size == 0 {
		ht.size = 1
	}

	for _, pair := range pairs {
		ht.pairs[pair.Key] = pair.Value
	}

	if err := ht.rebuild(); err != nil {
		return nil, err
	}
	return ht, nil
}

func (ht *PerfectHashTable) Set(key string, value float64) error {
	if ht.pairs == nil {
		ht.pairs = make(map[string]float64)
	}
	if ht.baseHash == nil {
		ht.baseHash = randomBaseHashFunc()
	}

	ht.pairs[key] = value
	if uint64(len(ht.pairs)) > ht.size {
		ht.size = uint64(len(ht.pairs))
	}
	return ht.rebuild()
}

func (ht *PerfectHashTable) Get(key string) (float64, bool) {
	if len(ht.buckets) == 0 {
		return 0, false
	}

	primaryIdx := hashWithSeed(ht.baseHash, key, ht.primarySeed) % uint64(len(ht.buckets))
	b := ht.buckets[primaryIdx]

	if len(b.slots) == 0 {
		return 0, false
	}

	if len(b.slots) == 1 {
		e := b.slots[0]
		if e.used && e.key == key {
			return e.value, true
		}
		return 0, false
	}

	secondaryIdx := hashWithSeed(ht.baseHash, key, b.seed) % uint64(len(b.slots))
	e := b.slots[secondaryIdx]
	if !e.used || e.key != key {
		return 0, false
	}

	return e.value, true
}

func (ht *PerfectHashTable) Len() int {
	return len(ht.pairs)
}

func (ht *PerfectHashTable) Stats() (primaryBuckets int, secondarySlots int) {
	primaryBuckets = len(ht.buckets)
	for _, b := range ht.buckets {
		secondarySlots += len(b.slots)
	}
	return primaryBuckets, secondarySlots
}

func (ht *PerfectHashTable) rebuild() error {
	if ht.baseHash == nil {
		ht.baseHash = randomBaseHashFunc()
	}
	if ht.size == 0 {
		ht.size = 1
	}

	pairs := make([]Pair, 0, len(ht.pairs))
	for key, value := range ht.pairs {
		pairs = append(pairs, Pair{Key: key, Value: value})
	}

	if len(pairs) == 0 {
		ht.primarySeed = 0
		ht.buckets = make([]bucket, 1)
		return nil
	}

	primarySize := uint64(len(pairs))
	if primarySize == 0 {
		primarySize = 1
	}

	bestSeed := uint64(1)
	bestGroups := make([][]Pair, primarySize)
	bestScore := ^uint64(0)
	limit := uint64(4 * len(pairs))
	if limit == 0 {
		limit = 1
	}

	for attempt := uint64(0); attempt < 256; attempt++ {
		seed := mix64(attempt + 1)
		groups := make([][]Pair, primarySize)
		var sumSquares uint64

		for _, pair := range pairs {
			idx := hashWithSeed(ht.baseHash, pair.Key, seed) % primarySize
			groups[idx] = append(groups[idx], pair)
		}

		for _, group := range groups {
			m := uint64(len(group))
			sumSquares += m * m
		}

		if sumSquares < bestScore {
			bestScore = sumSquares
			bestSeed = seed
			bestGroups = groups
		}
		if sumSquares <= limit {
			break
		}
	}

	buckets := make([]bucket, primarySize)
	for i, group := range bestGroups {
		built, err := buildSecondary(group, ht.baseHash)
		if err != nil {
			return fmt.Errorf("build secondary bucket %d: %w", i, err)
		}
		buckets[i] = built
	}

	ht.primarySeed = bestSeed
	ht.buckets = buckets
	ht.size = primarySize
	return nil
}

func buildSecondary(group []Pair, baseHash baseHashFunc) (bucket, error) {
	if len(group) == 0 {
		return bucket{}, nil
	}

	if len(group) == 1 {
		return bucket{
			seed:  0,
			count: 1,
			slots: []entry{{key: group[0].Key, value: group[0].Value, used: true}},
		}, nil
	}

	slotsCount := uint64(len(group) * len(group))
	if slotsCount == 0 {
		return bucket{}, errors.New("invalid secondary size")
	}

	for attempt := uint64(0); attempt < 10000; attempt++ {
		seed := mix64(uint64(len(group)) + attempt + 1)
		slots := make([]entry, slotsCount)
		collision := false

		for _, pair := range group {
			idx := hashWithSeed(baseHash, pair.Key, seed) % slotsCount
			if slots[idx].used {
				collision = true
				break
			}
			slots[idx] = entry{key: pair.Key, value: pair.Value, used: true}
		}

		if !collision {
			return bucket{seed: seed, slots: slots, count: len(group)}, nil
		}
	}

	return bucket{}, fmt.Errorf("failed to find collision-free secondary seed for %d keys", len(group))
}
