package items

import (
	"bitbucket.org/Thinkofdeath/netherrack/nbt"
	"Soulsand"
	"sync"
)

//Compile time checks
var _ Soulsand.ItemStack = &ItemStack{}

type ItemStack struct {
	ID       int16
	Count    byte
	Damage   int16
	Tag      *nbt.Compound
	metaLock sync.RWMutex
	metadata map[string]interface{}
}

func (i *ItemStack) GetID() int16 {
	return i.ID
}

func (i *ItemStack) GetData() int16 {
	return i.Damage
}

func (i *ItemStack) GetCount() byte {
	return i.Count
}

func (i *ItemStack) SetDisplayName(name string) {
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
