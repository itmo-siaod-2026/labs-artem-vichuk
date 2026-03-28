package kdtree

import (
	"fmt"
	"sort"
)

type nearestResult struct {
	neighbor Neighbor
	ok       bool
}

func prepareQuery(q Point) (Point, error) {
	q.Lng = normalizeLng(q.Lng)
	if !q.Valid() {
		return Point{}, fmt.Errorf("invalid query point: lat=%v lng=%v", q.Lat, q.Lng)
	}
	return q, nil
}

func (n *node) orderedChildren(q Point) (near, far *node) {
	if n == nil {
		return nil, nil
	}

	qv := n.axis.value(q)
	nv := n.axis.value(n.item.Point)

	if qv < nv {
		return n.left, n.right
	}
	return n.right, n.left
}

func chooseBetter(best nearestResult, candidate Neighbor) nearestResult {
	if !best.ok {
		return nearestResult{neighbor: candidate, ok: true}
	}

	if candidate.DistanceMeters < best.neighbor.DistanceMeters {
		return nearestResult{neighbor: candidate, ok: true}
	}
	if candidate.DistanceMeters > best.neighbor.DistanceMeters {
		return best
	}

	if candidate.Item.ID < best.neighbor.Item.ID {
		return nearestResult{neighbor: candidate, ok: true}
	}

	return best
}

func (n *node) nearest(q Point, best nearestResult) nearestResult {
	if n == nil {
		return best
	}

	if best.ok && n.minPossibleDistanceMeters(q) > best.neighbor.DistanceMeters {
		return best
	}

	candidate := Neighbor{
		Item:           n.item,
		DistanceMeters: haversineDistanceMeters(q, n.item.Point),
	}
	best = chooseBetter(best, candidate)

	near, far := n.orderedChildren(q)

	best = near.nearest(q, best)

	if far != nil && (!best.ok || far.minPossibleDistanceMeters(q) <= best.neighbor.DistanceMeters) {
		best = far.nearest(q, best)
	}

	return best
}

func (t *Tree) Nearest(q Point) (Neighbor, bool, error) {
	q, err := prepareQuery(q)
	if err != nil {
		return Neighbor{}, false, err
	}

	if t == nil || t.root == nil {
		return Neighbor{}, false, nil
	}

	res := t.root.nearest(q, nearestResult{})
	return res.neighbor, res.ok, nil
}

func (n *node) withinRadius(q Point, radiusMeters float64, out *[]Neighbor) {
	if n == nil {
		return
	}

	if n.minPossibleDistanceMeters(q) > radiusMeters {
		return
	}

	d := haversineDistanceMeters(q, n.item.Point)
	if d <= radiusMeters {
		*out = append(*out, Neighbor{
			Item:           n.item,
			DistanceMeters: d,
		})
	}

	near, far := n.orderedChildren(q)

	near.withinRadius(q, radiusMeters, out)

	if far != nil && far.minPossibleDistanceMeters(q) <= radiusMeters {
		far.withinRadius(q, radiusMeters, out)
	}
}

func (t *Tree) WithinRadius(q Point, radiusMeters float64) ([]Neighbor, error) {
	if radiusMeters < 0 {
		return nil, fmt.Errorf("radiusMeters must be >= 0")
	}

	q, err := prepareQuery(q)
	if err != nil {
		return nil, err
	}

	if t == nil || t.root == nil {
		return nil, nil
	}

	out := make([]Neighbor, 0)
	t.root.withinRadius(q, radiusMeters, &out)

	sort.Slice(out, func(i, j int) bool {
		if out[i].DistanceMeters < out[j].DistanceMeters {
			return true
		}
		if out[i].DistanceMeters > out[j].DistanceMeters {
			return false
		}
		return out[i].Item.ID < out[j].Item.ID
	})

	return out, nil
}
