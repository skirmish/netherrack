package event

import (
	"Netherrack/internal"
	"Soulsand"
	soulevent "Soulsand/event"
	"sync"
)

//Compile time checks
var _ Soulsand.EventSource = &Source{}
var _ internal.Event = &Event{}

type Source struct {
	handlers     map[soulevent.Type]map[int]chan Soulsand.Event
	handlersLock sync.RWMutex
	handlePos    int
}

func (es *Source) Init() {
	es.handlers = make(map[soulevent.Type]map[int]chan Soulsand.Event)
}

func (es *Source) Register(eventType soulevent.Type, callback chan Soulsand.Event) int {
	es.handlersLock.Lock()
	defer es.handlersLock.Unlock()
	m, ok := es.handlers[eventType]
	if !ok {
		m = make(map[int]chan Soulsand.Event)
		es.handlers[eventType] = m
	}
	m[es.handlePos] = callback
	es.handlePos++
	return es.handlePos - 1
}

func (es *Source) Unregister(eventType soulevent.Type, id int) {
	es.handlersLock.Lock()
	defer es.handlersLock.Unlock()
	m, ok := es.handlers[eventType]
	if !ok {
		return
	}
	delete(m, id)
}

func (es *Source) Fire(eventType soulevent.Type, event internal.Event) bool {
	es.handlersLock.RLock()
	defer es.handlersLock.RUnlock()
	event.Set(eventType, es)
	if m, ok := es.handlers[eventType]; ok {
		for _, cb := range m {
			event.Add()
			cb <- event
			event.Wait()
		}
	}
	return event.IsCanceled()
}

type Event struct {
	canceled  bool
	waitGroup sync.WaitGroup
	source *Source
	eventType soulevent.Type
}

func (e *Event) Wait() {
	e.waitGroup.Wait()
}

func (e *Event) Add() {
	e.waitGroup.Add(1)
}

func (e *Event) Done() {
	e.waitGroup.Done()
}

func (e *Event) Cancel() {
	e.canceled = true
}

func (e *Event) IsCanceled() bool {
	return e.canceled
}

func (e *Event) Set(eventType soulevent.Type, source Soulsand.EventSource) {
	e.source = source.(*Source)
	e.eventType = eventType
}

func (e *Event) Remove(id int) {
	m, ok := e.source.handlers[e.eventType]
	if !ok {
		return
	}
	delete(m, id)
}