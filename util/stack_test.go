package util

import (
	"testing"
)

func TestStackNoOverflow(t *testing.T) {
	s := NewStack(20)
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

func TestStackOverflow1(t *testing.T) {
	s := NewStack(10)
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

func TestStackOverflow4(t *testing.T) {
	s := NewStack(5)
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

func BenchmarkStackNoOverflowPush(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := NewStack(200)
		for j := 0; j < 200; j++ {
			s.Push(j)
		}
	}
}

func BenchmarkStackNoOverflowPop(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := NewStack(200)
		b.StopTimer()
		for j := 0; j < 200; j++ {
			s.Push(j)
		}
		b.StartTimer()
		for j := 0; j < 200; j++ {
			_ = s.Pop().(int)
		}
	}
}

func BenchmarkStackOverflowPush(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := NewStack(50)
		for j := 0; j < 200; j++ {
			s.Push(j)
		}
	}
}

func BenchmarkStackOverflowPop(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := NewStack(50)
		b.StopTimer()
		for j := 0; j < 200; j++ {
			s.Push(j)
		}
		b.StartTimer()
		for j := 0; j < 200; j++ {
			_ = s.Pop().(int)
		}
	}
}
