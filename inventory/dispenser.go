package inventory

import (
	"github.com/thinkofdeath/soulsand"
)

var _ soulsand.DispenserInventory = &DispenserInventory{}

type DispenserInventory struct {
	Type
}

func CreateDispenserInventory(name string) *DispenserInventory {
	return &DispenserInventory{
		Type: Type{
			items: make([]soulsand.ItemStack, 45),
			Id:    3,
			Name:  name,
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

func (di *DispenserInventory) GetPlayerInventorySlot(slot int) soulsand.ItemStack {
	return di.GetSlot(9 + slot)
}

func (di *DispenserInventory) SetPlayerInventorySlot(slot int, item soulsand.ItemStack) {
	di.SetSlot(9+slot, item)
}

func (di *DispenserInventory) GetPlayerInventorySize() int {
	return 27
}

func (di *DispenserInventory) GetHotbarSlot(slot int) soulsand.ItemStack {
	return di.GetSlot(9 + 27 + slot)
}

func (di *DispenserInventory) SetHotbarSlot(slot int, item soulsand.ItemStack) {
	di.SetSlot(9+27+slot, item)
}

func (di *DispenserInventory) GetHotbarSize() int {
	return 9
}
