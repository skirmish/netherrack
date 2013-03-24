package Soulsand

import ()

type EventPlayerMessage interface {
	Event
	GetMessage() string
	SetMessage(msg string)
	GetPlayer() SyncPlayer
}

type EventPlayerJoin interface {
	Event
	GetPlayer() SyncPlayer
	Disconnect(reason string)
}

type EventPlayerLeave interface {
	Event
	GetPlayer() SyncPlayer
}
