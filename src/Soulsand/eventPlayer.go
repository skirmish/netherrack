package Soulsand

import ()

type EventPlayerMessage interface {
	Event
	GetMessage() string
	GetPlayer() Player
}

type EventPlayerJoin interface {
	Event
	GetPlayer() SyncPlayer
}