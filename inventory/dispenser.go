package inventory

import (
	"github.com/NetherrackDev/soulsand"
)

var _ soulsand.DispenserInventory = &DispenserInventory{}

type DispenserInventory struct {
	Type
}

func CreateDispenserInventory(name string) *DispenserInventory {
	return &DispenserInventory{
		Type: Type{
			items:    make([]soulsand.ItemStack, 9),
			Id:       3,
			Name:     name,
			watchers: make(map[string]soulsand.Player),
		},
	}
}

func (di *DispenserInventory) GetInventorySlot(slot int) soulsand.ItemStack {
	return di.GetSlot(slot)
}

func (di *DispenserInventory) SetInventorySlot(slot int, item soulsand.ItemStack) {
	di.SetSlot(slot, item)
}

func (di *DispenserInventory) GetInventorySize() int {
	return 9
}
