package kdtree

import "math"

const earthRadiusMeters = 6371008.8

func (p Point) Valid() bool {
	if math.IsNaN(p.Lat) || math.IsNaN(p.Lng) {
		return false
	}
	if math.IsInf(p.Lat, 0) || math.IsInf(p.Lng, 0) {
		return false
	}

	return p.Lat >= minEarthLat && p.Lat <= maxEarthLat &&
		p.Lng >= minEarthLng && p.Lng <= maxEarthLng
}

func normalizeLng(lng float64) float64 {
	if math.IsNaN(lng) || math.IsInf(lng, 0) {
		return lng
	}

	lng = math.Mod(lng+180.0, 360.0)
	if lng < 0 {
		lng += 360.0
	}
	return lng - 180.0
}

func (a axis) value(p Point) float64 {
	if a == axisLat {
		return p.Lat
	}
	return p.Lng
}

func nextAxis(a axis) axis {
	if a == axisLat {
		return axisLng
	}
	return axisLat
}

func degToRad(v float64) float64 {
	return v * math.Pi / 180.0
}

func haversineDistanceMeters(a, b Point) float64 {
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
	return earthRadiusMeters * c
}

func newNode(item Item, split axis) *node {
	return &node{
		item:   item,
		axis:   split,
		minLat: item.Point.Lat,
		maxLat: item.Point.Lat,
		minLng: item.Point.Lng,
		maxLng: item.Point.Lng,
	}
}

func mergeChildBounds(n *node, child *node) {
	if child.minLat < n.minLat {
		n.minLat = child.minLat
	}
	if child.maxLat > n.maxLat {
		n.maxLat = child.maxLat
	}
	if child.minLng < n.minLng {
		n.minLng = child.minLng
	}
	if child.maxLng > n.maxLng {
		n.maxLng = child.maxLng
	}
}

func (n *node) recomputeBounds() {
	if n == nil {
		return
	}

	p := n.item.Point
	n.minLat, n.maxLat = p.Lat, p.Lat
	n.minLng, n.maxLng = p.Lng, p.Lng

	if n.left != nil {
		mergeChildBounds(n, n.left)
	}
	if n.right != nil {
		mergeChildBounds(n, n.right)
	}
}

// Консервативная нижняя оценка расстояния от точки до поддерева.
// Используем только широту, чтобы pruning оставался корректным.
// Это безопасно, но не максимально агрессивно около anti-meridian.
func latitudeIntervalLowerBoundMeters(lat, minLat, maxLat float64) float64 {
	switch {
	case lat < minLat:
		return earthRadiusMeters * degToRad(minLat-lat)
	case lat > maxLat:
		return earthRadiusMeters * degToRad(lat-maxLat)
	default:
		return 0
	}
}

func (n *node) minPossibleDistanceMeters(q Point) float64 {
	if n == nil {
		return math.Inf(1)
	}
	return latitudeIntervalLowerBoundMeters(q.Lat, n.minLat, n.maxLat)
}
