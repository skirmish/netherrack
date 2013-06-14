package items

import (
	"github.com/NetherrackDev/netherrack/metadata"
	"github.com/NetherrackDev/netherrack/nbt"
	"github.com/NetherrackDev/soulsand"
	"sync"
)

//Compile time checks
var _ soulsand.ItemStack = &ItemStack{}

func CreateItemStack(id, data int16, count byte) *ItemStack {
	return &ItemStack{
		id:     id,
		damage: data,
		count:  count,
	}
}

type ItemStack struct {
	Lock   sync.RWMutex
	id     int16
	count  byte
	damage int16
	Tag    nbt.Type
	metadata.Storage
}

func (i *ItemStack) ID() int16 {
	i.Lock.RLock()
	defer i.Lock.RUnlock()
	return i.id
}

func (i *ItemStack) SetID(id int16) {
	i.Lock.Lock()
	defer i.Lock.Unlock()
	i.id = id
}

func (i *ItemStack) Data() int16 {
	i.Lock.RLock()
	defer i.Lock.RUnlock()
	return i.damage
}

func (i *ItemStack) SetData(data int16) {
	i.Lock.Lock()
	defer i.Lock.Unlock()
	i.damage = data
}

func (i *ItemStack) Count() byte {
	i.Lock.RLock()
	defer i.Lock.RUnlock()
	return i.count
}

func (i *ItemStack) SetCount(count byte) {
	i.Lock.Lock()
	defer i.Lock.Unlock()
	i.count = count
}

func (i *ItemStack) SetDisplayName(name string) {
	i.Lock.Lock()
	defer i.Lock.Unlock()
	if i.Tag == nil {
		i.Tag = nbt.NewNBT()
	}

	display, _ := i.Tag.GetCompound("display", true)
	display.Set("Name", name)
}

func (i *ItemStack) ClearLore() {
	i.Lock.Lock()
	defer i.Lock.Unlock()
	if i.Tag == nil {
		return
	}
	if display, ok := i.Tag.GetCompound("display", false); ok {
		display.Remove("Lore")
	}
}

func (i *ItemStack) AddLore(line string) {
	i.Lock.Lock()
	defer i.Lock.Unlock()
	if i.Tag == nil {
		i.Tag = nbt.NewNBT()
	}
	display, _ := i.Tag.GetCompound("display", true)
	lore, _ := display.GetList("Lore", true)
	lore = append(lore, line)
	display.Set("Lore", lore)
}
