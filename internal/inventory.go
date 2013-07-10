package internal

import (
	"github.com/NetherrackDev/soulsand"
)

type Inventory interface {
	soulsand.Inventory
	WindowType() int8
	AddWatcher(p soulsand.Player)
	RemoveWatcher(p soulsand.Player)
}
