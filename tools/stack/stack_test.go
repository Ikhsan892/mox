package stack_test

import (
	"testing"

	"mox/tools/stack"
	"github.com/stretchr/testify/assert"
)

func TestNewStack(t *testing.T) {

	s := stack.New[int]()

	assert.NotNil(t, s)
}

func TestPushStack(t *testing.T) {
	s := stack.New[int]()

	s.Push(10)
	s.Push(11)
	s.Push(12)
	s.Push(13)

	assert.Equal(t, s.Len(), 4)
}

func TestPopStack(t *testing.T) {
	s := stack.New[int]()

	s.Push(10)
	s.Push(11)

	// now popping
	s.Pop()

	assert.Equal(t, s.Len(), 1)
	assert.Equal(t, s.Peek(), 10)
}

func TestPopByLength(t *testing.T) {
	s := stack.New[int]()

	s.Push(10)
	s.Push(11)

	// now popping by length
	res := s.PopByLength(s.Len())

	assert.EqualValues(t, []int{11, 10}, res)
}
