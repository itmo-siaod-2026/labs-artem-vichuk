package kdtree_test

import (
	kd "bdrs/lab-2-geo/kdtree"
	"fmt"
	"math"
	"math/rand"
	"testing"
)

const testEarthRadiusMeters = 6371008.8

func buildTestTree(tb testing.TB, items []kd.Item) *kd.Tree {
	tb.Helper()

	tree, err := kd.New(items)
	if err != nil {
		tb.Fatalf("build tree: %v", err)
	}
	return tree
}

func ExampleTree_Nearest() {
	tree, err := kd.New([]kd.Item{
		{ID: 1, Point: kd.Point{Lat: 55.751244, Lng: 37.618423}},
		{ID: 2, Point: kd.Point{Lat: 59.934280, Lng: 30.335099}},
		{ID: 3, Point: kd.Point{Lat: 51.507351, Lng: -0.127758}},
	})
	if err != nil {
		panic(err)
	}

	nn, ok, err := tree.Nearest(kd.Point{Lat: 55.7600, Lng: 37.6200})
	if err != nil {
		panic(err)
	}

	fmt.Println(ok)
	fmt.Println(nn.Item.ID)
	fmt.Println(nn.DistanceMeters < 2000)
}

func TestNewKDEmpty(t *testing.T) {
	tree, err := kd.New(nil)
	if err != nil {
		t.Fatalf("NewKD(nil): %v", err)
	}

	if tree == nil {
		t.Fatalf("tree is nil")
	}
	if !tree.Empty() {
		t.Fatalf("expected empty tree")
	}
	if tree.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", tree.Len())
	}
}

func TestNewKDInvalidPoint(t *testing.T) {
	_, err := kd.New([]kd.Item{
		{ID: 1, Point: kd.Point{Lat: 123, Lng: 10}},
	})
	if err == nil {
		t.Fatalf("expected error for invalid point")
	}
}

func TestNormalizeLngOnBuild(t *testing.T) {
	tree := buildTestTree(t, []kd.Item{
		{ID: 1, Point: kd.Point{Lat: 0, Lng: 190}},
	})

	nn, ok, err := tree.Nearest(kd.Point{Lat: 0, Lng: -170})
	if err != nil {
		t.Fatalf("Nearest: %v", err)
	}
	if !ok {
		t.Fatalf("Nearest: not found")
	}
	if nn.Item.ID != 1 {
		t.Fatalf("got id=%d want 1", nn.Item.ID)
	}
}

func TestNearestSingleItem(t *testing.T) {
	tree := buildTestTree(t, []kd.Item{
		{ID: 42, Point: kd.Point{Lat: 10, Lng: 20}},
	})

	nn, ok, err := tree.Nearest(kd.Point{Lat: 10, Lng: 20})
	if err != nil {
		t.Fatalf("Nearest: %v", err)
	}
	if !ok {
		t.Fatalf("Nearest: not found")
	}
	if nn.Item.ID != 42 {
		t.Fatalf("got id=%d want 42", nn.Item.ID)
	}
	if nn.DistanceMeters != 0 {
		t.Fatalf("distance=%v want 0", nn.DistanceMeters)
	}
}

func TestNearestEmptyTree(t *testing.T) {
	tree, err := kd.New(nil)
	if err != nil {
		t.Fatalf("NewKD(nil): %v", err)
	}

	_, ok, err := tree.Nearest(kd.Point{Lat: 0, Lng: 0})
	if err != nil {
		t.Fatalf("Nearest: %v", err)
	}
	if ok {
		t.Fatalf("expected no result")
	}
}

func TestNearestInvalidQuery(t *testing.T) {
	tree := buildTestTree(t, []kd.Item{
		{ID: 1, Point: kd.Point{Lat: 10, Lng: 20}},
	})

	_, _, err := tree.Nearest(kd.Point{Lat: 1000, Lng: 0})
	if err == nil {
		t.Fatalf("expected error for invalid query")
	}
}

func TestNearestKnownPoints(t *testing.T) {
	items := []kd.Item{
		{ID: 1, Point: kd.Point{Lat: 55.751244, Lng: 37.618423}},  // Moscow
		{ID: 2, Point: kd.Point{Lat: 59.934280, Lng: 30.335099}},  // SPb
		{ID: 3, Point: kd.Point{Lat: 56.838011, Lng: 60.597465}},  // Ekb
		{ID: 4, Point: kd.Point{Lat: 43.115536, Lng: 131.885485}}, // Vladivostok
	}

	tree := buildTestTree(t, items)

	tests := []struct {
		name string
		q    kd.Point
		want int64
	}{
		{
			name: "near moscow",
			q:    kd.Point{Lat: 55.76, Lng: 37.62},
			want: 1,
		},
		{
			name: "near spb",
			q:    kd.Point{Lat: 59.93, Lng: 30.31},
			want: 2,
		},
		{
			name: "near ekb",
			q:    kd.Point{Lat: 56.84, Lng: 60.60},
			want: 3,
		},
		{
			name: "near vladivostok",
			q:    kd.Point{Lat: 43.12, Lng: 131.90},
			want: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nn, ok, err := tree.Nearest(tt.q)
			if err != nil {
				t.Fatalf("Nearest: %v", err)
			}
			if !ok {
				t.Fatalf("Nearest: not found")
			}
			if nn.Item.ID != tt.want {
				t.Fatalf("got id=%d want %d", nn.Item.ID, tt.want)
			}
		})
	}
}

func TestNearestTieBreakByID(t *testing.T) {
	tree := buildTestTree(t, []kd.Item{
		{ID: 20, Point: kd.Point{Lat: 10, Lng: 20}},
		{ID: 10, Point: kd.Point{Lat: 10, Lng: 20}},
	})

	nn, ok, err := tree.Nearest(kd.Point{Lat: 10, Lng: 20})
	if err != nil {
		t.Fatalf("Nearest: %v", err)
	}
	if !ok {
		t.Fatalf("Nearest: not found")
	}
	if nn.Item.ID != 10 {
		t.Fatalf("got id=%d want 10", nn.Item.ID)
	}
}

func TestNearestAcrossAntiMeridian(t *testing.T) {
	tree := buildTestTree(t, []kd.Item{
		{ID: 1, Point: kd.Point{Lat: 0, Lng: 179.9}},
		{ID: 2, Point: kd.Point{Lat: 0, Lng: 0}},
	})

	nn, ok, err := tree.Nearest(kd.Point{Lat: 0, Lng: -179.95})
	if err != nil {
		t.Fatalf("Nearest: %v", err)
	}
	if !ok {
		t.Fatalf("Nearest: not found")
	}
	if nn.Item.ID != 1 {
		t.Fatalf("got id=%d want 1", nn.Item.ID)
	}
}

func TestWithinRadiusBasic(t *testing.T) {
	items := []kd.Item{
		{ID: 1, Point: kd.Point{Lat: 55.751244, Lng: 37.618423}},
		{ID: 2, Point: kd.Point{Lat: 55.760000, Lng: 37.620000}},
		{ID: 3, Point: kd.Point{Lat: 59.934280, Lng: 30.335099}},
	}
	tree := buildTestTree(t, items)

	got, err := tree.WithinRadius(kd.Point{Lat: 55.755, Lng: 37.618}, 2000)
	if err != nil {
		t.Fatalf("WithinRadius: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("len(got)=%d want 2", len(got))
	}

	ids := map[int64]bool{}
	for _, n := range got {
		ids[n.Item.ID] = true
		if n.DistanceMeters > 2000 {
			t.Fatalf("neighbor id=%d has distance=%f > 2000", n.Item.ID, n.DistanceMeters)
		}
	}

	if !ids[1] || !ids[2] {
		t.Fatalf("expected ids 1 and 2, got %#v", ids)
	}
}

func TestWithinRadiusZero(t *testing.T) {
	tree := buildTestTree(t, []kd.Item{
		{ID: 1, Point: kd.Point{Lat: 10, Lng: 20}},
		{ID: 2, Point: kd.Point{Lat: 10.0001, Lng: 20.0001}},
	})

	got, err := tree.WithinRadius(kd.Point{Lat: 10, Lng: 20}, 0)
	if err != nil {
		t.Fatalf("WithinRadius: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got)=%d want 1", len(got))
	}
	if got[0].Item.ID != 1 {
		t.Fatalf("got id=%d want 1", got[0].Item.ID)
	}
}

func TestWithinRadiusNegative(t *testing.T) {
	tree := buildTestTree(t, []kd.Item{
		{ID: 1, Point: kd.Point{Lat: 10, Lng: 20}},
	})

	_, err := tree.WithinRadius(kd.Point{Lat: 10, Lng: 20}, -1)
	if err == nil {
		t.Fatalf("expected error for negative radius")
	}
}

func TestWithinRadiusAcrossAntiMeridian(t *testing.T) {
	tree := buildTestTree(t, []kd.Item{
		{ID: 1, Point: kd.Point{Lat: 0, Lng: 179.9}},
		{ID: 2, Point: kd.Point{Lat: 0, Lng: -179.9}},
		{ID: 3, Point: kd.Point{Lat: 0, Lng: 0}},
	})

	got, err := tree.WithinRadius(kd.Point{Lat: 0, Lng: 180}, 30_000)
	if err != nil {
		t.Fatalf("WithinRadius: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("len(got)=%d want 2", len(got))
	}

	ids := map[int64]bool{}
	for _, n := range got {
		ids[n.Item.ID] = true
	}
	if !ids[1] || !ids[2] {
		t.Fatalf("expected ids 1 and 2, got %#v", ids)
	}
}

func TestNearestRandomizedAgainstLinear(t *testing.T) {
	rng := rand.New(rand.NewSource(42))

	for iter := 0; iter < 50; iter++ {
		n := 50 + rng.Intn(200)

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

		tree := buildTestTree(t, items)

		for qn := 0; qn < 100; qn++ {
			q := kd.Point{
				Lat: rng.Float64()*180 - 90,
				Lng: rng.Float64()*360 - 180,
			}

			got, ok, err := tree.Nearest(q)
			if err != nil {
				t.Fatalf("Nearest: %v", err)
			}
			if !ok {
				t.Fatalf("Nearest: not found")
			}

			want := linearNearestTest(items, q)

			if got.Item.ID != want.Item.ID {
				t.Fatalf("query=%+v got id=%d want id=%d", q, got.Item.ID, want.Item.ID)
			}

			if math.Abs(got.DistanceMeters-want.DistanceMeters) > 1e-6 {
				t.Fatalf("query=%+v got dist=%f want dist=%f", q, got.DistanceMeters, want.DistanceMeters)
			}
		}
	}
}

func TestWithinRadiusRandomizedAgainstLinear(t *testing.T) {
	rng := rand.New(rand.NewSource(99))

	for iter := 0; iter < 30; iter++ {
		n := 50 + rng.Intn(200)

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

		tree := buildTestTree(t, items)

		for qn := 0; qn < 50; qn++ {
			q := kd.Point{
				Lat: rng.Float64()*180 - 90,
				Lng: rng.Float64()*360 - 180,
			}
			radius := rng.Float64() * 200000

			got, err := tree.WithinRadius(q, radius)
			if err != nil {
				t.Fatalf("WithinRadius: %v", err)
			}

			wantCount := linearWithinRadiusCountTest(items, q, radius)
			if len(got) != wantCount {
				t.Fatalf("query=%+v radius=%f len(got)=%d want %d", q, radius, len(got), wantCount)
			}
		}
	}
}

func linearNearestTest(items []kd.Item, q kd.Point) kd.Neighbor {
	best := kd.Neighbor{
		DistanceMeters: math.Inf(1),
	}

	for _, it := range items {
		d := testHaversineDistanceMeters(q, it.Point)
		if d < best.DistanceMeters ||
			(math.Abs(d-best.DistanceMeters) <= 1e-12 && it.ID < best.Item.ID) {
			best = kd.Neighbor{
				Item:           it,
				DistanceMeters: d,
			}
		}
	}

	return best
}

func linearWithinRadiusCountTest(items []kd.Item, q kd.Point, radiusMeters float64) int {
	count := 0
	for _, it := range items {
		if testHaversineDistanceMeters(q, it.Point) <= radiusMeters {
			count++
		}
	}
	return count
}

func testDegToRad(v float64) float64 {
	return v * math.Pi / 180.0
}

func testHaversineDistanceMeters(a, b kd.Point) float64 {
	lat1 := testDegToRad(a.Lat)
	lng1 := testDegToRad(a.Lng)
	lat2 := testDegToRad(b.Lat)
	lng2 := testDegToRad(b.Lng)

	dLat := lat2 - lat1
	dLng := lng2 - lng1

	sinLat := math.Sin(dLat / 2)
	sinLng := math.Sin(dLng / 2)

	h := sinLat*sinLat + math.Cos(lat1)*math.Cos(lat2)*sinLng*sinLng
	if h > 1 {
		h = 1
	}

	c := 2 * math.Asin(math.Sqrt(h))
	return testEarthRadiusMeters * c
}

func FuzzNearestSingleItemSelf(f *testing.F) {
	f.Add(0.0, 0.0)
	f.Add(55.751244, 37.618423)
	f.Add(-33.8688, 151.2093)
	f.Add(0.0, 179.999)
	f.Add(0.0, -179.999)

	f.Fuzz(func(t *testing.T, lat, lng float64) {
		if math.IsNaN(lat) || math.IsNaN(lng) {
			t.Skip()
		}
		if math.IsInf(lat, 0) || math.IsInf(lng, 0) {
			t.Skip()
		}
		if lat < -90 || lat > 90 {
			t.Skip()
		}

		item := kd.Item{
			ID: 123,
			Point: kd.Point{
				Lat: lat,
				Lng: lng,
			},
		}

		tree, err := kd.New([]kd.Item{item})
		if err != nil {
			t.Fatalf("New: %v", err)
		}

		nn, ok, err := tree.Nearest(item.Point)
		if err != nil {
			t.Fatalf("Nearest: %v", err)
		}
		if !ok {
			t.Fatalf("Nearest: not found")
		}
		if nn.Item.ID != item.ID {
			t.Fatalf("got id=%d want %d", nn.Item.ID, item.ID)
		}
		if nn.DistanceMeters != 0 {
			t.Fatalf("distance=%v want 0", nn.DistanceMeters)
		}
	})
}

func FuzzWithinRadiusContainsSelf(f *testing.F) {
	f.Add(0.0, 0.0, 0.0)
	f.Add(55.751244, 37.618423, 1.0)
	f.Add(-33.8688, 151.2093, 1000.0)
	f.Add(0.0, 179.999, 10.0)

	f.Fuzz(func(t *testing.T, lat, lng, radius float64) {
		if math.IsNaN(lat) || math.IsNaN(lng) || math.IsNaN(radius) {
			t.Skip()
		}
		if math.IsInf(lat, 0) || math.IsInf(lng, 0) || math.IsInf(radius, 0) {
			t.Skip()
		}
		if lat < -90 || lat > 90 {
			t.Skip()
		}
		if radius < 0 {
			t.Skip()
		}

		item := kd.Item{
			ID: 1,
			Point: kd.Point{
				Lat: lat,
				Lng: lng,
			},
		}

		tree, err := kd.New([]kd.Item{item})
		if err != nil {
			t.Fatalf("New: %v", err)
		}

		got, err := tree.WithinRadius(item.Point, radius)
		if err != nil {
			t.Fatalf("WithinRadius: %v", err)
		}

		found := false
		for _, n := range got {
			if n.Item.ID == item.ID {
				found = true
				if n.DistanceMeters > radius {
					t.Fatalf("distance=%f > radius=%f", n.DistanceMeters, radius)
				}
			}
		}

		if !found {
			t.Fatalf("self item not found in radius result")
		}
	})
}

func FuzzNewRejectsInvalidLatitude(f *testing.F) {
	f.Add(123.0, 10.0)
	f.Add(-123.0, 10.0)
	f.Add(90.1, 0.0)
	f.Add(-90.1, 0.0)

	f.Fuzz(func(t *testing.T, lat, lng float64) {
		if math.IsNaN(lat) || math.IsNaN(lng) {
			t.Skip()
		}
		if math.IsInf(lat, 0) || math.IsInf(lng, 0) {
			t.Skip()
		}
		if lat >= -90 && lat <= 90 {
			t.Skip()
		}

		_, err := kd.New([]kd.Item{
			{
				ID: 1,
				Point: kd.Point{
					Lat: lat,
					Lng: lng,
				},
			},
		})
		if err == nil {
			t.Fatalf("expected error for invalid latitude lat=%f lng=%f", lat, lng)
		}
	})
}
