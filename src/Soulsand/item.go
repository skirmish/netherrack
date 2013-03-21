package Soulsand

import (

)

type ItemStack interface {
	//Gets the item's ID
	GetID() int16
	//Gets the item's data
	GetData() int16
	//Gets the number of items in this stack
	GetCount() byte 
	//Sets the item's display name
	SetDisplayName(name string)
	//Removes all lore from this item
	ClearLore()
	//Adds a line to the lore
	AddLore(line string)
	//Stores a custom value on the item
	SetMetadata(key string, value interface{})
	//Gets a custom value from the item
	GetMetadata(key string) interface{}
	//Removes a custom value from the item
	RemoveMetadata(key string)
}