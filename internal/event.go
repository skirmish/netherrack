package internal

import (
	"bitbucket.org/Thinkofdeath/soulsand"
)

type Event interface {
	soulsand.Event

	IsCanceled() bool
	Set(eventType string, source soulsand.EventSource)
}
