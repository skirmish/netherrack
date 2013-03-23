package internal

import (
	"Soulsand"
	"Soulsand/event"
)

type Event interface {
	Soulsand.Event

	Wait()
	IsCanceled() bool
	Add()
	Set(eventType event.Type, source Soulsand.EventSource)
}
