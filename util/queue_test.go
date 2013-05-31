package util

import (
	"testing"
)

func TestQueue(t *testing.T) {
	queue := NewQueue()
	for i := 0; i < 50; i++ {
		queue.Add(i)
	}
	if queue.IsEmpty() {
		t.Fatal("Queue empty")
	}
	for i := 0; i < 50; i++ {
		val := queue.Remove().(int)
		if val != i {
			t.Fatal("Incorrect return")
		}
	}
}

func BenchmarkQueueAdd(b *testing.B) {
	queue := NewQueue()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queue.Add(i)
	}
}

func BenchmarkQueueRemove(b *testing.B) {
	queue := NewQueue()
	for i := 0; i < b.N; i++ {
		queue.Add(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = queue.Remove().(int)
	}
}
