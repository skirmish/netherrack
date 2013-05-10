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
			items:    make([]soulsand.ItemStack, size+27+9),
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

func (ci *ChestInventory) GetPlayerInventorySlot(slot int) soulsand.ItemStack {
	return ci.GetSlot(ci.chestSlots + slot)
}

func (ci *ChestInventory) SetPlayerInventorySlot(slot int, item soulsand.ItemStack) {
	ci.SetSlot(ci.chestSlots+slot, item)
}

func (ci *ChestInventory) GetPlayerInventorySize() int {
	return 27
}

func (ci *ChestInventory) GetHotbarSlot(slot int) soulsand.ItemStack {
	return ci.GetSlot(ci.chestSlots + 27 + slot)
}

func (ci *ChestInventory) SetHotbarSlot(slot int, item soulsand.ItemStack) {
	ci.SetSlot(ci.chestSlots+27+slot, item)
}

func (ci *ChestInventory) GetHotbarSize() int {
	return 9
}
