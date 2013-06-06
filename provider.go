package netherrack

import (
	"github.com/NetherrackDev/netherrack/entity/creeper"
	"github.com/NetherrackDev/netherrack/inventory"
	"github.com/NetherrackDev/netherrack/items"
	"github.com/NetherrackDev/soulsand"
)

type provider struct{}

func (p provider) CreateChestInventory(name string, size int) soulsand.ChestInventory {
	return inventory.CreateChestInventory(name, size)
}
func (p provider) CreateCraftingInventory() soulsand.CraftingInventory {
	return inventory.CreateCraftingInventory()
}
func (p provider) CreateFurnaceInventory(name string) soulsand.FurnanceInventory {
	return inventory.CreateFurnaceInventory(name)
}
func (p provider) CreateDispenserInventory(name string) soulsand.DispenserInventory {
	return inventory.CreateDispenserInventory(name)
}
func (p provider) CreateEnchantingTableInventory(name string) soulsand.EnchantingTableInventory {
	return inventory.CreateEnchantingTableInventory(name)
}

func (p provider) CreateItemStack(id, data int16, count byte) soulsand.ItemStack {
	return items.CreateItemStack(id, data, count)
}

func (p provider) NewEntityCreeper(x, y, z float64, world soulsand.World) soulsand.EntityCreeper {
	return creeper.New(x, y, z, world)
}
