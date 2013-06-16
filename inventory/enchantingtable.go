package inventory

import (
	"github.com/NetherrackDev/soulsand"
)

type EnchantingTableInventory struct {
	Type
}

func CreateEnchantingTableInventory(name string) *EnchantingTableInventory {
	return &EnchantingTableInventory{
		Type: Type{
			items:    make([]soulsand.ItemStack, 1),
			Id:       4,
			Name:     name,
			watchers: make(map[string]soulsand.Player),
		},
	}
}

func (eti *EnchantingTableInventory) GetItem() soulsand.ItemStack {
	return eti.GetSlot(0)
}

func (eti *EnchantingTableInventory) SetItem(item soulsand.ItemStack) {
	eti.SetSlot(0, item)
}
