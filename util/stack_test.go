package util

import (
	"testing"
)

func TestStack(t *testing.T) {
	s := NewStack()
	for i := 0; i < 20; i++ {
		s.Push(i)
	}
	if s.IsEmpty() {
		t.Fatal("Stack empty")
	}
	for i := 19; i >= 0; i-- {
		j := s.Pop().(int)
		if i != j {
			t.Fatal("Incorrect return value")
		}
	}
}

func BenchmarkStackPush(b *testing.B) {
	queue := NewStack()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queue.Push(i)
	}
}

func BenchmarkStackRemove(b *testing.B) {
	queue := NewStack()
	for i := 0; i < b.N; i++ {
		queue.Push(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = queue.Pop().(int)
	}
}
