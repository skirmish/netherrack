package internal

import (
	"github.com/thinkofdeath/soulsand"
)

type Event interface {
	soulsand.Event

	IsCanceled() bool
	Set(eventType string, source soulsand.EventSource)
}
