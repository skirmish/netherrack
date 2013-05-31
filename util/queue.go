package util

import ()

type Queue interface {
	Add(interface{})
	Remove() interface{}
	IsEmpty() bool
}

type queue struct {
	head *node
	next *node
}

type node struct {
	value interface{}
	next  *node
}

func NewQueue() Queue {
	return &queue{}
}

func (q *queue) Add(value interface{}) {
	n := &node{value: value}
	if q.next == nil {
		q.next = n
	}
	if q.head != nil {
		q.head.next = n
	}
	q.head = n
}

func (q *queue) Remove() interface{} {
	val := q.next
	q.next = q.next.next
	if q.next == nil {
		q.head = nil
	}
	return val.value
}

func (q *queue) IsEmpty() bool {
	return q.next == nil
}
