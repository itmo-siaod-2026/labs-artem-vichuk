package kdtree

import (
	"fmt"
	"sort"
)

func prepareItems(items []Item) ([]Item, error) {
	if len(items) == 0 {
		return nil, nil
	}

	prepared := make([]Item, len(items))
	for i, it := range items {
		it.Point.Lng = normalizeLng(it.Point.Lng)
		if !it.Point.Valid() {
			return nil, fmt.Errorf(
				"invalid point at index %d (id=%d): lat=%v lng=%v",
				i, it.ID, it.Point.Lat, it.Point.Lng,
			)
		}
		prepared[i] = it
	}

	return prepared, nil
}

func build(items []Item, split axis) *node {
	if len(items) == 0 {
		return nil
	}

	other := nextAxis(split)

	sort.Slice(items, func(i, j int) bool {
		vi := split.value(items[i].Point)
		vj := split.value(items[j].Point)
		if vi < vj {
			return true
		}
		if vi > vj {
			return false
		}

		oi := other.value(items[i].Point)
		oj := other.value(items[j].Point)
		if oi < oj {
			return true
		}
		if oi > oj {
			return false
		}

		return items[i].ID < items[j].ID
	})

	mid := len(items) / 2

	n := newNode(items[mid], split)
	n.left = build(items[:mid], other)
	n.right = build(items[mid+1:], other)
	n.recomputeBounds()

	return n
}

func New(items []Item) (*Tree, error) {
	prepared, err := prepareItems(items)
	if err != nil {
		return nil, err
	}

	if len(prepared) == 0 {
		return &Tree{}, nil
	}

	return &Tree{
		root: build(prepared, axisLat),
		size: len(prepared),
	}, nil
}

func (t *Tree) Len() int {
	if t == nil {
		return 0
	}
	return t.size
}

func (t *Tree) Empty() bool {
	return t == nil || t.size == 0
}
