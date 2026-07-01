package stdx

type SliceQueue[T any] []T

func NewSliceQueue[T any]() SliceQueue[T] {
	return make(SliceQueue[T], 0)
}

func (q SliceQueue[T]) Len() int         { return len(q) }
func (q *SliceQueue[T]) PushBack(v T)    { *q = append(*q, v) }
func (q *SliceQueue[T]) PopBack() (v T)  { v, *q = (*q)[len(*q)-1], (*q)[:len(*q)-1]; return }
func (q *SliceQueue[T]) PopFront() (v T) { v, *q = (*q)[0], (*q)[1:]; return }
