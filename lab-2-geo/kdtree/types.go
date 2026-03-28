package kdtree

const (
	minEarthLat = -90.0
	maxEarthLat = 90.0
	minEarthLng = -180.0
	maxEarthLng = 180.0
)

const (
	axisLat axis = iota
	axisLng
)

type Point struct {
	Lat float64
	Lng float64
}

type Item struct {
	ID    int64
	Point Point
}

type Neighbor struct {
	Item           Item
	DistanceMeters float64
}

type axis uint8

type node struct {
	item  Item
	axis  axis
	left  *node
	right *node

	minLat float64
	maxLat float64
	minLng float64
	maxLng float64
}

type Tree struct {
	root *node
	size int
}
