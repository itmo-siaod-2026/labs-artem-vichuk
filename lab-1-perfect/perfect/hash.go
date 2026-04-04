package perfect

import (
	"encoding/binary"
	"hash/fnv"
	"math/rand"
	"sync"
	"time"
)

type baseHashFunc func(key string) uint64

var (
	hashRNG     *rand.Rand
	hashRNGOnce sync.Once
)

func initHashRNG() {
	hashRNG = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func randomBaseHashFunc() baseHashFunc {
	hashRNGOnce.Do(initHashRNG)

	switch hashRNG.Intn(3) {
	case 0:
		return hashFNV1a64
	case 1:
		return hashDJB264
	default:
		return hashSDBM64
	}
}

func hashFNV1a64(key string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(key))
	return h.Sum64()
}

func hashDJB264(key string) uint64 {
	var h uint64 = 5381
	for _, c := range key {
		h = (h << 5) + h + uint64(c)
	}
	return h
}

func hashSDBM64(key string) uint64 {
	var h uint64
	for _, c := range key {
		h = uint64(c) + (h << 6) + (h << 16) - h
	}
	return h
}

func hashWithSeed(baseHash baseHashFunc, key string, seed uint64) uint64 {
	if baseHash == nil {
		baseHash = hashFNV1a64
	}

	base := baseHash(key)

	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], seed)
	seeded := binary.LittleEndian.Uint64(buf[:])

	return mix64(base ^ mix64(seeded+0x9e3779b97f4a7c15))
}

func mix64(v uint64) uint64 {
	v ^= v >> 30
	v *= 0xbf58476d1ce4e5b9
	v ^= v >> 27
	v *= 0x94d049bb133111eb
	v ^= v >> 31
	return v
}
