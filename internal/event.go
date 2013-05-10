package internal

import (
	"github.com/NetherrackDev/soulsand"
)

type Event interface {
	soulsand.Event

	IsCanceled() bool
	Set(eventType string, source soulsand.EventSource)
}
