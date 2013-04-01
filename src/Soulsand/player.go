package Soulsand

import (
	"Soulsand/effect"
)

//A currently online player
type Player interface {
	Entity
	EventSource
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
	//Plays the the sound or particle effect at the location
	PlayEffect(x, y, z int, eff effect.Type, data int, relative bool)
}

type SyncPlayer interface {
	SyncEntity
	EventSource
	CommandSender
	//Returns the player's name
	GetName() string
	//Returns the player's display name
	GetDisplayNameSync() string
	//Sets the player's display name
	SetDisplayNameSync(name string)
	//Returns the player's view distance in chunks
	GetViewDistanceSync() int
	//UNSAFE: Returns a UnsafeConnection which provides ways to send packets to the player
	GetConnection() UnsafeConnection
}
