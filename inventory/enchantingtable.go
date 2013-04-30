package inventory

import (
	"github.com/thinkofdeath/soulsand"
)

type EnchantingTableInventory struct {
	Type
}

func CreateEnchantingTableInventory(name string) *EnchantingTableInventory {
	return &EnchantingTableInventory{
		Type: Type{
			items: make([]soulsand.ItemStack, 39),
			Id:    4,
			Name:  name,
		},
	}
}

func (eti *EnchantingTableInventory) GetItem() soulsand.ItemStack {
	return eti.GetSlot(0)
}

func (eti *EnchantingTableInventory) SetItem(item soulsand.ItemStack) {
	eti.SetSlot(0, item)
}

func (eti *EnchantingTableInventory) GetInventorySlot(slot int) soulsand.ItemStack {
	return eti.GetSlot(1 + slot)
}

func (eti *EnchantingTableInventory) SetInventorySlot(slot int, item soulsand.ItemStack) {
	eti.SetSlot(1+slot, item)
}

func (pu *EnchantingTableInventory) GetInventorySize() int {
	return 27
}

func (eti *EnchantingTableInventory) GetHotbarSlot(slot int) soulsand.ItemStack {
	return eti.GetSlot(28 + slot)
}

func (eti *EnchantingTableInventory) SetHotbarSlot(slot int, item soulsand.ItemStack) {
	eti.SetSlot(28+slot, item)
}

func (eti *EnchantingTableInventory) GetHotbarSize() int {
	return 9
}
