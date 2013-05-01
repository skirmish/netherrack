package items

import (
	"github.com/thinkofdeath/netherrack/nbt"
	"github.com/thinkofdeath/soulsand"
	"sync"
)

//Compile time checks
var _ soulsand.ItemStack = &ItemStack{}

func CreateItemStack(id, data int16, count byte) *ItemStack {
	return &ItemStack{
		ID:       id,
		Damage:   data,
		Count:    count,
		metadata: make(map[string]interface{}),
	}
}

type ItemStack struct {
	Lock     sync.RWMutex
	ID       int16
	Count    byte
	Damage   int16
	Tag      *nbt.Compound
	metaLock sync.RWMutex
	metadata map[string]interface{}
}

func (i *ItemStack) GetID() int16 {
	i.Lock.RLock()
	defer i.Lock.RUnlock()
	return i.ID
}

func (i *ItemStack) SetID(id int16) {
	i.Lock.Lock()
	defer i.Lock.Unlock()
	i.ID = id
}

func (i *ItemStack) GetData() int16 {
	i.Lock.RLock()
	defer i.Lock.RUnlock()
	return i.Damage
}

func (i *ItemStack) SetData(data int16) {
	i.Lock.Lock()
	defer i.Lock.Unlock()
	i.Damage = data
}

func (i *ItemStack) GetCount() byte {
	i.Lock.RLock()
	defer i.Lock.RUnlock()
	return i.Count
}

func (i *ItemStack) SetCount(count byte) {
	i.Lock.Lock()
	defer i.Lock.Unlock()
	i.Count = count
}

func (i *ItemStack) SetDisplayName(name string) {
	i.Lock.Lock()
	defer i.Lock.Unlock()
	if i.Tag == nil {
		i.Tag = nbt.NewCompound()
		i.Tag.Name = "tag"
	}
	if _, ok := i.Tag.Tags["display"]; !ok {
		display := nbt.NewCompound()
		display.Name = "display"
		i.Tag.Tags["display"] = display
	}
	display := i.Tag.Tags["display"].(*nbt.Compound)
	display.Tags["Name"] = nbt.String{"Name", name}
}

func (i *ItemStack) ClearLore() {
	i.Lock.Lock()
	defer i.Lock.Unlock()
	if i.Tag == nil {
		return
	}
	if _, ok := i.Tag.Tags["display"]; !ok {
		return
	}
	display := i.Tag.Tags["display"].(*nbt.Compound)
	delete(display.Tags, "Lore")
}

func (i *ItemStack) AddLore(line string) {
	i.Lock.Lock()
	defer i.Lock.Unlock()
	if i.Tag == nil {
		i.Tag = nbt.NewCompound()
		i.Tag.Name = "tag"
	}
	if _, ok := i.Tag.Tags["display"]; !ok {
		display := nbt.NewCompound()
		display.Name = "display"
		i.Tag.Tags["display"] = display
	}
	display := i.Tag.Tags["display"].(*nbt.Compound)
	if _, ok := display.Tags["Lore"]; !ok {
		lore := nbt.List{}
		lore.Name = "Lore"
		lore.Type = nbt.TypeString
		lore.Tags = make([]nbt.WriterTo, 0, 1)
		display.Tags["Lore"] = lore
	}
	lore := display.Tags["Lore"].(nbt.List)
	lore.Tags = append(lore.Tags, nbt.String{Value: line})
}

func (i *ItemStack) SetMetadata(key string, value interface{}) {
	i.metaLock.Lock()
	defer i.metaLock.Unlock()
	i.metadata[key] = value
}

func (i *ItemStack) GetMetadata(key string) interface{} {
	i.metaLock.RLock()
	defer i.metaLock.RUnlock()
	return i.metadata[key]
}

func (i *ItemStack) RemoveMetadata(key string) {
	i.metaLock.Lock()
	defer i.metaLock.Unlock()
	delete(i.metadata, key)
}
