package inventory

import (
	"github.com/thinkofdeath/soulsand"
)

type FurnaceInventory struct {
	Type
}

func CreateFurnaceInventory(name string) *FurnaceInventory {
	return &FurnaceInventory{
		Type: Type{
			items: make([]soulsand.ItemStack, 39),
			Id:    2,
			Name:  name,
		},
	}
}

func (fi *FurnaceInventory) GetOutput() soulsand.ItemStack {
	return fi.GetSlot(0)
}

func (fi *FurnaceInventory) SetOutput(item soulsand.ItemStack) {
	fi.SetSlot(0, item)
}

func (fi *FurnaceInventory) GetFuel() soulsand.ItemStack {
	return fi.GetSlot(1)
}

func (fi *FurnaceInventory) SetFuel(item soulsand.ItemStack) {
	fi.SetSlot(1, item)
}

func (fi *FurnaceInventory) GetInput() soulsand.ItemStack {
	return fi.GetSlot(2)
}

func (fi *FurnaceInventory) SetInput(item soulsand.ItemStack) {
	fi.SetSlot(2, item)
}

func (fi *FurnaceInventory) GetInventorySlot(slot int) soulsand.ItemStack {
	return fi.GetSlot(3 + slot)
}

func (fi *FurnaceInventory) SetInventorySlot(slot int, item soulsand.ItemStack) {
	fi.SetSlot(3+slot, item)
}

func (pu *FurnaceInventory) GetInventorySize() int {
	return 27
}

func (fi *FurnaceInventory) GetHotbarSlot(slot int) soulsand.ItemStack {
	return fi.GetSlot(30 + slot)
}

func (fi *FurnaceInventory) SetHotbarSlot(slot int, item soulsand.ItemStack) {
	fi.SetSlot(30+slot, item)
}

func (fi *FurnaceInventory) GetHotbarSize() int {
	return 9
}
