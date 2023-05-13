package utils

import (
	"container/list"
	"sync"
)

type ConditionalQueue[T any] interface {
	Enqueue(T)
	DequeueIf(predicate Predicate[T]) (T, bool)
	Drain()
}

type Predicate[T any] func(T) bool

type ConcurrentConditionalQueue[T any] struct {
	list         *list.List
	enqueueMutex *sync.Mutex
	dequeueMutex *sync.Mutex
}

func NewConcurrentConditionalQueue[T any]() ConditionalQueue[T] {
	return &ConcurrentConditionalQueue[T]{
		list:         list.New(),
		enqueueMutex: &sync.Mutex{},
		dequeueMutex: &sync.Mutex{},
	}
}

// Enqueue locks all enqueue operations and put the item in the back of the list.
// The operation can happen concurrently with Dequeue as Dequeue looks only on the first element,
// except for the case where the first element is the last element then it locks both mutexes
// O(1) /**
func (t *ConcurrentConditionalQueue[T]) Enqueue(item T) {
	if t.list.Len() == 1 {
		t.dequeueMutex.Lock()
		defer t.dequeueMutex.Unlock()
	}

	t.enqueueMutex.Lock()
	defer t.enqueueMutex.Unlock()
	t.list.PushBack(item)
}

// DequeueIf dequeues the top of the list if and only if the condition returns true
// Returns the top element and true in a successful case and nil and false otherwise
// O(1) /**
func (t *ConcurrentConditionalQueue[T]) DequeueIf(condition Predicate[T]) (T, bool) {
	var emptyPtr T

	t.dequeueMutex.Lock()
	defer t.dequeueMutex.Unlock()

	top := t.list.Front()
	if top == nil {
		return emptyPtr, false
	}

	val := top.Value.(T)
	if condition(val) {
		t.list.Remove(top)
		return val, true
	}

	return emptyPtr, false
}

// Drain removes all elements from the list and break when the list is empty.
// O(n) /**
func (t *ConcurrentConditionalQueue[T]) Drain() {
	for {
		_, exists := t.DequeueIf(func(_ T) bool {
			return true
		})

		if !exists {
			break
		}
	}
}
