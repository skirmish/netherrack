package inventory

import (
	"github.com/thinkofdeath/netherrack/internal"
	"github.com/thinkofdeath/soulsand"
	"sync"
)

var _ soulsand.Inventory = &Type{}
var _ internal.Inventory = &Type{}

type Type struct {
	lock        sync.RWMutex
	items       []soulsand.ItemStack
	Id          int8
	Name        string
	watcherLock sync.RWMutex
	watchers    map[string]soulsand.Player
}

func (inv *Type) GetWindowType() int8 {
	return inv.Id
}

func (inv *Type) GetSlot(slot int) soulsand.ItemStack {
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
			sp.GetConnection().WriteSetSlot(5, int16(slot), item)
		})
	}
}

func (inv *Type) GetSize() int {
	return len(inv.items)
}

func (inv *Type) GetName() string {
	return inv.Name
}

func (inv *Type) AddWatcher(p soulsand.Player) {
	inv.watcherLock.Lock()
	defer inv.watcherLock.Unlock()
	inv.watchers[p.GetName()] = p
	inv.lock.RLock()
	defer inv.lock.RUnlock()
	p.RunSync(func(se soulsand.SyncEntity) {
		sp := se.(soulsand.SyncPlayer)
		sp.GetConnection().WriteSetWindowItems(5, inv.items)
	})
}

func (inv *Type) RemoveWatcher(p soulsand.Player) {
	inv.watcherLock.Lock()
	defer inv.watcherLock.Unlock()
	delete(inv.watchers, p.GetName())
}
