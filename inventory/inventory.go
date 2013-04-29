package inventory

import (
	"github.com/thinkofdeath/soulsand"
)

type Type struct {
	items []soulsand.ItemStack
}

func (inv *Type) GetSlot(slot int) soulsand.ItemStack {
	return inv.items[slot]
}

func (inv *Type) GetSize() int {
	return len(inv.items)
}

type PlayerInventory struct {
	Type
}

func (pi *PlayerInventory) GetCraftingOutput() soulsand.ItemStack {
	return pi.GetSlot(0)
}

func (pi *PlayerInventory) GetCraftingInput(x, y int) soulsand.ItemStack {
	return pi.GetSlot(1 + x + y*2)
}

func (pi *PlayerInventory) GetArmourHead() soulsand.ItemStack {
	return pi.GetSlot(5)
}

func (pi *PlayerInventory) GetArmourChest() soulsand.ItemStack {
	return pi.GetSlot(6)
}

func (pi *PlayerInventory) GetArmourLegs() soulsand.ItemStack {
	return pi.GetSlot(7)
}

func (pi *PlayerInventory) GetArmourFeet() soulsand.ItemStack {
	return pi.GetSlot(8)
}

func (pi *PlayerInventory) GetInventorySlot(slot int) soulsand.ItemStack {
	return pi.GetSlot(9 + slot)
}

func (pu *PlayerInventory) GetInventorySlots() int {
	return 27
}

func (pi *PlayerInventory) GetHotbarSlot(slot int) soulsand.ItemStack {
	return pi.GetSlot(36 + slot)
}

func (pi *PlayerInventory) GetHotbarSlots() int {
	return 9
}
