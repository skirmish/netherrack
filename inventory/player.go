package inventory

import (
	"github.com/NetherrackDev/soulsand"
)

var _ soulsand.PlayerInventory = &PlayerInventory{}

type PlayerInventory struct {
	Type
}

func CreatePlayerInventory() *PlayerInventory {
	return &PlayerInventory{
		Type: Type{
			items:    make([]soulsand.ItemStack, 45),
			Id:       -1,
			watchers: make(map[string]soulsand.Player),
		},
	}
}

func (pi *PlayerInventory) SetSlot(slot int, item soulsand.ItemStack) {
	pi.lock.Lock()
	defer pi.lock.Unlock()
	pi.items[slot] = item
	pi.watcherLock.RLock()
	defer pi.watcherLock.RUnlock()
	for _, p := range pi.watchers {
		p.RunSync(func(se soulsand.SyncEntity) {
			sp := se.(soulsand.SyncPlayer)
			sp.Connection().WriteSetSlot(0, int16(slot), item)
		})
	}
}

func (pi *PlayerInventory) AddWatcher(p soulsand.Player) {
	pi.watcherLock.Lock()
	defer pi.watcherLock.Unlock()
	pi.watchers[p.Name()] = p
	pi.lock.RLock()
	defer pi.lock.RUnlock()
	p.RunSync(func(se soulsand.SyncEntity) {
		sp := se.(soulsand.SyncPlayer)
		sp.Connection().WriteSetWindowItems(0, pi.items)
	})
}

func (pi *PlayerInventory) CraftingOutput() soulsand.ItemStack {
	return pi.Slot(0)
}

func (pi *PlayerInventory) SetCraftingOutput(item soulsand.ItemStack) {
	pi.SetSlot(0, item)
}

func (pi *PlayerInventory) CraftingInput(x, y int) soulsand.ItemStack {
	return pi.Slot(1 + x + y*2)
}

func (pi *PlayerInventory) SetCraftingInput(x, y int, item soulsand.ItemStack) {
	pi.SetSlot(1+x+y*2, item)
}

func (pi *PlayerInventory) ArmourHead() soulsand.ItemStack {
	return pi.Slot(5)
}

func (pi *PlayerInventory) SetArmourHead(item soulsand.ItemStack) {
	pi.SetSlot(5, item)
}

func (pi *PlayerInventory) ArmourChest() soulsand.ItemStack {
	return pi.Slot(6)
}

func (pi *PlayerInventory) SetArmourChest(item soulsand.ItemStack) {
	pi.SetSlot(6, item)
}

func (pi *PlayerInventory) ArmourLegs() soulsand.ItemStack {
	return pi.Slot(7)
}

func (pi *PlayerInventory) SetArmourLegs(item soulsand.ItemStack) {
	pi.SetSlot(7, item)
}

func (pi *PlayerInventory) ArmourFeet() soulsand.ItemStack {
	return pi.Slot(8)
}

func (pi *PlayerInventory) SetArmourFeet(item soulsand.ItemStack) {
	pi.SetSlot(8, item)
}

func (pi *PlayerInventory) PlayerInventorySlot(slot int) soulsand.ItemStack {
	return pi.Slot(9 + slot)
}

func (pi *PlayerInventory) SetPlayerInventorySlot(slot int, item soulsand.ItemStack) {
	pi.SetSlot(9+slot, item)
}

func (pi *PlayerInventory) PlayerInventorySize() int {
	return 27
}

func (pi *PlayerInventory) HotbarSlot(slot int) soulsand.ItemStack {
	return pi.Slot(36 + slot)
}

func (pi *PlayerInventory) SetHotbarSlot(slot int, item soulsand.ItemStack) {
	pi.SetSlot(36+slot, item)
}

func (pi *PlayerInventory) HotbarSize() int {
	return 9
}
