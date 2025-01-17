package priority_queue

// Credit to
// https://github.com/gansidui/priority_queue

import (
	"container/heap"
)

type Interface interface {
	Less(other interface{}) bool
}

type sorter []Interface

// Implement heap.Interface: Push, Pop, Len, Less, Swap
func (s *sorter) Push(x interface{}) {
	*s = append(*s, x.(Interface))
}

func (s *sorter) Pop() interface{} {
	n := len(*s)
	if n > 0 {
		x := (*s)[n-1]
		*s = (*s)[0 : n-1]
		return x
	}
	return nil
}

func (s *sorter) Len() int {
	return len(*s)
}

func (s *sorter) Less(i, j int) bool {
	return (*s)[i].Less((*s)[j])
}

func (s *sorter) Swap(i, j int) {
	(*s)[i], (*s)[j] = (*s)[j], (*s)[i]
}

// Define priority queue struct
type PQ struct {
	s *sorter
}

func New() *PQ {
	q := &PQ{s: new(sorter)}
	heap.Init(q.s)
	return q
}

func (q *PQ) Push(x Interface) {
	heap.Push(q.s, x)
}

func (q *PQ) Pop() Interface {
	return heap.Pop(q.s).(Interface)
}

func (q *PQ) Top() Interface {
	if len(*q.s) > 0 {
		return (*q.s)[0].(Interface)
	}
	return nil
}

func (q *PQ) Fix(x Interface, i int) {
	(*q.s)[i] = x
	heap.Fix(q.s, i)
}

func (q *PQ) Remove(i int) Interface {
	return heap.Remove(q.s, i).(Interface)
}

func (q *PQ) Len() int {
	return q.s.Len()
}
