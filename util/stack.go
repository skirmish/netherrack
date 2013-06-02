package util

type stack struct {
	head *node
}

type Stack interface {
	Push(interface{})
	Pop() interface{}
	IsEmpty() bool
}

func NewStack() Stack {
	return &stack{}
}

func (s *stack) Push(val interface{}) {
	n := &node{value: val}
	n.link = s.head
	s.head = n
}

func (s *stack) Pop() (val interface{}) {
	val = s.head.value
	old := s.head
	s.head = old.link
	old.link = nil
	return
}

func (s *stack) IsEmpty() bool {
	return s.head == nil
}
