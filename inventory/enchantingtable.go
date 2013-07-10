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
			name:     name,
			watchers: make(map[string]soulsand.Player),
		},
	}
}

func (eti *EnchantingTableInventory) Item() soulsand.ItemStack {
	return eti.Slot(0)
}

func (eti *EnchantingTableInventory) SetItem(item soulsand.ItemStack) {
	eti.SetSlot(0, item)
}
