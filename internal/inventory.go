package internal

import (
	"github.com/thinkofdeath/soulsand"
)

type Inventory interface {
	soulsand.Inventory
	GetWindowType() int8
}
