package event

import (
	"github.com/NetherrackDev/netherrack/internal"
	"github.com/NetherrackDev/soulsand"
	"reflect"
	"sync"
)

//Compile time checks
var _ soulsand.EventSource = &Source{}
var _ internal.Event = &Event{}
var _ soulsand.Event = &Event{}

type Source struct {
	handlers       map[string]map[int]reflect.Value
	handlersIntMap map[int]string
	handlersLock   sync.Mutex
	handlePos      int
}

func (es *Source) Init() {
	es.handlers = make(map[string]map[int]reflect.Value)
	es.handlersIntMap = make(map[int]string)
}

func (es *Source) Register(callback interface{}) int {
	fun := reflect.ValueOf(callback)
	t := fun.Type()
	if t.NumIn() != 1 {
		panic("Event: Must have one argument")
	}
	eventType := t.In(0).Name()
	es.handlersLock.Lock()
	defer es.handlersLock.Unlock()
	m, ok := es.handlers[eventType]
	if !ok {
		m = make(map[int]reflect.Value)
		es.handlers[eventType] = m
	}
	m[es.handlePos] = fun
	es.handlePos++
	return es.handlePos - 1
}

func (es *Source) Unregister(id int) {
	es.handlersLock.Lock()
	defer es.handlersLock.Unlock()
	eventType, ok := es.handlersIntMap[id]
	if !ok {
		return
	}
	m, ok := es.handlers[eventType]
	delete(m, id)
	delete(es.handlersIntMap, id)
}

func (es *Source) Fire(eventType string, event internal.Event) bool {
	es.handlersLock.Lock()
	defer es.handlersLock.Unlock()
	event.Set(eventType, es)
	if m, ok := es.handlers[eventType]; ok {
		eValue := []reflect.Value{reflect.ValueOf(event)}
		for _, cb := range m {
			cb.Call(eValue)
		}
	}
	return event.IsCanceled()
}

type Event struct {
	canceled  bool
	source    *Source
	eventType string
}

func (e *Event) Cancel() {
	e.canceled = true
}

func (e *Event) IsCanceled() bool {
	return e.canceled
}

func (e *Event) Set(eventType string, source soulsand.EventSource) {
	e.source = source.(*Source)
	e.eventType = eventType
}

func (e *Event) Remove(id int) {
	m, ok := e.source.handlers[e.eventType]
	if !ok {
		return
	}
	delete(m, id)
	delete(e.source.handlersIntMap, id)
}
