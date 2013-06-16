package inventory

import (
	"github.com/NetherrackDev/soulsand"
)

var _ soulsand.FurnanceInventory = &FurnaceInventory{}

type FurnaceInventory struct {
	Type
}

func CreateFurnaceInventory(name string) *FurnaceInventory {
	return &FurnaceInventory{
		Type: Type{
			items:    make([]soulsand.ItemStack, 3),
			Id:       2,
			Name:     name,
			watchers: make(map[string]soulsand.Player),
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
