package stack

import "sync"

type Stack[T any] struct {
	data []T
	mu   *sync.Mutex
}

func New[T any]() *Stack[T] {
	return &Stack[T]{
		mu: &sync.Mutex{},
	}
}

func (stack *Stack[T]) Push(data T) {
	stack.mu.Lock()
	stack.data = append(stack.data, data)
	stack.mu.Unlock()
}

func (stack *Stack[T]) Pop() T {
	stack.mu.Lock()
	last := stack.data[len(stack.data)-1]
	stack.data = stack.data[:len(stack.data)-1]
	stack.mu.Unlock()

	return last
}

func (stack *Stack[T]) PopByLength(len int) []T {
	result := make([]T, 0)

	for i := 0; i < len; i++ {
		result = append(result, stack.Pop())
	}

	return result
}

func (stack *Stack[T]) Peek() T {
	stack.mu.Lock()
	peek := stack.data[0]
	stack.mu.Unlock()

	return peek
}

func (stack *Stack[T]) Len() int {

	l2 := *stack

	stack.mu.Lock()
	l2.data = make([]T, len(stack.data))
	copy(l2.data, stack.data)
	stack.mu.Unlock()

	return len(l2.data)
}
