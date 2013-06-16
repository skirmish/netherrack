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

func (ci *CraftingInventory) GetCraftingOutput() soulsand.ItemStack {
	return ci.GetSlot(0)
}

func (ci *CraftingInventory) SetCraftingOutput(item soulsand.ItemStack) {
	ci.SetSlot(0, item)
}

func (ci *CraftingInventory) GetCraftingInput(x, y int) soulsand.ItemStack {
	return ci.GetSlot(1 + x + y*3)
}

func (ci *CraftingInventory) SetCraftingInput(x, y int, item soulsand.ItemStack) {
	ci.SetSlot(1+x+y*3, item)
}
