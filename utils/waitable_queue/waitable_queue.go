package waitable_queue

import "sync"

type WaitableQueue[T any] struct {
	later       chan T
	mu          sync.Mutex
	handleFirst []T
}

func NewWaitableQueue[T any]() *WaitableQueue[T] {
	return &WaitableQueue[T]{
		later:       make(chan T, 1),
		handleFirst: make([]T, 0),
		mu:          sync.Mutex{},
	}
}

func (q *WaitableQueue[T]) Put(data T) {
	q.mu.Lock()
	defer q.mu.Unlock()
	select {
	case notHandled := <-q.later:
		q.handleFirst = append(q.handleFirst, notHandled)
	default:
	}
	q.later <- data
}

func (q *WaitableQueue[T]) Get() (data T) {
	q.mu.Lock()
	if len(q.handleFirst) > 0 {
		e := q.handleFirst[0] 
		q.handleFirst = q.handleFirst[1:]
		q.mu.Unlock()
		return e
	}
	q.mu.Unlock()
	return <-q.later
}

func (q *WaitableQueue[T]) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.handleFirst = make([]T, 0)
	select {
	case <-q.later:
	default:
	}
}
