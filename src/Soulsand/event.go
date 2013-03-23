package Soulsand

import (
	"Soulsand/event"
)

type EventSource interface {
	//Regigster the eventType with this EventSource.
	//When the event occurs it will be put into the
	//channel. The callback must call Done on the event
	//on completion. Returns the handlers ID
	Register(eventType event.Type, callback chan Event) int
	//Unregisters the event handler from the event. An event
	//cannot be unregistered during the event, use Event.Remove
	Unregister(eventType event.Type, id int)
}

type Event interface {
	//An event handler must call this on completion.
	//	defer e.Done()
	Done()
	//Trys to cancel the event
	Cancel()
	//Removes the handler from the event.
	Remove(id int)
}
