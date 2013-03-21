package Soulsand

import (
)

//A currently online player
type Player interface {
	Entity
	//Returns the player's name
	GetName() string
	//Sends a message to the player
	SendMessage(message string)
	//Runs the function in the player's goroutine
	RunSync(func(SyncPlayer))
	//Returns the player's display name
	GetDisplayName() string
	//Sets the player's display name
	SetDisplayName(name string)
	//Sets the player's level
	SetLevel(level int)
	//Sets the player's experience bar position
	SetExperienceBar(position float32)
	//Gets the player's locale
	GetLocale() string
	//Gets the player's gamemode
	GetGamemode() Gamemode
	//Sets the player's gamemode
	SetGamemode(mode Gamemode)
}

type SyncPlayer interface {
	SyncEntity
	//Returns the player's name
	GetName() string
	//Returns the player's view distance in chunks
	GetViewDistanceSync() int
	//Returns the player's locale
	GetLocaleSync() string 
	//Sends a message to the player
	SendMessageSync(msg string)
	//UNSAFE: Returns a UnsafeConnection which provides ways to send packets to the player
	GetConnection() UnsafeConnection
}