package event

import (
	"Soulsand"
)

var _ Soulsand.EventPlayerMessage = &Message{}
var _ Soulsand.EventPlayerJoin = &Join{}

type Message struct {
	Event

	player  Soulsand.SyncPlayer
	message string
}

func NewMessage(player Soulsand.SyncPlayer, message string) *Message {
	return &Message{
		Event:   Event{},
		player:  player,
		message: message,
	}
}

func (m *Message) SetMessage(msg string) {
	m.message = msg
}

func (m *Message) GetMessage() string {
	return m.message
}

func (m *Message) GetPlayer() Soulsand.SyncPlayer {
	return m.player
}

type Join struct {
	Event

	player Soulsand.SyncPlayer
	Reason string
}

func NewJoin(player Soulsand.SyncPlayer, reason string) *Join {
	return &Join{
		Event:  Event{},
		player: player,
		Reason: reason,
	}
}

func (j *Join) GetPlayer() Soulsand.SyncPlayer {
	return j.player
}

func (j *Join) Disconnect(reason string) {
	j.Reason = reason
	j.Cancel()
}

type Leave struct {
	Event

	player Soulsand.SyncPlayer
}

func NewLeave(player Soulsand.SyncPlayer) *Leave {
	return &Leave{
		Event:  Event{},
		player: player,
	}
}

func (l *Leave) GetPlayer() Soulsand.SyncPlayer {
	return l.player
}
