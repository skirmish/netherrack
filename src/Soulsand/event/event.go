package event

import ()

type Type int

const (
	PLAYER_MESSAGE Type = iota
	PLAYER_JOIN
	PLAYER_LEAVE
)
