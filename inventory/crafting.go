package inventory

import (
	"github.com/NetherrackDev/soulsand"
)

var _ soulsand.CraftingInventory = &CraftingInventory{}

type CraftingInventory struct {
	Type
}

func CreateCraftingInventory() *CraftingInventory {
	return &CraftingInventory{
		Type: Type{
			items:    make([]soulsand.ItemStack, 10),
			Id:       1,
			watchers: make(map[string]soulsand.Player),
		},
	}
}

func (ci *CraftingInventory) CraftingOutput() soulsand.ItemStack {
	return ci.Slot(0)
}

func (ci *CraftingInventory) SetCraftingOutput(item soulsand.ItemStack) {
	ci.SetSlot(0, item)
}

func (ci *CraftingInventory) CraftingInput(x, y int) soulsand.ItemStack {
	return ci.Slot(1 + x + y*3)
}

func (ci *CraftingInventory) SetCraftingInput(x, y int, item soulsand.ItemStack) {
	ci.SetSlot(1+x+y*3, item)
}
