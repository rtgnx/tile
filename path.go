// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package tile

type costFn = func(Tile) uint16

// Edge represents an edge of the path
type edge struct {
	Point
	Cost uint32
}

// Around performs a breadth first search around a point.
func (m *Map) Around(from Point, distance uint32, costOf costFn, fn Iterator) {
	start, ok := m.At(from.X, from.Y)
	if !ok {
		return
	}

	fn(from, start)
	frontier := newHeap32()
	frontier.Push(from.Integer(), 0)
	reached := make(map[uint32]struct{}, distance*distance) // TODO: too much here
	reached[from.Integer()] = struct{}{}

	for !frontier.IsEmpty() {
		pCurr, _ := frontier.Pop()
		current := unpackPoint(pCurr)

		// Get all of the neighbors
		m.Neighbors(current.X, current.Y, func(next Point, nextTile Tile) {
			if d := from.DistanceTo(next); d > distance {
				return // Too far
			}

			if cost := costOf(nextTile); cost == 0 {
				return // Blocked tile, ignore completely
			}

			// Add to the search queue
			pNext := next.Integer()
			if _, ok := reached[pNext]; !ok {
				frontier.Push(pNext, 1)
				reached[pNext] = struct{}{}
				fn(next, nextTile)
			}
		})
	}
}

// Path calculates a short path and the distance between the two locations
func (m *Map) Path(from, to Point, costOf costFn) ([]Point, int, bool) {
	frontier := newHeap32()
	frontier.Push(from.Integer(), 0)

	// Add the first edge
	capacity := int(float32(from.DistanceTo(to)) * 1.5)
	edges := make(map[uint32]edge, capacity)
	edges[from.Integer()] = edge{
		Point: from,
		Cost:  0,
	}

	for !frontier.IsEmpty() {
		pCurr, _ := frontier.Pop()
		current := unpackPoint(pCurr)

		// We have a path to the goal
		if current.Equal(to) {
			dist := int(edges[current.Integer()].Cost)
			path := make([]Point, 0, dist)
			curr, _ := edges[current.Integer()]
			for !curr.Point.Equal(from) {
				path = append(path, curr.Point)
				curr = edges[curr.Point.Integer()]
			}

			return path, dist, true
		}

		// Get all of the neighbors
		m.Neighbors(current.X, current.Y, func(next Point, nextTile Tile) {
			cNext := costOf(nextTile)
			if cNext == 0 {
				return // Blocked tile, ignore completely
			}

			pNext := next.Integer()
			newCost := edges[pCurr].Cost + uint32(cNext) // cost(current, next)

			if e, ok := edges[pNext]; !ok || newCost < e.Cost {
				priority := newCost + next.DistanceTo(to) // heuristic
				frontier.Push(next.Integer(), priority)

				edges[pNext] = edge{
					Point: current,
					Cost:  newCost,
				}
			}

		})
	}

	return nil, 0, false
}

// -----------------------------------------------------------------------------

// heapNode represents a ranked node for the heap.
type heapNode struct {
	Value uint32 // The value of the ranked node.
	Rank  uint32 // The rank associated with the ranked node.
}

type heap32 []heapNode

func newHeap32() heap32 {
	return make(heap32, 0, 16)
}

// Push pushes the element x onto the heap.
// The complexity is O(log n) where n = h.Len().
func (h *heap32) Push(v, rank uint32) {
	*h = append(*h, heapNode{
		Value: v,
		Rank:  rank,
	})
	h.up(h.Len() - 1)
}

// Pop removes and returns the minimum element (according to Less) from the heap.
// The complexity is O(log n) where n = h.Len().
// Pop is equivalent to Remove(h, 0).
func (h *heap32) Pop() (uint32, bool) {
	n := h.Len() - 1
	if n < 0 {
		return 0, false
	}

	h.Swap(0, n)
	h.down(0, n)
	return h.pop(), true
}

// Remove removes and returns the element at index i from the heap.
// The complexity is O(log n) where n = h.Len().
func (h *heap32) Remove(i int) uint32 {
	n := h.Len() - 1
	if n != i {
		h.Swap(i, n)
		if !h.down(i, n) {
			h.up(i)
		}
	}
	return h.pop()
}

func (h *heap32) pop() uint32 {
	old := *h
	n := len(old)
	no := old[n-1]
	*h = old[0 : n-1]
	return no.Value
}

func (h *heap32) up(j int) {
	for {
		i := (j - 1) / 2 // parent
		if i == j || !h.Less(j, i) {
			break
		}
		h.Swap(i, j)
		j = i
	}
}

func (h *heap32) down(i0, n int) bool {
	i := i0
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && h.Less(j2, j1) {
			j = j2 // = 2*i + 2  // right child
		}
		if !h.Less(j, i) {
			break
		}
		h.Swap(i, j)
		i = j
	}
	return i > i0
}

func (h heap32) Len() int {
	return len(h)
}

func (h heap32) IsEmpty() bool {
	return len(h) == 0
}

func (h heap32) Less(i, j int) bool {
	return h[i].Rank < h[j].Rank
}

func (h *heap32) Swap(i, j int) {
	(*h)[i], (*h)[j] = (*h)[j], (*h)[i]
}
