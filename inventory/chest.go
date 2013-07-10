package inventory

import (
	"github.com/NetherrackDev/soulsand"
)

var _ soulsand.ChestInventory = &ChestInventory{}

type ChestInventory struct {
	Type
	chestSlots int
}

func CreateChestInventory(name string, size int) *ChestInventory {
	return &ChestInventory{
		Type: Type{
			items:    make([]soulsand.ItemStack, size),
			Id:       0,
			name:     name,
			watchers: make(map[string]soulsand.Player),
		},
		chestSlots: size,
	}
}

func (ci *ChestInventory) InventorySlot(slot int) soulsand.ItemStack {
	return ci.Slot(slot)
}

func (ci *ChestInventory) SetInventorySlot(slot int, item soulsand.ItemStack) {
	ci.SetSlot(slot, item)
}

func (ci *ChestInventory) InventorySize() int {
	return ci.chestSlots
}
