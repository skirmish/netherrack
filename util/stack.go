package util

type stack struct {
	data    [][]interface{}
	gaps    int
	pointer int
}

type Stack interface {
	Push(interface{})
	Pop() interface{}
	IsEmpty() bool
}

func NewStack(gaps int) Stack {
	s := &stack{
		data: make([][]interface{}, 1),
		gaps: gaps,
	}
	s.data[0] = make([]interface{}, gaps)
	return s
}

func (s *stack) Push(val interface{}) {
	if s.pointer == len(s.data)*s.gaps {
		old := s.data
		s.data = make([][]interface{}, len(old)+1)
		copy(s.data, old)
		s.data[len(old)] = make([]interface{}, s.gaps)
	}
	s.data[s.pointer/s.gaps][s.pointer%s.gaps] = val
	s.pointer++
}

func (s *stack) Pop() (val interface{}) {
	s.pointer--
	val = s.data[s.pointer/s.gaps][s.pointer%s.gaps]
	if s.pointer == len(s.data)*s.gaps {
		old := s.data
		s.data = make([][]interface{}, len(old)-1)
		copy(s.data, old)
	}
	return
}

func (s *stack) IsEmpty() bool {
	return s.pointer == 0
}
