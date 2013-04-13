package internal

import (
	"bitbucket.org/Thinkofdeath/soulsand"
	"bitbucket.org/Thinkofdeath/soulsand/event"
)

type Event interface {
	soulsand.Event

	Wait()
	IsCanceled() bool
	Add()
	Set(eventType event.Type, source soulsand.EventSource)
}
