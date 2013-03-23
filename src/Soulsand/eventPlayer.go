package Soulsand

import ()

type EventPlayerMessage interface {
	Event
	GetMessage() string
	SetMessage(msg string)
	GetPlayer() Player
}

type EventPlayerJoin interface {
	Event
	GetPlayer() SyncPlayer
	Disconnect(reason string)
}
