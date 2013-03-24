package event

import ()

type Type int

const (
	//Fired by (Sync)Player. Returns an EventPlayerMessage
	PLAYER_MESSAGE Type = iota
	//Fired by Server. Returns an EventPlayerJoin
	PLAYER_JOIN
	//Fire by (Sync)Player. Returns an EventPlayerLeave
	PLAYER_LEAVE
)
