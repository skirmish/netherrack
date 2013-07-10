package inventory

import (
	"github.com/NetherrackDev/netherrack/internal"
	"github.com/NetherrackDev/soulsand"
	"sync"
)

var _ soulsand.Inventory = &Type{}
var _ internal.Inventory = &Type{}

type Type struct {
	lock        sync.RWMutex
	items       []soulsand.ItemStack
	Id          int8
	name        string
	watcherLock sync.RWMutex
	watchers    map[string]soulsand.Player
}

func (inv *Type) WindowType() int8 {
	return inv.Id
}

func (inv *Type) Slot(slot int) soulsand.ItemStack {
	inv.lock.RLock()
	defer inv.lock.RUnlock()
	return inv.items[slot]
}

func (inv *Type) SetSlot(slot int, item soulsand.ItemStack) {
	inv.lock.Lock()
	defer inv.lock.Unlock()
	inv.items[slot] = item
	inv.watcherLock.RLock()
	defer inv.watcherLock.RUnlock()
	for _, p := range inv.watchers {
		p.RunSync(func(se soulsand.SyncEntity) {
			sp := se.(soulsand.SyncPlayer)
			sp.Connection().WriteSetSlot(5, int16(slot), item)
		})
	}
}

func (inv *Type) Size() int {
	return len(inv.items)
}

func (inv *Type) Name() string {
	return inv.name
}

func (inv *Type) AddWatcher(p soulsand.Player) {
	inv.watcherLock.Lock()
	defer inv.watcherLock.Unlock()
	inv.watchers[p.Name()] = p
	inv.lock.RLock()
	defer inv.lock.RUnlock()
	p.RunSync(func(se soulsand.SyncEntity) {
		sp := se.(soulsand.SyncPlayer)
		playerInv := sp.Inventory().(*PlayerInventory)
		playerInv.lock.Lock()
		defer playerInv.lock.Unlock()
		inv.lock.RLock()
		defer inv.lock.RUnlock()
		items := make([]soulsand.ItemStack, len(inv.items)+playerInv.PlayerInventorySize()+playerInv.HotbarSize())
		copy(items, inv.items)
		copy(items[len(inv.items):], playerInv.items[9:])
		sp.Connection().WriteSetWindowItems(5, items)
	})
}

func (inv *Type) RemoveWatcher(p soulsand.Player) {
	inv.watcherLock.Lock()
	defer inv.watcherLock.Unlock()
	delete(inv.watchers, p.Name())
}
