package inventory

import (
	"github.com/thinkofdeath/soulsand"
)

type Type struct {
	items []soulsand.ItemStack
	Id    int8
	Name  string
}

func (inv *Type) GetSlot(slot int) soulsand.ItemStack {
	return inv.items[slot]
}

func (inv *Type) SetSlot(slot int, item soulsand.ItemStack) {
	inv.items[slot] = item
}

func (inv *Type) GetSize() int {
	return len(inv.items)
}

func (inv *Type) GetName() string {
	return inv.Name
}
