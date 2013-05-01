package internal

import (
	"github.com/thinkofdeath/soulsand"
)

type Inventory interface {
	soulsand.Inventory
	GetWindowType() int8
	AddWatcher(p soulsand.Player)
	RemoveWatcher(p soulsand.Player)
}
