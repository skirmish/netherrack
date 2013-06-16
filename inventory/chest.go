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
			Name:     name,
			watchers: make(map[string]soulsand.Player),
		},
		chestSlots: size,
	}
}

func (ci *ChestInventory) GetInventorySlot(slot int) soulsand.ItemStack {
	return ci.GetSlot(slot)
}

func (ci *ChestInventory) SetInventorySlot(slot int, item soulsand.ItemStack) {
	ci.SetSlot(slot, item)
}

func (ci *ChestInventory) GetInventorySize() int {
	return ci.chestSlots
}
