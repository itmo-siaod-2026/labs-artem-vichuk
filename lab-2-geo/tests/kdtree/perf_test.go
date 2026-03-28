package kdtree_test

import (
	kd "bdrs/lab-2-geo/kdtree"
	"fmt"
	"math"
	"math/rand"
	"testing"
)

const benchEarthRadiusMeters = 6371008.8

var (
	defaultKDTreeSizes      = []int{100_000, 500_000, 1_000_000, 10_000_000, 50_000_000}
	defaultKDTreeQueryCount = 8192
	defaultKDTreeRadii      = []float64{1000, 10000, 50000}

	kdTreeSinkInt       int
	kdTreeSinkInt64     int64
	kdTreeSinkFloat64   float64
	kdTreeSinkBool      bool
	kdTreeSinkNeighbors []kd.Neighbor
	kdTreeSinkTree      *kd.Tree
)

func kdTreeSizes() []int {
	return defaultKDTreeSizes
}

func kdTreeQueryCount() int {
	return defaultKDTreeQueryCount
}

func kdTreeRadii() []float64 {
	return defaultKDTreeRadii
}

func makeKDTreeItems(n int) []kd.Item {
	rng := rand.New(rand.NewSource(42))
	items := make([]kd.Item, n)
	for i := 0; i < n; i++ {
		items[i] = kd.Item{
			ID: int64(i + 1),
			Point: kd.Point{
				Lat: rng.Float64()*180 - 90,
				Lng: rng.Float64()*360 - 180,
			},
		}
	}
	return items
}

func makeKDTreeQueries(n int) []kd.Point {
	rng := rand.New(rand.NewSource(99))
	queries := make([]kd.Point, n)
	for i := 0; i < n; i++ {
		queries[i] = kd.Point{
			Lat: rng.Float64()*170 - 85,
			Lng: rng.Float64()*340 - 170,
		}
	}
	return queries
}

func BenchmarkKDTreeBuild(b *testing.B) {
	for _, size := range kdTreeSizes() {
		items := makeKDTreeItems(size)

		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				tree, err := kd.New(items)
				if err != nil {
					b.Fatalf("build kd-tree: %v", err)
				}
				if tree.Len() != size {
					b.Fatalf("unexpected size: got %d want %d", tree.Len(), size)
				}
				kdTreeSinkTree = tree
			}

			b.ReportMetric(float64(size*b.N)/b.Elapsed().Seconds(), "items/s")
		})
	}
}

func BenchmarkKDTreeNearest(b *testing.B) {
	queryCount := kdTreeQueryCount()

	for _, size := range kdTreeSizes() {
		items := makeKDTreeItems(size)
		queries := makeKDTreeQueries(queryCount)

		tree, err := kd.New(items)
		if err != nil {
			b.Fatalf("build kd-tree: %v", err)
		}

		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			var sumID int64
			var sumDist float64

			for i := 0; i < b.N; i++ {
				neighbor, ok, err := tree.Nearest(queries[i%len(queries)])
				if err != nil {
					b.Fatalf("nearest: %v", err)
				}
				if !ok {
					b.Fatalf("nearest: not found")
				}
				sumID += neighbor.Item.ID
				sumDist += neighbor.DistanceMeters
			}

			kdTreeSinkInt64 = sumID
			kdTreeSinkFloat64 = sumDist

			b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "ops/s")
		})
	}
}

func BenchmarkKDTreeWithinRadius(b *testing.B) {
	queryCount := kdTreeQueryCount()
	radii := kdTreeRadii()

	for _, size := range kdTreeSizes() {
		items := makeKDTreeItems(size)
		queries := makeKDTreeQueries(queryCount)

		tree, err := kd.New(items)
		if err != nil {
			b.Fatalf("build kd-tree: %v", err)
		}

		for _, radius := range radii {
			b.Run(fmt.Sprintf("size=%d/radius=%.0f", size, radius), func(b *testing.B) {
				b.ReportAllocs()
				b.ResetTimer()

				var total int
				var sumIDs int64
				var last []kd.Neighbor

				for i := 0; i < b.N; i++ {
					neighbors, err := tree.WithinRadius(queries[i%len(queries)], radius)
					if err != nil {
						b.Fatalf("within radius: %v", err)
					}
					total += len(neighbors)
					for j := range neighbors {
						sumIDs += neighbors[j].Item.ID
					}
					last = neighbors
				}

				kdTreeSinkInt = total
				kdTreeSinkInt64 = sumIDs
				kdTreeSinkNeighbors = last

				b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "ops/s")
			})
		}
	}
}

func BenchmarkLinearNearest(b *testing.B) {
	queryCount := kdTreeQueryCount()

	for _, size := range kdTreeSizes() {
		items := makeKDTreeItems(size)
		queries := makeKDTreeQueries(queryCount)

		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			var sumID int64
			var sumDist float64

			for i := 0; i < b.N; i++ {
				neighbor := linearNearest(items, queries[i%len(queries)])
				sumID += neighbor.Item.ID
				sumDist += neighbor.DistanceMeters
			}

			kdTreeSinkInt64 = sumID
			kdTreeSinkFloat64 = sumDist

			b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "ops/s")
		})
	}
}

func BenchmarkLinearWithinRadius(b *testing.B) {
	queryCount := kdTreeQueryCount()
	radii := kdTreeRadii()

	for _, size := range kdTreeSizes() {
		items := makeKDTreeItems(size)
		queries := makeKDTreeQueries(queryCount)

		for _, radius := range radii {
			b.Run(fmt.Sprintf("size=%d/radius=%.0f", size, radius), func(b *testing.B) {
				b.ReportAllocs()
				b.ResetTimer()

				var total int

				for i := 0; i < b.N; i++ {
					total += linearWithinRadiusCount(items, queries[i%len(queries)], radius)
				}

				kdTreeSinkInt = total

				b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "ops/s")
			})
		}
	}
}

func linearNearest(items []kd.Item, q kd.Point) kd.Neighbor {
	best := kd.Neighbor{
		DistanceMeters: math.Inf(1),
	}

	for _, item := range items {
		distance := haversineDistanceMeters(q, item.Point)
		if distance < best.DistanceMeters ||
			(math.Abs(distance-best.DistanceMeters) <= 1e-12 && item.ID < best.Item.ID) {
			best = kd.Neighbor{
				Item:           item,
				DistanceMeters: distance,
			}
		}
	}

	return best
}

func linearWithinRadiusCount(items []kd.Item, q kd.Point, radiusMeters float64) int {
	count := 0
	for _, item := range items {
		if haversineDistanceMeters(q, item.Point) <= radiusMeters {
			count++
		}
	}
	return count
}

func degToRad(v float64) float64 {
	return v * math.Pi / 180
}

func haversineDistanceMeters(a, b kd.Point) float64 {
	lat1 := degToRad(a.Lat)
	lng1 := degToRad(a.Lng)
	lat2 := degToRad(b.Lat)
	lng2 := degToRad(b.Lng)

	dLat := lat2 - lat1
	dLng := lng2 - lng1

	sinLat := math.Sin(dLat / 2)
	sinLng := math.Sin(dLng / 2)

	h := sinLat*sinLat + math.Cos(lat1)*math.Cos(lat2)*sinLng*sinLng
	if h > 1 {
		h = 1
	}

	c := 2 * math.Asin(math.Sqrt(h))
	return benchEarthRadiusMeters * c
}
